package workflow

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/alertmanager/template"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kubecano/cano-collector/config/workflow"
	"github.com/kubecano/cano-collector/pkg/core/event"
)

// Test helper functions

// createTestTemplateData creates a basic template.Data for testing
func createTestTemplateData(status, alertname, severity, namespace string) template.Data {
	return template.Data{
		Receiver: "test-receiver",
		Status:   status,
		Alerts: []template.Alert{
			{
				Status: status,
				Labels: map[string]string{
					"alertname": alertname,
					"severity":  severity,
					"namespace": namespace,
				},
				Annotations: map[string]string{
					"summary": "Test alert summary",
				},
				StartsAt: time.Now(),
			},
		},
	}
}

// createTestWorkflowEvent creates a WorkflowEvent from basic parameters
func createTestWorkflowEvent(status, alertname, severity, namespace string) event.WorkflowEvent {
	templateData := createTestTemplateData(status, alertname, severity, namespace)
	alertEvent := event.NewAlertManagerEvent(templateData)
	return event.NewAlertManagerWorkflowEvent(alertEvent)
}

// createBasicWorkflowConfig creates a basic workflow config for testing
func createBasicWorkflowConfig() *workflow.WorkflowConfig {
	return &workflow.WorkflowConfig{
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
}

// createSpecificWorkflowConfig creates a workflow config with specific trigger criteria
func createSpecificWorkflowConfig(alertName, status, severity, namespace string) *workflow.WorkflowConfig {
	return &workflow.WorkflowConfig{
		ActiveWorkflows: []workflow.WorkflowDefinition{
			{
				Name: "specific-alert-workflow",
				Triggers: []workflow.TriggerDefinition{
					{
						OnAlertmanagerAlert: &workflow.AlertmanagerAlertTrigger{
							AlertName: alertName,
							Status:    status,
							Severity:  severity,
							Namespace: namespace,
						},
					},
				},
				Actions: []workflow.ActionDefinition{
					{
						ActionType: "escalate",
						RawData: map[string]interface{}{
							"action_type": "escalate",
						},
					},
				},
			},
		},
	}
}

// createTestEngine creates a WorkflowEngine with nil executor for testing
func createTestEngine(config *workflow.WorkflowConfig) *WorkflowEngine {
	return NewWorkflowEngine(config, nil)
}

// Test functions

func TestWorkflowEngine_SelectWorkflows_EventBased(t *testing.T) {
	config := createBasicWorkflowConfig()
	engine := createTestEngine(config)

	// Test firing workflow
	workflowEvent := createTestWorkflowEvent("firing", "TestAlert", "warning", "default")
	matchingWorkflows := engine.SelectWorkflows(workflowEvent)

	// Should match the firing workflow
	require.Len(t, matchingWorkflows, 1)
	assert.Equal(t, "test-firing-workflow", matchingWorkflows[0].Name)

	// Test resolved workflow
	workflowEvent = createTestWorkflowEvent("resolved", "TestAlert", "warning", "default")
	matchingWorkflows = engine.SelectWorkflows(workflowEvent)

	// Should match the resolved workflow
	require.Len(t, matchingWorkflows, 1)
	assert.Equal(t, "test-resolved-workflow", matchingWorkflows[0].Name)
}

func TestWorkflowEngine_SelectWorkflows_MultipleFilters(t *testing.T) {
	config := createSpecificWorkflowConfig("HighCPU", "firing", "critical", "production")
	engine := createTestEngine(config)

	tests := []struct {
		name            string
		status          string
		alertname       string
		severity        string
		namespace       string
		expectWorkflows int
	}{
		{
			name:            "matches all criteria",
			status:          "firing",
			alertname:       "HighCPU",
			severity:        "critical",
			namespace:       "production",
			expectWorkflows: 1,
		},
		{
			name:            "wrong alert name",
			status:          "firing",
			alertname:       "LowMemory",
			severity:        "critical",
			namespace:       "production",
			expectWorkflows: 0,
		},
		{
			name:            "wrong severity",
			status:          "firing",
			alertname:       "HighCPU",
			severity:        "warning",
			namespace:       "production",
			expectWorkflows: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workflowEvent := createTestWorkflowEvent(tt.status, tt.alertname, tt.severity, tt.namespace)
			matchingWorkflows := engine.SelectWorkflows(workflowEvent)
			assert.Len(t, matchingWorkflows, tt.expectWorkflows)
		})
	}
}

func TestWorkflowEngine_AlertEventConversion(t *testing.T) {
	// Create test template data with enhanced labels
	templateData := createTestTemplateData("firing", "HighCPUUsage", "critical", "production")
	templateData.Alerts[0].Labels["pod"] = "app-pod-123"
	templateData.Alerts[0].Annotations["description"] = "CPU usage is above 90%"

	// Convert to AlertManagerEvent
	alertEvent := event.NewAlertManagerEvent(templateData)

	// Verify event structure
	assert.Equal(t, event.EventTypeAlertManager, alertEvent.Type)
	assert.Equal(t, "alertmanager", alertEvent.Source)
	assert.Equal(t, "HighCPUUsage", alertEvent.GetAlertName())
	assert.Equal(t, "firing", alertEvent.GetStatus())
	assert.Equal(t, "critical", alertEvent.GetSeverity())
	assert.Equal(t, "production", alertEvent.GetNamespace())
	assert.Equal(t, "Test alert summary", alertEvent.GetSummary())
	assert.Equal(t, "app-pod-123", alertEvent.GetLabels()["pod"])

	// Verify event has proper base structure
	assert.NotZero(t, alertEvent.ID)
	assert.NotZero(t, alertEvent.Timestamp)
}

func TestWorkflowEngine_TriggerMatching(t *testing.T) {
	config := createSpecificWorkflowConfig("HighCPUUsage", "firing", "critical", "production")
	engine := createTestEngine(config)

	// Test matching event
	workflowEvent := createTestWorkflowEvent("firing", "HighCPUUsage", "critical", "production")
	matchingWorkflows := engine.SelectWorkflows(workflowEvent)
	assert.Len(t, matchingWorkflows, 1)

	// Test non-matching event (different alert name)
	workflowEvent = createTestWorkflowEvent("firing", "HighMemoryUsage", "critical", "production")
	matchingWorkflows = engine.SelectWorkflows(workflowEvent)
	assert.Empty(t, matchingWorkflows)
}

func TestWorkflowEngine_ExecuteWorkflow(t *testing.T) {
	config := createBasicWorkflowConfig()
	engine := createTestEngine(config)

	workflowEvent := createTestWorkflowEvent("firing", "TestAlert", "warning", "default")
	workflow := &config.ActiveWorkflows[0]

	// Test ExecuteWorkflow - should return error due to missing executor
	ctx := context.Background()
	err := engine.ExecuteWorkflow(ctx, workflow, workflowEvent)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "action executor is not configured")
}

func TestWorkflowEngine_MatchesTrigger_NoAlertmanagerTrigger(t *testing.T) {
	engine := createTestEngine(&workflow.WorkflowConfig{})
	workflowEvent := createTestWorkflowEvent("firing", "TestAlert", "warning", "default")

	// Test trigger without AlertManager alert (should return false)
	trigger := &workflow.TriggerDefinition{
		// No OnAlertmanagerAlert set
	}

	result := engine.matchesTrigger(trigger, workflowEvent)
	assert.False(t, result)
}

func TestWorkflowEngine_MatchesAlertmanagerAlertTrigger(t *testing.T) {
	engine := createTestEngine(&workflow.WorkflowConfig{})
	workflowEvent := createTestWorkflowEvent("firing", "TestAlert", "warning", "default")

	tests := []struct {
		name        string
		trigger     *workflow.AlertmanagerAlertTrigger
		shouldMatch bool
	}{
		{
			name:        "empty trigger matches all",
			trigger:     &workflow.AlertmanagerAlertTrigger{},
			shouldMatch: true,
		},
		{
			name: "matches alert name",
			trigger: &workflow.AlertmanagerAlertTrigger{
				AlertName: "TestAlert",
			},
			shouldMatch: true,
		},
		{
			name: "doesn't match alert name",
			trigger: &workflow.AlertmanagerAlertTrigger{
				AlertName: "OtherAlert",
			},
			shouldMatch: false,
		},
		{
			name: "matches status",
			trigger: &workflow.AlertmanagerAlertTrigger{
				Status: "firing",
			},
			shouldMatch: true,
		},
		{
			name: "doesn't match status",
			trigger: &workflow.AlertmanagerAlertTrigger{
				Status: "resolved",
			},
			shouldMatch: false,
		},
		{
			name: "matches severity",
			trigger: &workflow.AlertmanagerAlertTrigger{
				Severity: "warning",
			},
			shouldMatch: true,
		},
		{
			name: "doesn't match severity",
			trigger: &workflow.AlertmanagerAlertTrigger{
				Severity: "critical",
			},
			shouldMatch: false,
		},
		{
			name: "matches namespace",
			trigger: &workflow.AlertmanagerAlertTrigger{
				Namespace: "default",
			},
			shouldMatch: true,
		},
		{
			name: "doesn't match namespace",
			trigger: &workflow.AlertmanagerAlertTrigger{
				Namespace: "production",
			},
			shouldMatch: false,
		},
		{
			name: "matches multiple criteria",
			trigger: &workflow.AlertmanagerAlertTrigger{
				AlertName: "TestAlert",
				Status:    "firing",
				Severity:  "warning",
			},
			shouldMatch: true,
		},
		{
			name: "partial match fails",
			trigger: &workflow.AlertmanagerAlertTrigger{
				AlertName: "TestAlert",
				Status:    "resolved", // Wrong status
			},
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.matchesAlertmanagerAlertTrigger(tt.trigger, workflowEvent)
			assert.Equal(t, tt.shouldMatch, result)
		})
	}
}
