package actions

import (
	"fmt"

	pod_logs_config "github.com/kubecano/cano-collector/config/workflow/actions"
	logger_interfaces "github.com/kubecano/cano-collector/pkg/logger/interfaces"
	metric_interfaces "github.com/kubecano/cano-collector/pkg/metric/interfaces"
	actions_interfaces "github.com/kubecano/cano-collector/pkg/workflow/actions/interfaces"
)

// PodLogsActionFactory creates PodLogsAction instances
type PodLogsActionFactory struct {
	logger     logger_interfaces.LoggerInterface
	metrics    metric_interfaces.MetricsInterface
	kubeClient KubernetesClient
}

// NewPodLogsActionFactory creates a new PodLogsActionFactory
func NewPodLogsActionFactory(
	logger logger_interfaces.LoggerInterface,
	metrics metric_interfaces.MetricsInterface,
	kubeClient KubernetesClient,
) *PodLogsActionFactory {
	return &PodLogsActionFactory{
		logger:     logger,
		metrics:    metrics,
		kubeClient: kubeClient,
	}
}

// Create creates a new PodLogsAction instance from configuration
func (f *PodLogsActionFactory) Create(config actions_interfaces.ActionConfig) (actions_interfaces.WorkflowAction, error) {
	// Create PodLogsActionConfig with defaults from environment
	podLogsConfig := pod_logs_config.NewPodLogsActionConfigWithDefaults(config)

	// Update with parameters from action config
	if err := podLogsConfig.UpdateFromParameters(config.Parameters); err != nil {
		return nil, fmt.Errorf("failed to update configuration from parameters: %w", err)
	}

	// Validate the configuration
	if err := podLogsConfig.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	// Create the action
	action := NewPodLogsAction(podLogsConfig, f.logger, f.metrics, f.kubeClient)

	return action, nil
}

// GetActionType returns the action type this factory creates
func (f *PodLogsActionFactory) GetActionType() string {
	return "pod_logs"
}

// ValidateConfig validates the action configuration
func (f *PodLogsActionFactory) ValidateConfig(config actions_interfaces.ActionConfig) error {
	if config.Type != "pod_logs" {
		return fmt.Errorf("invalid action type for PodLogsActionFactory: %s", config.Type)
	}

	// Create a temporary config to validate
	podLogsConfig := pod_logs_config.NewPodLogsActionConfigWithDefaults(config)

	// Update with parameters to test their validity
	if err := podLogsConfig.UpdateFromParameters(config.Parameters); err != nil {
		return fmt.Errorf("invalid parameters: %w", err)
	}

	// Validate the complete configuration
	if err := podLogsConfig.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	return nil
}
