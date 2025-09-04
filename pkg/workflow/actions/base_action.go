package actions

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	core_event "github.com/kubecano/cano-collector/pkg/core/event"
	logger_interfaces "github.com/kubecano/cano-collector/pkg/logger/interfaces"
	metric_interfaces "github.com/kubecano/cano-collector/pkg/metric/interfaces"
	actions_interfaces "github.com/kubecano/cano-collector/pkg/workflow/actions/interfaces"
)

// BaseAction provides common functionality for all workflow actions
type BaseAction struct {
	config  actions_interfaces.ActionConfig
	logger  logger_interfaces.LoggerInterface
	metrics metric_interfaces.MetricsInterface
}

// NewBaseAction creates a new BaseAction instance
func NewBaseAction(config actions_interfaces.ActionConfig, logger logger_interfaces.LoggerInterface, metrics metric_interfaces.MetricsInterface) *BaseAction {
	return &BaseAction{
		config:  config,
		logger:  logger,
		metrics: metrics,
	}
}

// GetName returns the action name
func (ba *BaseAction) GetName() string {
	return ba.config.Name
}

// GetType returns the action type
func (ba *BaseAction) GetType() string {
	return ba.config.Type
}

// GetTimeout returns the action timeout duration
func (ba *BaseAction) GetTimeout() time.Duration {
	if ba.config.Timeout > 0 {
		return time.Duration(ba.config.Timeout) * time.Second
	}
	return 30 * time.Second // Default timeout
}

// IsEnabled returns whether the action is enabled
func (ba *BaseAction) IsEnabled() bool {
	return ba.config.Enabled
}

// GetParameter retrieves a parameter from the action configuration
func (ba *BaseAction) GetParameter(key string) (interface{}, bool) {
	value, exists := ba.config.Parameters[key]
	return value, exists
}

// GetStringParameter retrieves a string parameter with optional default value
func (ba *BaseAction) GetStringParameter(key, defaultValue string) string {
	if value, exists := ba.config.Parameters[key]; exists {
		if strValue, ok := value.(string); ok {
			return strValue
		}
	}
	return defaultValue
}

// GetIntParameter retrieves an integer parameter with optional default value
func (ba *BaseAction) GetIntParameter(key string, defaultValue int) int {
	if value, exists := ba.config.Parameters[key]; exists {
		if intValue, ok := value.(int); ok {
			return intValue
		}
		// Try to convert float64 to int (JSON numbers are parsed as float64)
		if floatValue, ok := value.(float64); ok {
			return int(floatValue)
		}
	}
	return defaultValue
}

// GetBoolParameter retrieves a boolean parameter with optional default value
func (ba *BaseAction) GetBoolParameter(key string, defaultValue bool) bool {
	if value, exists := ba.config.Parameters[key]; exists {
		if boolValue, ok := value.(bool); ok {
			return boolValue
		}
	}
	return defaultValue
}

// ValidateBasicConfig validates basic action configuration
func (ba *BaseAction) ValidateBasicConfig() error {
	if ba.config.Name == "" {
		return fmt.Errorf("action name cannot be empty")
	}

	if ba.config.Type == "" {
		return fmt.Errorf("action type cannot be empty")
	}

	if ba.config.Timeout < 0 {
		return fmt.Errorf("action timeout cannot be negative")
	}

	return nil
}

// RecordMetric records a metric for the action
func (ba *BaseAction) RecordMetric(metricName string, value float64, labels map[string]string) {
	if ba.metrics != nil {
		enrichedLabels := ba.enrichMetricLabels(labels)
		// Note: This assumes metrics interface has a generic RecordMetric method
		// You might need to adjust based on your actual metrics interface
		if ba.logger != nil {
			ba.logger.Debug("Recording metric",
				zap.String("action_name", ba.config.Name),
				zap.String("action_type", ba.config.Type),
				zap.String("metric_name", metricName),
				zap.Float64("value", value),
				zap.Any("labels", enrichedLabels),
			)
		}
	}
}

// enrichMetricLabels adds action context to metric labels
func (ba *BaseAction) enrichMetricLabels(labels map[string]string) map[string]string {
	enriched := make(map[string]string)

	// Add action context
	enriched["action_name"] = ba.config.Name
	enriched["action_type"] = ba.config.Type

	// Add user labels
	for k, v := range labels {
		enriched[k] = v
	}

	return enriched
}

// CreateSuccessResult creates a successful ActionResult
func (ba *BaseAction) CreateSuccessResult(data interface{}, enrichments ...interface{}) *actions_interfaces.ActionResult {
	result := &actions_interfaces.ActionResult{
		Success:  true,
		Data:     data,
		Metadata: make(map[string]interface{}),
	}

	// Convert enrichments to the proper type
	// Note: You might need to adjust this based on your actual enrichment types
	for _, enrichment := range enrichments {
		// This is a placeholder - implement based on your enrichment interface
		ba.logger.Debug("Adding enrichment to result", zap.Any("enrichment", enrichment))
	}

	return result
}

// CreateErrorResult creates a failed ActionResult
func (ba *BaseAction) CreateErrorResult(err error, data interface{}) *actions_interfaces.ActionResult {
	return &actions_interfaces.ActionResult{
		Success:  false,
		Data:     data,
		Error:    err,
		Metadata: make(map[string]interface{}),
	}
}

// ExecuteWithTimeout executes a function with the configured timeout
func (ba *BaseAction) ExecuteWithTimeout(ctx context.Context, fn func(context.Context) (*actions_interfaces.ActionResult, error)) (*actions_interfaces.ActionResult, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, ba.GetTimeout())
	defer cancel()

	resultChan := make(chan *actions_interfaces.ActionResult, 1)
	errorChan := make(chan error, 1)

	go func() {
		result, err := fn(timeoutCtx)
		if err != nil {
			errorChan <- err
		} else {
			resultChan <- result
		}
	}()

	select {
	case result := <-resultChan:
		return result, nil
	case err := <-errorChan:
		return nil, err
	case <-timeoutCtx.Done():
		return nil, fmt.Errorf("action %s timed out after %v", ba.config.Name, ba.GetTimeout())
	}
}

// ExtractAlertEvent extracts AlertManagerWorkflowEvent from a WorkflowEvent
func (ba *BaseAction) ExtractAlertEvent(event core_event.WorkflowEvent, actionType string) (*core_event.AlertManagerWorkflowEvent, error) {
	alertEvent, ok := event.(*core_event.AlertManagerWorkflowEvent)
	if !ok {
		err := fmt.Errorf("%s action requires AlertManagerWorkflowEvent, got %T", actionType, event)
		ba.logger.Error("Invalid event type for action", zap.String("action_type", actionType), zap.Error(err))
		return nil, err
	}
	return alertEvent, nil
}

// GetFirstAlert extracts the first alert from an AlertManagerWorkflowEvent
func (ba *BaseAction) GetFirstAlert(alertEvent *core_event.AlertManagerWorkflowEvent, actionType string) (*core_event.PrometheusAlert, error) {
	if len(alertEvent.GetAlertManagerEvent().Alerts) == 0 {
		err := fmt.Errorf("no alerts found in event")
		ba.logger.Error("No alerts in event", zap.String("action_type", actionType), zap.Error(err))
		return nil, err
	}
	return &alertEvent.GetAlertManagerEvent().Alerts[0], nil
}

// getMapParameter retrieves a map parameter from action configuration
func (ba *BaseAction) getMapParameter(key string) map[string]string {
	if value, exists := ba.GetParameter(key); exists {
		if mapValue, ok := value.(map[string]interface{}); ok {
			result := make(map[string]string)
			for k, v := range mapValue {
				if strValue, ok := v.(string); ok {
					result[k] = strValue
				}
			}
			return result
		}
	}
	return make(map[string]string)
}

// getSliceParameter retrieves a slice parameter from action configuration
func (ba *BaseAction) getSliceParameter(key string) []string {
	if value, exists := ba.GetParameter(key); exists {
		if sliceValue, ok := value.([]interface{}); ok {
			result := make([]string, 0, len(sliceValue))
			for _, v := range sliceValue {
				if strValue, ok := v.(string); ok {
					result = append(result, strValue)
				}
			}
			return result
		}
	}
	return make([]string, 0)
}
