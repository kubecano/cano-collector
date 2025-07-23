package actions

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"

	"github.com/kubecano/cano-collector/pkg/core/event"
	logger_interfaces "github.com/kubecano/cano-collector/pkg/logger/interfaces"
	metric_interfaces "github.com/kubecano/cano-collector/pkg/metric/interfaces"
	actions_interfaces "github.com/kubecano/cano-collector/pkg/workflow/actions/interfaces"
)

// DefaultActionRegistry implements ActionRegistry interface
type DefaultActionRegistry struct {
	factories map[string]actions_interfaces.ActionFactory
	logger    logger_interfaces.LoggerInterface
	mu        sync.RWMutex
}

// NewDefaultActionRegistry creates a new default action registry
func NewDefaultActionRegistry(logger logger_interfaces.LoggerInterface) *DefaultActionRegistry {
	return &DefaultActionRegistry{
		factories: make(map[string]actions_interfaces.ActionFactory),
		logger:    logger,
	}
}

// Register registers a new action factory
func (r *DefaultActionRegistry) Register(actionType string, factory actions_interfaces.ActionFactory) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if actionType == "" {
		return fmt.Errorf("action type cannot be empty")
	}

	if factory == nil {
		return fmt.Errorf("action factory cannot be nil")
	}

	if _, exists := r.factories[actionType]; exists {
		return fmt.Errorf("action type '%s' is already registered", actionType)
	}

	r.factories[actionType] = factory
	r.logger.Info("Registered workflow action factory", zap.String("action_type", actionType))

	return nil
}

// Create creates an action instance from configuration
func (r *DefaultActionRegistry) Create(config actions_interfaces.ActionConfig) (actions_interfaces.WorkflowAction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if config.Type == "" {
		return nil, fmt.Errorf("action type cannot be empty in config")
	}

	factory, exists := r.factories[config.Type]
	if !exists {
		return nil, fmt.Errorf("no factory registered for action type '%s'", config.Type)
	}

	return factory.Create(config)
}

// GetRegisteredTypes returns all registered action types
func (r *DefaultActionRegistry) GetRegisteredTypes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]string, 0, len(r.factories))
	for actionType := range r.factories {
		types = append(types, actionType)
	}

	return types
}

// DefaultActionExecutor implements ActionExecutor interface
type DefaultActionExecutor struct {
	registry actions_interfaces.ActionRegistry
	logger   logger_interfaces.LoggerInterface
	metrics  metric_interfaces.MetricsInterface
}

// NewDefaultActionExecutor creates a new default action executor
func NewDefaultActionExecutor(
	registry actions_interfaces.ActionRegistry,
	logger logger_interfaces.LoggerInterface,
	metrics metric_interfaces.MetricsInterface,
) *DefaultActionExecutor {
	return &DefaultActionExecutor{
		registry: registry,
		logger:   logger,
		metrics:  metrics,
	}
}

// ExecuteAction executes a workflow action
func (e *DefaultActionExecutor) ExecuteAction(ctx context.Context, action actions_interfaces.WorkflowAction, event event.WorkflowEvent) (*actions_interfaces.ActionResult, error) {
	if action == nil {
		return nil, fmt.Errorf("action cannot be nil")
	}

	if event == nil {
		return nil, fmt.Errorf("event cannot be nil")
	}

	e.logger.Info("Executing workflow action",
		zap.String("action_name", action.GetName()),
		zap.String("event_id", event.GetID().String()),
	)

	// Validate action before execution
	if err := action.Validate(); err != nil {
		e.logger.Error("Action validation failed",
			zap.Error(err),
			zap.String("action_name", action.GetName()),
		)
		return nil, fmt.Errorf("action validation failed: %w", err)
	}

	// Execute action with timeout and error handling
	result, err := action.Execute(ctx, event)
	if err != nil {
		e.logger.Error("Action execution failed",
			zap.Error(err),
			zap.String("action_name", action.GetName()),
			zap.String("event_id", event.GetID().String()),
		)
		return nil, fmt.Errorf("action execution failed: %w", err)
	}

	// Log execution result
	if result != nil && result.Success {
		e.logger.Info("Action executed successfully",
			zap.String("action_name", action.GetName()),
			zap.String("event_id", event.GetID().String()),
			zap.Int("enrichments", len(result.Enrichments)),
		)
	} else {
		var resultErr error
		if result != nil {
			resultErr = result.Error
		}
		e.logger.Error("Action execution was not successful",
			zap.Error(resultErr),
			zap.String("action_name", action.GetName()),
			zap.String("event_id", event.GetID().String()),
		)
	}

	return result, nil
}

// RegisterAction registers a new action instance (deprecated - use registry directly)
func (e *DefaultActionExecutor) RegisterAction(actionType string, action actions_interfaces.WorkflowAction) error {
	return fmt.Errorf("RegisterAction is deprecated, use ActionRegistry.Register with ActionFactory instead")
}

// GetAction retrieves an action by type (deprecated - use registry directly)
func (e *DefaultActionExecutor) GetAction(actionType string) (actions_interfaces.WorkflowAction, error) {
	return nil, fmt.Errorf("GetAction is deprecated, use ActionRegistry.Create instead")
}

// CreateActionsFromConfig creates workflow actions from action configurations
func (e *DefaultActionExecutor) CreateActionsFromConfig(configs []actions_interfaces.ActionConfig) ([]actions_interfaces.WorkflowAction, error) {
	actions := make([]actions_interfaces.WorkflowAction, 0, len(configs))

	for i, config := range configs {
		action, err := e.registry.Create(config)
		if err != nil {
			return nil, fmt.Errorf("failed to create action %d (%s): %w", i, config.Type, err)
		}

		actions = append(actions, action)
	}

	return actions, nil
}

// ExecuteActions executes multiple workflow actions in sequence
func (e *DefaultActionExecutor) ExecuteActions(ctx context.Context, actions []actions_interfaces.WorkflowAction, event event.WorkflowEvent) ([]*actions_interfaces.ActionResult, error) {
	if len(actions) == 0 {
		return []*actions_interfaces.ActionResult{}, nil
	}

	results := make([]*actions_interfaces.ActionResult, 0, len(actions))

	for i, action := range actions {
		result, err := e.ExecuteAction(ctx, action, event)
		if err != nil {
			e.logger.Error("Failed to execute action in sequence",
				zap.Error(err),
				zap.Int("action_index", i),
				zap.String("action_name", action.GetName()),
				zap.String("event_id", event.GetID().String()),
			)
			// Continue executing other actions even if one fails
			result = &actions_interfaces.ActionResult{
				Success: false,
				Error:   err,
				Metadata: map[string]interface{}{
					"action_name":  action.GetName(),
					"action_index": i,
					"error":        err.Error(),
				},
			}
		}

		results = append(results, result)
	}

	return results, nil
}
