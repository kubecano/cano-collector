package actions

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"

	core_event "github.com/kubecano/cano-collector/pkg/core/event"
	"github.com/kubecano/cano-collector/pkg/core/issue"
	logger_interfaces "github.com/kubecano/cano-collector/pkg/logger/interfaces"
	metric_interfaces "github.com/kubecano/cano-collector/pkg/metric/interfaces"
	actions_interfaces "github.com/kubecano/cano-collector/pkg/workflow/actions/interfaces"
)

// LabelFilterAction filters alerts based on label matching rules
type LabelFilterAction struct {
	*BaseAction
}

// NewLabelFilterAction creates a new LabelFilterAction
func NewLabelFilterAction(
	config actions_interfaces.ActionConfig,
	logger logger_interfaces.LoggerInterface,
	metrics metric_interfaces.MetricsInterface,
) *LabelFilterAction {
	baseAction := NewBaseAction(config, logger, metrics)
	return &LabelFilterAction{
		BaseAction: baseAction,
	}
}

// Execute filters the alert based on configured label rules
func (a *LabelFilterAction) Execute(ctx context.Context, event core_event.WorkflowEvent) (*actions_interfaces.ActionResult, error) {
	a.logger.Info("Starting label filter action execution",
		zap.String("action_name", a.GetName()),
		zap.String("event_type", string(event.GetType())),
	)

	// Extract alert from event using common helper
	alertEvent, err := a.ExtractAlertEvent(event, "label filter")
	if err != nil {
		return a.CreateErrorResult(err, nil), err
	}

	// Get the first alert using common helper
	alert, err := a.GetFirstAlert(alertEvent, "label filter")
	if err != nil {
		return a.CreateErrorResult(err, nil), err
	}

	// Check if alert passes filter
	passed, reason := a.shouldPassFilter(alert.Labels)

	result := &actions_interfaces.ActionResult{
		Success: passed,
		Data: map[string]interface{}{
			"filter_passed": passed,
			"reason":        reason,
			"labels_count":  len(alert.Labels),
		},
		Metadata: map[string]interface{}{
			"action_type":   "label_filter",
			"filter_result": passed,
		},
	}

	if passed {
		a.logger.Info("Alert passed label filter",
			zap.String("reason", reason),
			zap.Int("labels_count", len(alert.Labels)),
		)
		a.RecordMetric("label_filter_passed", 1, map[string]string{"result": "passed"})
	} else {
		a.logger.Info("Alert filtered out by label filter",
			zap.String("reason", reason),
			zap.Int("labels_count", len(alert.Labels)),
		)
		a.RecordMetric("label_filter_passed", 0, map[string]string{"result": "filtered"})
	}

	return result, nil
}

// shouldPassFilter checks if alert labels match the configured filter rules
func (a *LabelFilterAction) shouldPassFilter(labels map[string]string) (bool, string) {
	// Get include labels filter
	if includeLabels := a.getMapParameter("include_labels"); len(includeLabels) > 0 {
		for labelKey, expectedValue := range includeLabels {
			actualValue, exists := labels[labelKey]
			if !exists {
				return false, "missing required include label: " + labelKey
			}
			if expectedValue != "" && actualValue != expectedValue {
				return false, fmt.Sprintf("include label %s has value %s, expected %s", labelKey, actualValue, expectedValue)
			}
		}
	}

	// Get exclude labels filter
	if excludeLabels := a.getMapParameter("exclude_labels"); len(excludeLabels) > 0 {
		for labelKey, excludeValue := range excludeLabels {
			actualValue, exists := labels[labelKey]
			if exists {
				if excludeValue == "" || actualValue == excludeValue {
					return false, fmt.Sprintf("alert has excluded label: %s=%s", labelKey, actualValue)
				}
			}
		}
	}

	// Get required labels filter
	if requiredLabels := a.getSliceParameter("required_labels"); len(requiredLabels) > 0 {
		for _, requiredLabel := range requiredLabels {
			if _, exists := labels[requiredLabel]; !exists {
				return false, "missing required label: " + requiredLabel
			}
		}
	}

	return true, "all label filters passed"
}

// Validate validates the LabelFilterAction configuration
func (a *LabelFilterAction) Validate() error {
	if err := a.ValidateBasicConfig(); err != nil {
		return err
	}

	// Validate that at least one filter type is configured
	includeLabels := a.getMapParameter("include_labels")
	excludeLabels := a.getMapParameter("exclude_labels")
	requiredLabels := a.getSliceParameter("required_labels")

	if len(includeLabels) == 0 && len(excludeLabels) == 0 && len(requiredLabels) == 0 {
		return fmt.Errorf("label filter action must have at least one filter configured (include_labels, exclude_labels, or required_labels)")
	}

	return nil
}

// SeverityRouterAction routes alerts to different destinations based on severity level
type SeverityRouterAction struct {
	*BaseAction
}

// NewSeverityRouterAction creates a new SeverityRouterAction
func NewSeverityRouterAction(
	config actions_interfaces.ActionConfig,
	logger logger_interfaces.LoggerInterface,
	metrics metric_interfaces.MetricsInterface,
) *SeverityRouterAction {
	baseAction := NewBaseAction(config, logger, metrics)
	return &SeverityRouterAction{
		BaseAction: baseAction,
	}
}

// Execute routes the alert based on its severity level
func (a *SeverityRouterAction) Execute(ctx context.Context, event core_event.WorkflowEvent) (*actions_interfaces.ActionResult, error) {
	a.logger.Info("Starting severity router action execution",
		zap.String("action_name", a.GetName()),
		zap.String("event_type", string(event.GetType())),
	)

	// Extract alert from event using common helper
	alertEvent, err := a.ExtractAlertEvent(event, "severity router")
	if err != nil {
		return a.CreateErrorResult(err, nil), err
	}

	// Get the first alert using common helper
	alert, err := a.GetFirstAlert(alertEvent, "severity router")
	if err != nil {
		return a.CreateErrorResult(err, nil), err
	}

	// Determine severity from alert labels
	severity, originalLabel := a.extractSeverityFromAlert(*alert)

	// Get destination mapping for this severity
	destination := a.getDestinationForSeverity(severity, originalLabel)

	result := &actions_interfaces.ActionResult{
		Success: true,
		Data: map[string]interface{}{
			"severity":    severity.String(),
			"destination": destination,
			"routed":      destination != "",
		},
		Metadata: map[string]interface{}{
			"action_type":       "severity_router",
			"severity":          severity.String(),
			"destination":       destination,
			"routing_performed": destination != "",
		},
	}

	if destination != "" {
		a.logger.Info("Alert routed based on severity",
			zap.String("severity", severity.String()),
			zap.String("destination", destination),
		)
		a.RecordMetric("severity_routing_performed", 1, map[string]string{
			"severity":    severity.String(),
			"destination": destination,
		})
	} else {
		a.logger.Info("No specific routing for severity, using default",
			zap.String("severity", severity.String()),
		)
		a.RecordMetric("severity_routing_performed", 0, map[string]string{
			"severity": severity.String(),
			"reason":   "no_mapping",
		})
	}

	return result, nil
}

// extractSeverityFromAlert extracts severity from alert labels
func (a *SeverityRouterAction) extractSeverityFromAlert(alert core_event.PrometheusAlert) (issue.Severity, string) {
	// Try to get severity from common label keys
	severityKeys := []string{"severity", "priority", "level"}

	for _, key := range severityKeys {
		if value, exists := alert.Labels[key]; exists {
			// Try Prometheus label mapping first (handles critical -> HIGH)
			severity := issue.SeverityFromPrometheusLabel(value)
			if severity != issue.SeverityInfo || value == "info" {
				return severity, value
			}
			// Fall back to direct string mapping
			if severity, err := issue.SeverityFromString(value); err == nil {
				return severity, value
			}
		}
	}

	// Default to Info if no severity found
	return issue.SeverityInfo, ""
}

// getDestinationForSeverity gets the destination mapping for a given severity
func (a *SeverityRouterAction) getDestinationForSeverity(severity issue.Severity, originalLabel string) string {
	severityMapping := a.getSeverityMapping()

	// Try original Prometheus label value first (e.g., "critical")
	if originalLabel != "" {
		if destination, exists := severityMapping[originalLabel]; exists {
			return destination
		}
		if destination, exists := severityMapping[strings.ToLower(originalLabel)]; exists {
			return destination
		}
	}

	// Try exact severity match (e.g., "HIGH")
	if destination, exists := severityMapping[severity.String()]; exists {
		return destination
	}

	// Try lowercase match (e.g., "high")
	if destination, exists := severityMapping[strings.ToLower(severity.String())]; exists {
		return destination
	}

	// Try default mapping
	if destination, exists := severityMapping["default"]; exists {
		return destination
	}

	return ""
}

// getSeverityMapping retrieves severity to destination mapping from configuration
func (a *SeverityRouterAction) getSeverityMapping() map[string]string {
	if value, exists := a.GetParameter("severity_mapping"); exists {
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

// Validate validates the SeverityRouterAction configuration
func (a *SeverityRouterAction) Validate() error {
	if err := a.ValidateBasicConfig(); err != nil {
		return err
	}

	// Validate that severity mapping is configured
	severityMapping := a.getSeverityMapping()
	if len(severityMapping) == 0 {
		return fmt.Errorf("severity router action must have severity_mapping parameter configured")
	}

	// Validate that all configured severities are valid
	validSeverities := map[string]bool{
		"debug": true, "info": true, "low": true, "high": true, "default": true,
		"DEBUG": true, "INFO": true, "LOW": true, "HIGH": true,
		// Also accept common Prometheus label values
		"critical": true, "warning": true, "CRITICAL": true, "WARNING": true,
	}

	for severity := range severityMapping {
		if !validSeverities[severity] {
			return fmt.Errorf("invalid severity level in mapping: %s (valid: debug, info, low, high, default)", severity)
		}
	}

	return nil
}
