package actions

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kubecano/cano-collector/pkg/logger"
)

func TestNewMockKubernetesClient(t *testing.T) {
	logger := logger.NewLogger("debug", "test")
	client := NewMockKubernetesClient(logger)

	assert.NotNil(t, client)
	assert.NotNil(t, client.logger)
}

func TestMockKubernetesClient_GetPodLogs(t *testing.T) {
	logger := logger.NewLogger("debug", "test")
	client := NewMockKubernetesClient(logger)

	ctx := context.Background()
	options := map[string]interface{}{
		"container": "test-container",
		"tailLines": 100,
	}

	logs, err := client.GetPodLogs(ctx, "test-ns", "test-pod", options)

	require.NoError(t, err)
	assert.Contains(t, logs, "[Mock Logs for test-ns/test-pod]")
	assert.Contains(t, logs, "Pod: test-pod, Namespace: test-ns")
	assert.Contains(t, logs, "Starting application...")
	assert.Contains(t, logs, "ERROR Failed to process request: timeout")
}

func TestNewPlaceholderKubernetesClient(t *testing.T) {
	logger := logger.NewLogger("debug", "test")
	client := NewPlaceholderKubernetesClient(logger)

	assert.NotNil(t, client)
	assert.NotNil(t, client.logger)
}

func TestPlaceholderKubernetesClient_GetPodLogs(t *testing.T) {
	logger := logger.NewLogger("debug", "test")
	client := NewPlaceholderKubernetesClient(logger)

	ctx := context.Background()
	options := map[string]interface{}{
		"container": "test-container",
	}

	logs, err := client.GetPodLogs(ctx, "test-ns", "test-pod", options)

	require.Error(t, err)
	assert.Empty(t, logs)
	assert.Contains(t, err.Error(), "kubernetes client not implemented")
	assert.Contains(t, err.Error(), "test-ns/test-pod")
}

// Note: RealKubernetesClient tests would require a test Kubernetes cluster
// or extensive mocking of kubernetes.Interface, which is beyond current scope.
// In production, integration tests would cover this functionality.
func TestNewRealKubernetesClient_OutOfCluster(t *testing.T) {
	logger := logger.NewLogger("debug", "test")

	// This test runs outside of Kubernetes cluster, so should fail gracefully
	client, err := NewRealKubernetesClient(logger)

	// Outside cluster - should return error about missing service account
	require.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "failed to create in-cluster config")
}
