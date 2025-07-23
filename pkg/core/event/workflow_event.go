package event

import (
	"time"

	"github.com/google/uuid"
)

// WorkflowEvent defines the interface for all events that can trigger workflows
type WorkflowEvent interface {
	GetID() uuid.UUID
	GetTimestamp() time.Time
	GetSource() string
	GetType() EventType
	GetEventData() interface{}

	// Methods for extracting common alert information
	GetAlertName() string
	GetStatus() string
	GetSeverity() string
	GetNamespace() string
}

// AlertManagerWorkflowEvent wraps AlertManagerEvent to implement WorkflowEvent interface
type AlertManagerWorkflowEvent struct {
	*AlertManagerEvent
}

// NewAlertManagerWorkflowEvent creates a new AlertManagerWorkflowEvent
func NewAlertManagerWorkflowEvent(alertEvent *AlertManagerEvent) *AlertManagerWorkflowEvent {
	return &AlertManagerWorkflowEvent{
		AlertManagerEvent: alertEvent,
	}
}

// GetID returns the event ID
func (e *AlertManagerWorkflowEvent) GetID() uuid.UUID {
	return e.ID
}

// GetTimestamp returns the event timestamp
func (e *AlertManagerWorkflowEvent) GetTimestamp() time.Time {
	return e.Timestamp
}

// GetSource returns the event source
func (e *AlertManagerWorkflowEvent) GetSource() string {
	return e.Source
}

// GetType returns the event type
func (e *AlertManagerWorkflowEvent) GetType() EventType {
	return e.Type
}

// GetEventData returns the underlying AlertManagerEvent
func (e *AlertManagerWorkflowEvent) GetEventData() interface{} {
	return e.AlertManagerEvent
}

// GetAlertManagerEvent returns the underlying AlertManagerEvent for internal use
func (e *AlertManagerWorkflowEvent) GetAlertManagerEvent() *AlertManagerEvent {
	return e.AlertManagerEvent
}

// GetAlertName returns the alert name
func (e *AlertManagerWorkflowEvent) GetAlertName() string {
	return e.AlertManagerEvent.GetAlertName()
}

// GetStatus returns the alert status
func (e *AlertManagerWorkflowEvent) GetStatus() string {
	return e.Status
}

// GetSeverity returns the alert severity
func (e *AlertManagerWorkflowEvent) GetSeverity() string {
	return e.AlertManagerEvent.GetSeverity()
}

// GetNamespace returns the alert namespace
func (e *AlertManagerWorkflowEvent) GetNamespace() string {
	return e.AlertManagerEvent.GetNamespace()
}
