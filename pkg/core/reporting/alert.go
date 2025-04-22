package reporting

import "time"

// AlertDetails contains the details of an alert
type AlertDetails struct {
	Title       string
	Description string
	Severity    string
	Links       []LinkProp
	Logs        []string
	Events      []EventInfo
	Metadata    map[string]string
}

// EventInfo contains information about an event
type EventInfo struct {
	Timestamp time.Time
	Type      string
	Message   string
	Data      map[string]interface{}
}
