package alert

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func createTestAlertManagerEventForFormatter() *AlertManagerEvent {
	now := time.Now()
	return &AlertManagerEvent{
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
}

func TestAlertFormatter_FormatAlert_BasicAlert(t *testing.T) {
	formatter := NewAlertFormatter()

	alert := createTestAlertManagerEventForFormatter()

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
	assert.Contains(t, result, "**Alert:** HighCPUUsage")
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
					"alertname": "SpecialChars",
					"severity":  "critical",
					"special":   "test@example.com",
					"unicode":   "🚨🔥💻",
				},
				Annotations: map[string]string{
					"description": "Unicode symbols: 🚨🔥💻",
				},
			},
		},
	}

	result := formatter.FormatAlert(alert)

	assert.Contains(t, result, "**Alert:** SpecialChars")
	assert.Contains(t, result, "Unicode symbols: 🚨🔥💻")
}

func TestAlertFormatter_FormatAlert_NewlinesInContent(t *testing.T) {
	formatter := NewAlertFormatter()

	alert := &AlertManagerEvent{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []PrometheusAlert{
			{
				Status: "firing",
				Labels: map[string]string{"alertname": "MultilineAlert"},
				Annotations: map[string]string{
					"description": "Line 1\nLine 2\nLine 3",
				},
			},
		},
	}

	result := formatter.FormatAlert(alert)

	assert.Contains(t, result, "**Description:** Line 1\nLine 2\nLine 3")
}

func TestAlertFormatter_FormatAlert_EmptyStringValues(t *testing.T) {
	formatter := NewAlertFormatter()

	alert := &AlertManagerEvent{
		Receiver: "test-receiver",
		Status:   "firing",
		GroupLabels: map[string]string{
			"empty": "",
			"valid": "value",
		},
		Alerts: []PrometheusAlert{
			{
				Status: "firing",
				Labels: map[string]string{
					"alertname": "EmptyValues",
					"empty":     "",
				},
				Annotations: map[string]string{
					"summary":     "",
					"description": "Valid description",
				},
			},
		},
	}

	result := formatter.FormatAlert(alert)

	assert.Contains(t, result, "**valid:** value")
	assert.NotContains(t, result, "**empty:**")
	assert.Contains(t, result, "**Description:** Valid description")
	assert.NotContains(t, result, "**Summary:**")
}

func TestAlertFormatter_FormatAlert_MessageStructure(t *testing.T) {
	formatter := NewAlertFormatter()

	alert := createTestAlertManagerEventForFormatter()

	result := formatter.FormatAlert(alert)

	lines := strings.Split(result, "\n")
	assert.GreaterOrEqual(t, len(lines), 5, "Expected at least 5 lines in the formatted message")

	// First line should be the alert status
	assert.Contains(t, lines[0], "🚨 **Alert: firing**")

	// There should be a blank line between sections
	foundBlankLine := false
	for _, line := range lines {
		if line == "" {
			foundBlankLine = true
			break
		}
	}
	assert.True(t, foundBlankLine, "Expected at least one blank line in the formatted message")
}

func TestAlertFormatter_FormatAlert_InvalidAlertType(t *testing.T) {
	formatter := NewAlertFormatter()

	// Pass a string instead of AlertManagerEvent
	result := formatter.FormatAlert(&AlertManagerEvent{})

	assert.Equal(t, "Error: Invalid alert format", result)
}
