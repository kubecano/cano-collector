package actions

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	actions_interfaces "github.com/kubecano/cano-collector/pkg/workflow/actions/interfaces"
)

func TestNewPodLogsActionConfigWithDefaults(t *testing.T) {
	baseConfig := actions_interfaces.ActionConfig{
		Name: "test-pod-logs",
		Type: "pod_logs",
	}

	config := NewPodLogsActionConfigWithDefaults(baseConfig)

	assert.Equal(t, "test-pod-logs", config.Name)
	assert.Equal(t, "pod_logs", config.Type)
	assert.Equal(t, 1000, config.MaxLines)
	assert.Equal(t, 100, config.TailLines)
	assert.Equal(t, false, config.Previous)
	assert.Equal(t, true, config.Timestamps)
	assert.Equal(t, true, config.IncludeTimestamp)
	assert.Equal(t, true, config.IncludeContainer)
	assert.Equal(t, "20060102-150405", config.TimestampFormat)
}

func TestNewPodLogsActionConfigWithDefaultsEnvironmentVariables(t *testing.T) {
	// Set environment variables
	os.Setenv("WORKFLOW_POD_LOGS_DEFAULT_MAX_LINES", "2000")
	os.Setenv("WORKFLOW_POD_LOGS_DEFAULT_TAIL_LINES", "200")
	os.Setenv("WORKFLOW_POD_LOGS_DEFAULT_PREVIOUS", "true")
	os.Setenv("WORKFLOW_POD_LOGS_DEFAULT_TIMESTAMPS", "false")
	os.Setenv("WORKFLOW_POD_LOGS_INCLUDE_TIMESTAMP", "false")
	os.Setenv("WORKFLOW_POD_LOGS_INCLUDE_CONTAINER", "false")
	os.Setenv("WORKFLOW_POD_LOGS_TIMESTAMP_FORMAT", "2006-01-02_15-04-05")

	defer func() {
		os.Unsetenv("WORKFLOW_POD_LOGS_DEFAULT_MAX_LINES")
		os.Unsetenv("WORKFLOW_POD_LOGS_DEFAULT_TAIL_LINES")
		os.Unsetenv("WORKFLOW_POD_LOGS_DEFAULT_PREVIOUS")
		os.Unsetenv("WORKFLOW_POD_LOGS_DEFAULT_TIMESTAMPS")
		os.Unsetenv("WORKFLOW_POD_LOGS_INCLUDE_TIMESTAMP")
		os.Unsetenv("WORKFLOW_POD_LOGS_INCLUDE_CONTAINER")
		os.Unsetenv("WORKFLOW_POD_LOGS_TIMESTAMP_FORMAT")
	}()

	baseConfig := actions_interfaces.ActionConfig{
		Name: "test-pod-logs",
		Type: "pod_logs",
	}

	config := NewPodLogsActionConfigWithDefaults(baseConfig)

	assert.Equal(t, 2000, config.MaxLines)
	assert.Equal(t, 200, config.TailLines)
	assert.Equal(t, true, config.Previous)
	assert.Equal(t, false, config.Timestamps)
	assert.Equal(t, false, config.IncludeTimestamp)
	assert.Equal(t, false, config.IncludeContainer)
	assert.Equal(t, "2006-01-02_15-04-05", config.TimestampFormat)
}

func TestPodLogsActionConfigApplyJavaDefaults(t *testing.T) {
	// Set Java-specific environment variables
	os.Setenv("WORKFLOW_POD_LOGS_JAVA_DEFAULT_MAX_LINES", "10000")
	os.Setenv("WORKFLOW_POD_LOGS_JAVA_DEFAULT_TAIL_LINES", "1000")

	defer func() {
		os.Unsetenv("WORKFLOW_POD_LOGS_JAVA_DEFAULT_MAX_LINES")
		os.Unsetenv("WORKFLOW_POD_LOGS_JAVA_DEFAULT_TAIL_LINES")
	}()

	baseConfig := actions_interfaces.ActionConfig{
		Name: "test-pod-logs",
		Type: "pod_logs",
	}

	config := NewPodLogsActionConfigWithDefaults(baseConfig)
	config.ApplyJavaDefaults()

	assert.Equal(t, true, config.JavaSpecific)
	assert.Equal(t, 10000, config.MaxLines)
	assert.Equal(t, 1000, config.TailLines)
}

func TestPodLogsActionConfigValidate(t *testing.T) {
	tests := []struct {
		name      string
		config    PodLogsActionConfig
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid configuration",
			config: PodLogsActionConfig{
				ActionConfig: actions_interfaces.ActionConfig{
					Name: "test-action",
					Type: "pod_logs",
				},
				MaxLines:        1000,
				TailLines:       100,
				TimestampFormat: "20060102-150405",
			},
			wantError: false,
		},
		{
			name: "empty name",
			config: PodLogsActionConfig{
				ActionConfig: actions_interfaces.ActionConfig{
					Name: "",
					Type: "pod_logs",
				},
				TimestampFormat: "20060102-150405",
			},
			wantError: true,
			errorMsg:  "action name cannot be empty",
		},
		{
			name: "negative max_lines",
			config: PodLogsActionConfig{
				ActionConfig: actions_interfaces.ActionConfig{
					Name: "test-action",
					Type: "pod_logs",
				},
				MaxLines:        -1,
				TimestampFormat: "20060102-150405",
			},
			wantError: true,
			errorMsg:  "max_lines must be non-negative",
		},
		{
			name: "negative tail_lines",
			config: PodLogsActionConfig{
				ActionConfig: actions_interfaces.ActionConfig{
					Name: "test-action",
					Type: "pod_logs",
				},
				MaxLines:        1000,
				TailLines:       -1,
				TimestampFormat: "20060102-150405",
			},
			wantError: true,
			errorMsg:  "tail_lines must be non-negative",
		},
		{
			name: "empty timestamp format",
			config: PodLogsActionConfig{
				ActionConfig: actions_interfaces.ActionConfig{
					Name: "test-action",
					Type: "pod_logs",
				},
				MaxLines:        1000,
				TailLines:       100,
				TimestampFormat: "",
			},
			wantError: true,
			errorMsg:  "timestamp_format cannot be empty",
		},
		{
			name: "empty since_time string",
			config: PodLogsActionConfig{
				ActionConfig: actions_interfaces.ActionConfig{
					Name: "test-action",
					Type: "pod_logs",
				},
				MaxLines:        1000,
				TailLines:       100,
				SinceTime:       "   ",
				TimestampFormat: "20060102-150405",
			},
			wantError: true,
			errorMsg:  "since_time cannot be empty string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPodLogsActionConfigUpdateFromParameters(t *testing.T) {
	config := PodLogsActionConfig{
		ActionConfig: actions_interfaces.ActionConfig{
			Name: "test-action",
			Type: "pod_logs",
		},
		MaxLines:        1000,
		TailLines:       100,
		TimestampFormat: "20060102-150405",
	}

	parameters := map[string]interface{}{
		"max_lines":         2000,
		"tail_lines":        200,
		"previous":          true,
		"timestamps":        false,
		"container":         "app-container",
		"since_time":        "2023-01-01T00:00:00Z",
		"java_specific":     true,
		"include_timestamp": false,
		"include_container": false,
		"timestamp_format":  "2006-01-02_15-04-05",
	}

	err := config.UpdateFromParameters(parameters)
	require.NoError(t, err)

	assert.Equal(t, 2000, config.MaxLines)
	assert.Equal(t, 200, config.TailLines)
	assert.Equal(t, true, config.Previous)
	assert.Equal(t, false, config.Timestamps)
	assert.Equal(t, "app-container", config.Container)
	assert.Equal(t, "2023-01-01T00:00:00Z", config.SinceTime)
	assert.Equal(t, true, config.JavaSpecific)
	assert.Equal(t, false, config.IncludeTimestamp)
	assert.Equal(t, false, config.IncludeContainer)
	assert.Equal(t, "2006-01-02_15-04-05", config.TimestampFormat)
}

func TestPodLogsActionConfigUpdateFromParametersInvalidTypes(t *testing.T) {
	config := PodLogsActionConfig{
		ActionConfig: actions_interfaces.ActionConfig{
			Name: "test-action",
			Type: "pod_logs",
		},
		TimestampFormat: "20060102-150405",
	}

	tests := []struct {
		name       string
		parameters map[string]interface{}
		errorMsg   string
	}{
		{
			name:       "invalid max_lines type",
			parameters: map[string]interface{}{"max_lines": "invalid"},
			errorMsg:   "max_lines must be an integer",
		},
		{
			name:       "invalid tail_lines type",
			parameters: map[string]interface{}{"tail_lines": "invalid"},
			errorMsg:   "tail_lines must be an integer",
		},
		{
			name:       "invalid previous type",
			parameters: map[string]interface{}{"previous": "invalid"},
			errorMsg:   "previous must be a boolean",
		},
		{
			name:       "invalid timestamps type",
			parameters: map[string]interface{}{"timestamps": "invalid"},
			errorMsg:   "timestamps must be a boolean",
		},
		{
			name:       "invalid container type",
			parameters: map[string]interface{}{"container": 123},
			errorMsg:   "container must be a string",
		},
		{
			name:       "invalid since_time type",
			parameters: map[string]interface{}{"since_time": 123},
			errorMsg:   "since_time must be a string",
		},
		{
			name:       "invalid java_specific type",
			parameters: map[string]interface{}{"java_specific": "invalid"},
			errorMsg:   "java_specific must be a boolean",
		},
		{
			name:       "invalid include_timestamp type",
			parameters: map[string]interface{}{"include_timestamp": "invalid"},
			errorMsg:   "include_timestamp must be a boolean",
		},
		{
			name:       "invalid include_container type",
			parameters: map[string]interface{}{"include_container": "invalid"},
			errorMsg:   "include_container must be a boolean",
		},
		{
			name:       "invalid timestamp_format type",
			parameters: map[string]interface{}{"timestamp_format": 123},
			errorMsg:   "timestamp_format must be a string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := config.UpdateFromParameters(tt.parameters)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errorMsg)
		})
	}
}

func TestIsJavaContainer(t *testing.T) {
	tests := []struct {
		name          string
		containerName string
		imageName     string
		expected      bool
	}{
		{
			name:          "java in image name",
			containerName: "app",
			imageName:     "openjdk:11",
			expected:      true,
		},
		{
			name:          "spring in image name",
			containerName: "app",
			imageName:     "springboot/app:latest",
			expected:      true,
		},
		{
			name:          "java in container name",
			containerName: "java-app",
			imageName:     "myregistry/app:v1.0",
			expected:      true,
		},
		{
			name:          "no java indicators",
			containerName: "nginx",
			imageName:     "nginx:alpine",
			expected:      false,
		},
		{
			name:          "case insensitive match",
			containerName: "APP",
			imageName:     "OPENJDK:11",
			expected:      true,
		},
		{
			name:          "kafka in image name",
			containerName: "broker",
			imageName:     "confluentinc/cp-kafka:latest",
			expected:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsJavaContainer(tt.containerName, tt.imageName)
			assert.Equal(t, tt.expected, result)
		})
	}
}
