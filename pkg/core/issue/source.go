package issue

import (
	"fmt"
	"strings"
)

// Source represents the origin of an issue
type Source int

const (
	SourceUnknown Source = iota
	SourcePrometheus
	SourceKubernetesAPIServer
	SourceCustom
	SourceWebhook
	SourceManual
	SourceOperator
)

// String returns the string representation of the source
func (s Source) String() string {
	switch s {
	case SourceUnknown:
		return "UNKNOWN"
	case SourcePrometheus:
		return "PROMETHEUS"
	case SourceKubernetesAPIServer:
		return "KUBERNETES_API_SERVER"
	case SourceCustom:
		return "CUSTOM"
	case SourceWebhook:
		return "WEBHOOK"
	case SourceManual:
		return "MANUAL"
	case SourceOperator:
		return "OPERATOR"
	default:
		return "UNKNOWN"
	}
}

// FromString converts a string to Source
func SourceFromString(s string) (Source, error) {
	switch strings.ToUpper(s) {
	case "UNKNOWN":
		return SourceUnknown, nil
	case "PROMETHEUS":
		return SourcePrometheus, nil
	case "KUBERNETES_API_SERVER":
		return SourceKubernetesAPIServer, nil
	case "CUSTOM":
		return SourceCustom, nil
	case "WEBHOOK":
		return SourceWebhook, nil
	case "MANUAL":
		return SourceManual, nil
	case "OPERATOR":
		return SourceOperator, nil
	default:
		return SourceUnknown, fmt.Errorf("unknown source: %s", s)
	}
}
