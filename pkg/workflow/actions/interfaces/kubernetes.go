package interfaces

import (
	"context"

	corev1 "k8s.io/api/core/v1"
)

// KubernetesClient represents a simplified kubernetes client interface
// Used for pod logs retrieval and pod information access
//
//go:generate mockgen -source=kubernetes.go -destination=../../../../mocks/kubernetes_mock.go -package=mocks
type KubernetesClient interface {
	// GetPodLogs retrieves pod logs with options (container, previous, timestamps, etc.)
	GetPodLogs(ctx context.Context, namespace, podName string, options map[string]interface{}) (string, error)

	// GetPod retrieves pod information
	GetPod(ctx context.Context, namespace, podName string) (*corev1.Pod, error)
}
