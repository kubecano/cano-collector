package alert

import (
	"errors"
	"fmt"
	"time"

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

// NewAlertManagerEventFromTemplateData converts template.Data to AlertManagerEvent
func NewAlertManagerEventFromTemplateData(data template.Data) *AlertManagerEvent {
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

	return &AlertManagerEvent{
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

// GetAlertName returns the alert name
func (a *AlertManagerEvent) GetAlertName() string {
	if len(a.Alerts) == 0 {
		return "unknown"
	}

	if name, exists := a.Alerts[0].Labels["alertname"]; exists {
		return name
	}

	return "unknown"
}

// GetSeverity returns the alert severity
func (a *AlertManagerEvent) GetSeverity() string {
	if len(a.Alerts) == 0 {
		return "unknown"
	}

	if severity, exists := a.Alerts[0].Labels["severity"]; exists {
		return severity
	}

	return "unknown"
}

// GetStartTime returns the alert start time
func (a *AlertManagerEvent) GetStartTime() time.Time {
	if len(a.Alerts) == 0 {
		return time.Time{}
	}

	return a.Alerts[0].StartsAt
}

// GetSummary returns the alert summary
func (a *AlertManagerEvent) GetSummary() string {
	if len(a.Alerts) == 0 {
		return ""
	}

	if summary, exists := a.Alerts[0].Annotations["summary"]; exists {
		return summary
	}

	return ""
}

// GetDescription returns the alert description
func (a *AlertManagerEvent) GetDescription() string {
	if len(a.Alerts) == 0 {
		return ""
	}

	if description, exists := a.Alerts[0].Annotations["description"]; exists {
		return description
	}

	return ""
}

// ToTemplateData converts AlertManagerEvent back to template.Data
// This function is useful for maintaining compatibility with existing code
func (a *AlertManagerEvent) ToTemplateData() template.Data {
	alerts := make([]template.Alert, 0, len(a.Alerts))
	for _, alert := range a.Alerts {
		alerts = append(alerts, template.Alert{
			EndsAt:       alert.EndsAt,
			GeneratorURL: alert.GeneratorURL,
			StartsAt:     alert.StartsAt,
			Fingerprint:  alert.Fingerprint,
			Status:       alert.Status,
			Labels:       alert.Labels,
			Annotations:  alert.Annotations,
		})
	}

	return template.Data{
		Alerts:            alerts,
		ExternalURL:       a.ExternalURL,
		CommonAnnotations: a.CommonAnnotations,
		CommonLabels:      a.CommonLabels,
		GroupLabels:       a.GroupLabels,
		Receiver:          a.Receiver,
		Status:            a.Status,
	}
}
