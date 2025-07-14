package event

import (
	"testing"
	"time"

	"github.com/prometheus/alertmanager/template"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAlertManagerEventFromTemplateData(t *testing.T) {
	// Create test template data
	templateData := template.Data{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []template.Alert{
			{
				Status:      "firing",
				Labels:      map[string]string{"alertname": "TestAlert", "severity": "warning"},
				Annotations: map[string]string{"summary": "Test alert"},
				StartsAt:    time.Now(),
				EndsAt:      time.Now().Add(time.Hour),
			},
		},
		ExternalURL:       "http://alertmanager:9093",
		CommonAnnotations: map[string]string{"common": "annotation"},
		CommonLabels:      map[string]string{"common": "label"},
		GroupLabels:       map[string]string{"group": "label"},
	}

	// Create AlertManagerEvent
	event := NewAlertManagerEvent(templateData)

	// Verify event structure
	assert.NotNil(t, event)
	assert.Equal(t, "test-receiver", event.Receiver)
	assert.Equal(t, "firing", event.Status)
	assert.Len(t, event.Alerts, 1)
	assert.Equal(t, "TestAlert", event.GetAlertName())
	assert.Equal(t, "warning", event.GetSeverity())
	assert.Equal(t, "Test alert", event.GetSummary())
	assert.Equal(t, EventTypeAlertManager, event.Type)
	assert.Equal(t, "alertmanager", event.Source)
	assert.NotZero(t, event.ID)
	assert.NotZero(t, event.Timestamp)
}

func TestAlertManagerEvent_Validate_ValidAlert(t *testing.T) {
	event := &AlertManagerEvent{
		BaseEvent: BaseEvent{
			ID:        [16]byte{},
			Timestamp: time.Now(),
			Source:    "alertmanager",
			Type:      EventTypeAlertManager,
		},
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []PrometheusAlert{
			{
				Status:      "firing",
				Labels:      map[string]string{"alertname": "TestAlert"},
				StartsAt:    time.Now(),
				EndsAt:      time.Now().Add(time.Hour),
				Fingerprint: "test-fingerprint",
			},
		},
	}

	err := event.Validate()
	require.NoError(t, err)
}

func TestAlertManagerEvent_Validate_MissingReceiver(t *testing.T) {
	event := &AlertManagerEvent{
		BaseEvent: BaseEvent{
			ID:        [16]byte{},
			Timestamp: time.Now(),
			Source:    "alertmanager",
			Type:      EventTypeAlertManager,
		},
		Status: "firing",
		Alerts: []PrometheusAlert{
			{
				Status:      "firing",
				Labels:      map[string]string{"alertname": "TestAlert"},
				StartsAt:    time.Now(),
				EndsAt:      time.Now().Add(time.Hour),
				Fingerprint: "test-fingerprint",
			},
		},
	}

	err := event.Validate()
	require.Error(t, err)
	assert.Equal(t, ErrMissingReceiver, err)
}

func TestAlertManagerEvent_Validate_MissingStatus(t *testing.T) {
	event := &AlertManagerEvent{
		BaseEvent: BaseEvent{
			ID:        [16]byte{},
			Timestamp: time.Now(),
			Source:    "alertmanager",
			Type:      EventTypeAlertManager,
		},
		Receiver: "test-receiver",
		Alerts: []PrometheusAlert{
			{
				Status:      "firing",
				Labels:      map[string]string{"alertname": "TestAlert"},
				StartsAt:    time.Now(),
				EndsAt:      time.Now().Add(time.Hour),
				Fingerprint: "test-fingerprint",
			},
		},
	}

	err := event.Validate()
	require.Error(t, err)
	assert.Equal(t, ErrMissingStatus, err)
}

func TestAlertManagerEvent_Validate_MissingAlerts(t *testing.T) {
	event := &AlertManagerEvent{
		BaseEvent: BaseEvent{
			ID:        [16]byte{},
			Timestamp: time.Now(),
			Source:    "alertmanager",
			Type:      EventTypeAlertManager,
		},
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts:   []PrometheusAlert{},
	}

	err := event.Validate()
	require.Error(t, err)
	assert.Equal(t, ErrMissingAlerts, err)
}

func TestAlertManagerEvent_Validate_InvalidAlertStatus(t *testing.T) {
	event := &AlertManagerEvent{
		BaseEvent: BaseEvent{
			ID:        [16]byte{},
			Timestamp: time.Now(),
			Source:    "alertmanager",
			Type:      EventTypeAlertManager,
		},
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []PrometheusAlert{
			{
				Status:      "invalid-status",
				Labels:      map[string]string{"alertname": "TestAlert"},
				StartsAt:    time.Now(),
				EndsAt:      time.Now().Add(time.Hour),
				Fingerprint: "test-fingerprint",
			},
		},
	}

	err := event.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid status value")
}

func TestAlertManagerEvent_Validate_MissingAlertName(t *testing.T) {
	event := &AlertManagerEvent{
		BaseEvent: BaseEvent{
			ID:        [16]byte{},
			Timestamp: time.Now(),
			Source:    "alertmanager",
			Type:      EventTypeAlertManager,
		},
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []PrometheusAlert{
			{
				Status:      "firing",
				Labels:      map[string]string{},
				StartsAt:    time.Now(),
				EndsAt:      time.Now().Add(time.Hour),
				Fingerprint: "test-fingerprint",
			},
		},
	}

	err := event.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing alertname label")
}

func TestAlertManagerEvent_Validate_MissingStartTime(t *testing.T) {
	event := &AlertManagerEvent{
		BaseEvent: BaseEvent{
			ID:        [16]byte{},
			Timestamp: time.Now(),
			Source:    "alertmanager",
			Type:      EventTypeAlertManager,
		},
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []PrometheusAlert{
			{
				Status:      "firing",
				Labels:      map[string]string{"alertname": "TestAlert"},
				StartsAt:    time.Time{},
				EndsAt:      time.Now().Add(time.Hour),
				Fingerprint: "test-fingerprint",
			},
		},
	}

	err := event.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing start time for firing alert")
}

func TestAlertManagerEvent_GetAlertName(t *testing.T) {
	event := &AlertManagerEvent{
		Alerts: []PrometheusAlert{
			{
				Labels: map[string]string{"alertname": "TestAlert"},
			},
		},
	}

	assert.Equal(t, "TestAlert", event.GetAlertName())
}

func TestAlertManagerEvent_GetAlertName_NoAlerts(t *testing.T) {
	event := &AlertManagerEvent{
		Alerts: []PrometheusAlert{},
	}

	assert.Empty(t, event.GetAlertName())
}

func TestAlertManagerEvent_GetAlertName_NoAlertName(t *testing.T) {
	event := &AlertManagerEvent{
		Alerts: []PrometheusAlert{
			{
				Labels: map[string]string{},
			},
		},
	}

	assert.Empty(t, event.GetAlertName())
}

func TestAlertManagerEvent_GetSeverity(t *testing.T) {
	event := &AlertManagerEvent{
		Alerts: []PrometheusAlert{
			{
				Labels: map[string]string{"severity": "critical"},
			},
		},
	}

	assert.Equal(t, "critical", event.GetSeverity())
}

func TestAlertManagerEvent_GetSeverity_NoAlerts(t *testing.T) {
	event := &AlertManagerEvent{
		Alerts: []PrometheusAlert{},
	}

	assert.Empty(t, event.GetSeverity())
}

func TestAlertManagerEvent_GetSeverity_NoSeverity(t *testing.T) {
	event := &AlertManagerEvent{
		Alerts: []PrometheusAlert{
			{
				Labels: map[string]string{},
			},
		},
	}

	assert.Empty(t, event.GetSeverity())
}

func TestAlertManagerEvent_GetStartTime(t *testing.T) {
	startTime := time.Now()
	event := &AlertManagerEvent{
		Alerts: []PrometheusAlert{
			{
				StartsAt: startTime,
			},
		},
	}

	assert.Equal(t, startTime, event.GetStartTime())
}

func TestAlertManagerEvent_GetStartTime_NoAlerts(t *testing.T) {
	event := &AlertManagerEvent{
		Alerts: []PrometheusAlert{},
	}

	assert.Equal(t, time.Time{}, event.GetStartTime())
}

func TestAlertManagerEvent_GetSummary(t *testing.T) {
	event := &AlertManagerEvent{
		Alerts: []PrometheusAlert{
			{
				Annotations: map[string]string{"summary": "Test summary"},
			},
		},
	}

	assert.Equal(t, "Test summary", event.GetSummary())
}

func TestAlertManagerEvent_GetSummary_NoAlerts(t *testing.T) {
	event := &AlertManagerEvent{
		Alerts: []PrometheusAlert{},
	}

	assert.Empty(t, event.GetSummary())
}

func TestAlertManagerEvent_GetSummary_NoSummary(t *testing.T) {
	event := &AlertManagerEvent{
		Alerts: []PrometheusAlert{
			{
				Annotations: map[string]string{},
			},
		},
	}

	assert.Empty(t, event.GetSummary())
}

func TestAlertManagerEvent_GetDescription(t *testing.T) {
	event := &AlertManagerEvent{
		Alerts: []PrometheusAlert{
			{
				Annotations: map[string]string{"description": "Test description"},
			},
		},
	}

	assert.Equal(t, "Test description", event.GetDescription())
}

func TestAlertManagerEvent_GetDescription_NoAlerts(t *testing.T) {
	event := &AlertManagerEvent{
		Alerts: []PrometheusAlert{},
	}

	assert.Empty(t, event.GetDescription())
}

func TestAlertManagerEvent_GetDescription_NoDescription(t *testing.T) {
	event := &AlertManagerEvent{
		Alerts: []PrometheusAlert{
			{
				Annotations: map[string]string{},
			},
		},
	}

	assert.Empty(t, event.GetDescription())
}

func TestAlertManagerEvent_GetLabels(t *testing.T) {
	labels := map[string]string{"alertname": "TestAlert", "severity": "critical"}
	event := &AlertManagerEvent{
		Alerts: []PrometheusAlert{
			{
				Labels: labels,
			},
		},
	}

	assert.Equal(t, labels, event.GetLabels())
}

func TestAlertManagerEvent_GetAnnotations(t *testing.T) {
	annotations := map[string]string{"summary": "Test summary", "description": "Test description"}
	event := &AlertManagerEvent{
		Alerts: []PrometheusAlert{
			{
				Annotations: annotations,
			},
		},
	}

	assert.Equal(t, annotations, event.GetAnnotations())
}

func TestAlertManagerEvent_GetNamespace(t *testing.T) {
	event := &AlertManagerEvent{
		Alerts: []PrometheusAlert{
			{
				Labels: map[string]string{"namespace": "production"},
			},
		},
	}

	assert.Equal(t, "production", event.GetNamespace())
}
