package alert

import (
	"strings"

	"github.com/kubecano/cano-collector/pkg/core/event"
)

// AlertFormatter formats alert into readable messages
type AlertFormatter struct{}

// NewAlertFormatter creates a new alert formatter
func NewAlertFormatter() *AlertFormatter {
	return &AlertFormatter{}
}

// FormatAlert converts alertmanager alert to a readable message
func (f *AlertFormatter) FormatAlert(alertEvent *event.AlertManagerEvent) string {
	var messages []string

	// Get alert name with fallback
	alertName := alertEvent.GetAlertName()
	if alertName == "" {
		alertName = "unknown"
	}

	// Get severity with fallback
	severity := alertEvent.GetSeverity()
	if severity == "" {
		severity = "unknown"
	}

	// Add alert header
	messages = append(messages, "🚨 Alert: "+alertName)
	messages = append(messages, "Status: "+alertEvent.Status)
	messages = append(messages, "Severity: "+severity)

	// Add individual alerts
	for _, alert := range alertEvent.Alerts {
		if summary, ok := alert.Annotations["summary"]; ok {
			messages = append(messages, "Summary: "+summary)
		}
		if description, ok := alert.Annotations["description"]; ok {
			messages = append(messages, "Description: "+description)
		}
	}

	return strings.Join(messages, "\n")
}
