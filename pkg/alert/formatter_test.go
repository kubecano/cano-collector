package alert

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAlertFormatter_FormatAlert_BasicAlert(t *testing.T) {
	formatter := NewAlertFormatter()
	now := time.Now()

	alert := &AlertManagerEvent{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []PrometheusAlert{
			{
				Status: "firing",
				Labels: map[string]string{
					"alertname": "HighCPUUsage",
					"severity":  "critical",
				},
				Annotations: map[string]string{
					"summary":     "High CPU usage detected",
					"description": "The CPU usage has exceeded the threshold",
				},
				StartsAt: now,
			},
		},
	}

	result := formatter.FormatAlert(alert)

	assert.Contains(t, result, "🚨 **Alert: firing**")
	assert.Contains(t, result, "**Alert:** HighCPUUsage")
	assert.Contains(t, result, "**Status:** firing")
	assert.Contains(t, result, "**Severity:** critical")
	assert.Contains(t, result, "**Summary:** High CPU usage detected")
	assert.Contains(t, result, "**Description:** The CPU usage has exceeded the threshold")
}

func TestAlertFormatter_FormatAlert_WithGroupLabels(t *testing.T) {
	formatter := NewAlertFormatter()

	alert := &AlertManagerEvent{
		Receiver: "test-receiver",
		Status:   "firing",
		GroupLabels: map[string]string{
			"namespace": "production",
			"service":   "api",
		},
		Alerts: []PrometheusAlert{
			{
				Status: "firing",
				Labels: map[string]string{
					"alertname": "HighCPUUsage",
					"severity":  "warning",
				},
			},
		},
	}

	result := formatter.FormatAlert(alert)

	assert.Contains(t, result, "**namespace:** production")
	assert.Contains(t, result, "**service:** api")
}

func TestAlertFormatter_FormatAlert_MultipleAlerts(t *testing.T) {
	formatter := NewAlertFormatter()

	alert := &AlertManagerEvent{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []PrometheusAlert{
			{
				Status: "firing",
				Labels: map[string]string{
					"alertname": "HighCPUUsage",
					"severity":  "critical",
				},
				Annotations: map[string]string{
					"summary": "First alert",
				},
			},
			{
				Status: "firing",
				Labels: map[string]string{
					"alertname": "HighMemoryUsage",
					"severity":  "warning",
				},
				Annotations: map[string]string{
					"summary": "Second alert",
				},
			},
		},
	}

	result := formatter.FormatAlert(alert)

	// Check if both alerts are in the result
	assert.Contains(t, result, "HighCPUUsage")
	assert.Contains(t, result, "HighMemoryUsage")
	assert.Contains(t, result, "**Summary:** First alert")
	assert.Contains(t, result, "**Summary:** Second alert")
}

func TestAlertFormatter_FormatAlert_MissingLabels(t *testing.T) {
	formatter := NewAlertFormatter()

	alert := &AlertManagerEvent{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []PrometheusAlert{
			{
				Status: "firing",
				Labels: map[string]string{
					"alertname": "HighCPUUsage",
					// missing severity
				},
			},
		},
	}

	result := formatter.FormatAlert(alert)

	assert.Contains(t, result, "**Alert:** HighCPUUsage")
	assert.Contains(t, result, "**Status:** firing")
	// Should not contain severity if it's missing
	assert.NotContains(t, result, "**Severity:**")
}

func TestAlertFormatter_FormatAlert_MissingAnnotations(t *testing.T) {
	formatter := NewAlertFormatter()

	alert := &AlertManagerEvent{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []PrometheusAlert{
			{
				Status: "firing",
				Labels: map[string]string{
					"alertname": "HighCPUUsage",
					"severity":  "critical",
				},
				// missing annotations
			},
		},
	}

	result := formatter.FormatAlert(alert)

	assert.Contains(t, result, "**Alert:** HighCPUUsage")
	assert.Contains(t, result, "**Status:** firing")
	assert.Contains(t, result, "**Severity:** critical")
	// Should not contain summary or description
	assert.NotContains(t, result, "**Summary:**")
	assert.NotContains(t, result, "**Description:**")
}

func TestAlertFormatter_FormatAlert_EmptyAlerts(t *testing.T) {
	formatter := NewAlertFormatter()

	alert := &AlertManagerEvent{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts:   []PrometheusAlert{}, // empty alerts list
	}

	result := formatter.FormatAlert(alert)

	assert.Contains(t, result, "🚨 **Alert: firing**")
	// Should not contain alert details
	assert.NotContains(t, result, "**Alert:**")
	assert.NotContains(t, result, "**Severity:**")
	assert.NotContains(t, result, "**Summary:**")
	assert.NotContains(t, result, "**Description:**")
}

func TestAlertFormatter_FormatAlert_ResolvedStatus(t *testing.T) {
	formatter := NewAlertFormatter()

	alert := &AlertManagerEvent{
		Receiver: "test-receiver",
		Status:   "resolved",
		Alerts: []PrometheusAlert{
			{
				Status: "resolved",
				Labels: map[string]string{
					"alertname": "HighCPUUsage",
					"severity":  "critical",
				},
			},
		},
	}

	result := formatter.FormatAlert(alert)

	assert.Contains(t, result, "🚨 **Alert: resolved**")
	assert.Contains(t, result, "**Status:** resolved")
}

func TestAlertFormatter_FormatAlert_SpecialCharacters(t *testing.T) {
	formatter := NewAlertFormatter()

	alert := &AlertManagerEvent{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []PrometheusAlert{
			{
				Status: "firing",
				Labels: map[string]string{
					"alertname": "High CPU Usage (API)",
					"severity":  "critical",
				},
				Annotations: map[string]string{
					"summary":     "CPU usage > 90% for 5 minutes",
					"description": "Alert with special chars: < > & \" '",
				},
			},
		},
	}

	result := formatter.FormatAlert(alert)

	assert.Contains(t, result, "**Alert:** High CPU Usage (API)")
	assert.Contains(t, result, "**Summary:** CPU usage > 90% for 5 minutes")
	assert.Contains(t, result, "**Description:** Alert with special chars: < > & \" '")
}

func TestAlertFormatter_FormatAlert_NewlinesInContent(t *testing.T) {
	formatter := NewAlertFormatter()

	alert := &AlertManagerEvent{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []PrometheusAlert{
			{
				Status: "firing",
				Labels: map[string]string{
					"alertname": "HighCPUUsage",
					"severity":  "critical",
				},
				Annotations: map[string]string{
					"summary":     "Multi-line\nsummary",
					"description": "Multi-line\ndescription\nwith breaks",
				},
			},
		},
	}

	result := formatter.FormatAlert(alert)

	assert.Contains(t, result, "**Summary:** Multi-line\nsummary")
	assert.Contains(t, result, "**Description:** Multi-line\ndescription\nwith breaks")
}

func TestAlertFormatter_FormatAlert_EmptyStringValues(t *testing.T) {
	formatter := NewAlertFormatter()

	alert := &AlertManagerEvent{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []PrometheusAlert{
			{
				Status: "firing",
				Labels: map[string]string{
					"alertname": "HighCPUUsage",
					"severity":  "", // empty severity
				},
				Annotations: map[string]string{
					"summary":     "", // empty summary
					"description": "", // empty description
				},
			},
		},
	}

	result := formatter.FormatAlert(alert)

	assert.Contains(t, result, "🚨 **Alert: firing**")
	assert.Contains(t, result, "**Status:** firing")
	// Should not include empty fields
	assert.NotContains(t, result, "**Severity:** ")
	assert.NotContains(t, result, "**Summary:** ")
	assert.NotContains(t, result, "**Description:** ")
}

func TestAlertFormatter_FormatAlert_MessageStructure(t *testing.T) {
	formatter := NewAlertFormatter()

	alert := &AlertManagerEvent{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []PrometheusAlert{
			{
				Status: "firing",
				Labels: map[string]string{
					"alertname": "HighCPUUsage",
				},
			},
		},
	}

	result := formatter.FormatAlert(alert)

	// Check the structure: status at top, then group labels, then alerts
	lines := strings.Split(result, "\n")
	assert.True(t, strings.Contains(lines[0], "🚨 **Alert: firing**"))

	// Check that there's a blank line between sections
	hasBlankLine := false
	for _, line := range lines {
		if line == "" {
			hasBlankLine = true
			break
		}
	}
	assert.True(t, hasBlankLine)
}

func TestAlertFormatter_FormatAlert_InvalidAlertType(t *testing.T) {
	formatter := NewAlertFormatter()

	// Pass a string instead of AlertManagerEvent
	result := formatter.FormatAlert("not an alert")

	assert.Equal(t, "Error: Invalid alert format", result)
}
