package alert

import (
	"testing"
	"time"

	"github.com/prometheus/alertmanager/template"
	"github.com/stretchr/testify/assert"
)

func TestNewAlertManagerEventFromTemplateData(t *testing.T) {
	now := time.Now()
	templateData := template.Data{
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
				StartsAt: now,
			},
		},
		ExternalURL:       "http://alertmanager.example.com",
		CommonAnnotations: map[string]string{"team": "platform"},
		CommonLabels:      map[string]string{"datacenter": "eu-west-1"},
		GroupLabels:       map[string]string{"service": "api"},
	}

	event := NewAlertManagerEventFromTemplateData(templateData)

	assert.Equal(t, "test-receiver", event.Receiver)
	assert.Equal(t, "firing", event.Status)
	assert.Equal(t, 1, len(event.Alerts))
	assert.Equal(t, "HighCPUUsage", event.Alerts[0].Labels["alertname"])
	assert.Equal(t, "critical", event.Alerts[0].Labels["severity"])
	assert.Equal(t, "High CPU usage detected", event.Alerts[0].Annotations["summary"])
	assert.Equal(t, "The CPU usage has exceeded the threshold", event.Alerts[0].Annotations["description"])
	assert.Equal(t, now, event.Alerts[0].StartsAt)
	assert.Equal(t, "http://alertmanager.example.com", event.ExternalURL)
	assert.Equal(t, "platform", event.CommonAnnotations["team"])
	assert.Equal(t, "eu-west-1", event.CommonLabels["datacenter"])
	assert.Equal(t, "api", event.GroupLabels["service"])
}

func TestAlertManagerEvent_Validate_ValidAlert(t *testing.T) {
	now := time.Now()
	event := &AlertManagerEvent{
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

	err := event.Validate()
	assert.NoError(t, err)
}

func TestAlertManagerEvent_Validate_MissingReceiver(t *testing.T) {
	now := time.Now()
	event := &AlertManagerEvent{
		// Receiver field missing
		Status: "firing",
		Alerts: []PrometheusAlert{
			{
				Status: "firing",
				Labels: map[string]string{
					"alertname": "HighCPUUsage",
				},
				StartsAt: now,
			},
		},
	}

	err := event.Validate()
	assert.ErrorIs(t, err, ErrMissingReceiver)
}

func TestAlertManagerEvent_Validate_MissingStatus(t *testing.T) {
	now := time.Now()
	event := &AlertManagerEvent{
		Receiver: "test-receiver",
		// Status field missing
		Alerts: []PrometheusAlert{
			{
				Status: "firing",
				Labels: map[string]string{
					"alertname": "HighCPUUsage",
				},
				StartsAt: now,
			},
		},
	}

	err := event.Validate()
	assert.ErrorIs(t, err, ErrMissingStatus)
}

func TestAlertManagerEvent_Validate_MissingAlerts(t *testing.T) {
	event := &AlertManagerEvent{
		Receiver: "test-receiver",
		Status:   "firing",
		// Alerts field empty
	}

	err := event.Validate()
	assert.ErrorIs(t, err, ErrMissingAlerts)
}

func TestAlertManagerEvent_Validate_InvalidAlertStatus(t *testing.T) {
	now := time.Now()
	event := &AlertManagerEvent{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []PrometheusAlert{
			{
				Status: "invalid-status", // Invalid status
				Labels: map[string]string{
					"alertname": "HighCPUUsage",
				},
				StartsAt: now,
			},
		},
	}

	err := event.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid status value")
}

func TestAlertManagerEvent_Validate_MissingAlertName(t *testing.T) {
	now := time.Now()
	event := &AlertManagerEvent{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []PrometheusAlert{
			{
				Status: "firing",
				Labels: map[string]string{
					// alertname missing
					"severity": "critical",
				},
				StartsAt: now,
			},
		},
	}

	err := event.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing alertname label")
}

func TestAlertManagerEvent_Validate_MissingStartTime(t *testing.T) {
	event := &AlertManagerEvent{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []PrometheusAlert{
			{
				Status: "firing",
				Labels: map[string]string{
					"alertname": "HighCPUUsage",
				},
				// StartsAt missing
			},
		},
	}

	err := event.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing start time for firing alert")
}

func TestAlertManagerEvent_GetAlertName(t *testing.T) {
	event := &AlertManagerEvent{
		Alerts: []PrometheusAlert{
			{
				Labels: map[string]string{
					"alertname": "HighCPUUsage",
				},
			},
		},
	}

	assert.Equal(t, "HighCPUUsage", event.GetAlertName())
}

func TestAlertManagerEvent_GetAlertName_NoAlerts(t *testing.T) {
	event := &AlertManagerEvent{}
	assert.Equal(t, "unknown", event.GetAlertName())
}

func TestAlertManagerEvent_GetAlertName_NoAlertName(t *testing.T) {
	event := &AlertManagerEvent{
		Alerts: []PrometheusAlert{
			{
				Labels: map[string]string{},
			},
		},
	}

	assert.Equal(t, "unknown", event.GetAlertName())
}

func TestAlertManagerEvent_GetSeverity(t *testing.T) {
	event := &AlertManagerEvent{
		Alerts: []PrometheusAlert{
			{
				Labels: map[string]string{
					"severity": "critical",
				},
			},
		},
	}

	assert.Equal(t, "critical", event.GetSeverity())
}

func TestAlertManagerEvent_GetSeverity_NoAlerts(t *testing.T) {
	event := &AlertManagerEvent{}
	assert.Equal(t, "unknown", event.GetSeverity())
}

func TestAlertManagerEvent_GetSeverity_NoSeverity(t *testing.T) {
	event := &AlertManagerEvent{
		Alerts: []PrometheusAlert{
			{
				Labels: map[string]string{},
			},
		},
	}

	assert.Equal(t, "unknown", event.GetSeverity())
}

func TestAlertManagerEvent_GetStartTime(t *testing.T) {
	now := time.Now()
	event := &AlertManagerEvent{
		Alerts: []PrometheusAlert{
			{
				StartsAt: now,
			},
		},
	}

	assert.Equal(t, now, event.GetStartTime())
}

func TestAlertManagerEvent_GetStartTime_NoAlerts(t *testing.T) {
	event := &AlertManagerEvent{}
	assert.True(t, event.GetStartTime().IsZero())
}

func TestAlertManagerEvent_GetSummary(t *testing.T) {
	event := &AlertManagerEvent{
		Alerts: []PrometheusAlert{
			{
				Annotations: map[string]string{
					"summary": "High CPU usage detected",
				},
			},
		},
	}

	assert.Equal(t, "High CPU usage detected", event.GetSummary())
}

func TestAlertManagerEvent_GetSummary_NoAlerts(t *testing.T) {
	event := &AlertManagerEvent{}
	assert.Equal(t, "", event.GetSummary())
}

func TestAlertManagerEvent_GetSummary_NoSummary(t *testing.T) {
	event := &AlertManagerEvent{
		Alerts: []PrometheusAlert{
			{
				Annotations: map[string]string{},
			},
		},
	}

	assert.Equal(t, "", event.GetSummary())
}

func TestAlertManagerEvent_GetDescription(t *testing.T) {
	event := &AlertManagerEvent{
		Alerts: []PrometheusAlert{
			{
				Annotations: map[string]string{
					"description": "The CPU usage has exceeded the threshold",
				},
			},
		},
	}

	assert.Equal(t, "The CPU usage has exceeded the threshold", event.GetDescription())
}

func TestAlertManagerEvent_GetDescription_NoAlerts(t *testing.T) {
	event := &AlertManagerEvent{}
	assert.Equal(t, "", event.GetDescription())
}

func TestAlertManagerEvent_GetDescription_NoDescription(t *testing.T) {
	event := &AlertManagerEvent{
		Alerts: []PrometheusAlert{
			{
				Annotations: map[string]string{},
			},
		},
	}

	assert.Equal(t, "", event.GetDescription())
}

func TestAlertManagerEvent_ToTemplateData(t *testing.T) {
	now := time.Now()
	event := &AlertManagerEvent{
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
		ExternalURL:       "http://alertmanager.example.com",
		CommonAnnotations: map[string]string{"team": "platform"},
		CommonLabels:      map[string]string{"datacenter": "eu-west-1"},
		GroupLabels:       map[string]string{"service": "api"},
	}

	templateData := event.ToTemplateData()

	assert.Equal(t, "test-receiver", templateData.Receiver)
	assert.Equal(t, "firing", templateData.Status)
	assert.Equal(t, 1, len(templateData.Alerts))
	assert.Equal(t, "HighCPUUsage", templateData.Alerts[0].Labels["alertname"])
	assert.Equal(t, "critical", templateData.Alerts[0].Labels["severity"])
	assert.Equal(t, "High CPU usage detected", templateData.Alerts[0].Annotations["summary"])
	assert.Equal(t, "The CPU usage has exceeded the threshold", templateData.Alerts[0].Annotations["description"])
	assert.Equal(t, now, templateData.Alerts[0].StartsAt)
	assert.Equal(t, "http://alertmanager.example.com", templateData.ExternalURL)
	assert.Equal(t, "platform", templateData.CommonAnnotations["team"])
	assert.Equal(t, "eu-west-1", templateData.CommonLabels["datacenter"])
	assert.Equal(t, "api", templateData.GroupLabels["service"])
}
