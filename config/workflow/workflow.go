package workflow

import (
	"fmt"
	"time"
)

// WorkflowDefinition represents a complete workflow configuration
type WorkflowDefinition struct {
	Name     string              `yaml:"name" json:"name"`
	Triggers []TriggerDefinition `yaml:"triggers" json:"triggers"`
	Actions  []ActionDefinition  `yaml:"actions" json:"actions"`
	Stop     bool                `yaml:"stop,omitempty" json:"stop,omitempty"`
}

// TriggerDefinition represents a workflow trigger configuration
type TriggerDefinition struct {
	OnAlertmanagerAlert *AlertmanagerAlertTrigger `yaml:"on_alertmanager_alert,omitempty" json:"on_alertmanager_alert,omitempty"`
	// Future: OnKubernetesEvent will be added in later sprints
	// OnKubernetesEvent   *KubernetesEventTrigger   `yaml:"on_kubernetes_event,omitempty" json:"on_kubernetes_event,omitempty"`
}

// ActionDefinition represents a workflow action configuration
// NOTE: Concrete action implementations will be added in "Workflow Actions Foundation" task
type ActionDefinition struct {
	// Future actions will be added here in subsequent tasks
	// LogsEnricher    *LogsEnricherAction    `yaml:"logs_enricher,omitempty" json:"logs_enricher,omitempty"`
	// MetricsEnricher *MetricsEnricherAction `yaml:"metrics_enricher,omitempty" json:"metrics_enricher,omitempty"`

	// For now, we store raw action data for pass-through processing
	ActionType string                 `yaml:"-" json:"-"` // Internal field to track action type
	RawData    map[string]interface{} `yaml:",inline" json:",inline"`
}

// AlertmanagerAlertTrigger represents trigger conditions for Alertmanager alerts
type AlertmanagerAlertTrigger struct {
	AlertName string `yaml:"alert_name,omitempty" json:"alert_name,omitempty"`
	Status    string `yaml:"status,omitempty" json:"status,omitempty"`       // firing, resolved, all
	Severity  string `yaml:"severity,omitempty" json:"severity,omitempty"`   // critical, warning, info
	Namespace string `yaml:"namespace,omitempty" json:"namespace,omitempty"` // kubernetes namespace
	Instance  string `yaml:"instance,omitempty" json:"instance,omitempty"`   // prometheus instance
	PodName   string `yaml:"pod_name,omitempty" json:"pod_name,omitempty"`   // pod name prefix
}

// GetTriggerType returns the type identifier for this trigger
func (t *TriggerDefinition) GetTriggerType() string {
	if t.OnAlertmanagerAlert != nil {
		return "alertmanager_alert"
	}
	// Future trigger types will be added here
	return "unknown"
}

// Validate checks if the workflow definition is valid
func (w *WorkflowDefinition) Validate() error {
	if w.Name == "" {
		return fmt.Errorf("workflow name cannot be empty")
	}

	if len(w.Triggers) == 0 {
		return fmt.Errorf("workflow '%s' must have at least one trigger", w.Name)
	}

	if len(w.Actions) == 0 {
		return fmt.Errorf("workflow '%s' must have at least one action", w.Name)
	}

	// Validate each trigger
	for i, trigger := range w.Triggers {
		if err := trigger.Validate(); err != nil {
			return fmt.Errorf("workflow '%s' trigger %d validation failed: %w", w.Name, i, err)
		}
	}

	return nil
}

// Validate checks if the trigger definition is valid
func (t *TriggerDefinition) Validate() error {
	triggerCount := 0

	if t.OnAlertmanagerAlert != nil {
		triggerCount++
		if err := t.OnAlertmanagerAlert.Validate(); err != nil {
			return fmt.Errorf("alertmanager_alert trigger validation failed: %w", err)
		}
	}

	if triggerCount == 0 {
		return fmt.Errorf("trigger definition must specify exactly one trigger type")
	}

	if triggerCount > 1 {
		return fmt.Errorf("trigger definition must specify exactly one trigger type, found %d", triggerCount)
	}

	return nil
}

// Validate checks if the alertmanager alert trigger is valid
func (a *AlertmanagerAlertTrigger) Validate() error {
	// AlertName is optional (empty means match all)

	// Validate status if provided
	if a.Status != "" {
		validStatuses := []string{"firing", "resolved", "all"}
		valid := false
		for _, status := range validStatuses {
			if a.Status == status {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid status '%s', must be one of: firing, resolved, all", a.Status)
		}
	}

	// Validate severity if provided
	if a.Severity != "" {
		validSeverities := []string{"critical", "warning", "info"}
		valid := false
		for _, severity := range validSeverities {
			if a.Severity == severity {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid severity '%s', must be one of: critical, warning, info", a.Severity)
		}
	}

	return nil
}

// GetID returns a unique identifier for the workflow
func (w *WorkflowDefinition) GetID() string {
	return w.Name
}

// HasTriggerType checks if workflow has a trigger of the specified type
func (w *WorkflowDefinition) HasTriggerType(triggerType string) bool {
	for _, trigger := range w.Triggers {
		if trigger.GetTriggerType() == triggerType {
			return true
		}
	}
	return false
}

// WorkflowMetadata contains additional metadata about workflow execution
type WorkflowMetadata struct {
	ExecutionID  string    `json:"execution_id"`
	WorkflowName string    `json:"workflow_name"`
	StartTime    time.Time `json:"start_time"`
	TriggerType  string    `json:"trigger_type"`
	ActionsCount int       `json:"actions_count"`
}

// NewWorkflowMetadata creates new workflow metadata
func NewWorkflowMetadata(workflowName, triggerType string, actionsCount int) *WorkflowMetadata {
	return &WorkflowMetadata{
		ExecutionID:  fmt.Sprintf("%s-%d", workflowName, time.Now().UnixNano()),
		WorkflowName: workflowName,
		StartTime:    time.Now(),
		TriggerType:  triggerType,
		ActionsCount: actionsCount,
	}
}
