package workflow

import (
	"testing"
	"time"

	"github.com/prometheus/alertmanager/template"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kubecano/cano-collector/config/workflow"
	"github.com/kubecano/cano-collector/pkg/core/event"
)

func TestWorkflowEngine_SelectWorkflows_EventBased(t *testing.T) {
	// Create test configuration
	config := &workflow.WorkflowConfig{
		ActiveWorkflows: []workflow.WorkflowDefinition{
			{
				Name: "test-firing-workflow",
				Triggers: []workflow.TriggerDefinition{
					{
						OnAlertmanagerAlert: &workflow.AlertmanagerAlertTrigger{
							Status: "firing",
						},
					},
				},
				Actions: []workflow.ActionDefinition{
					{
						ActionType: "create_issue",
						RawData: map[string]interface{}{
							"action_type": "create_issue",
							"data": map[string]interface{}{
								"title": "{{.alert_name}}",
							},
						},
					},
				},
			},
			{
				Name: "test-resolved-workflow",
				Triggers: []workflow.TriggerDefinition{
					{
						OnAlertmanagerAlert: &workflow.AlertmanagerAlertTrigger{
							Status: "resolved",
						},
					},
				},
				Actions: []workflow.ActionDefinition{
					{
						ActionType: "resolve_issue",
						RawData: map[string]interface{}{
							"action_type": "resolve_issue",
							"data":        map[string]interface{}{},
						},
					},
				},
			},
		},
	}

	engine := NewWorkflowEngine(config)

	// Create test AlertManagerEvent
	templateData := template.Data{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []template.Alert{
			{
				Status:      "firing",
				Labels:      map[string]string{"alertname": "TestAlert", "severity": "warning", "namespace": "default"},
				Annotations: map[string]string{"summary": "Test alert summary"},
				StartsAt:    time.Now(),
			},
		},
	}

	alertEvent := event.NewAlertManagerEvent(templateData)

	// Test workflow selection
	matchingWorkflows := engine.SelectWorkflows(alertEvent)

	// Should match the firing workflow
	require.Len(t, matchingWorkflows, 1)
	assert.Equal(t, "test-firing-workflow", matchingWorkflows[0].Name)

	// Test with resolved alert
	templateData.Status = "resolved"
	templateData.Alerts[0].Status = "resolved"
	resolvedEvent := event.NewAlertManagerEvent(templateData)

	matchingWorkflows = engine.SelectWorkflows(resolvedEvent)

	// Should match the resolved workflow
	require.Len(t, matchingWorkflows, 1)
	assert.Equal(t, "test-resolved-workflow", matchingWorkflows[0].Name)
}

func TestWorkflowEngine_AlertEventConversion(t *testing.T) {
	// Create test template data
	templateData := template.Data{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []template.Alert{
			{
				Status:      "firing",
				Labels:      map[string]string{"alertname": "HighCPUUsage", "severity": "critical", "namespace": "production", "pod": "app-pod-123"},
				Annotations: map[string]string{"summary": "High CPU usage detected", "description": "CPU usage is above 90%"},
				StartsAt:    time.Now(),
			},
		},
	}

	// Convert to AlertManagerEvent
	alertEvent := event.NewAlertManagerEvent(templateData)

	// Verify event structure
	assert.Equal(t, event.EventTypeAlertManager, alertEvent.Type)
	assert.Equal(t, "alertmanager", alertEvent.Source)
	assert.Equal(t, "HighCPUUsage", alertEvent.GetAlertName())
	assert.Equal(t, "firing", alertEvent.GetStatus())
	assert.Equal(t, "critical", alertEvent.GetSeverity())
	assert.Equal(t, "production", alertEvent.GetNamespace())
	assert.Equal(t, "High CPU usage detected", alertEvent.GetSummary())
	assert.Equal(t, "app-pod-123", alertEvent.GetLabels()["pod"])

	// Verify event has proper base structure
	assert.NotZero(t, alertEvent.ID)
	assert.NotZero(t, alertEvent.Timestamp)
}

func TestWorkflowEngine_TriggerMatching(t *testing.T) {
	config := &workflow.WorkflowConfig{
		ActiveWorkflows: []workflow.WorkflowDefinition{
			{
				Name: "specific-alert-workflow",
				Triggers: []workflow.TriggerDefinition{
					{
						OnAlertmanagerAlert: &workflow.AlertmanagerAlertTrigger{
							AlertName: "HighCPUUsage",
							Status:    "firing",
							Severity:  "critical",
							Namespace: "production",
						},
					},
				},
				Actions: []workflow.ActionDefinition{
					{
						ActionType: "create_issue",
						RawData: map[string]interface{}{
							"action_type": "create_issue",
							"data": map[string]interface{}{
								"title": "Critical CPU Alert",
							},
						},
					},
				},
			},
		},
	}

	engine := NewWorkflowEngine(config)

	// Test matching event
	templateData := template.Data{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []template.Alert{
			{
				Status:   "firing",
				Labels:   map[string]string{"alertname": "HighCPUUsage", "severity": "critical", "namespace": "production"},
				StartsAt: time.Now(),
			},
		},
	}

	alertEvent := event.NewAlertManagerEvent(templateData)
	matchingWorkflows := engine.SelectWorkflows(alertEvent)
	assert.Len(t, matchingWorkflows, 1)

	// Test non-matching event (different alert name)
	templateData.Alerts[0].Labels["alertname"] = "HighMemoryUsage"
	alertEvent = event.NewAlertManagerEvent(templateData)
	matchingWorkflows = engine.SelectWorkflows(alertEvent)
	assert.Empty(t, matchingWorkflows)
}

func TestWorkflowEngine_ExecuteWorkflow(t *testing.T) {
	config := &workflow.WorkflowConfig{
		ActiveWorkflows: []workflow.WorkflowDefinition{
			{
				Name: "test-workflow",
				Triggers: []workflow.TriggerDefinition{
					{
						OnAlertmanagerAlert: &workflow.AlertmanagerAlertTrigger{
							Status: "firing",
						},
					},
				},
				Actions: []workflow.ActionDefinition{
					{
						ActionType: "create_issue",
						RawData: map[string]interface{}{
							"action_type": "create_issue",
							"data": map[string]interface{}{
								"title": "Test Issue",
							},
						},
					},
				},
			},
		},
	}

	engine := NewWorkflowEngine(config)

	// Create test AlertManagerEvent
	templateData := template.Data{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []template.Alert{
			{
				Status:   "firing",
				Labels:   map[string]string{"alertname": "TestAlert", "severity": "warning"},
				StartsAt: time.Now(),
			},
		},
	}

	alertEvent := event.NewAlertManagerEvent(templateData)
	workflow := &config.ActiveWorkflows[0]

	// Test ExecuteWorkflow - should return nil (no-op implementation)
	err := engine.ExecuteWorkflow(workflow, alertEvent)
	assert.NoError(t, err)
}

func TestWorkflowEngine_MatchesTrigger_NoAlertmanagerTrigger(t *testing.T) {
	config := &workflow.WorkflowConfig{
		ActiveWorkflows: []workflow.WorkflowDefinition{},
	}

	engine := NewWorkflowEngine(config)

	// Create trigger without OnAlertmanagerAlert
	trigger := &workflow.TriggerDefinition{
		OnAlertmanagerAlert: nil, // This will cause matchesTrigger to return false
	}

	// Create test AlertManagerEvent
	templateData := template.Data{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []template.Alert{
			{
				Status:   "firing",
				Labels:   map[string]string{"alertname": "TestAlert"},
				StartsAt: time.Now(),
			},
		},
	}

	alertEvent := event.NewAlertManagerEvent(templateData)

	// Test that trigger without OnAlertmanagerAlert returns false
	matches := engine.matchesTrigger(trigger, alertEvent)
	assert.False(t, matches)
}

func TestWorkflowEngine_MatchesAlertmanagerAlertTrigger_AllFields(t *testing.T) {
	config := &workflow.WorkflowConfig{
		ActiveWorkflows: []workflow.WorkflowDefinition{},
	}

	engine := NewWorkflowEngine(config)

	// Test trigger with all fields specified
	trigger := &workflow.AlertmanagerAlertTrigger{
		AlertName: "SpecificAlert",
		Status:    "firing",
		Severity:  "critical",
		Namespace: "production",
	}

	// Create matching AlertManagerEvent
	templateData := template.Data{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []template.Alert{
			{
				Status: "firing",
				Labels: map[string]string{
					"alertname": "SpecificAlert",
					"severity":  "critical",
					"namespace": "production",
				},
				StartsAt: time.Now(),
			},
		},
	}

	alertEvent := event.NewAlertManagerEvent(templateData)

	// Should match
	matches := engine.matchesAlertmanagerAlertTrigger(trigger, alertEvent)
	assert.True(t, matches)

	// Test non-matching alert name
	templateData.Alerts[0].Labels["alertname"] = "DifferentAlert"
	alertEvent = event.NewAlertManagerEvent(templateData)
	matches = engine.matchesAlertmanagerAlertTrigger(trigger, alertEvent)
	assert.False(t, matches)

	// Test non-matching status
	templateData.Alerts[0].Labels["alertname"] = "SpecificAlert"
	templateData.Alerts[0].Status = "resolved"
	alertEvent = event.NewAlertManagerEvent(templateData)
	matches = engine.matchesAlertmanagerAlertTrigger(trigger, alertEvent)
	assert.False(t, matches)

	// Test non-matching severity
	templateData.Alerts[0].Status = "firing"
	templateData.Alerts[0].Labels["severity"] = "warning"
	alertEvent = event.NewAlertManagerEvent(templateData)
	matches = engine.matchesAlertmanagerAlertTrigger(trigger, alertEvent)
	assert.False(t, matches)

	// Test non-matching namespace
	templateData.Alerts[0].Labels["severity"] = "critical"
	templateData.Alerts[0].Labels["namespace"] = "staging"
	alertEvent = event.NewAlertManagerEvent(templateData)
	matches = engine.matchesAlertmanagerAlertTrigger(trigger, alertEvent)
	assert.False(t, matches)
}

func TestWorkflowEngine_MatchesAlertmanagerAlertTrigger_EmptyTrigger(t *testing.T) {
	config := &workflow.WorkflowConfig{
		ActiveWorkflows: []workflow.WorkflowDefinition{},
	}

	engine := NewWorkflowEngine(config)

	// Test empty trigger (should match any alert)
	trigger := &workflow.AlertmanagerAlertTrigger{}

	// Create test AlertManagerEvent
	templateData := template.Data{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []template.Alert{
			{
				Status:   "firing",
				Labels:   map[string]string{"alertname": "AnyAlert", "severity": "warning"},
				StartsAt: time.Now(),
			},
		},
	}

	alertEvent := event.NewAlertManagerEvent(templateData)

	// Empty trigger should match any alert
	matches := engine.matchesAlertmanagerAlertTrigger(trigger, alertEvent)
	assert.True(t, matches)
}
