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
			if key != "" && value != "" {
				messages = append(messages, fmt.Sprintf("**%s:** %s", key, value))
			}
		}
	}

	messages = append(messages, "")

	for _, alertItem := range alert.Alerts {
		if alertname := alertItem.Labels["alertname"]; alertname != "" {
			messages = append(messages, "**Alert:** "+alertname)
		}
		if status := alertItem.Status; status != "" {
			messages = append(messages, "**Status:** "+status)
		}
		if severity := alertItem.Labels["severity"]; severity != "" {
			messages = append(messages, "**Severity:** "+severity)
		}

		if summary := alertItem.Annotations["summary"]; summary != "" {
			messages = append(messages, "**Summary:** "+summary)
		}

		if description := alertItem.Annotations["description"]; description != "" {
			messages = append(messages, "**Description:** "+description)
		}

		messages = append(messages, "")
	}

	return strings.Join(messages, "\n")
}
