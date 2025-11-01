package actions

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pod_logs_config "github.com/kubecano/cano-collector/config/workflow/actions"
	"github.com/kubecano/cano-collector/pkg/core/event"
	"github.com/kubecano/cano-collector/pkg/core/issue"
	"github.com/kubecano/cano-collector/pkg/logger"
	"github.com/kubecano/cano-collector/pkg/metric"
	actions_interfaces "github.com/kubecano/cano-collector/pkg/workflow/actions/interfaces"
)

// MockKubernetesClientForTest implements KubernetesClient for testing
type MockKubernetesClientForTest struct {
	logs      map[string]string // key: namespace/podName, value: logs
	shouldErr bool
}

func NewMockKubernetesClientForTest() *MockKubernetesClientForTest {
	return &MockKubernetesClientForTest{
		logs: make(map[string]string),
	}
}

func (m *MockKubernetesClientForTest) GetPodLogs(ctx context.Context, namespace, podName string, options map[string]interface{}) (string, error) {
	if m.shouldErr {
		return "", assert.AnError
	}

	key := namespace + "/" + podName
	if logs, exists := m.logs[key]; exists {
		return logs, nil
	}

	return "Mock logs for " + key, nil
}

func (m *MockKubernetesClientForTest) SetLogs(namespace, podName, logs string) {
	key := namespace + "/" + podName
	m.logs[key] = logs
}

func (m *MockKubernetesClientForTest) SetShouldError(shouldErr bool) {
	m.shouldErr = shouldErr
}

// Helper function to create test workflow event for pod logs tests
func createPodLogsTestWorkflowEvent() event.WorkflowEvent {
	alertEvent := &event.AlertManagerEvent{
		BaseEvent: event.BaseEvent{
			ID:        uuid.New(),
			Timestamp: time.Now(),
			Source:    "test",
			Type:      event.EventTypeAlertManager,
		},
		Status: "firing",
		Alerts: []event.PrometheusAlert{
			{
				Labels: map[string]string{
					"alertname": "TestAlert",
					"pod":       "test-pod",
					"namespace": "test-namespace",
					"container": "test-container",
				},
				Annotations: map[string]string{
					"summary": "Test alert",
				},
			},
		},
	}

	return event.NewAlertManagerWorkflowEvent(alertEvent)
}

func TestNewPodLogsActionConfigWithDefaults(t *testing.T) {
	// Set environment variables
	os.Setenv("WORKFLOW_POD_LOGS_DEFAULT_TAIL_LINES", "200")
	os.Setenv("WORKFLOW_POD_LOGS_DEFAULT_MAX_LINES", "2000")
	os.Setenv("WORKFLOW_POD_LOGS_DEFAULT_TIMESTAMPS", "false")
	os.Setenv("WORKFLOW_POD_LOGS_INCLUDE_TIMESTAMP", "false")
	defer func() {
		os.Unsetenv("WORKFLOW_POD_LOGS_DEFAULT_TAIL_LINES")
		os.Unsetenv("WORKFLOW_POD_LOGS_DEFAULT_MAX_LINES")
		os.Unsetenv("WORKFLOW_POD_LOGS_DEFAULT_TIMESTAMPS")
		os.Unsetenv("WORKFLOW_POD_LOGS_INCLUDE_TIMESTAMP")
	}()

	baseConfig := actions_interfaces.ActionConfig{
		Name: "test-action",
		Type: "pod_logs",
	}

	config := pod_logs_config.NewPodLogsActionConfigWithDefaults(baseConfig)

	assert.Equal(t, "test-action", config.Name)
	assert.Equal(t, "pod_logs", config.Type)
	assert.Equal(t, 200, config.TailLines)
	assert.Equal(t, 2000, config.MaxLines)
	assert.False(t, config.Timestamps)
	assert.False(t, config.IncludeTimestamp)
}

func TestPodLogsActionConfig_ApplyJavaDefaults(t *testing.T) {
	// Set Java-specific environment variables
	os.Setenv("WORKFLOW_POD_LOGS_JAVA_DEFAULT_TAIL_LINES", "600")
	os.Setenv("WORKFLOW_POD_LOGS_JAVA_DEFAULT_MAX_LINES", "6000")
	defer func() {
		os.Unsetenv("WORKFLOW_POD_LOGS_JAVA_DEFAULT_TAIL_LINES")
		os.Unsetenv("WORKFLOW_POD_LOGS_JAVA_DEFAULT_MAX_LINES")
	}()

	config := pod_logs_config.PodLogsActionConfig{
		TailLines: 100,
		MaxLines:  1000,
	}

	config.ApplyJavaDefaults()

	assert.True(t, config.JavaSpecific)
	assert.Equal(t, 600, config.TailLines)
	assert.Equal(t, 6000, config.MaxLines)
}

func TestPodLogsAction_Execute_Success(t *testing.T) {
	logger := logger.NewLogger("debug", "test")
	metrics := metric.NewMetricsCollector(logger)
	mockClient := NewMockKubernetesClientForTest()

	config := pod_logs_config.PodLogsActionConfig{
		ActionConfig: actions_interfaces.ActionConfig{
			Name:    "test-pod-logs",
			Type:    "pod_logs",
			Enabled: true,
		},
		MaxLines:         1000,
		TailLines:        100,
		Timestamps:       true,
		IncludeTimestamp: true,
		IncludeContainer: true,
		TimestampFormat:  "20060102-150405",
	}

	action := NewPodLogsAction(config, logger, metrics, mockClient)

	// Set mock logs
	mockClient.SetLogs("test-namespace", "test-pod", "Test log line 1\nTest log line 2")

	workflowEvent := createPodLogsTestWorkflowEvent()
	result, err := action.Execute(context.Background(), workflowEvent)

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "test-pod", result.Data.(map[string]interface{})["pod_name"])
	assert.Equal(t, "test-namespace", result.Data.(map[string]interface{})["namespace"])
	assert.Len(t, result.Enrichments, 1)

	// Check enrichment
	enrichment := result.Enrichments[0]
	assert.Equal(t, issue.EnrichmentTypeLogs, enrichment.Type)
	assert.Contains(t, enrichment.Title, "Pod Logs: test-namespace/test-pod")
}

func TestPodLogsAction_Execute_JavaContainer(t *testing.T) {
	logger := logger.NewLogger("debug", "test")
	metrics := metric.NewMetricsCollector(logger)
	mockClient := NewMockKubernetesClientForTest()

	// Set Java-specific env vars
	os.Setenv("WORKFLOW_POD_LOGS_JAVA_DEFAULT_TAIL_LINES", "500")
	os.Setenv("WORKFLOW_POD_LOGS_JAVA_DEFAULT_MAX_LINES", "5000")
	defer func() {
		os.Unsetenv("WORKFLOW_POD_LOGS_JAVA_DEFAULT_TAIL_LINES")
		os.Unsetenv("WORKFLOW_POD_LOGS_JAVA_DEFAULT_MAX_LINES")
	}()

	config := pod_logs_config.PodLogsActionConfig{
		ActionConfig: actions_interfaces.ActionConfig{
			Name:    "test-java-logs",
			Type:    "pod_logs",
			Enabled: true,
		},
		MaxLines:         1000,
		TailLines:        100,
		IncludeTimestamp: true,
		TimestampFormat:  "20060102-150405",
	}

	action := NewPodLogsAction(config, logger, metrics, mockClient)

	// Create event with Java container
	alertEvent := &event.AlertManagerEvent{
		BaseEvent: event.BaseEvent{
			ID:        uuid.New(),
			Timestamp: time.Now(),
			Source:    "test",
			Type:      event.EventTypeAlertManager,
		},
		Status: "firing",
		Alerts: []event.PrometheusAlert{
			{
				Labels: map[string]string{
					"alertname": "JavaAppAlert",
					"pod":       "java-app-pod",
					"namespace": "default",
					"container": "java-app", // Java container name
				},
			},
		},
	}

	workflowEvent := event.NewAlertManagerWorkflowEvent(alertEvent)

	mockClient.SetLogs("default", "java-app-pod", "Java app log line")

	result, err := action.Execute(context.Background(), workflowEvent)

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.True(t, result.Data.(map[string]interface{})["java_specific"].(bool))

	// Check that Java-specific title is used
	enrichment := result.Enrichments[0]
	assert.Contains(t, enrichment.Title, "Java Pod Logs")
}

func TestPodLogsAction_Execute_NoPodFound(t *testing.T) {
	logger := logger.NewLogger("debug", "test")
	metrics := metric.NewMetricsCollector(logger)
	mockClient := NewMockKubernetesClientForTest()

	config := pod_logs_config.PodLogsActionConfig{
		ActionConfig: actions_interfaces.ActionConfig{
			Name:    "test-pod-logs",
			Type:    "pod_logs",
			Enabled: true,
		},
	}

	action := NewPodLogsAction(config, logger, metrics, mockClient)

	// Create event without pod information
	alertEvent := &event.AlertManagerEvent{
		BaseEvent: event.BaseEvent{
			ID:        uuid.New(),
			Timestamp: time.Now(),
			Source:    "test",
			Type:      event.EventTypeAlertManager,
		},
		Status: "firing",
		Alerts: []event.PrometheusAlert{
			{
				Labels: map[string]string{
					"alertname": "TestAlert",
					// No pod label
				},
			},
		},
	}

	workflowEvent := event.NewAlertManagerWorkflowEvent(alertEvent)

	result, err := action.Execute(context.Background(), workflowEvent)

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Equal(t, "no pod information found in event", result.Data)
	assert.Empty(t, result.Enrichments)
}

func TestPodLogsAction_Execute_KubernetesError(t *testing.T) {
	logger := logger.NewLogger("debug", "test")
	metrics := metric.NewMetricsCollector(logger)
	mockClient := NewMockKubernetesClientForTest()
	mockClient.SetShouldError(true)

	config := pod_logs_config.PodLogsActionConfig{
		ActionConfig: actions_interfaces.ActionConfig{
			Name:    "test-pod-logs",
			Type:    "pod_logs",
			Enabled: true,
		},
	}

	action := NewPodLogsAction(config, logger, metrics, mockClient)
	workflowEvent := createPodLogsTestWorkflowEvent()

	result, err := action.Execute(context.Background(), workflowEvent)

	require.NoError(t, err)
	// After fix: graceful fallback instead of error - creates enrichment with explanation
	assert.True(t, result.Success)
	require.NoError(t, result.Error)
	assert.NotEmpty(t, result.Enrichments, "Should have enrichment with error explanation")

	// Check result data
	data := result.Data.(map[string]interface{})
	assert.Equal(t, true, data["logs_empty"], "Should indicate logs are empty")
	assert.Contains(t, data["error_message"], "assert.AnError", "Should contain error message")
}

func TestPodLogsAction_GenerateLogFilename(t *testing.T) {
	config := pod_logs_config.PodLogsActionConfig{
		IncludeTimestamp: true,
		IncludeContainer: true,
		TimestampFormat:  "20060102-150405",
		JavaSpecific:     false,
	}

	action := &PodLogsAction{config: config}

	filename := action.generateLogFilename("test-pod", "test-ns", "app-container")

	// Check basic structure (timestamp will vary)
	assert.Contains(t, filename, "pod-logs-test-ns-test-pod-app-container")
	assert.True(t, strings.HasSuffix(filename, ".log"))

	// Test Java container
	config.JavaSpecific = true
	action.config = config
	filename = action.generateLogFilename("java-pod", "java-ns", "java-app")

	assert.Contains(t, filename, "java-logs-java-ns-java-pod-java-app")
	assert.True(t, strings.HasSuffix(filename, ".log"))

	// Test without timestamp
	config.IncludeTimestamp = false
	action.config = config
	filename = action.generateLogFilename("test-pod", "test-ns", "app")
	assert.Equal(t, "java-logs-test-ns-test-pod-app.log", filename)
}

func TestPodLogsActionFactory_Create(t *testing.T) {
	logger := logger.NewLogger("debug", "test")
	metrics := metric.NewMetricsCollector(logger)
	mockClient := NewMockKubernetesClientForTest()

	factory := NewPodLogsActionFactory(logger, metrics, mockClient)

	config := actions_interfaces.ActionConfig{
		Name:    "test-pod-logs",
		Type:    "pod_logs",
		Enabled: true,
		Parameters: map[string]interface{}{
			"max_lines":         500,
			"tail_lines":        50,
			"java_specific":     true,
			"include_timestamp": false,
		},
	}

	action, err := factory.Create(config)

	require.NoError(t, err)
	assert.NotNil(t, action)

	podLogsAction, ok := action.(*PodLogsAction)
	require.True(t, ok)

	assert.Equal(t, "test-pod-logs", podLogsAction.GetName())
	assert.Equal(t, 5000, podLogsAction.config.MaxLines)
	assert.Equal(t, 500, podLogsAction.config.TailLines)
	assert.True(t, podLogsAction.config.JavaSpecific)
	assert.False(t, podLogsAction.config.IncludeTimestamp)
}

func TestPodLogsAction_Execute_MaxLinesLimit(t *testing.T) {
	logger := logger.NewLogger("debug", "test")
	metrics := metric.NewMetricsCollector(logger)
	mockClient := NewMockKubernetesClientForTest()

	config := pod_logs_config.PodLogsActionConfig{
		ActionConfig: actions_interfaces.ActionConfig{
			Name:    "test-pod-logs",
			Type:    "pod_logs",
			Enabled: true,
		},
		MaxLines:         2, // Limit to 2 lines
		TailLines:        100,
		Timestamps:       true,
		IncludeTimestamp: true,
		IncludeContainer: true,
		TimestampFormat:  "20060102-150405",
	}

	action := NewPodLogsAction(config, logger, metrics, mockClient)

	// Set mock logs with more than 2 lines
	mockClient.SetLogs("test-namespace", "test-pod", "Line 1\nLine 2\nLine 3\nLine 4")

	workflowEvent := createPodLogsTestWorkflowEvent()
	result, err := action.Execute(context.Background(), workflowEvent)

	require.NoError(t, err)
	assert.True(t, result.Success)

	// Check that logs were truncated to maxLines
	assert.Equal(t, 2, result.Data.(map[string]interface{})["log_lines"])
}

func TestPodLogsAction_Execute_SinceTimeValidation(t *testing.T) {
	logger := logger.NewLogger("debug", "test")
	metrics := metric.NewMetricsCollector(logger)
	mockClient := NewMockKubernetesClientForTest()

	config := pod_logs_config.PodLogsActionConfig{
		ActionConfig: actions_interfaces.ActionConfig{
			Name:    "test-pod-logs",
			Type:    "pod_logs",
			Enabled: true,
		},
		SinceTime: "invalid-time-format", // Invalid RFC3339 format
		MaxLines:  1000,
		TailLines: 100,
	}

	action := NewPodLogsAction(config, logger, metrics, mockClient)

	workflowEvent := createPodLogsTestWorkflowEvent()
	result, err := action.Execute(context.Background(), workflowEvent)

	require.NoError(t, err)
	// After fix: graceful fallback - creates enrichment with error explanation
	assert.True(t, result.Success)
	assert.NotEmpty(t, result.Enrichments, "Should have enrichment with error explanation")

	// Check that error message mentions invalid time format
	data := result.Data.(map[string]interface{})
	assert.Equal(t, true, data["logs_empty"], "Should indicate logs are empty")
	assert.Contains(t, data["error_message"], "invalid since_time format", "Should contain time format error")
}

func TestPodLogsAction_ExtractPodInfo_AlertManagerEvent(t *testing.T) {
	logger := logger.NewLogger("debug", "test")
	metrics := metric.NewMetricsCollector(logger)
	mockClient := NewMockKubernetesClientForTest()

	config := pod_logs_config.PodLogsActionConfig{
		ActionConfig: actions_interfaces.ActionConfig{
			Name:    "test-pod-logs",
			Type:    "pod_logs",
			Enabled: true,
		},
	}

	action := NewPodLogsAction(config, logger, metrics, mockClient)

	// Create AlertManagerWorkflowEvent with labels
	alertEvent := &event.AlertManagerEvent{
		BaseEvent: event.BaseEvent{
			ID:        uuid.New(),
			Timestamp: time.Now(),
			Source:    "test",
			Type:      event.EventTypeAlertManager,
		},
		Status: "firing",
		Alerts: []event.PrometheusAlert{
			{
				Labels: map[string]string{
					"alertname": "TestAlert",
					"pod":       "test-pod-from-labels",
					"container": "test-container",
					"instance":  "pod-with-port:8080",
					"severity":  "critical",
					"namespace": "alert-namespace",
				},
				Annotations: map[string]string{
					"summary": "Test alert",
				},
			},
		},
	}

	workflowEvent := event.NewAlertManagerWorkflowEvent(alertEvent)

	podName, namespace, containerName := action.extractPodInfo(workflowEvent)

	assert.Equal(t, "test-pod-from-labels", podName)
	assert.Equal(t, "alert-namespace", namespace)
	assert.Equal(t, "test-container", containerName)
}

func TestPodLogsAction_ExtractPodInfo_InstanceLabel(t *testing.T) {
	logger := logger.NewLogger("debug", "test")
	metrics := metric.NewMetricsCollector(logger)
	mockClient := NewMockKubernetesClientForTest()

	config := pod_logs_config.PodLogsActionConfig{
		ActionConfig: actions_interfaces.ActionConfig{
			Name:    "test-pod-logs",
			Type:    "pod_logs",
			Enabled: true,
		},
	}

	action := NewPodLogsAction(config, logger, metrics, mockClient)

	// Create AlertManagerWorkflowEvent with instance label only
	alertEvent := &event.AlertManagerEvent{
		BaseEvent: event.BaseEvent{
			ID:        uuid.New(),
			Timestamp: time.Now(),
			Source:    "test",
			Type:      event.EventTypeAlertManager,
		},
		Status: "firing",
		Alerts: []event.PrometheusAlert{
			{
				Labels: map[string]string{
					"alertname": "TestAlert",
					"instance":  "pod-instance-name:9090",
					"severity":  "warning",
					"namespace": "default",
				},
				Annotations: map[string]string{
					"summary": "Test alert",
				},
			},
		},
	}

	workflowEvent := event.NewAlertManagerWorkflowEvent(alertEvent)

	podName, namespace, containerName := action.extractPodInfo(workflowEvent)

	assert.Equal(t, "pod-instance-name", podName) // Should extract pod name from instance
	assert.Equal(t, "default", namespace)
	assert.Empty(t, containerName)
}

func TestPodLogsAction_ExtractPodInfo_ActionParameters(t *testing.T) {
	logger := logger.NewLogger("debug", "test")
	metrics := metric.NewMetricsCollector(logger)
	mockClient := NewMockKubernetesClientForTest()

	config := pod_logs_config.PodLogsActionConfig{
		ActionConfig: actions_interfaces.ActionConfig{
			Name:    "test-pod-logs",
			Type:    "pod_logs",
			Enabled: true,
			Parameters: map[string]interface{}{
				"pod_name":  "param-pod",
				"container": "param-container",
			},
		},
		Container: "config-container",
	}

	action := NewPodLogsAction(config, logger, metrics, mockClient)

	// Create minimal workflow event
	alertEvent := &event.AlertManagerEvent{
		BaseEvent: event.BaseEvent{
			ID:        uuid.New(),
			Timestamp: time.Now(),
			Source:    "test",
			Type:      event.EventTypeAlertManager,
		},
		Status: "firing",
		Alerts: []event.PrometheusAlert{
			{
				Labels: map[string]string{
					"alertname": "TestAlert",
					"severity":  "info",
					"namespace": "param-namespace",
				},
				Annotations: map[string]string{
					"summary": "Test alert",
				},
			},
		},
	}

	workflowEvent := event.NewAlertManagerWorkflowEvent(alertEvent)

	podName, namespace, containerName := action.extractPodInfo(workflowEvent)

	assert.Equal(t, "param-pod", podName) // Should get from parameters
	assert.Equal(t, "param-namespace", namespace)
	assert.Equal(t, "param-container", containerName) // Should get from parameters
}

func TestPodLogsAction_Validate_Errors(t *testing.T) {
	logger := logger.NewLogger("debug", "test")
	metrics := metric.NewMetricsCollector(logger)

	tests := []struct {
		name       string
		config     pod_logs_config.PodLogsActionConfig
		kubeClient KubernetesClient
		wantErr    string
	}{
		{
			name: "nil kubernetes client",
			config: pod_logs_config.PodLogsActionConfig{
				ActionConfig: actions_interfaces.ActionConfig{
					Name:    "test",
					Type:    "pod_logs",
					Enabled: true,
				},
				MaxLines:  1000,
				TailLines: 100,
			},
			kubeClient: nil,
			wantErr:    "kubernetes client is required",
		},
		{
			name: "negative max lines",
			config: pod_logs_config.PodLogsActionConfig{
				ActionConfig: actions_interfaces.ActionConfig{
					Name:    "test",
					Type:    "pod_logs",
					Enabled: true,
				},
				MaxLines:  -1,
				TailLines: 100,
			},
			kubeClient: NewMockKubernetesClientForTest(),
			wantErr:    "max_lines must be non-negative",
		},
		{
			name: "negative tail lines",
			config: pod_logs_config.PodLogsActionConfig{
				ActionConfig: actions_interfaces.ActionConfig{
					Name:    "test",
					Type:    "pod_logs",
					Enabled: true,
				},
				MaxLines:  1000,
				TailLines: -1,
			},
			kubeClient: NewMockKubernetesClientForTest(),
			wantErr:    "tail_lines must be non-negative",
		},
		{
			name: "invalid base config",
			config: pod_logs_config.PodLogsActionConfig{
				ActionConfig: actions_interfaces.ActionConfig{
					Name:    "", // Empty name should trigger ValidateBasicConfig error
					Type:    "pod_logs",
					Enabled: true,
				},
				MaxLines:  1000,
				TailLines: 100,
			},
			kubeClient: NewMockKubernetesClientForTest(),
			wantErr:    "action name cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := NewPodLogsAction(tt.config, logger, metrics, tt.kubeClient)
			err := action.Validate()

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPodLogsAction_GenerateLogFilename_EdgeCases(t *testing.T) {
	tests := []struct {
		name             string
		config           pod_logs_config.PodLogsActionConfig
		podName          string
		namespace        string
		containerName    string
		expectedContains []string
	}{
		{
			name: "no timestamp, no container",
			config: pod_logs_config.PodLogsActionConfig{
				IncludeTimestamp: false,
				IncludeContainer: false,
				JavaSpecific:     false,
			},
			podName:          "simple-pod",
			namespace:        "simple-ns",
			containerName:    "container",
			expectedContains: []string{"pod-logs-simple-ns-simple-pod.log"},
		},
		{
			name: "empty container name",
			config: pod_logs_config.PodLogsActionConfig{
				IncludeTimestamp: false,
				IncludeContainer: true,
				JavaSpecific:     false,
			},
			podName:          "pod",
			namespace:        "ns",
			containerName:    "",
			expectedContains: []string{"pod-logs-ns-pod.log"}, // No container in name
		},
		{
			name: "java with empty container",
			config: pod_logs_config.PodLogsActionConfig{
				IncludeTimestamp: false,
				IncludeContainer: true,
				JavaSpecific:     true,
			},
			podName:          "java-pod",
			namespace:        "java-ns",
			containerName:    "",
			expectedContains: []string{"java-logs-java-ns-java-pod.log"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := &PodLogsAction{config: tt.config}
			filename := action.generateLogFilename(tt.podName, tt.namespace, tt.containerName)

			for _, expected := range tt.expectedContains {
				assert.Contains(t, filename, expected)
			}
		})
	}
}

func TestPodLogsActionFactory_ValidateConfig(t *testing.T) {
	logger := logger.NewLogger("debug", "test")
	metrics := metric.NewMetricsCollector(logger)
	mockClient := NewMockKubernetesClientForTest()

	factory := NewPodLogsActionFactory(logger, metrics, mockClient)

	tests := []struct {
		name    string
		config  actions_interfaces.ActionConfig
		wantErr string
	}{
		{
			name: "valid config",
			config: actions_interfaces.ActionConfig{
				Name:    "test-pod-logs",
				Type:    "pod_logs",
				Enabled: true,
				Parameters: map[string]interface{}{
					"max_lines":  500,
					"tail_lines": 50,
				},
			},
			wantErr: "",
		},
		{
			name: "invalid action type",
			config: actions_interfaces.ActionConfig{
				Name:    "test-action",
				Type:    "invalid_type",
				Enabled: true,
			},
			wantErr: "invalid action type for PodLogsActionFactory",
		},
		{
			name: "invalid parameter type",
			config: actions_interfaces.ActionConfig{
				Name:    "test-pod-logs",
				Type:    "pod_logs",
				Enabled: true,
				Parameters: map[string]interface{}{
					"max_lines": "not-a-number", // Should be int
				},
			},
			wantErr: "max_lines must be an integer",
		},
		{
			name: "empty action name",
			config: actions_interfaces.ActionConfig{
				Name:    "",
				Type:    "pod_logs",
				Enabled: true,
			},
			wantErr: "action name cannot be empty",
		},
		{
			name: "negative values",
			config: actions_interfaces.ActionConfig{
				Name:    "test-pod-logs",
				Type:    "pod_logs",
				Enabled: true,
				Parameters: map[string]interface{}{
					"max_lines": -100,
				},
			},
			wantErr: "max_lines must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := factory.ValidateConfig(tt.config)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPodLogsActionFactory_Create_Errors(t *testing.T) {
	logger := logger.NewLogger("debug", "test")
	metrics := metric.NewMetricsCollector(logger)
	mockClient := NewMockKubernetesClientForTest()

	factory := NewPodLogsActionFactory(logger, metrics, mockClient)

	tests := []struct {
		name    string
		config  actions_interfaces.ActionConfig
		wantErr string
	}{
		{
			name: "invalid parameter in create",
			config: actions_interfaces.ActionConfig{
				Name:    "test-pod-logs",
				Type:    "pod_logs",
				Enabled: true,
				Parameters: map[string]interface{}{
					"timestamps": "invalid-bool", // Should be bool
				},
			},
			wantErr: "timestamps must be a boolean",
		},
		{
			name: "empty timestamp format",
			config: actions_interfaces.ActionConfig{
				Name:    "test-pod-logs",
				Type:    "pod_logs",
				Enabled: true,
				Parameters: map[string]interface{}{
					"timestamp_format": "",
				},
			},
			wantErr: "timestamp_format cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action, err := factory.Create(tt.config)

			require.Error(t, err)
			assert.Nil(t, action)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestPodLogsActionFactory_GetActionType(t *testing.T) {
	logger := logger.NewLogger("debug", "test")
	metrics := metric.NewMetricsCollector(logger)
	mockClient := NewMockKubernetesClientForTest()

	factory := NewPodLogsActionFactory(logger, metrics, mockClient)

	assert.Equal(t, "pod_logs", factory.GetActionType())
}

func TestPodLogsAction_GetActionType(t *testing.T) {
	logger := logger.NewLogger("debug", "test")
	metrics := metric.NewMetricsCollector(logger)
	mockClient := NewMockKubernetesClientForTest()

	config := pod_logs_config.PodLogsActionConfig{
		ActionConfig: actions_interfaces.ActionConfig{
			Name:    "test-pod-logs",
			Type:    "pod_logs",
			Enabled: true,
		},
	}

	action := NewPodLogsAction(config, logger, metrics, mockClient)

	assert.Equal(t, "pod_logs", action.GetActionType())
}

func TestPodLogsAction_GetPodLogs_ValidSinceTime(t *testing.T) {
	logger := logger.NewLogger("debug", "test")
	metrics := metric.NewMetricsCollector(logger)
	mockClient := NewMockKubernetesClientForTest()

	config := pod_logs_config.PodLogsActionConfig{
		ActionConfig: actions_interfaces.ActionConfig{
			Name:    "test-pod-logs",
			Type:    "pod_logs",
			Enabled: true,
		},
		SinceTime: "2025-01-08T10:00:00Z", // Valid RFC3339 format
		MaxLines:  1000,
		TailLines: 100,
	}

	action := NewPodLogsAction(config, logger, metrics, mockClient)

	// Set mock logs
	mockClient.SetLogs("test-namespace", "test-pod", "Valid time test logs")

	workflowEvent := createPodLogsTestWorkflowEvent()
	result, err := action.Execute(context.Background(), workflowEvent)

	require.NoError(t, err)
	assert.True(t, result.Success) // Should succeed with valid time format
}

func TestPodLogsAction_ExtractPodInfo_FromAlertName(t *testing.T) {
	logger := logger.NewLogger("debug", "test")
	metrics := metric.NewMetricsCollector(logger)
	mockClient := NewMockKubernetesClientForTest()

	config := pod_logs_config.PodLogsActionConfig{
		ActionConfig: actions_interfaces.ActionConfig{
			Name:    "test-pod-logs",
			Type:    "pod_logs",
			Enabled: true,
		},
	}

	action := NewPodLogsAction(config, logger, metrics, mockClient)

	// Create AlertManagerWorkflowEvent with Pod in alert name
	alertEvent := &event.AlertManagerEvent{
		BaseEvent: event.BaseEvent{
			ID:        uuid.New(),
			Timestamp: time.Now(),
			Source:    "test",
			Type:      event.EventTypeAlertManager,
		},
		Status: "firing",
		Alerts: []event.PrometheusAlert{
			{
				Labels: map[string]string{
					"alertname": "PodCrashLooping", // Contains "Pod"
					"namespace": "test-namespace",
				},
				Annotations: map[string]string{
					"summary": "Pod is crash looping",
				},
			},
		},
	}

	workflowEvent := event.NewAlertManagerWorkflowEvent(alertEvent)

	podName, namespace, containerName := action.extractPodInfo(workflowEvent)

	// Should extract from action parameters when alert name contains "Pod"
	assert.Empty(t, podName) // No pod_name parameter set
	assert.Equal(t, "test-namespace", namespace)
	assert.Empty(t, containerName)
}
