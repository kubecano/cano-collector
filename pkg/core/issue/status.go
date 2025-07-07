package issue

import (
	"fmt"
	"strings"
)

// Status represents the current state of an issue
type Status int

const (
	StatusFiring Status = iota
	StatusResolved
)

// String returns the string representation of the status
func (s Status) String() string {
	switch s {
	case StatusFiring:
		return "FIRING"
	case StatusResolved:
		return "RESOLVED"
	default:
		return "UNKNOWN"
	}
}

// ToEmoji returns the emoji representation of the status
func (s Status) ToEmoji() string {
	switch s {
	case StatusResolved:
		return "✅"
	case StatusFiring:
		return "🔥"
	default:
		return "🔥"
	}
}

// FromString converts a string to Status
func StatusFromString(s string) (Status, error) {
	switch strings.ToUpper(s) {
	case "FIRING":
		return StatusFiring, nil
	case "RESOLVED":
		return StatusResolved, nil
	default:
		return StatusFiring, fmt.Errorf("unknown status: %s", s)
	}
}

// StatusFromPrometheusStatus maps Prometheus alert status to Issue status
func StatusFromPrometheusStatus(status string) Status {
	switch strings.ToLower(status) {
	case "firing":
		return StatusFiring
	case "resolved":
		return StatusResolved
	default:
		return StatusFiring
	}
}
