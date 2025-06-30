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

	// Sprawdź czy oba alerty są w wyniku
	assert.Contains(t, result, "**Alert:** HighCPUUsage")
	assert.Contains(t, result, "**Alert:** HighMemoryUsage")
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
					// brak severity
				},
			},
		},
	}

	result := formatter.FormatAlert(alert)

	assert.Contains(t, result, "**Alert:** HighCPUUsage")
	assert.Contains(t, result, "**Status:** firing")
	assert.Contains(t, result, "**Severity:** ") // pusty severity
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
				// brak annotations
			},
		},
	}

	result := formatter.FormatAlert(alert)

	assert.Contains(t, result, "**Alert:** HighCPUUsage")
	assert.Contains(t, result, "**Status:** firing")
	assert.Contains(t, result, "**Severity:** critical")
	// Nie powinno zawierać summary ani description
	assert.NotContains(t, result, "**Summary:**")
	assert.NotContains(t, result, "**Description:**")
}

func TestAlertFormatter_FormatAlert_EmptyAlerts(t *testing.T) {
	formatter := NewAlertFormatter()

	alert := template.Data{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts:   []template.Alert{}, // pusta lista alertów
	}

	result := formatter.FormatAlert(alert)

	assert.Contains(t, result, "🚨 **Alert: firing**")
	// Nie powinno zawierać szczegółów alertów
	assert.NotContains(t, result, "**Alert:**")
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

	// Sprawdź czy newlines są zachowane
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

	assert.Contains(t, result, "**Alert:** ")
	assert.Contains(t, result, "**Severity:** ")
	// Nie powinno być linii summary/description jeśli są puste
	assert.NotContains(t, result, "**Summary:**")
	assert.NotContains(t, result, "**Description:**")
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

	// Sprawdź strukturę wiadomości
	lines := strings.Split(result, "\n")

	// Pierwsza linia powinna być nagłówkiem
	assert.Contains(t, lines[0], "🚨 **Alert: firing**")

	// Powinna być pusta linia po group labels
	assert.Contains(t, result, "**namespace:** prod")

	// Powinna być pusta linia przed alertami
	assert.Contains(t, result, "**Alert:** TestAlert")
}
