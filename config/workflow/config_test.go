package workflow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestConfig creates a temporary config file and returns the path
func setupTestConfig(t *testing.T, content string) string {
	t.Helper()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_config.yaml")

	err := os.WriteFile(configPath, []byte(content), 0o644)
	require.NoError(t, err)

	return configPath
}

func TestLoadWorkflowConfigFromString(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
		errMsg  string
		check   func(*testing.T, *WorkflowConfig)
	}{
		{
			name: "valid configuration",
			yaml: `
active_workflows:
  - name: "test-workflow"
    triggers:
      - on_alertmanager_alert:
          alert_name: "TestAlert"
          status: "firing"
    actions:
      - test_action: {}
`,
			wantErr: false,
			check: func(t *testing.T, config *WorkflowConfig) {
				t.Helper()
				require.Len(t, config.ActiveWorkflows, 1)
				workflow := config.ActiveWorkflows[0]
				assert.Equal(t, "test-workflow", workflow.Name)
				assert.Len(t, workflow.Triggers, 1)
				assert.Len(t, workflow.Actions, 1)
				assert.Equal(t, "TestAlert", workflow.Triggers[0].OnAlertmanagerAlert.AlertName)
				assert.Equal(t, "firing", workflow.Triggers[0].OnAlertmanagerAlert.Status)
			},
		},
		{
			name: "empty configuration",
			yaml: `
active_workflows: []
`,
			wantErr: true,
			errMsg:  "no active workflows configured",
		},
		{
			name: "invalid YAML",
			yaml: `
active_workflows:
  - name: "test-workflow"
    triggers:
      - on_alertmanager_alert:
          alert_name: "TestAlert"
          status: "firing"
    actions:
      invalid_yaml_here: [
`,
			wantErr: true,
			errMsg:  "failed to parse workflow config YAML",
		},
		{
			name: "multiple workflows",
			yaml: `
active_workflows:
  - name: "workflow1"
    triggers:
      - on_alertmanager_alert:
          alert_name: "Alert1"
    actions:
      - action1: {}
  - name: "workflow2"
    triggers:
      - on_alertmanager_alert:
          alert_name: "Alert2"
    actions:
      - action2: {}
`,
			wantErr: false,
			check: func(t *testing.T, config *WorkflowConfig) {
				t.Helper()
				require.Len(t, config.ActiveWorkflows, 2)
				assert.Equal(t, "workflow1", config.ActiveWorkflows[0].Name)
				assert.Equal(t, "workflow2", config.ActiveWorkflows[1].Name)
			},
		},
		{
			name: "workflow with stop flag",
			yaml: `
active_workflows:
  - name: "stop-workflow"
    triggers:
      - on_alertmanager_alert:
          alert_name: "StopAlert"
    actions:
      - stop_action: {}
    stop: true
`,
			wantErr: false,
			check: func(t *testing.T, config *WorkflowConfig) {
				t.Helper()
				require.Len(t, config.ActiveWorkflows, 1)
				assert.True(t, config.ActiveWorkflows[0].Stop)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := LoadWorkflowConfigFromString(tt.yaml)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
				require.NotNil(t, config)
				if tt.check != nil {
					tt.check(t, config)
				}
			}
		})
	}
}

func TestValidateWorkflowConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *WorkflowConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: &WorkflowConfig{
				ActiveWorkflows: []WorkflowDefinition{
					{
						Name: "test-workflow",
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
				},
			},
			wantErr: false,
		},
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
			errMsg:  "workflow config cannot be nil",
		},
		{
			name: "empty workflows",
			config: &WorkflowConfig{
				ActiveWorkflows: []WorkflowDefinition{},
			},
			wantErr: true,
			errMsg:  "no active workflows configured",
		},
		{
			name: "duplicate workflow names",
			config: &WorkflowConfig{
				ActiveWorkflows: []WorkflowDefinition{
					{
						Name: "duplicate-name",
						Triggers: []TriggerDefinition{
							{
								OnAlertmanagerAlert: &AlertmanagerAlertTrigger{
									AlertName: "Alert1",
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
					},
					{
						Name: "duplicate-name",
						Triggers: []TriggerDefinition{
							{
								OnAlertmanagerAlert: &AlertmanagerAlertTrigger{
									AlertName: "Alert2",
								},
							},
						},
						Actions: []ActionDefinition{
							{
								ActionType: "action2",
								RawData: map[string]interface{}{
									"action2": map[string]interface{}{},
								},
							},
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "duplicate workflow name 'duplicate-name' found",
		},
		{
			name: "invalid workflow",
			config: &WorkflowConfig{
				ActiveWorkflows: []WorkflowDefinition{
					{
						Name:     "", // Invalid: empty name
						Triggers: []TriggerDefinition{},
						Actions:  []ActionDefinition{},
					},
				},
			},
			wantErr: true,
			errMsg:  "workflow 0 validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWorkflowConfig(tt.config)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestLoadWorkflowConfig(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
		errMsg  string
		check   func(*testing.T, *WorkflowConfig)
	}{
		{
			name: "valid config file",
			content: `
active_workflows:
  - name: "file-workflow"
    triggers:
      - on_alertmanager_alert:
          alert_name: "FileAlert"
    actions:
      - file_action: {}
`,
			wantErr: false,
			check: func(t *testing.T, config *WorkflowConfig) {
				t.Helper()
				require.Len(t, config.ActiveWorkflows, 1)
				assert.Equal(t, "file-workflow", config.ActiveWorkflows[0].Name)
			},
		},
		{
			name: "invalid config file",
			content: `
active_workflows:
  - name: ""
    triggers: []
    actions: []
`,
			wantErr: true,
			errMsg:  "workflow config validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := setupTestConfig(t, tt.content)

			config, err := LoadWorkflowConfig(configPath)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
				require.NotNil(t, config)
				if tt.check != nil {
					tt.check(t, config)
				}
			}
		})
	}
}

func TestLoadWorkflowConfig_FileNotFound(t *testing.T) {
	config, err := LoadWorkflowConfig("/nonexistent/path/config.yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workflow config file not found")
	assert.Nil(t, config)
}

func TestLoadWorkflowConfig_EmptyPath(t *testing.T) {
	config, err := LoadWorkflowConfig("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "config path cannot be empty")
	assert.Nil(t, config)
}

func TestWorkflowConfig_GetWorkflowsByTriggerType(t *testing.T) {
	config := &WorkflowConfig{
		ActiveWorkflows: []WorkflowDefinition{
			{
				Name: "alertmanager-workflow",
				Triggers: []TriggerDefinition{
					{
						OnAlertmanagerAlert: &AlertmanagerAlertTrigger{
							AlertName: "Alert1",
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
			},
			{
				Name: "another-alertmanager-workflow",
				Triggers: []TriggerDefinition{
					{
						OnAlertmanagerAlert: &AlertmanagerAlertTrigger{
							AlertName: "Alert2",
						},
					},
				},
				Actions: []ActionDefinition{
					{
						ActionType: "action2",
						RawData: map[string]interface{}{
							"action2": map[string]interface{}{},
						},
					},
				},
			},
		},
	}

	alertmanagerWorkflows := config.GetWorkflowsByTriggerType("alertmanager_alert")
	assert.Len(t, alertmanagerWorkflows, 2)

	kubernetesWorkflows := config.GetWorkflowsByTriggerType("kubernetes_event")
	assert.Empty(t, kubernetesWorkflows)
}

func TestWorkflowConfig_GetWorkflowByName(t *testing.T) {
	config := &WorkflowConfig{
		ActiveWorkflows: []WorkflowDefinition{
			{
				Name: "test-workflow",
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
		},
	}

	workflow, err := config.GetWorkflowByName("test-workflow")
	require.NoError(t, err)
	assert.Equal(t, "test-workflow", workflow.Name)

	workflow, err = config.GetWorkflowByName("nonexistent-workflow")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "workflow 'nonexistent-workflow' not found")
	assert.Nil(t, workflow)
}

func TestWorkflowConfig_GetWorkflowCount(t *testing.T) {
	config := &WorkflowConfig{
		ActiveWorkflows: []WorkflowDefinition{
			{Name: "workflow1"},
			{Name: "workflow2"},
			{Name: "workflow3"},
		},
	}

	assert.Equal(t, 3, config.GetWorkflowCount())

	emptyConfig := &WorkflowConfig{
		ActiveWorkflows: []WorkflowDefinition{},
	}
	assert.Equal(t, 0, emptyConfig.GetWorkflowCount())
}

func TestWorkflowConfig_GetTriggerTypeCounts(t *testing.T) {
	config := &WorkflowConfig{
		ActiveWorkflows: []WorkflowDefinition{
			{
				Name: "workflow1",
				Triggers: []TriggerDefinition{
					{
						OnAlertmanagerAlert: &AlertmanagerAlertTrigger{
							AlertName: "Alert1",
						},
					},
					{
						OnAlertmanagerAlert: &AlertmanagerAlertTrigger{
							AlertName: "Alert2",
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
			},
			{
				Name: "workflow2",
				Triggers: []TriggerDefinition{
					{
						OnAlertmanagerAlert: &AlertmanagerAlertTrigger{
							AlertName: "Alert3",
						},
					},
				},
				Actions: []ActionDefinition{
					{
						ActionType: "action2",
						RawData: map[string]interface{}{
							"action2": map[string]interface{}{},
						},
					},
				},
			},
		},
	}

	counts := config.GetTriggerTypeCounts()
	assert.Equal(t, 3, counts["alertmanager_alert"])
}

func TestDefaultWorkflowConfig(t *testing.T) {
	config := DefaultWorkflowConfig()

	require.NotNil(t, config)
	require.Len(t, config.ActiveWorkflows, 1)

	workflow := config.ActiveWorkflows[0]
	assert.Equal(t, "AlertManagerAlertEnrichment", workflow.Name)
	assert.Len(t, workflow.Triggers, 1)
	assert.Len(t, workflow.Actions, 1)
	assert.False(t, workflow.Stop)

	err := ValidateWorkflowConfig(config)
	assert.NoError(t, err)
}

func TestConfigLoader(t *testing.T) {
	content := `
active_workflows:
  - name: "loader-test-workflow"
    triggers:
      - on_alertmanager_alert:
          alert_name: "LoaderTestAlert"
    actions:
      - loader_action: {}
`

	configPath := setupTestConfig(t, content)

	loader := NewConfigLoader(configPath)
	config, err := loader.LoadConfig()

	require.NoError(t, err)
	require.NotNil(t, config)
	require.Len(t, config.ActiveWorkflows, 1)
	assert.Equal(t, "loader-test-workflow", config.ActiveWorkflows[0].Name)
}

func TestLoadWorkflowConfigFromReader(t *testing.T) {
	yamlContent := `
active_workflows:
  - name: "reader-test-workflow"
    triggers:
      - on_alertmanager_alert:
          alert_name: "ReaderTestAlert"
    actions:
      - reader_action: {}
`

	reader := strings.NewReader(yamlContent)
	config, err := LoadWorkflowConfigFromReader(reader)

	require.NoError(t, err)
	require.NotNil(t, config)
	require.Len(t, config.ActiveWorkflows, 1)
	assert.Equal(t, "reader-test-workflow", config.ActiveWorkflows[0].Name)
}
