package issue

import (
	"fmt"
	"strings"
)

// Severity represents the severity level of an issue
type Severity int

const (
	SeverityDebug Severity = iota
	SeverityInfo
	SeverityLow
	SeverityHigh
)

// String returns the string representation of the severity
func (s Severity) String() string {
	switch s {
	case SeverityDebug:
		return "DEBUG"
	case SeverityInfo:
		return "INFO"
	case SeverityLow:
		return "LOW"
	case SeverityHigh:
		return "HIGH"
	default:
		return "UNKNOWN"
	}
}

// FromString converts a string to Severity
func SeverityFromString(s string) (Severity, error) {
	switch strings.ToUpper(s) {
	case "DEBUG":
		return SeverityDebug, nil
	case "INFO":
		return SeverityInfo, nil
	case "LOW":
		return SeverityLow, nil
	case "HIGH":
		return SeverityHigh, nil
	default:
		return SeverityInfo, fmt.Errorf("unknown severity: %s", s)
	}
}

// SeverityFromPrometheusLabel maps Prometheus severity labels to Issue severity
func SeverityFromPrometheusLabel(label string) Severity {
	switch strings.ToLower(label) {
	case "critical":
		return SeverityHigh
	case "high":
		return SeverityHigh
	case "error":
		return SeverityHigh
	case "warning":
		return SeverityLow
	case "low":
		return SeverityLow
	case "info":
		return SeverityInfo
	case "debug":
		return SeverityDebug
	default:
		return SeverityInfo
	}
}
