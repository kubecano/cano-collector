package actions

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	logger_interfaces "github.com/kubecano/cano-collector/pkg/logger/interfaces"
)

// RealKubernetesClient implements KubernetesClient using kubernetes client-go
type RealKubernetesClient struct {
	clientset kubernetes.Interface
	logger    logger_interfaces.LoggerInterface
}

// NewRealKubernetesClient creates a new real Kubernetes client using in-cluster config
func NewRealKubernetesClient(logger logger_interfaces.LoggerInterface) (*RealKubernetesClient, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create in-cluster config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}

	return &RealKubernetesClient{
		clientset: clientset,
		logger:    logger,
	}, nil
}

// GetPodLogs retrieves pod logs using kubernetes client-go (inspired by Robusta implementation)
func (r *RealKubernetesClient) GetPodLogs(ctx context.Context, namespace, podName string, options map[string]interface{}) (string, error) {
	r.logger.Info("Fetching pod logs",
		zap.String("namespace", namespace),
		zap.String("pod_name", podName),
		zap.Any("options", options),
	)

	// Build log options from the provided map
	podLogOpts := &corev1.PodLogOptions{}

	// Parse options similar to Robusta's approach
	if container, ok := options["container"].(string); ok && container != "" {
		podLogOpts.Container = container
	}

	if previous, ok := options["previous"].(bool); ok {
		podLogOpts.Previous = previous
	}

	if timestamps, ok := options["timestamps"].(bool); ok {
		podLogOpts.Timestamps = timestamps
	}

	if tailLines, ok := options["tailLines"].(int); ok && tailLines > 0 {
		tailLines64 := int64(tailLines)
		podLogOpts.TailLines = &tailLines64
	}

	if sinceTime, ok := options["sinceTime"].(string); ok && sinceTime != "" {
		if parsedTime, err := time.Parse(time.RFC3339, sinceTime); err == nil {
			metaTime := metav1.NewTime(parsedTime)
			podLogOpts.SinceTime = &metaTime
		}
	}

	if sinceSeconds, ok := options["sinceSeconds"].(int); ok && sinceSeconds > 0 {
		sinceSeconds64 := int64(sinceSeconds)
		podLogOpts.SinceSeconds = &sinceSeconds64
	}

	// Request logs from Kubernetes API
	req := r.clientset.CoreV1().Pods(namespace).GetLogs(podName, podLogOpts)

	podLogs, err := req.Stream(ctx)
	if err != nil {
		// Handle 404 errors gracefully like Robusta does
		if errors.IsNotFound(err) {
			r.logger.Warn("Pod not found",
				zap.String("namespace", namespace),
				zap.String("pod_name", podName),
			)
			return "", fmt.Errorf("pod %s/%s not found", namespace, podName)
		}
		r.logger.Error("Failed to get pod logs stream",
			zap.Error(err),
			zap.String("namespace", namespace),
			zap.String("pod_name", podName),
		)
		return "", fmt.Errorf("failed to get pod logs: %w", err)
	}
	defer podLogs.Close()

	// Read logs from stream (inspired by Robusta's approach)
	buf, err := io.ReadAll(podLogs)
	if err != nil {
		r.logger.Error("Failed to read pod logs",
			zap.Error(err),
			zap.String("namespace", namespace),
			zap.String("pod_name", podName),
		)
		return "", fmt.Errorf("failed to read pod logs: %w", err)
	}

	logs := string(buf)

	// Check if logs are empty - this can happen when:
	// 1. Pod hasn't started yet
	// 2. previous=true but pod hasn't restarted
	// 3. Container has no output
	if len(logs) == 0 {
		r.logger.Warn("Retrieved empty pod logs",
			zap.String("namespace", namespace),
			zap.String("pod_name", podName),
			zap.Bool("previous", podLogOpts.Previous),
			zap.String("container", podLogOpts.Container),
		)
		return "", fmt.Errorf("pod logs are empty (pod may not have started, or previous logs don't exist for container %s)", podLogOpts.Container)
	}

	r.logger.Debug("Successfully retrieved pod logs",
		zap.String("namespace", namespace),
		zap.String("pod_name", podName),
		zap.Int("log_size", len(logs)),
		zap.Int("log_lines", len(strings.Split(logs, "\n"))),
	)

	return logs, nil
}

// MockKubernetesClient implements KubernetesClient for testing and development
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
