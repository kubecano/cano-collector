package interfaces

import (
	"context"

	"github.com/kubecano/cano-collector/pkg/core/event"
	"github.com/kubecano/cano-collector/pkg/core/issue"
)

//go:generate mockgen -source=actions.go -destination=../../../../mocks/workflow_action_mock.go -package=mocks

// WorkflowAction defines the interface for all workflow actions
type WorkflowAction interface {
	// Execute runs the action with the provided context and event
	Execute(ctx context.Context, event event.WorkflowEvent) (*ActionResult, error)

	// GetName returns the unique name of the action
	GetName() string

	// Validate checks if the action configuration is valid
	Validate() error
}

// ActionResult represents the result of executing a workflow action
type ActionResult struct {
	// Success indicates if the action executed successfully
	Success bool

	// Data contains action-specific result data
	Data interface{}

	// Error contains any error that occurred during execution
	Error error

	// Enrichments contains issue enrichments to be added to the issue
	Enrichments []issue.Enrichment

	// Metadata contains additional metadata about the action execution
	Metadata map[string]interface{}
}

// ActionConfig represents common configuration for all actions
type ActionConfig struct {
	// Name is the action name
	Name string `yaml:"name" json:"name"`

	// Type is the action type
	Type string `yaml:"type" json:"type"`

	// Enabled indicates if the action is enabled
	Enabled bool `yaml:"enabled" json:"enabled"`

	// Timeout specifies action execution timeout in seconds
	Timeout int `yaml:"timeout" json:"timeout"`

	// Parameters contains action-specific parameters
	Parameters map[string]interface{} `yaml:"parameters" json:"parameters"`
}

// ActionExecutor handles the execution of workflow actions
type ActionExecutor interface {
	// ExecuteAction executes a workflow action
	ExecuteAction(ctx context.Context, action WorkflowAction, event event.WorkflowEvent) (*ActionResult, error)

	// RegisterAction registers a new action type
	RegisterAction(actionType string, action WorkflowAction) error

	// GetAction retrieves an action by type
	GetAction(actionType string) (WorkflowAction, error)

	// CreateActionsFromConfig creates workflow actions from action configurations
	CreateActionsFromConfig(configs []ActionConfig) ([]WorkflowAction, error)

	// ExecuteActions executes multiple workflow actions in sequence
	ExecuteActions(ctx context.Context, actions []WorkflowAction, event event.WorkflowEvent) ([]*ActionResult, error)
}

// ActionRegistry manages available workflow actions
type ActionRegistry interface {
	// Register registers a new action
	Register(actionType string, factory ActionFactory) error

	// Create creates an action instance from configuration
	Create(config ActionConfig) (WorkflowAction, error)

	// GetRegisteredTypes returns all registered action types
	GetRegisteredTypes() []string
}

// ActionFactory creates instances of workflow actions
type ActionFactory interface {
	// Create creates a new action instance from configuration
	Create(config ActionConfig) (WorkflowAction, error)

	// GetActionType returns the action type this factory creates
	GetActionType() string

	// ValidateConfig validates the action configuration
	ValidateConfig(config ActionConfig) error
}
