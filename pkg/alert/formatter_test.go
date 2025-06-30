package alert

import (
	"strings"
	"testing"
	"time"

	"github.com/prometheus/alertmanager/template"
	"github.com/stretchr/testify/assert"
)

func TestAlertFormatter_FormatAlert_BasicAlert(t *testing.T) {
	formatter := NewAlertFormatter()

	alert := template.Data{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []template.Alert{
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
				StartsAt: time.Now(),
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

	alert := template.Data{
		Receiver: "test-receiver",
		Status:   "firing",
		GroupLabels: map[string]string{
			"namespace": "production",
			"service":   "api",
		},
		Alerts: []template.Alert{
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

	alert := template.Data{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []template.Alert{
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

	alert := template.Data{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []template.Alert{
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

	alert := template.Data{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []template.Alert{
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

	alert := template.Data{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts:   []template.Alert{}, // empty alerts list
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

	alert := template.Data{
		Receiver: "test-receiver",
		Status:   "resolved",
		Alerts: []template.Alert{
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

	alert := template.Data{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []template.Alert{
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

	alert := template.Data{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []template.Alert{
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

	// Check if newlines are preserved
	assert.Contains(t, result, "**Summary:** Multi-line\nsummary")
	assert.Contains(t, result, "**Description:** Multi-line\ndescription\nwith breaks")
}

func TestAlertFormatter_FormatAlert_EmptyStringValues(t *testing.T) {
	formatter := NewAlertFormatter()

	alert := template.Data{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []template.Alert{
			{
				Status: "firing",
				Labels: map[string]string{
					"alertname": "",
					"severity":  "",
				},
				Annotations: map[string]string{
					"summary":     "",
					"description": "",
				},
			},
		},
	}

	result := formatter.FormatAlert(alert)

	// Should not contain empty fields
	assert.NotContains(t, result, "**Alert:**")
	assert.NotContains(t, result, "**Severity:**")
	assert.NotContains(t, result, "**Summary:**")
	assert.NotContains(t, result, "**Description:**")
	// Should only contain status and header
	assert.Contains(t, result, "🚨 **Alert: firing**")
	assert.Contains(t, result, "**Status:** firing")
}

func TestAlertFormatter_FormatAlert_MessageStructure(t *testing.T) {
	formatter := NewAlertFormatter()

	alert := template.Data{
		Receiver: "test-receiver",
		Status:   "firing",
		GroupLabels: map[string]string{
			"namespace": "prod",
		},
		Alerts: []template.Alert{
			{
				Status: "firing",
				Labels: map[string]string{
					"alertname": "TestAlert",
					"severity":  "critical",
				},
				Annotations: map[string]string{
					"summary": "Test summary",
				},
			},
		},
	}

	result := formatter.FormatAlert(alert)

	// Check message structure
	lines := strings.Split(result, "\n")

	// First line should be header
	assert.Contains(t, lines[0], "🚨 **Alert: firing**")

	// Second line should be group label
	assert.Contains(t, lines[1], "**namespace:** prod")

	// Third line should be empty (separator)
	assert.Empty(t, lines[2])

	// Fourth line should be alert name
	assert.Contains(t, lines[3], "**Alert:** TestAlert")

	// Fifth line should be status
	assert.Contains(t, lines[4], "**Status:** firing")

	// Sixth line should be severity
	assert.Contains(t, lines[5], "**Severity:** critical")

	// Seventh line should be summary
	assert.Contains(t, lines[6], "**Summary:** Test summary")

	// Eighth line should be empty (separator)
	assert.Empty(t, lines[7])
}
