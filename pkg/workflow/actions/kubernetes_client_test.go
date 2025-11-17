package actions

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kubecano/cano-collector/mocks"
	"github.com/kubecano/cano-collector/pkg/logger"
)

func TestMockKubernetesClient_GetPodLogs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockKubernetesClient(ctrl)
	ctx := context.Background()
	options := map[string]interface{}{
		"container": "test-container",
		"tailLines": 100,
	}

	expectedLogs := `[Mock Logs for test-ns/test-pod]
2025-01-08T10:00:00Z INFO  Starting application...
2025-01-08T10:00:01Z INFO  Connected to database
2025-01-08T10:00:02Z WARN  High memory usage detected
2025-01-08T10:00:03Z ERROR Failed to process request: timeout
2025-01-08T10:00:04Z INFO  Application running normally`

	// Set expectation
	mockClient.EXPECT().
		GetPodLogs(ctx, "test-ns", "test-pod", options).
		Return(expectedLogs, nil)

	// Execute
	logs, err := mockClient.GetPodLogs(ctx, "test-ns", "test-pod", options)

	// Verify
	require.NoError(t, err)
	assert.Equal(t, expectedLogs, logs)
	assert.Contains(t, logs, "[Mock Logs for test-ns/test-pod]")
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
