package alert

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kubecano/cano-collector/pkg/alert/model"
)

func createTestAlertManagerEventForFormatter() *model.AlertManagerEvent {
	now := time.Now()
	return &model.AlertManagerEvent{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []model.PrometheusAlert{
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

	assert.Contains(t, result, "🚨 Alert: HighCPUUsage")
	assert.Contains(t, result, "Status: firing")
	assert.Contains(t, result, "Severity: critical")
	assert.Contains(t, result, "Summary: High CPU usage detected")
	assert.Contains(t, result, "Description: The CPU usage has exceeded the threshold")
}

func TestAlertFormatter_FormatAlert_WithGroupLabels(t *testing.T) {
	formatter := NewAlertFormatter()

	alert := &model.AlertManagerEvent{
		Receiver: "test-receiver",
		Status:   "firing",
		GroupLabels: map[string]string{
			"namespace": "production",
			"service":   "api",
		},
		Alerts: []model.PrometheusAlert{
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

	// New formatter doesn't display GroupLabels, just check basic format
	assert.Contains(t, result, "🚨 Alert: HighCPUUsage")
	assert.Contains(t, result, "Status: firing")
	assert.Contains(t, result, "Severity: warning")
}

func TestAlertFormatter_FormatAlert_MultipleAlerts(t *testing.T) {
	formatter := NewAlertFormatter()

	alert := &model.AlertManagerEvent{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []model.PrometheusAlert{
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

	// Formatter shows first alert name in header, but all summaries
	assert.Contains(t, result, "🚨 Alert: HighCPUUsage")
	assert.Contains(t, result, "Status: firing")
	assert.Contains(t, result, "Severity: critical")
	assert.Contains(t, result, "Summary: First alert")
	assert.Contains(t, result, "Summary: Second alert")
}

func TestAlertFormatter_FormatAlert_MissingLabels(t *testing.T) {
	formatter := NewAlertFormatter()

	alert := &model.AlertManagerEvent{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []model.PrometheusAlert{
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

	assert.Contains(t, result, "🚨 Alert: HighCPUUsage")
	assert.Contains(t, result, "Status: firing")
	assert.Contains(t, result, "Severity: unknown")
}

func TestAlertFormatter_FormatAlert_MissingAnnotations(t *testing.T) {
	formatter := NewAlertFormatter()

	alert := &model.AlertManagerEvent{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []model.PrometheusAlert{
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

	assert.Contains(t, result, "🚨 Alert: HighCPUUsage")
	assert.Contains(t, result, "Status: firing")
	assert.Contains(t, result, "Severity: critical")
	// Should not contain summary or description
	assert.NotContains(t, result, "Summary:")
	assert.NotContains(t, result, "Description:")
}

func TestAlertFormatter_FormatAlert_EmptyAlerts(t *testing.T) {
	formatter := NewAlertFormatter()

	alert := &model.AlertManagerEvent{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts:   []model.PrometheusAlert{}, // empty alerts list
	}

	result := formatter.FormatAlert(alert)

	assert.Contains(t, result, "🚨 Alert: unknown")
	assert.Contains(t, result, "Status: firing")
	assert.Contains(t, result, "Severity: unknown")
	// Should not contain summary or description for empty alerts
	assert.NotContains(t, result, "Summary:")
	assert.NotContains(t, result, "Description:")
}

func TestAlertFormatter_FormatAlert_ResolvedStatus(t *testing.T) {
	formatter := NewAlertFormatter()

	alert := &model.AlertManagerEvent{
		Receiver: "test-receiver",
		Status:   "resolved",
		Alerts: []model.PrometheusAlert{
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

	assert.Contains(t, result, "🚨 Alert: HighCPUUsage")
	assert.Contains(t, result, "Status: resolved")
	assert.Contains(t, result, "Severity: critical")
}

func TestAlertFormatter_FormatAlert_SpecialCharacters(t *testing.T) {
	formatter := NewAlertFormatter()

	alert := &model.AlertManagerEvent{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []model.PrometheusAlert{
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

	assert.Contains(t, result, "🚨 Alert: SpecialChars")
	assert.Contains(t, result, "Unicode symbols: 🚨🔥💻")
}

func TestAlertFormatter_FormatAlert_NewlinesInContent(t *testing.T) {
	formatter := NewAlertFormatter()

	alert := &model.AlertManagerEvent{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []model.PrometheusAlert{
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

	assert.Contains(t, result, "Description: Line 1\nLine 2\nLine 3")
}

func TestAlertFormatter_FormatAlert_EmptyStringValues(t *testing.T) {
	formatter := NewAlertFormatter()

	alert := &model.AlertManagerEvent{
		Receiver: "test-receiver",
		Status:   "firing",
		GroupLabels: map[string]string{
			"empty": "",
			"valid": "value",
		},
		Alerts: []model.PrometheusAlert{
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

	// New formatter doesn't display GroupLabels, check basic format
	assert.Contains(t, result, "🚨 Alert: EmptyValues")
	assert.Contains(t, result, "Status: firing")
	assert.Contains(t, result, "Severity: unknown")
	assert.Contains(t, result, "Summary: ")
	assert.Contains(t, result, "Description: Valid description")
}

func TestAlertFormatter_FormatAlert_MessageStructure(t *testing.T) {
	formatter := NewAlertFormatter()

	alert := createTestAlertManagerEventForFormatter()

	result := formatter.FormatAlert(alert)

	lines := strings.Split(result, "\n")
	assert.GreaterOrEqual(t, len(lines), 3, "Expected at least 3 lines in the formatted message")

	// First line should be the alert header
	assert.Contains(t, lines[0], "🚨 Alert: HighCPUUsage")

	// Check basic structure
	assert.Contains(t, result, "Status:")
	assert.Contains(t, result, "Severity:")
}

func TestAlertFormatter_FormatAlert_EmptyAlertEvent(t *testing.T) {
	formatter := NewAlertFormatter()

	// Pass empty AlertManagerEvent
	result := formatter.FormatAlert(&model.AlertManagerEvent{})

	expected := "🚨 Alert: unknown\nStatus: \nSeverity: unknown"
	assert.Equal(t, expected, result)
}
