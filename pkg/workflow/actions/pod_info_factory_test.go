package actions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kubecano/cano-collector/pkg/logger"
	"github.com/kubecano/cano-collector/pkg/metric"
	actions_interfaces "github.com/kubecano/cano-collector/pkg/workflow/actions/interfaces"
)

func TestNewPodInfoActionFactory(t *testing.T) {
	testLogger := logger.NewLogger("debug", "test")
	testMetrics := metric.NewMetricsCollector(testLogger)
	mockClient := NewMockKubernetesClientForTest()

	factory := NewPodInfoActionFactory(testLogger, testMetrics, mockClient)

	assert.NotNil(t, factory)
	assert.Equal(t, testLogger, factory.logger)
	assert.Equal(t, testMetrics, factory.metrics)
	assert.Equal(t, mockClient, factory.kubeClient)
}

func TestPodInfoActionFactory_GetActionType(t *testing.T) {
	testLogger := logger.NewLogger("debug", "test")
	testMetrics := metric.NewMetricsCollector(testLogger)
	mockClient := NewMockKubernetesClientForTest()

	factory := NewPodInfoActionFactory(testLogger, testMetrics, mockClient)

	assert.Equal(t, "pod_info", factory.GetActionType())
}

func TestPodInfoActionFactory_Create_Success(t *testing.T) {
	testLogger := logger.NewLogger("debug", "test")
	testMetrics := metric.NewMetricsCollector(testLogger)
	mockClient := NewMockKubernetesClientForTest()

	factory := NewPodInfoActionFactory(testLogger, testMetrics, mockClient)

	config := actions_interfaces.ActionConfig{
		Name: "test-pod-info",
		Type: "pod_info",
		Parameters: map[string]interface{}{
			"min_restart_count": 1,
		},
	}

	action, err := factory.Create(config)

	require.NoError(t, err)
	require.NotNil(t, action)

	// Verify it's a PodInfoAction
	podInfoAction, ok := action.(*PodInfoAction)
	require.True(t, ok)
	assert.Equal(t, "pod_info", podInfoAction.GetActionType())
}

func TestPodInfoActionFactory_Create_WithInvalidType(t *testing.T) {
	testLogger := logger.NewLogger("debug", "test")
	testMetrics := metric.NewMetricsCollector(testLogger)
	mockClient := NewMockKubernetesClientForTest()

	factory := NewPodInfoActionFactory(testLogger, testMetrics, mockClient)

	config := actions_interfaces.ActionConfig{
		Name: "test-action",
		Type: "invalid_type",
	}

	action, err := factory.Create(config)

	require.Error(t, err)
	assert.Nil(t, action)
	assert.Contains(t, err.Error(), "invalid action type for PodInfoActionFactory")
}

func TestPodInfoActionFactory_Create_WithEmptyConfig(t *testing.T) {
	testLogger := logger.NewLogger("debug", "test")
	testMetrics := metric.NewMetricsCollector(testLogger)
	mockClient := NewMockKubernetesClientForTest()

	factory := NewPodInfoActionFactory(testLogger, testMetrics, mockClient)

	// Empty config should use defaults
	config := actions_interfaces.ActionConfig{
		Name: "test-pod-info",
		Type: "pod_info",
	}

	action, err := factory.Create(config)

	require.NoError(t, err)
	require.NotNil(t, action)

	// Verify it's a PodInfoAction with default values
	podInfoAction, ok := action.(*PodInfoAction)
	require.True(t, ok)
	assert.Equal(t, int32(0), podInfoAction.config.MinRestartCount) // Default is 0
	assert.False(t, podInfoAction.config.IncludeInitContainers)     // Default is false
}

func TestPodInfoActionFactory_Create_WithAllParameters(t *testing.T) {
	testLogger := logger.NewLogger("debug", "test")
	testMetrics := metric.NewMetricsCollector(testLogger)
	mockClient := NewMockKubernetesClientForTest()

	factory := NewPodInfoActionFactory(testLogger, testMetrics, mockClient)

	config := actions_interfaces.ActionConfig{
		Name: "test-pod-info",
		Type: "pod_info",
		Parameters: map[string]interface{}{
			"min_restart_count":       3,
			"include_init_containers": true,
			"container":               "app-container",
		},
	}

	action, err := factory.Create(config)

	require.NoError(t, err)
	require.NotNil(t, action)

	// Verify it's a PodInfoAction
	podInfoAction, ok := action.(*PodInfoAction)
	require.True(t, ok)
	assert.Equal(t, int32(3), podInfoAction.config.MinRestartCount)
	assert.True(t, podInfoAction.config.IncludeInitContainers)
	assert.Equal(t, "app-container", podInfoAction.config.Container)
}
