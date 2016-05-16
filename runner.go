package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/byuoitav/touchpanel-update-runner/helpers"
)

// Starts the TP Update process
func startRun(curTP tpStatus) {
	curTP.Attempts++

	curTP.Steps = getTPSteps()

	curTP.Attempts = 0 // We haven't tried yet

	// Get the hostname

	response, err := sendCommand(curTP, "hostname", true)

	if err != nil {
		fmt.Printf("Could not retrieve hostname.")
	}
	if strings.Contains(response, "Host Name:") {
		response = strings.Split(response, "Host Name:")[1]
	}

	curTP.Hostname = strings.TrimSpace(response)

	updateChannel <- curTP

	evaluateNextStep(curTP)
}

func startWait(curTP tpStatus) error {
	fmt.Printf("%s Sending to wait\n", curTP.IPAddress)

	var req = waitRequest{IPAddressHostname: curTP.IPAddress, Port: 41795, CallbackAddress: os.Getenv("TOUCHPANEL_UPDATE_RUNNER_ADDRESS") + "/callbacks/afterWait"}

	req.Identifier = curTP.UUID

	bits, _ := json.Marshal(req)

	// fmt.Printf("Payload being send: \n %s \n", string(bits))

	// we have to wait for the thing to actually restart - otherwise we'll return
	// before it gets in a non-communicative state
	time.Sleep(10 * time.Second) // TODO: Shift this into our wait microservice

	resp, err := http.Post(os.Getenv("WAIT_FOR_REBOOT_MICROSERVICE_ADDRESS"), "application/json", bytes.NewBuffer(bits))

	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)

	defer resp.Body.Close()

	if !strings.Contains(string(body), "Added to queue") {
		return errors.New(string(body))
	}

	return nil
}

func reportNotNeeded(tp tpStatus, status string) {
	fmt.Printf("%s Not needed\n", tp.IPAddress)

	tp.CurrentStatus = status
	tp.EndTime = time.Now()
	updateChannel <- tp

	helpers.SendToElastic(tp, 0)
}

func reportSuccess(tp tpStatus) {
	fmt.Printf("%s Success!\n", tp.IPAddress)

	tp.CurrentStatus = "Success"
	tp.EndTime = time.Now()
	updateChannel <- tp

	helpers.SendToElastic(tp, 0)
}

func reportError(tp tpStatus, err error) {
	fmt.Printf("%s Reporting a failure  %s ...\n", tp.IPAddress, err.Error())

	ipTable := false

	// if we want to retry
	fmt.Printf("%s Attempts: %v\n", tp.IPAddress, tp.Attempts)
	if tp.Attempts < 2 {
		tp.Attempts++

		fmt.Printf("%s Retring process.\n", tp.IPAddress)
		if tp.Steps[0].Completed {
			ipTable = true
		}

		tp.Steps = getTPSteps() // reset the steps

		if ipTable { // if the iptable was already populated
			tp.Steps[0].Completed = true
		}

		updateChannel <- tp

		startWait(tp) // Who knows what state, run a wait on them
		return
	}

	tp.CurrentStatus = "Error"
	tp.EndTime = time.Now()
	tp.ErrorInfo = append(tp.ErrorInfo, err.Error())
	updateChannel <- tp

	helpers.SendToElastic(tp, 0)
}

func getIPTable(IPAddress string) (IPTable, error) {
	var toReturn = IPTable{}
	// TODO: Make the prompt generic
	var req = telnetRequest{IPAddress: IPAddress, Command: "iptable"}

	bits, _ := json.Marshal(req)

	resp, err := http.Post(os.Getenv("TELNET_MICROSERVICE_ADDRESS"), "application/json", bytes.NewBuffer(bits))

	if err != nil {
		return toReturn, err
	}

	bits, err = ioutil.ReadAll(resp.Body)

	if err != nil {
		return toReturn, err
	}

	err = json.Unmarshal(bits, &toReturn)
	if err != nil {
		return toReturn, err
	}

	if len(toReturn.Entries) < 1 {
		return toReturn, errors.New("There were no entries in the IP Table, error.")
	}

	return toReturn, nil
}
