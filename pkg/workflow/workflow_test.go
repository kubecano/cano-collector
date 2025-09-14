package workflow

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/prometheus/alertmanager/template"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kubecano/cano-collector/config/workflow"
	"github.com/kubecano/cano-collector/pkg/core/event"
	"github.com/kubecano/cano-collector/pkg/core/issue"
	"github.com/kubecano/cano-collector/pkg/logger"
	"github.com/kubecano/cano-collector/pkg/metric"
	actions_interfaces "github.com/kubecano/cano-collector/pkg/workflow/actions/interfaces"
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
	return NewWorkflowEngine(config, nil, nil, nil)
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

func TestWorkflowEngine_ExecuteWorkflowWithEnrichments_NoExecutor(t *testing.T) {
	config := createBasicWorkflowConfig()
	engine := createTestEngine(config)

	workflowEvent := createTestWorkflowEvent("firing", "TestAlert", "warning", "default")
	workflow := &config.ActiveWorkflows[0]

	// Test ExecuteWorkflowWithEnrichments - should return error due to missing executor
	ctx := context.Background()
	enrichments, err := engine.ExecuteWorkflowWithEnrichments(ctx, workflow, workflowEvent)
	require.Error(t, err)
	assert.Nil(t, enrichments)
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

// TestWorkflowEngine_ActionTypeInference_Deterministic tests that action type inference is deterministic
func TestWorkflowEngine_ActionTypeInference_Deterministic(t *testing.T) {
	// Create a mock executor that captures action configs
	capturedConfigs := []actions_interfaces.ActionConfig{}
	mockExecutor := &mockActionExecutor{
		createActionsFromConfigFunc: func(configs []actions_interfaces.ActionConfig) ([]actions_interfaces.WorkflowAction, error) {
			capturedConfigs = append(capturedConfigs, configs...)
			return []actions_interfaces.WorkflowAction{}, nil
		},
		executeActionsFunc: func(ctx context.Context, actions []actions_interfaces.WorkflowAction, event event.WorkflowEvent) ([]*actions_interfaces.ActionResult, error) {
			return []*actions_interfaces.ActionResult{}, nil
		},
	}

	// Create a workflow with action that has multiple keys in RawData (to test deterministic inference)
	workflowConfig := &workflow.WorkflowConfig{
		ActiveWorkflows: []workflow.WorkflowDefinition{
			{
				Name: "test-inference-workflow",
				Triggers: []workflow.TriggerDefinition{
					{
						OnAlertmanagerAlert: &workflow.AlertmanagerAlertTrigger{
							Status: "firing",
						},
					},
				},
				Actions: []workflow.ActionDefinition{
					{
						// ActionType is empty, so inference should kick in
						ActionType: "",
						RawData: map[string]interface{}{
							"zebra_action": map[string]interface{}{"param1": "value1"},
							"alpha_action": map[string]interface{}{"param2": "value2"},
							"beta_action":  map[string]interface{}{"param3": "value3"},
							"action_type":  "", // This should be ignored
						},
					},
				},
			},
		},
	}

	engine := NewWorkflowEngine(workflowConfig, mockExecutor, nil, nil)
	workflowEvent := createTestWorkflowEvent("firing", "TestAlert", "warning", "default")

	// Execute workflow multiple times to ensure deterministic behavior
	ctx := context.Background()
	workflow := &workflowConfig.ActiveWorkflows[0]

	var inferredTypes []string
	for i := 0; i < 5; i++ {
		capturedConfigs = []actions_interfaces.ActionConfig{} // Reset
		_, err := engine.ExecuteWorkflowWithEnrichments(ctx, workflow, workflowEvent)
		require.NoError(t, err)

		require.Len(t, capturedConfigs, 1)
		inferredTypes = append(inferredTypes, capturedConfigs[0].Type)
	}

	// All inferred types should be the same (deterministic)
	for i := 1; i < len(inferredTypes); i++ {
		assert.Equal(t, inferredTypes[0], inferredTypes[i],
			"Action type inference should be deterministic, but got different types: %v", inferredTypes)
	}

	// The inferred type should be the alphabetically first key (excluding "action_type")
	expectedType := "alpha_action" // alphabetically first among: alpha_action, beta_action, zebra_action
	assert.Equal(t, expectedType, inferredTypes[0],
		"Action type should be inferred as alphabetically first key")
}

// mockActionExecutor for testing
type mockActionExecutor struct {
	createActionsFromConfigFunc func(configs []actions_interfaces.ActionConfig) ([]actions_interfaces.WorkflowAction, error)
	executeActionsFunc          func(ctx context.Context, actions []actions_interfaces.WorkflowAction, event event.WorkflowEvent) ([]*actions_interfaces.ActionResult, error)
}

func (m *mockActionExecutor) ExecuteAction(ctx context.Context, action actions_interfaces.WorkflowAction, event event.WorkflowEvent) (*actions_interfaces.ActionResult, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockActionExecutor) RegisterAction(actionType string, action actions_interfaces.WorkflowAction) error {
	return fmt.Errorf("not implemented")
}

func (m *mockActionExecutor) GetAction(actionType string) (actions_interfaces.WorkflowAction, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *mockActionExecutor) CreateActionsFromConfig(configs []actions_interfaces.ActionConfig) ([]actions_interfaces.WorkflowAction, error) {
	if m.createActionsFromConfigFunc != nil {
		return m.createActionsFromConfigFunc(configs)
	}
	return []actions_interfaces.WorkflowAction{}, nil
}

func (m *mockActionExecutor) ExecuteActions(ctx context.Context, actions []actions_interfaces.WorkflowAction, event event.WorkflowEvent) ([]*actions_interfaces.ActionResult, error) {
	if m.executeActionsFunc != nil {
		return m.executeActionsFunc(ctx, actions, event)
	}
	return []*actions_interfaces.ActionResult{}, nil
}

// Test createActionConfigs method
func TestWorkflowEngine_CreateActionConfigs(t *testing.T) {
	engine := createTestEngine(&workflow.WorkflowConfig{})

	tests := []struct {
		name          string
		workflowDef   *workflow.WorkflowDefinition
		expectedCount int
		expectError   bool
		expectedTypes []string
	}{
		{
			name: "single action with explicit type",
			workflowDef: &workflow.WorkflowDefinition{
				Name: "test-workflow",
				Actions: []workflow.ActionDefinition{
					{
						ActionType: "pod_logs",
						RawData: map[string]interface{}{
							"namespace": "default",
						},
					},
				},
			},
			expectedCount: 1,
			expectError:   false,
			expectedTypes: []string{"pod_logs"},
		},
		{
			name: "action with inferred type",
			workflowDef: &workflow.WorkflowDefinition{
				Name: "test-workflow",
				Actions: []workflow.ActionDefinition{
					{
						ActionType: "",
						RawData: map[string]interface{}{
							"pod_logs": map[string]interface{}{
								"namespace": "default",
							},
							"action_type": "", // Should be ignored
						},
					},
				},
			},
			expectedCount: 1,
			expectError:   false,
			expectedTypes: []string{"pod_logs"},
		},
		{
			name: "action with no type",
			workflowDef: &workflow.WorkflowDefinition{
				Name: "test-workflow",
				Actions: []workflow.ActionDefinition{
					{
						ActionType: "",
						RawData:    map[string]interface{}{},
					},
				},
			},
			expectedCount: 0,
			expectError:   true,
		},
		{
			name: "multiple actions with deterministic inference",
			workflowDef: &workflow.WorkflowDefinition{
				Name: "test-workflow",
				Actions: []workflow.ActionDefinition{
					{
						ActionType: "",
						RawData: map[string]interface{}{
							"zebra_action": map[string]interface{}{},
							"alpha_action": map[string]interface{}{},
						},
					},
				},
			},
			expectedCount: 1,
			expectError:   false,
			expectedTypes: []string{"alpha_action"}, // Should be alphabetically first
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configs, err := engine.createActionConfigs(tt.workflowDef)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, configs, tt.expectedCount)

			for i, expectedType := range tt.expectedTypes {
				assert.Equal(t, expectedType, configs[i].Type)
				assert.Equal(t, fmt.Sprintf("%s-action-%d", tt.workflowDef.Name, i), configs[i].Name)
				assert.True(t, configs[i].Enabled)
				assert.Equal(t, 30, configs[i].Timeout)
			}
		})
	}
}

// Test ExecuteWorkflowWithEnrichments method
func TestWorkflowEngine_ExecuteWorkflowWithEnrichments(t *testing.T) {
	// Create a mock executor that returns enrichments
	mockExecutor := &mockActionExecutor{
		createActionsFromConfigFunc: func(configs []actions_interfaces.ActionConfig) ([]actions_interfaces.WorkflowAction, error) {
			return []actions_interfaces.WorkflowAction{}, nil
		},
		executeActionsFunc: func(ctx context.Context, actions []actions_interfaces.WorkflowAction, event event.WorkflowEvent) ([]*actions_interfaces.ActionResult, error) {
			enrichment := issue.NewEnrichmentWithType(issue.EnrichmentTypeTextFile, "Test Enrichment")
			enrichment.AddBlock(issue.NewMarkdownBlock("Test content"))
			return []*actions_interfaces.ActionResult{
				{
					Success:     true,
					Enrichments: []issue.Enrichment{*enrichment},
				},
			}, nil
		},
	}

	config := createBasicWorkflowConfig()
	log := logger.NewLogger("debug", "test")
	metrics := metric.NewMetricsCollector(log)
	engine := NewWorkflowEngine(config, mockExecutor, log, metrics)

	ctx := context.Background()
	workflowEvent := createTestWorkflowEvent("firing", "TestAlert", "warning", "default")
	workflow := &config.ActiveWorkflows[0]

	enrichments, err := engine.ExecuteWorkflowWithEnrichments(ctx, workflow, workflowEvent)
	require.NoError(t, err)
	assert.Len(t, enrichments, 1)
	assert.Equal(t, "Test Enrichment", *enrichments[0].Title)
}

// Test ExecuteWorkflowsWithEnrichments method
func TestWorkflowEngine_ExecuteWorkflowsWithEnrichments(t *testing.T) {
	// Create a mock executor that returns enrichments
	mockExecutor := &mockActionExecutor{
		createActionsFromConfigFunc: func(configs []actions_interfaces.ActionConfig) ([]actions_interfaces.WorkflowAction, error) {
			return []actions_interfaces.WorkflowAction{}, nil
		},
		executeActionsFunc: func(ctx context.Context, actions []actions_interfaces.WorkflowAction, event event.WorkflowEvent) ([]*actions_interfaces.ActionResult, error) {
			enrichment := issue.NewEnrichmentWithType(issue.EnrichmentTypeTextFile, "Multiple Enrichment")
			enrichment.AddBlock(issue.NewMarkdownBlock("Multiple content"))
			return []*actions_interfaces.ActionResult{
				{
					Success:     true,
					Enrichments: []issue.Enrichment{*enrichment},
				},
			}, nil
		},
	}

	config := createBasicWorkflowConfig()
	log := logger.NewLogger("debug", "test")
	metrics := metric.NewMetricsCollector(log)
	engine := NewWorkflowEngine(config, mockExecutor, log, metrics)

	ctx := context.Background()
	workflowEvent := createTestWorkflowEvent("firing", "TestAlert", "warning", "default")
	workflows := []*workflow.WorkflowDefinition{
		&config.ActiveWorkflows[0],
		&config.ActiveWorkflows[1],
	}

	enrichments, err := engine.ExecuteWorkflowsWithEnrichments(ctx, workflows, workflowEvent)
	require.NoError(t, err)
	assert.Len(t, enrichments, 2) // One from each workflow
	assert.Equal(t, "Multiple Enrichment", *enrichments[0].Title)
}

// Test ExecuteWorkflowWithEnrichments error cases
func TestWorkflowEngine_ExecuteWorkflowWithEnrichments_Errors(t *testing.T) {
	config := createBasicWorkflowConfig()
	log := logger.NewLogger("debug", "test")
	metrics := metric.NewMetricsCollector(log)
	engine := NewWorkflowEngine(config, nil, log, metrics)

	ctx := context.Background()
	workflowEvent := createTestWorkflowEvent("firing", "TestAlert", "warning", "default")

	// Test nil workflow
	enrichments, err := engine.ExecuteWorkflowWithEnrichments(ctx, nil, workflowEvent)
	require.Error(t, err)
	assert.Nil(t, enrichments)
	assert.Contains(t, err.Error(), "workflow definition cannot be nil")

	// Test nil event
	workflow := &config.ActiveWorkflows[0]
	enrichments, err = engine.ExecuteWorkflowWithEnrichments(ctx, workflow, nil)
	require.Error(t, err)
	assert.Nil(t, enrichments)
	assert.Contains(t, err.Error(), "workflow event cannot be nil")

	// Test nil executor
	enrichments, err = engine.ExecuteWorkflowWithEnrichments(ctx, workflow, workflowEvent)
	require.Error(t, err)
	assert.Nil(t, enrichments)
	assert.Contains(t, err.Error(), "action executor is not configured")
}

// Test ExecuteWorkflowsWithEnrichments with empty workflow list
func TestWorkflowEngine_ExecuteWorkflowsWithEnrichments_EmptyList(t *testing.T) {
	config := createBasicWorkflowConfig()
	log := logger.NewLogger("debug", "test")
	metrics := metric.NewMetricsCollector(log)
	engine := NewWorkflowEngine(config, nil, log, metrics)

	ctx := context.Background()
	workflowEvent := createTestWorkflowEvent("firing", "TestAlert", "warning", "default")

	enrichments, err := engine.ExecuteWorkflowsWithEnrichments(ctx, []*workflow.WorkflowDefinition{}, workflowEvent)
	require.NoError(t, err)
	assert.Empty(t, enrichments)
}
