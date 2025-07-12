package workflow

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkflowDefinition_Validate(t *testing.T) {
	tests := []struct {
		name     string
		workflow WorkflowDefinition
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid workflow",
			workflow: WorkflowDefinition{
				Name: "test-workflow",
				Triggers: []TriggerDefinition{
					{
						OnAlertmanagerAlert: &AlertmanagerAlertTrigger{
							AlertName: "TestAlert",
							Status:    "firing",
						},
					},
				},
				Actions: []ActionDefinition{
					{
						ActionType: "test_action",
						RawData: map[string]interface{}{
							"test_action": map[string]interface{}{},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "empty name",
			workflow: WorkflowDefinition{
				Name: "",
				Triggers: []TriggerDefinition{
					{
						OnAlertmanagerAlert: &AlertmanagerAlertTrigger{
							AlertName: "TestAlert",
						},
					},
				},
				Actions: []ActionDefinition{
					{
						ActionType: "test_action",
						RawData: map[string]interface{}{
							"test_action": map[string]interface{}{},
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "workflow name cannot be empty",
		},
		{
			name: "no triggers",
			workflow: WorkflowDefinition{
				Name:     "test-workflow",
				Triggers: []TriggerDefinition{},
				Actions: []ActionDefinition{
					{
						ActionType: "test_action",
						RawData: map[string]interface{}{
							"test_action": map[string]interface{}{},
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "must have at least one trigger",
		},
		{
			name: "no actions",
			workflow: WorkflowDefinition{
				Name: "test-workflow",
				Triggers: []TriggerDefinition{
					{
						OnAlertmanagerAlert: &AlertmanagerAlertTrigger{
							AlertName: "TestAlert",
						},
					},
				},
				Actions: []ActionDefinition{},
			},
			wantErr: true,
			errMsg:  "must have at least one action",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.workflow.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTriggerDefinition_GetTriggerType(t *testing.T) {
	tests := []struct {
		name     string
		trigger  TriggerDefinition
		expected string
	}{
		{
			name: "alertmanager alert trigger",
			trigger: TriggerDefinition{
				OnAlertmanagerAlert: &AlertmanagerAlertTrigger{
					AlertName: "TestAlert",
				},
			},
			expected: "alertmanager_alert",
		},
		{
			name:     "unknown trigger",
			trigger:  TriggerDefinition{},
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.trigger.GetTriggerType()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTriggerDefinition_Validate(t *testing.T) {
	tests := []struct {
		name    string
		trigger TriggerDefinition
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid alertmanager trigger",
			trigger: TriggerDefinition{
				OnAlertmanagerAlert: &AlertmanagerAlertTrigger{
					AlertName: "TestAlert",
					Status:    "firing",
				},
			},
			wantErr: false,
		},
		{
			name:    "no trigger specified",
			trigger: TriggerDefinition{},
			wantErr: true,
			errMsg:  "must specify exactly one trigger type",
		},
		{
			name: "invalid alertmanager trigger",
			trigger: TriggerDefinition{
				OnAlertmanagerAlert: &AlertmanagerAlertTrigger{
					Status: "invalid-status",
				},
			},
			wantErr: true,
			errMsg:  "alertmanager_alert trigger validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.trigger.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAlertmanagerAlertTrigger_Validate(t *testing.T) {
	tests := []struct {
		name    string
		trigger AlertmanagerAlertTrigger
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid trigger with all fields",
			trigger: AlertmanagerAlertTrigger{
				AlertName: "TestAlert",
				Status:    "firing",
				Severity:  "critical",
				Namespace: "default",
				Instance:  "localhost:9090",
				PodName:   "test-pod",
			},
			wantErr: false,
		},
		{
			name: "valid trigger with minimal fields",
			trigger: AlertmanagerAlertTrigger{
				AlertName: "TestAlert",
			},
			wantErr: false,
		},
		{
			name:    "empty trigger (valid - matches all)",
			trigger: AlertmanagerAlertTrigger{},
			wantErr: false,
		},
		{
			name: "invalid status",
			trigger: AlertmanagerAlertTrigger{
				Status: "invalid-status",
			},
			wantErr: true,
			errMsg:  "invalid status 'invalid-status'",
		},
		{
			name: "invalid severity",
			trigger: AlertmanagerAlertTrigger{
				Severity: "invalid-severity",
			},
			wantErr: true,
			errMsg:  "invalid severity 'invalid-severity'",
		},
		{
			name: "valid status 'resolved'",
			trigger: AlertmanagerAlertTrigger{
				Status: "resolved",
			},
			wantErr: false,
		},
		{
			name: "valid status 'all'",
			trigger: AlertmanagerAlertTrigger{
				Status: "all",
			},
			wantErr: false,
		},
		{
			name: "valid severity 'warning'",
			trigger: AlertmanagerAlertTrigger{
				Severity: "warning",
			},
			wantErr: false,
		},
		{
			name: "valid severity 'info'",
			trigger: AlertmanagerAlertTrigger{
				Severity: "info",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.trigger.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestWorkflowDefinition_GetID(t *testing.T) {
	workflow := WorkflowDefinition{
		Name: "test-workflow",
	}

	assert.Equal(t, "test-workflow", workflow.GetID())
}

func TestWorkflowDefinition_HasTriggerType(t *testing.T) {
	workflow := WorkflowDefinition{
		Name: "test-workflow",
		Triggers: []TriggerDefinition{
			{
				OnAlertmanagerAlert: &AlertmanagerAlertTrigger{
					AlertName: "TestAlert",
				},
			},
		},
	}

	assert.True(t, workflow.HasTriggerType("alertmanager_alert"))
	assert.False(t, workflow.HasTriggerType("kubernetes_event"))
	assert.False(t, workflow.HasTriggerType("unknown"))
}

func TestNewWorkflowMetadata(t *testing.T) {
	startTime := time.Now()
	metadata := NewWorkflowMetadata("test-workflow", "alertmanager_alert", 3)

	assert.Equal(t, "test-workflow", metadata.WorkflowName)
	assert.Equal(t, "alertmanager_alert", metadata.TriggerType)
	assert.Equal(t, 3, metadata.ActionsCount)
	assert.NotEmpty(t, metadata.ExecutionID)
	assert.Contains(t, metadata.ExecutionID, "test-workflow-")
	assert.True(t, metadata.StartTime.After(startTime) || metadata.StartTime.Equal(startTime))
}

func TestWorkflowDefinition_ValidateComplexScenarios(t *testing.T) {
	t.Run("workflow with multiple triggers", func(t *testing.T) {
		workflow := WorkflowDefinition{
			Name: "multi-trigger-workflow",
			Triggers: []TriggerDefinition{
				{
					OnAlertmanagerAlert: &AlertmanagerAlertTrigger{
						AlertName: "Alert1",
						Status:    "firing",
					},
				},
				{
					OnAlertmanagerAlert: &AlertmanagerAlertTrigger{
						AlertName: "Alert2",
						Status:    "resolved",
					},
				},
			},
			Actions: []ActionDefinition{
				{
					ActionType: "action1",
					RawData: map[string]interface{}{
						"action1": map[string]interface{}{},
					},
				},
			},
		}

		err := workflow.Validate()
		assert.NoError(t, err)
	})

	t.Run("workflow with stop flag", func(t *testing.T) {
		workflow := WorkflowDefinition{
			Name: "stop-workflow",
			Triggers: []TriggerDefinition{
				{
					OnAlertmanagerAlert: &AlertmanagerAlertTrigger{
						AlertName: "StopAlert",
					},
				},
			},
			Actions: []ActionDefinition{
				{
					ActionType: "stop_action",
					RawData: map[string]interface{}{
						"stop_action": map[string]interface{}{},
					},
				},
			},
			Stop: true,
		}

		err := workflow.Validate()
		require.NoError(t, err)
		assert.True(t, workflow.Stop)
	})
}
