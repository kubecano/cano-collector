package alert

import (
	"fmt"
	"strings"

	"github.com/prometheus/alertmanager/template"
)

//go:generate mockgen -destination=../../mocks/alert_formatter_mock.go -package=mocks github.com/kubecano/cano-collector/pkg/alert AlertFormatterInterface
type AlertFormatterInterface interface {
	FormatAlert(alert template.Data) string
}

// AlertFormatter formats alertmanager alerts to readable messages
type AlertFormatter struct{}

// NewAlertFormatter creates a new alert formatter
func NewAlertFormatter() *AlertFormatter {
	return &AlertFormatter{}
}

// FormatAlert converts alertmanager alert to a readable message
func (f *AlertFormatter) FormatAlert(alert template.Data) string {
	var messages []string

	messages = append(messages, fmt.Sprintf("🚨 **Alert: %s**", alert.Status))

	if alert.GroupLabels != nil {
		for key, value := range alert.GroupLabels {
			messages = append(messages, fmt.Sprintf("**%s:** %s", key, value))
		}
	}

	messages = append(messages, "")

	for _, alertItem := range alert.Alerts {
		messages = append(messages, fmt.Sprintf("**Alert:** %s", alertItem.Labels["alertname"]))
		messages = append(messages, fmt.Sprintf("**Status:** %s", alertItem.Status))
		messages = append(messages, fmt.Sprintf("**Severity:** %s", alertItem.Labels["severity"]))

		if alertItem.Annotations["summary"] != "" {
			messages = append(messages, fmt.Sprintf("**Summary:** %s", alertItem.Annotations["summary"]))
		}

		if alertItem.Annotations["description"] != "" {
			messages = append(messages, fmt.Sprintf("**Description:** %s", alertItem.Annotations["description"]))
		}

		messages = append(messages, "")
	}

	return strings.Join(messages, "\n")
}
