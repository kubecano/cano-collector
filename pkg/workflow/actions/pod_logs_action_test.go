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
	assert.Equal(t, issue.EnrichmentTypeTextFile, *enrichment.EnrichmentType)
	assert.Contains(t, *enrichment.Title, "Pod Logs: test-namespace/test-pod")
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
	assert.Contains(t, *enrichment.Title, "Java Pod Logs")
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
	assert.False(t, result.Success)
	assert.Error(t, result.Error)
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

func TestPodLogsActionFactory_ValidateConfig(t *testing.T) {
	logger := logger.NewLogger("debug", "test")
	metrics := metric.NewMetricsCollector(logger)
	mockClient := NewMockKubernetesClientForTest()

	factory := NewPodLogsActionFactory(logger, metrics, mockClient)

	tests := []struct {
		name        string
		config      actions_interfaces.ActionConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid config",
			config: actions_interfaces.ActionConfig{
				Name: "test",
				Type: "pod_logs",
			},
			expectError: false,
		},
		{
			name: "Invalid action type",
			config: actions_interfaces.ActionConfig{
				Name: "test",
				Type: "invalid",
			},
			expectError: true,
			errorMsg:    "invalid action type",
		},
		{
			name: "Empty name",
			config: actions_interfaces.ActionConfig{
				Name: "",
				Type: "pod_logs",
			},
			expectError: true,
			errorMsg:    "action name cannot be empty",
		},
		{
			name: "Invalid max_lines",
			config: actions_interfaces.ActionConfig{
				Name: "test",
				Type: "pod_logs",
				Parameters: map[string]interface{}{
					"max_lines": -1,
				},
			},
			expectError: true,
			errorMsg:    "max_lines must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := factory.ValidateConfig(tt.config)
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPodLogsAction_Validate(t *testing.T) {
	logger := logger.NewLogger("debug", "test")
	metrics := metric.NewMetricsCollector(logger)

	tests := []struct {
		name        string
		kubeClient  KubernetesClient
		config      pod_logs_config.PodLogsActionConfig
		expectError bool
		errorMsg    string
	}{
		{
			name:       "Valid action",
			kubeClient: NewMockKubernetesClientForTest(),
			config: pod_logs_config.PodLogsActionConfig{
				ActionConfig: actions_interfaces.ActionConfig{
					Name: "test",
					Type: "pod_logs",
				},
				MaxLines:  1000,
				TailLines: 100,
			},
			expectError: false,
		},
		{
			name:       "Nil Kubernetes client",
			kubeClient: nil,
			config: pod_logs_config.PodLogsActionConfig{
				ActionConfig: actions_interfaces.ActionConfig{
					Name: "test",
					Type: "pod_logs",
				},
			},
			expectError: true,
			errorMsg:    "kubernetes client is required",
		},
		{
			name:       "Negative max_lines",
			kubeClient: NewMockKubernetesClientForTest(),
			config: pod_logs_config.PodLogsActionConfig{
				ActionConfig: actions_interfaces.ActionConfig{
					Name: "test",
					Type: "pod_logs",
				},
				MaxLines: -1,
			},
			expectError: true,
			errorMsg:    "max_lines must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action := NewPodLogsAction(tt.config, logger, metrics, tt.kubeClient)
			err := action.Validate()
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
