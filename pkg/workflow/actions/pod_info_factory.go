package actions

import (
	"fmt"

	pod_info_config "github.com/kubecano/cano-collector/config/workflow/actions"
	logger_interfaces "github.com/kubecano/cano-collector/pkg/logger/interfaces"
	metric_interfaces "github.com/kubecano/cano-collector/pkg/metric/interfaces"
	actions_interfaces "github.com/kubecano/cano-collector/pkg/workflow/actions/interfaces"
)

// PodInfoActionFactory creates PodInfoAction instances
type PodInfoActionFactory struct {
	logger     logger_interfaces.LoggerInterface
	metrics    metric_interfaces.MetricsInterface
	kubeClient KubernetesClient
}

// NewPodInfoActionFactory creates a new PodInfoActionFactory
func NewPodInfoActionFactory(
	logger logger_interfaces.LoggerInterface,
	metrics metric_interfaces.MetricsInterface,
	kubeClient KubernetesClient,
) *PodInfoActionFactory {
	return &PodInfoActionFactory{
		logger:     logger,
		metrics:    metrics,
		kubeClient: kubeClient,
	}
}

// Create creates a new PodInfoAction instance from configuration
func (f *PodInfoActionFactory) Create(config actions_interfaces.ActionConfig) (actions_interfaces.WorkflowAction, error) {
	// Validate action type
	if config.Type != "pod_info" {
		return nil, fmt.Errorf("invalid action type for PodInfoActionFactory: %s (expected: pod_info)", config.Type)
	}

	// Create PodInfoActionConfig with defaults
	podInfoConfig := pod_info_config.NewPodInfoActionConfigWithDefaults(config)

	// Update with parameters from action config
	if err := podInfoConfig.UpdateFromParameters(config.Parameters); err != nil {
		return nil, fmt.Errorf("failed to update configuration from parameters: %w", err)
	}

	// Validate the configuration
	if err := podInfoConfig.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	// Create the action
	action := NewPodInfoAction(podInfoConfig, f.logger, f.metrics, f.kubeClient)

	return action, nil
}

// GetActionType returns the action type this factory creates
func (f *PodInfoActionFactory) GetActionType() string {
	return "pod_info"
}
