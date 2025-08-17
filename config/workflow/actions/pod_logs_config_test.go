package actions

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	actions_interfaces "github.com/kubecano/cano-collector/pkg/workflow/actions/interfaces"
)

func TestIsJavaContainer_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		containerName string
		imageName     string
		expected      bool
	}{
		{
			name:          "empty names",
			containerName: "",
			imageName:     "",
			expected:      false,
		},
		{
			name:          "case insensitive java",
			containerName: "JAVA-app",
			imageName:     "",
			expected:      true,
		},
		{
			name:          "openjdk in image",
			containerName: "",
			imageName:     "docker.io/openjdk:11",
			expected:      true,
		},
		{
			name:          "spring boot",
			containerName: "spring-boot-app",
			imageName:     "springio/spring-boot:latest",
			expected:      true,
		},
		{
			name:          "elasticsearch",
			containerName: "es-container",
			imageName:     "elasticsearch:7.15.0",
			expected:      true,
		},
		{
			name:          "kafka",
			containerName: "kafka-broker",
			imageName:     "confluentinc/cp-kafka:latest",
			expected:      true,
		},
		{
			name:          "non-java container",
			containerName: "nginx",
			imageName:     "nginx:alpine",
			expected:      false,
		},
		{
			name:          "python app",
			containerName: "python-api",
			imageName:     "python:3.9",
			expected:      false,
		},
		{
			name:          "partial match should not work",
			containerName: "javanese-app",
			imageName:     "",
			expected:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsJavaContainer(tt.containerName, tt.imageName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPodLogsActionConfig_UpdateFromParameters_AllTypes(t *testing.T) {
	config := PodLogsActionConfig{}

	parameters := map[string]interface{}{
		"max_lines":         2000,
		"tail_lines":        200,
		"previous":          true,
		"timestamps":        false,
		"container":         "test-container",
		"since_time":        "2025-01-08T10:00:00Z",
		"java_specific":     true,
		"include_timestamp": false,
		"include_container": false,
		"timestamp_format":  "2006-01-02",
	}

	err := config.UpdateFromParameters(parameters)

	require.NoError(t, err)
	assert.Equal(t, 5000, config.MaxLines) // Should be overridden by Java defaults
	assert.Equal(t, 500, config.TailLines) // Should be overridden by Java defaults
	assert.True(t, config.Previous)
	assert.False(t, config.Timestamps)
	assert.Equal(t, "test-container", config.Container)
	assert.Equal(t, "2025-01-08T10:00:00Z", config.SinceTime)
	assert.True(t, config.JavaSpecific)
	assert.False(t, config.IncludeTimestamp)
	assert.False(t, config.IncludeContainer)
	assert.Equal(t, "2006-01-02", config.TimestampFormat)
}

func TestPodLogsActionConfig_UpdateFromParameters_TypeErrors(t *testing.T) {
	tests := []struct {
		name       string
		parameters map[string]interface{}
		wantErr    string
	}{
		{
			name: "max_lines not int",
			parameters: map[string]interface{}{
				"max_lines": "not-int",
			},
			wantErr: "max_lines must be an integer",
		},
		{
			name: "tail_lines not int",
			parameters: map[string]interface{}{
				"tail_lines": 3.14,
			},
			wantErr: "tail_lines must be an integer",
		},
		{
			name: "previous not bool",
			parameters: map[string]interface{}{
				"previous": "yes",
			},
			wantErr: "previous must be a boolean",
		},
		{
			name: "timestamps not bool",
			parameters: map[string]interface{}{
				"timestamps": 1,
			},
			wantErr: "timestamps must be a boolean",
		},
		{
			name: "container not string",
			parameters: map[string]interface{}{
				"container": 123,
			},
			wantErr: "container must be a string",
		},
		{
			name: "since_time not string",
			parameters: map[string]interface{}{
				"since_time": 123456,
			},
			wantErr: "since_time must be a string",
		},
		{
			name: "java_specific not bool",
			parameters: map[string]interface{}{
				"java_specific": "true",
			},
			wantErr: "java_specific must be a boolean",
		},
		{
			name: "include_timestamp not bool",
			parameters: map[string]interface{}{
				"include_timestamp": "false",
			},
			wantErr: "include_timestamp must be a boolean",
		},
		{
			name: "include_container not bool",
			parameters: map[string]interface{}{
				"include_container": 0,
			},
			wantErr: "include_container must be a boolean",
		},
		{
			name: "timestamp_format not string",
			parameters: map[string]interface{}{
				"timestamp_format": []string{"format"},
			},
			wantErr: "timestamp_format must be a string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := PodLogsActionConfig{}
			err := config.UpdateFromParameters(tt.parameters)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestPodLogsActionConfig_Validate_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		config  PodLogsActionConfig
		wantErr string
	}{
		{
			name: "valid minimal config",
			config: PodLogsActionConfig{
				ActionConfig: actions_interfaces.ActionConfig{
					Name: "test",
				},
				MaxLines:        0,
				TailLines:       0,
				TimestampFormat: "20060102",
			},
			wantErr: "",
		},
		{
			name: "empty name",
			config: PodLogsActionConfig{
				ActionConfig: actions_interfaces.ActionConfig{
					Name: "",
				},
				TimestampFormat: "20060102",
			},
			wantErr: "action name cannot be empty",
		},
		{
			name: "negative max_lines",
			config: PodLogsActionConfig{
				ActionConfig: actions_interfaces.ActionConfig{
					Name: "test",
				},
				MaxLines:        -1,
				TimestampFormat: "20060102",
			},
			wantErr: "max_lines must be non-negative",
		},
		{
			name: "negative tail_lines",
			config: PodLogsActionConfig{
				ActionConfig: actions_interfaces.ActionConfig{
					Name: "test",
				},
				TailLines:       -5,
				TimestampFormat: "20060102",
			},
			wantErr: "tail_lines must be non-negative",
		},
		{
			name: "empty since_time string",
			config: PodLogsActionConfig{
				ActionConfig: actions_interfaces.ActionConfig{
					Name: "test",
				},
				SinceTime:       "   ",
				TimestampFormat: "20060102",
			},
			wantErr: "since_time cannot be empty string",
		},
		{
			name: "empty timestamp_format",
			config: PodLogsActionConfig{
				ActionConfig: actions_interfaces.ActionConfig{
					Name: "test",
				},
				TimestampFormat: "",
			},
			wantErr: "timestamp_format cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestEnvHelperFunctions(t *testing.T) {
	// Test getEnvInt
	os.Setenv("TEST_INT_VAR", "42")
	defer os.Unsetenv("TEST_INT_VAR")

	result := getEnvInt("TEST_INT_VAR", 10)
	assert.Equal(t, 42, result)

	// Test with invalid int
	os.Setenv("TEST_INVALID_INT", "not-a-number")
	defer os.Unsetenv("TEST_INVALID_INT")

	result = getEnvInt("TEST_INVALID_INT", 10)
	assert.Equal(t, 10, result) // Should return default

	// Test with missing env
	result = getEnvInt("MISSING_VAR", 5)
	assert.Equal(t, 5, result)

	// Test getEnvBool
	os.Setenv("TEST_BOOL_VAR", "true")
	defer os.Unsetenv("TEST_BOOL_VAR")

	boolResult := getEnvBool("TEST_BOOL_VAR", false)
	assert.True(t, boolResult)

	// Test with "false"
	os.Setenv("TEST_BOOL_FALSE", "false")
	defer os.Unsetenv("TEST_BOOL_FALSE")

	boolResult = getEnvBool("TEST_BOOL_FALSE", true)
	assert.False(t, boolResult)

	// Test with invalid bool
	os.Setenv("TEST_INVALID_BOOL", "maybe")
	defer os.Unsetenv("TEST_INVALID_BOOL")

	boolResult = getEnvBool("TEST_INVALID_BOOL", true)
	assert.True(t, boolResult) // Should return default

	// Test getEnvString
	os.Setenv("TEST_STRING_VAR", "test-value")
	defer os.Unsetenv("TEST_STRING_VAR")

	stringResult := getEnvString("TEST_STRING_VAR", "default")
	assert.Equal(t, "test-value", stringResult)

	// Test with missing env
	stringResult = getEnvString("MISSING_STRING_VAR", "default")
	assert.Equal(t, "default", stringResult)
}

func TestNewPodLogsActionConfigWithDefaults_CustomEnvs(t *testing.T) {
	// Set custom environment variables
	os.Setenv("WORKFLOW_POD_LOGS_DEFAULT_MAX_LINES", "2000")
	os.Setenv("WORKFLOW_POD_LOGS_DEFAULT_TAIL_LINES", "150")
	os.Setenv("WORKFLOW_POD_LOGS_DEFAULT_PREVIOUS", "true")
	os.Setenv("WORKFLOW_POD_LOGS_DEFAULT_TIMESTAMPS", "false")
	os.Setenv("WORKFLOW_POD_LOGS_INCLUDE_TIMESTAMP", "false")
	os.Setenv("WORKFLOW_POD_LOGS_INCLUDE_CONTAINER", "false")
	os.Setenv("WORKFLOW_POD_LOGS_TIMESTAMP_FORMAT", "2006-01-02")

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
		Name:    "test",
		Type:    "pod_logs",
		Enabled: true,
	}

	config := NewPodLogsActionConfigWithDefaults(baseConfig)

	assert.Equal(t, 2000, config.MaxLines)
	assert.Equal(t, 150, config.TailLines)
	assert.True(t, config.Previous)
	assert.False(t, config.Timestamps)
	assert.False(t, config.IncludeTimestamp)
	assert.False(t, config.IncludeContainer)
	assert.Equal(t, "2006-01-02", config.TimestampFormat)
}

func TestContainsWord(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		word     string
		expected bool
	}{
		{
			name:     "exact match",
			text:     "java",
			word:     "java",
			expected: true,
		},
		{
			name:     "word at start",
			text:     "java-app",
			word:     "java",
			expected: true,
		},
		{
			name:     "word at end",
			text:     "my-java",
			word:     "java",
			expected: true,
		},
		{
			name:     "word in middle",
			text:     "my-java-app",
			word:     "java",
			expected: true,
		},
		{
			name:     "partial match should fail",
			text:     "javanese",
			word:     "java",
			expected: false,
		},
		{
			name:     "partial match at end should fail",
			text:     "prejava",
			word:     "java",
			expected: false,
		},
		{
			name:     "word not present",
			text:     "python-app",
			word:     "java",
			expected: false,
		},
		{
			name:     "empty text",
			text:     "",
			word:     "java",
			expected: false,
		},
		{
			name:     "empty word",
			text:     "java-app",
			word:     "",
			expected: true, // Empty string is contained in any string
		},
		{
			name:     "case sensitive",
			text:     "Java-app",
			word:     "java",
			expected: false, // Case sensitive
		},
		{
			name:     "with numbers",
			text:     "openjdk11",
			word:     "openjdk",
			expected: false, // "openjdk" followed by alphanumeric
		},
		{
			name:     "with numbers separated",
			text:     "openjdk:11",
			word:     "openjdk",
			expected: true, // "openjdk" followed by non-alphanumeric
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsWord(tt.text, tt.word)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsAlphaNumeric(t *testing.T) {
	tests := []struct {
		name     string
		char     byte
		expected bool
	}{
		{"lowercase letter", 'a', true},
		{"uppercase letter", 'Z', true},
		{"digit", '5', true},
		{"hyphen", '-', false},
		{"underscore", '_', false},
		{"dot", '.', false},
		{"colon", ':', false},
		{"space", ' ', false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isAlphaNumeric(tt.char)
			assert.Equal(t, tt.expected, result)
		})
	}
}
