package actions

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	pod_logs_config "github.com/kubecano/cano-collector/config/workflow/actions"
	"github.com/kubecano/cano-collector/pkg/core/event"
	"github.com/kubecano/cano-collector/pkg/core/issue"
	logger_interfaces "github.com/kubecano/cano-collector/pkg/logger/interfaces"
	metric_interfaces "github.com/kubecano/cano-collector/pkg/metric/interfaces"
	actions_interfaces "github.com/kubecano/cano-collector/pkg/workflow/actions/interfaces"
)

// KubernetesClient represents a simplified kubernetes client interface for testing
// In real implementation, this would be replaced with kubernetes.Interface
type KubernetesClient interface {
	GetPodLogs(ctx context.Context, namespace, podName string, options map[string]interface{}) (string, error)
}

// PodLogsAction retrieves pod logs for alerts
type PodLogsAction struct {
	*BaseAction
	config     pod_logs_config.PodLogsActionConfig
	kubeClient KubernetesClient
}

// NewPodLogsAction creates a new PodLogsAction
func NewPodLogsAction(
	config pod_logs_config.PodLogsActionConfig,
	logger logger_interfaces.LoggerInterface,
	metrics metric_interfaces.MetricsInterface,
	kubeClient KubernetesClient,
) *PodLogsAction {
	baseAction := NewBaseAction(config.ActionConfig, logger, metrics)

	return &PodLogsAction{
		BaseAction: baseAction,
		config:     config,
		kubeClient: kubeClient,
	}
}

// Execute retrieves pod logs based on the alert information
func (a *PodLogsAction) Execute(ctx context.Context, event event.WorkflowEvent) (*actions_interfaces.ActionResult, error) {
	a.logger.Info("Starting pod logs action execution",
		zap.String("action_name", a.GetName()),
		zap.String("action_type", a.GetType()),
		zap.String("event_id", event.GetID().String()),
		zap.String("event_type", string(event.GetType())),
	)

	// Extract pod information from the event
	podName, namespace, containerName := a.extractPodInfo(event)
	if podName == "" {
		msg := "no pod information found in event"
		a.logger.Info(msg,
			zap.String("action_name", a.GetName()),
			zap.String("event_id", event.GetID().String()),
			zap.String("namespace", namespace),
		)
		return a.CreateSuccessResult(msg), nil
	}

	// Check if Java container and apply Java defaults if needed
	if containerName != "" && pod_logs_config.IsJavaContainer(containerName, "") {
		a.logger.Info("Java container detected, applying Java-specific configuration",
			zap.String("container_name", containerName),
		)
		a.config.ApplyJavaDefaults()
	}

	a.logger.Info("Extracting logs for pod",
		zap.String("action_name", a.GetName()),
		zap.String("pod_name", podName),
		zap.String("namespace", namespace),
		zap.String("container", containerName),
		zap.Bool("java_specific", a.config.JavaSpecific),
	)

	// Get pod logs
	logs, err := a.getPodLogs(ctx, podName, namespace, containerName)
	if err != nil {
		a.logger.Error("Failed to retrieve pod logs",
			zap.Error(err),
			zap.String("action_name", a.GetName()),
			zap.String("pod_name", podName),
			zap.String("namespace", namespace),
		)
		return a.CreateErrorResult(err, nil), nil
	}

	// Create enrichment with logs
	enrichment := a.createLogsEnrichment(podName, namespace, containerName, logs)

	result := &actions_interfaces.ActionResult{
		Success: true,
		Data: map[string]interface{}{
			"pod_name":      podName,
			"namespace":     namespace,
			"container":     containerName,
			"log_lines":     len(strings.Split(logs, "\n")),
			"java_specific": a.config.JavaSpecific,
		},
		Enrichments: []issue.Enrichment{*enrichment},
		Metadata: map[string]interface{}{
			"action_type": "pod_logs",
			"timestamp":   time.Now(),
		},
	}

	a.logger.Info("Pod logs action completed successfully",
		zap.String("action_name", a.GetName()),
		zap.String("pod_name", podName),
		zap.String("namespace", namespace),
		zap.String("container", containerName),
		zap.Int("log_lines", len(strings.Split(logs, "\n"))),
	)

	return result, nil
}

// Validate checks if the action configuration is valid
func (a *PodLogsAction) Validate() error {
	if err := a.ValidateBasicConfig(); err != nil {
		return err
	}

	if a.kubeClient == nil {
		return fmt.Errorf("kubernetes client is required for PodLogsAction")
	}

	if a.config.MaxLines < 0 {
		return fmt.Errorf("max_lines must be non-negative")
	}

	if a.config.TailLines < 0 {
		return fmt.Errorf("tail_lines must be non-negative")
	}

	return nil
}

// extractPodInfo extracts pod name, namespace and container from the workflow event
func (a *PodLogsAction) extractPodInfo(workflowEvent event.WorkflowEvent) (string, string, string) {
	// Get namespace from event - this uses the WorkflowEvent interface method
	namespace := workflowEvent.GetNamespace()
	if namespace == "" {
		namespace = "default"
	}

	var containerName string

	// Check if pod name is provided in action parameters
	if podName := a.GetStringParameter("pod_name", ""); podName != "" {
		containerName = a.GetStringParameter("container", a.config.Container)
		return podName, namespace, containerName
	}

	// Try to extract from alert labels if this is an AlertManager event
	if alertEvent, ok := workflowEvent.(*event.AlertManagerWorkflowEvent); ok {
		labels := alertEvent.GetAlertManagerEvent().GetLabels()

		// Check for pod label
		if podName, exists := labels["pod"]; exists && podName != "" {
			// Also check for container label
			if container, exists := labels["container"]; exists {
				containerName = container
			}
			return podName, namespace, containerName
		}

		// Check for instance label (sometimes contains pod name)
		if instance, exists := labels["instance"]; exists && instance != "" {
			// Instance might be in format "pod-name:port" or just "pod-name"
			parts := strings.Split(instance, ":")
			if len(parts) > 0 && parts[0] != "" {
				return parts[0], namespace, containerName
			}
		}
	}

	// Fallback: try to extract from alert name (simplified approach)
	alertName := workflowEvent.GetAlertName()
	if strings.Contains(alertName, "Pod") || strings.Contains(alertName, "pod") {
		// Look for pod name in parameters
		return a.GetStringParameter("pod_name", ""), namespace, containerName
	}

	return "", namespace, containerName
}

// getPodLogs retrieves logs from the specified pod
func (a *PodLogsAction) getPodLogs(ctx context.Context, podName, namespace, containerName string) (string, error) {
	// Build log options
	logOptions := map[string]interface{}{
		"timestamps": a.config.Timestamps,
		"previous":   a.config.Previous,
	}

	// Set container if specified
	if containerName != "" {
		logOptions["container"] = containerName
	} else if a.config.Container != "" {
		logOptions["container"] = a.config.Container
	}

	// Set tail lines
	if a.config.TailLines > 0 {
		logOptions["tailLines"] = a.config.TailLines
	}

	// Set since time if specified
	if a.config.SinceTime != "" {
		_, err := time.Parse(time.RFC3339, a.config.SinceTime)
		if err != nil {
			return "", fmt.Errorf("invalid since_time format: %w", err)
		}
		logOptions["sinceTime"] = a.config.SinceTime
	}

	// Use the simplified kubernetes client interface
	logs, err := a.kubeClient.GetPodLogs(ctx, namespace, podName, logOptions)
	if err != nil {
		return "", fmt.Errorf("failed to get pod logs: %w", err)
	}

	// Apply MaxLines limit if specified
	if a.config.MaxLines > 0 {
		lines := strings.Split(logs, "\n")
		if len(lines) > a.config.MaxLines {
			lines = lines[:a.config.MaxLines]
			logs = strings.Join(lines, "\n")
		}
	}

	return logs, nil
}

// createLogsEnrichment creates an enrichment with the pod logs using proper naming with timestamps
func (a *PodLogsAction) createLogsEnrichment(podName, namespace, containerName, logs string) *issue.Enrichment {
	// Generate filename with timestamp and container name
	filename := a.generateLogFilename(podName, namespace, containerName)
	fileBlock := issue.NewFileBlock(filename, []byte(logs), "text/plain")

	// Create title
	title := fmt.Sprintf("Pod Logs: %s/%s", namespace, podName)
	if containerName != "" {
		title = fmt.Sprintf("Pod Logs: %s/%s (%s)", namespace, podName, containerName)
	}
	if a.config.JavaSpecific {
		title = fmt.Sprintf("Java %s", title)
	}

	// Create enrichment with the file block
	enrichment := issue.NewEnrichmentWithType(
		issue.EnrichmentTypeTextFile,
		title,
	)
	enrichment.AddBlock(fileBlock)

	return enrichment
}

// generateLogFilename generates appropriate filename with timestamp and container support
func (a *PodLogsAction) generateLogFilename(podName, namespace, containerName string) string {
	var filename string

	// Base filename
	if a.config.JavaSpecific {
		filename = fmt.Sprintf("java-logs-%s-%s", namespace, podName)
	} else {
		filename = fmt.Sprintf("pod-logs-%s-%s", namespace, podName)
	}

	// Add container name if specified and config allows it
	if a.config.IncludeContainer && containerName != "" {
		filename = fmt.Sprintf("%s-%s", filename, containerName)
	}

	// Add timestamp if config allows it
	if a.config.IncludeTimestamp {
		timestamp := time.Now().Format(a.config.TimestampFormat)
		filename = fmt.Sprintf("%s-%s", filename, timestamp)
	}

	return filename + ".log"
}

// GetActionType returns the action type for registration
func (a *PodLogsAction) GetActionType() string {
	return "pod_logs"
}
