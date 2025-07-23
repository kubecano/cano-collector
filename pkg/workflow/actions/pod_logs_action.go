package actions

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/kubecano/cano-collector/pkg/core/event"
	"github.com/kubecano/cano-collector/pkg/core/issue"
	logger_interfaces "github.com/kubecano/cano-collector/pkg/logger/interfaces"
	metric_interfaces "github.com/kubecano/cano-collector/pkg/metric/interfaces"
	actions_interfaces "github.com/kubecano/cano-collector/pkg/workflow/actions/interfaces"
)

// PodLogsActionConfig contains configuration for PodLogsAction
type PodLogsActionConfig struct {
	actions_interfaces.ActionConfig `yaml:",inline"`

	// MaxLines maximum number of lines to retrieve
	MaxLines int `yaml:"max_lines" json:"max_lines"`

	// SinceTime retrieve logs since this time (RFC3339 format)
	SinceTime string `yaml:"since_time" json:"since_time"`

	// TailLines number of lines from the end of the logs to show
	TailLines int `yaml:"tail_lines" json:"tail_lines"`

	// Container name to get logs from (empty means all containers)
	Container string `yaml:"container" json:"container"`

	// Previous get logs from previous instance of the container
	Previous bool `yaml:"previous" json:"previous"`

	// Timestamps add timestamps to each log line
	Timestamps bool `yaml:"timestamps" json:"timestamps"`
}

// KubernetesClient represents a simplified kubernetes client interface for testing
// In real implementation, this would be replaced with kubernetes.Interface
type KubernetesClient interface {
	GetPodLogs(ctx context.Context, namespace, podName string, options map[string]interface{}) (string, error)
}

// PodLogsAction retrieves pod logs for alerts
type PodLogsAction struct {
	*BaseAction
	config     PodLogsActionConfig
	kubeClient KubernetesClient
}

// NewPodLogsAction creates a new PodLogsAction
func NewPodLogsAction(
	config PodLogsActionConfig,
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
	podName, namespace := a.extractPodInfo(event)
	if podName == "" {
		msg := "no pod information found in event"
		a.logger.Info(msg,
			zap.String("action_name", a.GetName()),
			zap.String("event_id", event.GetID().String()),
			zap.String("namespace", namespace),
		)
		return a.CreateSuccessResult(msg), nil
	}

	a.logger.Info("Extracting logs for pod",
		zap.String("action_name", a.GetName()),
		zap.String("pod_name", podName),
		zap.String("namespace", namespace),
		zap.String("container", a.config.Container),
	)

	// Get pod logs
	logs, err := a.getPodLogs(ctx, podName, namespace)
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
	enrichment := a.createLogsEnrichment(podName, namespace, logs)

	result := &actions_interfaces.ActionResult{
		Success: true,
		Data: map[string]interface{}{
			"pod_name":  podName,
			"namespace": namespace,
			"log_lines": len(strings.Split(logs, "\n")),
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

// extractPodInfo extracts pod name and namespace from the workflow event
func (a *PodLogsAction) extractPodInfo(event event.WorkflowEvent) (string, string) {
	// Get namespace from event - this uses the WorkflowEvent interface method
	namespace := event.GetNamespace()
	if namespace == "" {
		namespace = "default"
	}

	// Check if pod name is provided in action parameters
	if podName := a.GetStringParameter("pod_name", ""); podName != "" {
		return podName, namespace
	}

	// Try to extract from alert name (simplified approach)
	alertName := event.GetAlertName()
	if strings.Contains(alertName, "Pod") || strings.Contains(alertName, "pod") {
		// Look for pod name in parameters or try to parse from alert name
		return a.GetStringParameter("pod_name", ""), namespace
	}

	return "", namespace
}

// getPodLogs retrieves logs from the specified pod
func (a *PodLogsAction) getPodLogs(ctx context.Context, podName, namespace string) (string, error) {
	// Build log options
	logOptions := map[string]interface{}{
		"timestamps": a.config.Timestamps,
		"previous":   a.config.Previous,
	}

	// Set container if specified
	if a.config.Container != "" {
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

// createLogsEnrichment creates an enrichment with the pod logs
func (a *PodLogsAction) createLogsEnrichment(podName, namespace, logs string) *issue.Enrichment {
	// Create a FileBlock for the logs
	filename := fmt.Sprintf("pod-logs-%s-%s.log", namespace, podName)
	fileBlock := issue.NewFileBlock(filename, []byte(logs), "text/plain")

	// Create enrichment with the file block
	enrichment := issue.NewEnrichmentWithType(
		issue.EnrichmentTypeTextFile,
		fmt.Sprintf("Pod Logs: %s/%s", namespace, podName),
	)
	enrichment.AddBlock(fileBlock)

	return enrichment
}

// GetActionType returns the action type for registration
func (a *PodLogsAction) GetActionType() string {
	return "pod_logs"
}
