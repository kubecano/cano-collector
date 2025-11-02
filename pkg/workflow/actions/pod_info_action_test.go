package actions

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	pod_info_config "github.com/kubecano/cano-collector/config/workflow/actions"
	"github.com/kubecano/cano-collector/pkg/core/event"
	"github.com/kubecano/cano-collector/pkg/core/issue"
	"github.com/kubecano/cano-collector/pkg/logger"
	"github.com/kubecano/cano-collector/pkg/metric"
	actions_interfaces "github.com/kubecano/cano-collector/pkg/workflow/actions/interfaces"
)

// Helper to create test workflow event with pod info
func createPodInfoTestWorkflowEvent() event.WorkflowEvent {
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
					"alertname": "KubePodCrashLooping",
					"pod":       "crash-test-pod",
					"namespace": "test-namespace",
				},
				Annotations: map[string]string{
					"summary": "Pod is crash looping",
				},
			},
		},
	}

	return event.NewAlertManagerWorkflowEvent(alertEvent)
}

// Helper to create mock pod with crash info
func createMockCrashingPod(namespace, name string) *corev1.Pod {
	now := metav1.Now()
	fiveMinAgo := metav1.NewTime(now.Add(-5 * time.Minute))
	tenMinAgo := metav1.NewTime(now.Add(-10 * time.Minute))

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:         "main-container",
					RestartCount: 3,
					State: corev1.ContainerState{
						Waiting: &corev1.ContainerStateWaiting{
							Reason:  "CrashLoopBackOff",
							Message: "Back-off 5m0s restarting failed container",
						},
					},
					LastTerminationState: corev1.ContainerState{
						Terminated: &corev1.ContainerStateTerminated{
							Reason:     "Error",
							ExitCode:   1,
							StartedAt:  tenMinAgo,
							FinishedAt: fiveMinAgo,
						},
					},
				},
			},
		},
	}
}

func TestPodInfoAction_Execute_Success(t *testing.T) {
	mockClient := NewMockKubernetesClientForTest()
	mockClient.SetShouldError(false)

	// Override GetPod to return crashing pod
	crashingPod := createMockCrashingPod("test-namespace", "crash-test-pod")

	config := pod_info_config.PodInfoActionConfig{
		ActionConfig: actions_interfaces.ActionConfig{
			Name: "test-pod-info",
			Type: "pod_info",
		},
		IncludePreviousState:  true,
		IncludeInitContainers: false,
		MinRestartCount:       0,
	}

	testLogger := logger.NewLogger("debug", "test")
	testMetrics := metric.NewMetricsCollector(testLogger)

	action := NewPodInfoAction(config, testLogger, testMetrics, mockClient)

	// Create custom mock that returns our crashing pod
	customMock := &customPodMock{
		pod: crashingPod,
		err: nil,
	}
	action.kubeClient = customMock

	workflowEvent := createPodInfoTestWorkflowEvent()
	result, err := action.Execute(context.Background(), workflowEvent)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Success)

	data, ok := result.Data.(map[string]interface{})
	require.True(t, ok, "result.Data should be map[string]interface{}")
	assert.Equal(t, "crash-test-pod", data["pod_name"])
	assert.Equal(t, "test-namespace", data["namespace"])
	assert.Equal(t, 1, data["crash_count"])

	assert.Len(t, result.Enrichments, 1)
	assert.Equal(t, issue.EnrichmentTypeCrashInfo, result.Enrichments[0].Type)
}

func TestPodInfoAction_Execute_PodNotFound(t *testing.T) {
	mockClient := NewMockKubernetesClientForTest()
	mockClient.SetShouldError(true)

	config := pod_info_config.PodInfoActionConfig{
		ActionConfig: actions_interfaces.ActionConfig{
			Name: "test-pod-info",
			Type: "pod_info",
		},
		IncludePreviousState: true,
	}

	testLogger := logger.NewLogger("debug", "test")
	testMetrics := metric.NewMetricsCollector(testLogger)

	action := NewPodInfoAction(config, testLogger, testMetrics, mockClient)
	workflowEvent := createPodInfoTestWorkflowEvent()

	result, err := action.Execute(context.Background(), workflowEvent)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Success) // Should still succeed with error enrichment
	assert.Len(t, result.Enrichments, 1)
}

func TestPodInfoAction_Execute_NoPodInfo(t *testing.T) {
	mockClient := NewMockKubernetesClientForTest()

	config := pod_info_config.PodInfoActionConfig{
		ActionConfig: actions_interfaces.ActionConfig{
			Name: "test-pod-info",
			Type: "pod_info",
		},
	}

	testLogger := logger.NewLogger("debug", "test")
	testMetrics := metric.NewMetricsCollector(testLogger)

	action := NewPodInfoAction(config, testLogger, testMetrics, mockClient)

	// Create event without pod info
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
				},
			},
		},
	}
	workflowEvent := event.NewAlertManagerWorkflowEvent(alertEvent)

	result, err := action.Execute(context.Background(), workflowEvent)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Success)
}

func TestPodInfoAction_ExtractCrashInfo_WithRestarts(t *testing.T) {
	config := pod_info_config.PodInfoActionConfig{
		ActionConfig: actions_interfaces.ActionConfig{
			Name: "test",
			Type: "pod_info",
		},
		IncludePreviousState: true,
		MinRestartCount:      0,
	}

	testLogger := logger.NewLogger("debug", "test")
	testMetrics := metric.NewMetricsCollector(testLogger)
	mockClient := NewMockKubernetesClientForTest()

	action := NewPodInfoAction(config, testLogger, testMetrics, mockClient)
	pod := createMockCrashingPod("test-namespace", "test-pod")

	crashInfos := action.extractCrashInfo(pod)

	require.Len(t, crashInfos, 1)
	assert.Equal(t, "main-container", crashInfos[0].Container)
	assert.Equal(t, int32(3), crashInfos[0].RestartCount)
	assert.Equal(t, "WAITING", crashInfos[0].Status)
	assert.Equal(t, "CrashLoopBackOff", crashInfos[0].Reason)
	require.NotNil(t, crashInfos[0].LastStateInfo)
	assert.Equal(t, "Error", crashInfos[0].LastStateInfo.Reason)
	assert.Equal(t, int32(1), crashInfos[0].LastStateInfo.ExitCode)
}

func TestPodInfoAction_ExtractCrashInfo_WithMinRestartCount(t *testing.T) {
	config := pod_info_config.PodInfoActionConfig{
		ActionConfig: actions_interfaces.ActionConfig{
			Name: "test",
			Type: "pod_info",
		},
		MinRestartCount: 5, // Require at least 5 restarts
	}

	testLogger := logger.NewLogger("debug", "test")
	testMetrics := metric.NewMetricsCollector(testLogger)
	mockClient := NewMockKubernetesClientForTest()

	action := NewPodInfoAction(config, testLogger, testMetrics, mockClient)
	pod := createMockCrashingPod("test-namespace", "test-pod")

	crashInfos := action.extractCrashInfo(pod)

	// Should filter out container with only 3 restarts
	assert.Empty(t, crashInfos)
}

func TestPodInfoAction_ExtractCrashInfo_WithInitContainers(t *testing.T) {
	config := pod_info_config.PodInfoActionConfig{
		ActionConfig: actions_interfaces.ActionConfig{
			Name: "test",
			Type: "pod_info",
		},
		IncludeInitContainers: true,
		MinRestartCount:       0,
	}

	testLogger := logger.NewLogger("debug", "test")
	testMetrics := metric.NewMetricsCollector(testLogger)
	mockClient := NewMockKubernetesClientForTest()

	action := NewPodInfoAction(config, testLogger, testMetrics, mockClient)

	pod := &corev1.Pod{
		Status: corev1.PodStatus{
			InitContainerStatuses: []corev1.ContainerStatus{
				{
					Name:         "init-container",
					RestartCount: 2,
					State: corev1.ContainerState{
						Waiting: &corev1.ContainerStateWaiting{
							Reason: "CrashLoopBackOff",
						},
					},
				},
			},
		},
	}

	crashInfos := action.extractCrashInfo(pod)

	require.Len(t, crashInfos, 1)
	assert.Equal(t, "init-container", crashInfos[0].Container)
	assert.Equal(t, int32(2), crashInfos[0].RestartCount)
}

func TestPodInfoAction_ExtractCrashInfo_RunningWithRestarts(t *testing.T) {
	config := pod_info_config.PodInfoActionConfig{
		ActionConfig: actions_interfaces.ActionConfig{
			Name: "test",
			Type: "pod_info",
		},
		MinRestartCount: 0,
	}

	testLogger := logger.NewLogger("debug", "test")
	testMetrics := metric.NewMetricsCollector(testLogger)
	mockClient := NewMockKubernetesClientForTest()

	action := NewPodInfoAction(config, testLogger, testMetrics, mockClient)

	pod := &corev1.Pod{
		Status: corev1.PodStatus{
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:         "container",
					RestartCount: 5,
					State: corev1.ContainerState{
						Running: &corev1.ContainerStateRunning{},
					},
				},
			},
		},
	}

	crashInfos := action.extractCrashInfo(pod)

	require.Len(t, crashInfos, 1)
	assert.Equal(t, "RUNNING", crashInfos[0].Status)
	assert.Equal(t, "Restarted", crashInfos[0].Reason)
	assert.Equal(t, int32(5), crashInfos[0].RestartCount)
}

func TestPodInfoAction_CreateCrashInfoEnrichment(t *testing.T) {
	config := pod_info_config.PodInfoActionConfig{
		ActionConfig: actions_interfaces.ActionConfig{
			Name: "test",
			Type: "pod_info",
		},
		IncludePreviousState: true,
	}

	testLogger := logger.NewLogger("debug", "test")
	testMetrics := metric.NewMetricsCollector(testLogger)
	mockClient := NewMockKubernetesClientForTest()

	action := NewPodInfoAction(config, testLogger, testMetrics, mockClient)

	now := time.Now()
	crashInfos := []ContainerCrashInfo{
		{
			Container:    "test-container",
			RestartCount: 3,
			Status:       "WAITING",
			Reason:       "CrashLoopBackOff",
			LastStateInfo: &PreviousContainerInfo{
				Reason:     "Error",
				ExitCode:   1,
				StartedAt:  now.Add(-10 * time.Minute),
				FinishedAt: now.Add(-5 * time.Minute),
			},
		},
	}

	enrichment := action.createCrashInfoEnrichment(crashInfos, "")

	require.NotNil(t, enrichment)
	assert.Equal(t, issue.EnrichmentTypeCrashInfo, enrichment.Type)
	assert.Equal(t, "Container Crash Information", enrichment.Title)
	assert.NotEmpty(t, enrichment.Blocks)
}

func TestPodInfoAction_CreateCrashInfoEnrichment_WithError(t *testing.T) {
	config := pod_info_config.PodInfoActionConfig{
		ActionConfig: actions_interfaces.ActionConfig{
			Name: "test",
			Type: "pod_info",
		},
	}

	testLogger := logger.NewLogger("debug", "test")
	testMetrics := metric.NewMetricsCollector(testLogger)
	mockClient := NewMockKubernetesClientForTest()

	action := NewPodInfoAction(config, testLogger, testMetrics, mockClient)

	enrichment := action.createCrashInfoEnrichment([]ContainerCrashInfo{}, "Error message")

	require.NotNil(t, enrichment)
	assert.Len(t, enrichment.Blocks, 1) // Should have markdown block with error
}

func TestPodInfoAction_Validate_Success(t *testing.T) {
	config := pod_info_config.PodInfoActionConfig{
		ActionConfig: actions_interfaces.ActionConfig{
			Name: "test",
			Type: "pod_info",
		},
	}

	testLogger := logger.NewLogger("debug", "test")
	testMetrics := metric.NewMetricsCollector(testLogger)
	mockClient := NewMockKubernetesClientForTest()

	action := NewPodInfoAction(config, testLogger, testMetrics, mockClient)

	err := action.Validate()
	assert.NoError(t, err)
}

func TestPodInfoAction_Validate_NoKubernetesClient(t *testing.T) {
	config := pod_info_config.PodInfoActionConfig{
		ActionConfig: actions_interfaces.ActionConfig{
			Name: "test",
			Type: "pod_info",
		},
	}

	testLogger := logger.NewLogger("debug", "test")
	testMetrics := metric.NewMetricsCollector(testLogger)

	action := NewPodInfoAction(config, testLogger, testMetrics, nil)

	err := action.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "kubernetes client is required")
}

func TestPodInfoAction_GetActionType(t *testing.T) {
	config := pod_info_config.PodInfoActionConfig{
		ActionConfig: actions_interfaces.ActionConfig{
			Name: "test",
			Type: "pod_info",
		},
	}

	testLogger := logger.NewLogger("debug", "test")
	testMetrics := metric.NewMetricsCollector(testLogger)
	mockClient := NewMockKubernetesClientForTest()

	action := NewPodInfoAction(config, testLogger, testMetrics, mockClient)

	assert.Equal(t, "pod_info", action.GetActionType())
}

func TestPodInfoAction_ExtractPodInfo_FromPodLabel(t *testing.T) {
	config := pod_info_config.PodInfoActionConfig{
		ActionConfig: actions_interfaces.ActionConfig{
			Name: "test",
			Type: "pod_info",
		},
	}

	testLogger := logger.NewLogger("debug", "test")
	testMetrics := metric.NewMetricsCollector(testLogger)
	mockClient := NewMockKubernetesClientForTest()

	action := NewPodInfoAction(config, testLogger, testMetrics, mockClient)

	workflowEvent := createPodInfoTestWorkflowEvent()
	podName, namespace := action.extractPodInfo(workflowEvent)

	assert.Equal(t, "crash-test-pod", podName)
	assert.Equal(t, "test-namespace", namespace)
}

func TestPodInfoAction_ExtractPodInfo_FromInstanceLabel(t *testing.T) {
	config := pod_info_config.PodInfoActionConfig{
		ActionConfig: actions_interfaces.ActionConfig{
			Name: "test",
			Type: "pod_info",
		},
	}

	testLogger := logger.NewLogger("debug", "test")
	testMetrics := metric.NewMetricsCollector(testLogger)
	mockClient := NewMockKubernetesClientForTest()

	action := NewPodInfoAction(config, testLogger, testMetrics, mockClient)

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
					"instance":  "test-pod-123:8080",
					"namespace": "default",
				},
			},
		},
	}
	workflowEvent := event.NewAlertManagerWorkflowEvent(alertEvent)

	podName, namespace := action.extractPodInfo(workflowEvent)

	assert.Equal(t, "test-pod-123", podName)
	assert.Equal(t, "default", namespace)
}

func TestPodInfoAction_ExtractPodInfo_FromConfig(t *testing.T) {
	config := pod_info_config.PodInfoActionConfig{
		ActionConfig: actions_interfaces.ActionConfig{
			Name: "test",
			Type: "pod_info",
		},
		PodName: "config-pod",
	}

	testLogger := logger.NewLogger("debug", "test")
	testMetrics := metric.NewMetricsCollector(testLogger)
	mockClient := NewMockKubernetesClientForTest()

	action := NewPodInfoAction(config, testLogger, testMetrics, mockClient)

	workflowEvent := createPodInfoTestWorkflowEvent()
	podName, namespace := action.extractPodInfo(workflowEvent)

	assert.Equal(t, "config-pod", podName)
	assert.Equal(t, "test-namespace", namespace)
}

// Custom mock for testing with specific pod
type customPodMock struct {
	pod *corev1.Pod
	err error
}

func (m *customPodMock) GetPodLogs(ctx context.Context, namespace, podName string, options map[string]interface{}) (string, error) {
	return "", nil
}

func (m *customPodMock) GetPod(ctx context.Context, namespace, podName string) (*corev1.Pod, error) {
	return m.pod, m.err
}
