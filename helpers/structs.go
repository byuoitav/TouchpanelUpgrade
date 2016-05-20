package helpers

import "time"

type TouchpanelStatus struct {
	CurrentStatus   string // The current status (Step) of the touchpanel
	Hostname        string
	UUID            string // UUID that is assigned to each touchpanel
	RoomName        string // the name of the room associate with this touchpanel
	Type            string
	IPAddress       string    // IPAddress of the touchpanel
	StartTime       time.Time // Time the update process was started
	EndTime         time.Time // Time the update process finished, or errored out
	IPTable         IPTable   // The IPTable associated with this touchpanel
	FirmwareVersion string    // The version of the firmware loaded on the touchpanel
	ProjectDate     string    // The compile date of the project loaded on the device
	Information     modelInformation
	Batch           string // Batch for uploading to Elastic Search
	Attempts        int    // number of times to attempt the update
	Force           bool   // optional flag to bypass the validation and force the update
	ErrorInfo       []string
	Steps           []step // List of steps in the update process
}

type jobInformation struct {
	Type                 []string // HDTec, TecLite, fliptop
	IPAddress            string
	Force                bool
	Batch                string
	HDConfiguration      modelInformation // The information for the HDTec panels
	TecLiteConfiguraiton modelInformation // the information for the TecLite panels
	FliptopConfiguration modelInformation // The information for the fliptop panels
}

type multiJobInformation struct {
	HDConfiguration      modelInformation // The information for the HDTec panels
	TecLiteConfiguraiton modelInformation // the information for the TecLite panels
	FliptopConfiguration modelInformation // The information for the fliptop panels
	Info                 []jobInformation
}

type submissionRequest struct {
	CallbackAddress string
}

type ftpRequest struct {
	IPAddressHostname string    `json:",omitempty"`
	CallbackAddress   string    `json:",omitempty"`
	Path              string    `json:",omitempty"`
	File              string    `json:",omitempty"`
	Identifier        string    `json:",omitempty"`
	Timeout           int       `json:",omitempty"`
	Username          string    `json:",omitempty"`
	Password          string    `json:",omitempty"`
	SubmissionTime    time.Time `json:",omitempty"`
	CompletionTime    time.Time `json:",omitempty"`
	Status            string    `json:",omitempty"`
	Error             string    `json:",omitempty"`
}

type WaitRequest struct {
	IPAddressHostname string    // hostname to be pinged
	Port              int       // port to be used when testing connection
	Timeout           int       // Time in seconds to wait. Optional, will default to 300 seconds if not present or is 0
	CallbackAddress   string    // complete address to send the notification that the host is responding
	SubmissionTime    time.Time // Will be filled by the server as the time the process started pinging
	CompletionTime    time.Time // Will be filled by the service as the time that a) Sucessfully responded or b) timed out
	Status            string    // Timeout or Success
	Identifier        string    `json:",omitempty"` // Optional value to be passed in so the requester can identify the host when it's sent back
}

// Represents information needed to update the touchpanels
type modelInformation struct {
	FirmwareLocation string // The location of the .puf file to be loaded
	ProjectLocation  string // The locaton of the compiled project file to be loaded
	ProjectDate      string // The compile date of the project to be loaded
	FirmwareVersion  string // The version of the firmeware to be loaded
}

// Defines one step, it's completion status, as well as any information gathered from the step
type step struct {
	StepName  string // Name of the step
	Completed bool   // if the step has been completed
	Info      string // Any information gathered from the step. Usually the JSON body retrieved
	Attempts  int
}

// IPTable represents an IPTable returend from a crestron device
type IPTable struct {
	Entries []IPEntry
}

// IPEntry represents a single entry in the IPTable
type IPEntry struct {
	CipID             string `json:"CIP_ID"`
	Type              string
	Status            string
	DevID             string
	Port              string
	IPAddressSitename string
}

// Equals checks if two iptabels are equivalent
func (i *IPTable) Equals(compare IPTable) bool {
	if len(i.Entries) != len(compare.Entries) {
		return false
	}

	for r := range i.Entries {
		if !i.Entries[r].Equals(compare.Entries[r]) {
			return false
		}
	}
	return true
}

// Equals compares two IPEntries to see if they're equivalent
func (e *IPEntry) Equals(compare IPEntry) bool {
	if e.CipID != compare.CipID ||
		e.DevID != compare.DevID ||
		e.IPAddressSitename != compare.IPAddressSitename ||
		e.Port != compare.Port ||
		e.Status != compare.Status ||
		e.Type != compare.Type {
		return false
	}
	return true
}