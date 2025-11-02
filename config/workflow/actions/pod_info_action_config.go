package actions

import (
	"fmt"

	actions_interfaces "github.com/kubecano/cano-collector/pkg/workflow/actions/interfaces"
)

// PodInfoActionConfig defines the configuration for the pod_info workflow action
type PodInfoActionConfig struct {
	actions_interfaces.ActionConfig `yaml:",inline"`

	// IncludePreviousState includes information about the previous container state
	IncludePreviousState bool `yaml:"include_previous_state" json:"include_previous_state"`

	// IncludeInitContainers includes crash info for init containers
	IncludeInitContainers bool `yaml:"include_init_containers" json:"include_init_containers"`

	// MinRestartCount only report containers with at least this many restarts
	// Default is 0 (report all containers with any restarts)
	MinRestartCount int32 `yaml:"min_restart_count" json:"min_restart_count"`

	// PodName optionally override the pod name (defaults to extracting from event)
	PodName string `yaml:"pod_name,omitempty" json:"pod_name,omitempty"`

	// Container optionally filter to specific container name
	Container string `yaml:"container,omitempty" json:"container,omitempty"`
}

// Validate checks if the configuration is valid
func (c *PodInfoActionConfig) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("action name cannot be empty")
	}

	if c.MinRestartCount < 0 {
		return fmt.Errorf("min_restart_count cannot be negative")
	}

	return nil
}

// UpdateFromParameters updates the configuration from workflow parameters
func (c *PodInfoActionConfig) UpdateFromParameters(params map[string]interface{}) error {
	if params == nil {
		return nil
	}

	if val, ok := params["include_previous_state"].(bool); ok {
		c.IncludePreviousState = val
	}

	if val, ok := params["include_init_containers"].(bool); ok {
		c.IncludeInitContainers = val
	}

	if val, ok := params["min_restart_count"].(int); ok {
		c.MinRestartCount = int32(val)
	}

	if val, ok := params["min_restart_count"].(float64); ok {
		c.MinRestartCount = int32(val)
	}

	if val, ok := params["pod_name"].(string); ok {
		c.PodName = val
	}

	if val, ok := params["container"].(string); ok {
		c.Container = val
	}

	return nil
}

// GetActionType returns the action type identifier
func (c *PodInfoActionConfig) GetActionType() string {
	return "pod_info"
}

// NewPodInfoActionConfigWithDefaults creates a new PodInfoActionConfig with default values
func NewPodInfoActionConfigWithDefaults(baseConfig actions_interfaces.ActionConfig) PodInfoActionConfig {
	return PodInfoActionConfig{
		ActionConfig:          baseConfig,
		IncludePreviousState:  true,
		IncludeInitContainers: false,
		MinRestartCount:       0,
	}
}
