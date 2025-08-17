package actions

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	logger_interfaces "github.com/kubecano/cano-collector/pkg/logger/interfaces"
)

// MockKubernetesClient implements KubernetesClient for testing and development
// TODO: Replace with real kubernetes client-go implementation when dependencies are added
type MockKubernetesClient struct {
	logger logger_interfaces.LoggerInterface
}

// NewMockKubernetesClient creates a new mock Kubernetes client
func NewMockKubernetesClient(logger logger_interfaces.LoggerInterface) *MockKubernetesClient {
	return &MockKubernetesClient{
		logger: logger,
	}
}

// GetPodLogs returns mock pod logs for testing
func (m *MockKubernetesClient) GetPodLogs(ctx context.Context, namespace, podName string, options map[string]interface{}) (string, error) {
	m.logger.Info("Mock: Fetching pod logs",
		zap.String("namespace", namespace),
		zap.String("pod_name", podName),
		zap.Any("options", options),
	)

	// TODO: Replace with real implementation using kubernetes client-go
	// For now, return mock logs for development/testing
	mockLogs := fmt.Sprintf(`[Mock Logs for %s/%s]
2025-01-08T10:00:00Z INFO  Starting application...
2025-01-08T10:00:01Z INFO  Connected to database
2025-01-08T10:00:02Z WARN  High memory usage detected
2025-01-08T10:00:03Z ERROR Failed to process request: timeout
2025-01-08T10:00:04Z INFO  Application running normally

Note: These are mock logs. Real implementation requires kubernetes client-go dependencies.
Pod: %s, Namespace: %s
Options: %v`, namespace, podName, podName, namespace, options)

	return mockLogs, nil
}

// PlaceholderKubernetesClient implements KubernetesClient returning errors for production
// This signals that real implementation is needed
type PlaceholderKubernetesClient struct {
	logger logger_interfaces.LoggerInterface
}

// NewPlaceholderKubernetesClient creates a placeholder client that returns errors
func NewPlaceholderKubernetesClient(logger logger_interfaces.LoggerInterface) *PlaceholderKubernetesClient {
	return &PlaceholderKubernetesClient{
		logger: logger,
	}
}

// GetPodLogs returns an error indicating real implementation is needed
func (p *PlaceholderKubernetesClient) GetPodLogs(ctx context.Context, namespace, podName string, options map[string]interface{}) (string, error) {
	p.logger.Error("Placeholder KubernetesClient used - real implementation required",
		zap.String("namespace", namespace),
		zap.String("pod_name", podName),
	)

	return "", fmt.Errorf("kubernetes client not implemented - requires kubernetes client-go dependencies. Pod: %s/%s", namespace, podName)
}
