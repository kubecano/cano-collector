package event

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/alertmanager/template"
)

var (
	// ErrMissingReceiver indicates missing receiver field in alert
	ErrMissingReceiver = errors.New("missing receiver field")
	// ErrMissingStatus indicates missing status field in alert
	ErrMissingStatus = errors.New("missing status field")
	// ErrMissingAlerts indicates missing alerts in alert
	ErrMissingAlerts = errors.New("missing alerts")
	// ErrInvalidAlert indicates an invalid alert
	ErrInvalidAlert = errors.New("invalid alert")
)

// EventType represents the type of event
type EventType string

const (
	EventTypeAlertManager EventType = "alertmanager"
	EventTypeKubernetes   EventType = "kubernetes"
	EventTypeScheduled    EventType = "scheduled"
)

// BaseEvent represents the base structure for all events
type BaseEvent struct {
	ID        uuid.UUID `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"`
	Type      EventType `json:"type"`
}

// PrometheusAlert represents a single alert from Prometheus/Alertmanager
type PrometheusAlert struct {
	EndsAt       time.Time         `json:"endsAt"`
	GeneratorURL string            `json:"generatorURL"`
	StartsAt     time.Time         `json:"startsAt"`
	Fingerprint  string            `json:"fingerprint"`
	Status       string            `json:"status"` // "firing" or "resolved"
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
}

// AlertManagerEvent represents an event from Alertmanager containing one or more alerts
type AlertManagerEvent struct {
	BaseEvent
	Alerts            []PrometheusAlert `json:"alerts"`
	ExternalURL       string            `json:"externalURL"`
	GroupKey          string            `json:"groupKey"`
	Version           string            `json:"version"`
	CommonAnnotations map[string]string `json:"commonAnnotations,omitempty"`
	CommonLabels      map[string]string `json:"commonLabels,omitempty"`
	GroupLabels       map[string]string `json:"groupLabels,omitempty"`
	Receiver          string            `json:"receiver"`
	Status            string            `json:"status"`
}

// NewAlertManagerEvent creates a new AlertManagerEvent from template.Data
func NewAlertManagerEvent(data template.Data) *AlertManagerEvent {
	alerts := make([]PrometheusAlert, 0, len(data.Alerts))
	for _, alert := range data.Alerts {
		alerts = append(alerts, PrometheusAlert{
			EndsAt:       alert.EndsAt,
			GeneratorURL: alert.GeneratorURL,
			StartsAt:     alert.StartsAt,
			Fingerprint:  alert.Fingerprint,
			Status:       alert.Status,
			Labels:       alert.Labels,
			Annotations:  alert.Annotations,
		})
	}

	baseEvent := BaseEvent{
		ID:        uuid.New(),
		Timestamp: time.Now(),
		Source:    "alertmanager",
		Type:      EventTypeAlertManager,
	}

	return &AlertManagerEvent{
		BaseEvent:         baseEvent,
		Alerts:            alerts,
		ExternalURL:       data.ExternalURL,
		CommonAnnotations: data.CommonAnnotations,
		CommonLabels:      data.CommonLabels,
		GroupLabels:       data.GroupLabels,
		Receiver:          data.Receiver,
		Status:            data.Status,
	}
}

// Validate checks if the AlertManager event is valid
func (a *AlertManagerEvent) Validate() error {
	if a.Receiver == "" {
		return ErrMissingReceiver
	}

	if a.Status == "" {
		return ErrMissingStatus
	}

	if len(a.Alerts) == 0 {
		return ErrMissingAlerts
	}

	// Validate each alert in the collection
	for i, alert := range a.Alerts {
		if err := validateAlert(alert); err != nil {
			return fmt.Errorf("alert at index %d: %w", i, err)
		}
	}

	return nil
}

// validateAlert checks if a single alert is valid
func validateAlert(alert PrometheusAlert) error {
	if alert.Status == "" {
		return errors.New("missing status field in alert")
	}

	// Check if status has a valid value
	if alert.Status != "firing" && alert.Status != "resolved" {
		return fmt.Errorf("invalid status value: %s", alert.Status)
	}

	// Check if alert has a name
	if _, exists := alert.Labels["alertname"]; !exists {
		return errors.New("missing alertname label")
	}

	// Check if StartsAt is not zero for firing alerts
	if alert.Status == "firing" && alert.StartsAt.IsZero() {
		return errors.New("missing start time for firing alert")
	}

	return nil
}

// GetAlertName returns the alert name from the first alert
func (a *AlertManagerEvent) GetAlertName() string {
	if len(a.Alerts) == 0 {
		return ""
	}
	return a.Alerts[0].Labels["alertname"]
}

// GetSeverity returns the severity from the first alert
func (a *AlertManagerEvent) GetSeverity() string {
	if len(a.Alerts) == 0 {
		return ""
	}
	return a.Alerts[0].Labels["severity"]
}

// GetStartTime returns the start time from the first alert
func (a *AlertManagerEvent) GetStartTime() time.Time {
	if len(a.Alerts) == 0 {
		return time.Time{}
	}
	return a.Alerts[0].StartsAt
}

// GetSummary returns the summary annotation from the first alert
func (a *AlertManagerEvent) GetSummary() string {
	if len(a.Alerts) == 0 {
		return ""
	}
	return a.Alerts[0].Annotations["summary"]
}

// GetDescription returns the description annotation from the first alert
func (a *AlertManagerEvent) GetDescription() string {
	if len(a.Alerts) == 0 {
		return ""
	}
	return a.Alerts[0].Annotations["description"]
}

// GetLabels returns labels from the first alert
func (a *AlertManagerEvent) GetLabels() map[string]string {
	if len(a.Alerts) == 0 {
		return nil
	}
	return a.Alerts[0].Labels
}

// GetAnnotations returns annotations from the first alert
func (a *AlertManagerEvent) GetAnnotations() map[string]string {
	if len(a.Alerts) == 0 {
		return nil
	}
	return a.Alerts[0].Annotations
}

// GetNamespace returns the namespace from the first alert
func (a *AlertManagerEvent) GetNamespace() string {
	if len(a.Alerts) == 0 {
		return ""
	}
	return a.Alerts[0].Labels["namespace"]
}

// GetStatus returns the status from the first alert
func (a *AlertManagerEvent) GetStatus() string {
	if len(a.Alerts) == 0 {
		return ""
	}
	return a.Alerts[0].Status
}
