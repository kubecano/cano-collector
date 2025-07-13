package workflow

import (
	"testing"
	"time"

	"github.com/kubecano/cano-collector/config/workflow"
	"github.com/kubecano/cano-collector/pkg/core/event"
	"github.com/prometheus/alertmanager/template"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	assert.Len(t, matchingWorkflows, 0)
}
