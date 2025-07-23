package actions

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kubecano/cano-collector/pkg/core/event"
	"github.com/kubecano/cano-collector/pkg/logger"
	"github.com/kubecano/cano-collector/pkg/metric"
	actions_interfaces "github.com/kubecano/cano-collector/pkg/workflow/actions/interfaces"
)

// Mock implementations for testing
type mockActionFactory struct {
	actionType string
	createErr  error
	action     actions_interfaces.WorkflowAction
}

func (m *mockActionFactory) Create(config actions_interfaces.ActionConfig) (actions_interfaces.WorkflowAction, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	return m.action, nil
}

func (m *mockActionFactory) GetActionType() string {
	return m.actionType
}

func (m *mockActionFactory) ValidateConfig(config actions_interfaces.ActionConfig) error {
	return nil
}

type mockWorkflowAction struct {
	name        string
	validateErr error
	executeErr  error
	result      *actions_interfaces.ActionResult
}

func (m *mockWorkflowAction) Execute(ctx context.Context, event event.WorkflowEvent) (*actions_interfaces.ActionResult, error) {
	if m.executeErr != nil {
		return nil, m.executeErr
	}
	return m.result, nil
}

func (m *mockWorkflowAction) GetName() string {
	return m.name
}

func (m *mockWorkflowAction) Validate() error {
	return m.validateErr
}

// Simple mock event for testing
type mockWorkflowEvent struct {
	id        uuid.UUID
	alertName string
	status    string
	severity  string
	namespace string
}

func (m *mockWorkflowEvent) GetID() uuid.UUID          { return m.id }
func (m *mockWorkflowEvent) GetTimestamp() time.Time   { return time.Now() }
func (m *mockWorkflowEvent) GetSource() string         { return "test" }
func (m *mockWorkflowEvent) GetType() event.EventType  { return event.EventTypeAlertManager }
func (m *mockWorkflowEvent) GetEventData() interface{} { return nil }
func (m *mockWorkflowEvent) GetAlertName() string      { return m.alertName }
func (m *mockWorkflowEvent) GetStatus() string         { return m.status }
func (m *mockWorkflowEvent) GetSeverity() string       { return m.severity }
func (m *mockWorkflowEvent) GetNamespace() string      { return m.namespace }

// Test helper function
func createTestWorkflowEvent(status, alertname, severity, namespace string) event.WorkflowEvent {
	return &mockWorkflowEvent{
		id:        uuid.New(),
		alertName: alertname,
		status:    status,
		severity:  severity,
		namespace: namespace,
	}
}

// Tests for DefaultActionRegistry

func TestDefaultActionRegistry_NewDefaultActionRegistry(t *testing.T) {
	logger := logger.NewLogger("debug", "test")
	registry := NewDefaultActionRegistry(logger)

	assert.NotNil(t, registry)
	assert.Empty(t, registry.GetRegisteredTypes())
}

func TestDefaultActionRegistry_Register(t *testing.T) {
	logger := logger.NewLogger("debug", "test")
	registry := NewDefaultActionRegistry(logger)

	factory := &mockActionFactory{actionType: "test_action"}

	// Test successful registration
	err := registry.Register("test_action", factory)
	require.NoError(t, err)

	types := registry.GetRegisteredTypes()
	assert.Len(t, types, 1)
	assert.Contains(t, types, "test_action")

	// Test duplicate registration
	err = registry.Register("test_action", factory)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")

	// Test empty action type
	err = registry.Register("", factory)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "action type cannot be empty")

	// Test nil factory
	err = registry.Register("nil_action", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "action factory cannot be nil")
}

func TestDefaultActionRegistry_Create(t *testing.T) {
	logger := logger.NewLogger("debug", "test")
	registry := NewDefaultActionRegistry(logger)

	mockAction := &mockWorkflowAction{name: "test-action"}
	factory := &mockActionFactory{
		actionType: "test_action",
		action:     mockAction,
	}

	// Register factory
	err := registry.Register("test_action", factory)
	require.NoError(t, err)

	// Test successful creation
	config := actions_interfaces.ActionConfig{
		Name: "test-action",
		Type: "test_action",
	}

	action, err := registry.Create(config)
	require.NoError(t, err)
	assert.Equal(t, mockAction, action)

	// Test empty type
	config.Type = ""
	action, err = registry.Create(config)
	require.Error(t, err)
	assert.Nil(t, action)
	assert.Contains(t, err.Error(), "action type cannot be empty")

	// Test unknown type
	config.Type = "unknown_action"
	action, err = registry.Create(config)
	require.Error(t, err)
	assert.Nil(t, action)
	assert.Contains(t, err.Error(), "no factory registered")

	// Test factory error
	factoryWithError := &mockActionFactory{
		actionType: "error_action",
		createErr:  errors.New("factory error"),
	}
	err = registry.Register("error_action", factoryWithError)
	require.NoError(t, err)

	config.Type = "error_action"
	action, err = registry.Create(config)
	require.Error(t, err)
	assert.Nil(t, action)
	assert.Contains(t, err.Error(), "factory error")
}

func TestDefaultActionRegistry_ThreadSafety(t *testing.T) {
	logger := logger.NewLogger("debug", "test")
	registry := NewDefaultActionRegistry(logger)

	// Test concurrent registrations and creations
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(i int) {
			defer func() { done <- true }()

			actionType := fmt.Sprintf("action_%d", i)
			factory := &mockActionFactory{
				actionType: actionType,
				action:     &mockWorkflowAction{name: actionType},
			}

			err := registry.Register(actionType, factory)
			require.NoError(t, err)

			config := actions_interfaces.ActionConfig{
				Type: actionType,
				Name: actionType,
			}

			action, err := registry.Create(config)
			require.NoError(t, err)
			assert.NotNil(t, action)
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	assert.Len(t, registry.GetRegisteredTypes(), 10)
}

// Tests for DefaultActionExecutor

func TestDefaultActionExecutor_NewDefaultActionExecutor(t *testing.T) {
	logger := logger.NewLogger("debug", "test")
	metrics := metric.NewMetricsCollector(logger)
	registry := NewDefaultActionRegistry(logger)

	executor := NewDefaultActionExecutor(registry, logger, metrics)

	assert.NotNil(t, executor)
}

func TestDefaultActionExecutor_ExecuteAction(t *testing.T) {
	logger := logger.NewLogger("debug", "test")
	metrics := metric.NewMetricsCollector(logger)
	registry := NewDefaultActionRegistry(logger)
	executor := NewDefaultActionExecutor(registry, logger, metrics)

	// Create test event
	event := createTestWorkflowEvent("firing", "TestAlert", "warning", "default")
	ctx := context.Background()

	// Test nil action
	result, err := executor.ExecuteAction(ctx, nil, event)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "action cannot be nil")

	// Test nil event
	mockAction := &mockWorkflowAction{name: "test-action"}
	result, err = executor.ExecuteAction(ctx, mockAction, nil)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "event cannot be nil")

	// Test validation error
	mockAction = &mockWorkflowAction{
		name:        "test-action",
		validateErr: errors.New("validation failed"),
	}
	result, err = executor.ExecuteAction(ctx, mockAction, event)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "action validation failed")

	// Test execution error
	mockAction = &mockWorkflowAction{
		name:       "test-action",
		executeErr: errors.New("execution failed"),
	}
	result, err = executor.ExecuteAction(ctx, mockAction, event)
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "action execution failed")

	// Test successful execution
	expectedResult := &actions_interfaces.ActionResult{
		Success: true,
		Data:    "test data",
	}
	mockAction = &mockWorkflowAction{
		name:   "test-action",
		result: expectedResult,
	}
	result, err = executor.ExecuteAction(ctx, mockAction, event)
	require.NoError(t, err)
	assert.Equal(t, expectedResult, result)

	// Test unsuccessful result
	expectedResult = &actions_interfaces.ActionResult{
		Success: false,
		Error:   errors.New("action failed"),
	}
	mockAction = &mockWorkflowAction{
		name:   "test-action",
		result: expectedResult,
	}
	result, err = executor.ExecuteAction(ctx, mockAction, event)
	require.NoError(t, err)
	assert.Equal(t, expectedResult, result)
}

func TestDefaultActionExecutor_CreateActionsFromConfig(t *testing.T) {
	logger := logger.NewLogger("debug", "test")
	metrics := metric.NewMetricsCollector(logger)
	registry := NewDefaultActionRegistry(logger)
	executor := NewDefaultActionExecutor(registry, logger, metrics)

	// Register test factory
	mockAction := &mockWorkflowAction{name: "test-action"}
	factory := &mockActionFactory{
		actionType: "test_action",
		action:     mockAction,
	}
	err := registry.Register("test_action", factory)
	require.NoError(t, err)

	// Test empty configs
	actions, err := executor.CreateActionsFromConfig([]actions_interfaces.ActionConfig{})
	require.NoError(t, err)
	assert.Empty(t, actions)

	// Test successful creation
	configs := []actions_interfaces.ActionConfig{
		{Name: "action1", Type: "test_action"},
		{Name: "action2", Type: "test_action"},
	}

	actions, err = executor.CreateActionsFromConfig(configs)
	require.NoError(t, err)
	assert.Len(t, actions, 2)

	// Test creation failure
	configs = []actions_interfaces.ActionConfig{
		{Name: "action1", Type: "unknown_action"},
	}

	actions, err = executor.CreateActionsFromConfig(configs)
	require.Error(t, err)
	assert.Nil(t, actions)
	assert.Contains(t, err.Error(), "failed to create action")
}

func TestDefaultActionExecutor_ExecuteActions(t *testing.T) {
	logger := logger.NewLogger("debug", "test")
	metrics := metric.NewMetricsCollector(logger)
	registry := NewDefaultActionRegistry(logger)
	executor := NewDefaultActionExecutor(registry, logger, metrics)

	event := createTestWorkflowEvent("firing", "TestAlert", "warning", "default")
	ctx := context.Background()

	// Test empty actions
	results, err := executor.ExecuteActions(ctx, []actions_interfaces.WorkflowAction{}, event)
	require.NoError(t, err)
	assert.Empty(t, results)

	// Test successful execution
	mockAction1 := &mockWorkflowAction{
		name:   "action1",
		result: &actions_interfaces.ActionResult{Success: true, Data: "result1"},
	}
	mockAction2 := &mockWorkflowAction{
		name:   "action2",
		result: &actions_interfaces.ActionResult{Success: true, Data: "result2"},
	}

	actions := []actions_interfaces.WorkflowAction{mockAction1, mockAction2}
	results, err = executor.ExecuteActions(ctx, actions, event)
	require.NoError(t, err) // Function doesn't return error, continues execution
	assert.Len(t, results, 2)
	assert.True(t, results[0].Success)
	assert.True(t, results[1].Success)

	// Test mixed success/failure
	mockAction3 := &mockWorkflowAction{
		name:       "action3",
		executeErr: errors.New("execution failed"),
	}

	actions = []actions_interfaces.WorkflowAction{mockAction1, mockAction3}
	results, err = executor.ExecuteActions(ctx, actions, event)
	require.NoError(t, err) // Function doesn't return error, continues execution
	assert.Len(t, results, 2)
	assert.True(t, results[0].Success)
	assert.False(t, results[1].Success)
	assert.Error(t, results[1].Error)
}

func TestDefaultActionExecutor_DeprecatedMethods(t *testing.T) {
	logger := logger.NewLogger("debug", "test")
	metrics := metric.NewMetricsCollector(logger)
	registry := NewDefaultActionRegistry(logger)
	executor := NewDefaultActionExecutor(registry, logger, metrics)

	mockAction := &mockWorkflowAction{name: "test-action"}

	// Test deprecated RegisterAction
	err := executor.RegisterAction("test_action", mockAction)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "deprecated")

	// Test deprecated GetAction
	action, err := executor.GetAction("test_action")
	require.Error(t, err)
	assert.Nil(t, action)
	assert.Contains(t, err.Error(), "deprecated")
}
