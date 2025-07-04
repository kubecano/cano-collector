package alert

import (
	"strings"

	"github.com/kubecano/cano-collector/pkg/alert/model"
)

// AlertFormatter formats alert into readable messages
type AlertFormatter struct{}

// NewAlertFormatter creates a new alert formatter
func NewAlertFormatter() *AlertFormatter {
	return &AlertFormatter{}
}

// FormatAlert converts alertmanager alert to a readable message
func (f *AlertFormatter) FormatAlert(alertEvent *model.AlertManagerEvent) string {
	var messages []string

	// Add alert header
	messages = append(messages, "🚨 Alert: "+alertEvent.GetAlertName())
	messages = append(messages, "Status: "+alertEvent.Status)
	messages = append(messages, "Severity: "+alertEvent.GetSeverity())

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
