package actions

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"

	pod_info_config "github.com/kubecano/cano-collector/config/workflow/actions"
	"github.com/kubecano/cano-collector/pkg/core/event"
	"github.com/kubecano/cano-collector/pkg/core/issue"
	logger_interfaces "github.com/kubecano/cano-collector/pkg/logger/interfaces"
	metric_interfaces "github.com/kubecano/cano-collector/pkg/metric/interfaces"
	actions_interfaces "github.com/kubecano/cano-collector/pkg/workflow/actions/interfaces"
)

// ContainerCrashInfo represents crash information extracted from a container
type ContainerCrashInfo struct {
	Container     string
	RestartCount  int32
	Status        string // "WAITING" or "TERMINATED"
	Reason        string // "CrashLoopBackOff", "Error", etc.
	LastStateInfo *PreviousContainerInfo
}

// PreviousContainerInfo contains information about the previous container state
type PreviousContainerInfo struct {
	Reason     string
	ExitCode   int32
	StartedAt  time.Time
	FinishedAt time.Time
}

// PodInfoAction retrieves pod crash information for alerts
type PodInfoAction struct {
	*BaseAction
	config     pod_info_config.PodInfoActionConfig
	kubeClient actions_interfaces.KubernetesClient
}

// NewPodInfoAction creates a new PodInfoAction
func NewPodInfoAction(
	config pod_info_config.PodInfoActionConfig,
	logger logger_interfaces.LoggerInterface,
	metrics metric_interfaces.MetricsInterface,
	kubeClient actions_interfaces.KubernetesClient,
) *PodInfoAction {
	baseAction := NewBaseAction(config.ActionConfig, logger, metrics)

	return &PodInfoAction{
		BaseAction: baseAction,
		config:     config,
		kubeClient: kubeClient,
	}
}

// Execute retrieves pod information and extracts crash info
func (a *PodInfoAction) Execute(ctx context.Context, event event.WorkflowEvent) (*actions_interfaces.ActionResult, error) {
	a.logger.Info("Starting pod info action execution",
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

	a.logger.Info("Fetching pod information",
		zap.String("action_name", a.GetName()),
		zap.String("pod_name", podName),
		zap.String("namespace", namespace),
	)

	// Fetch Pod from Kubernetes API
	podObj, err := a.kubeClient.GetPod(ctx, namespace, podName)
	if err != nil {
		a.logger.Warn("Failed to fetch pod, creating empty crash info enrichment",
			zap.Error(err),
			zap.String("action_name", a.GetName()),
			zap.String("pod_name", podName),
			zap.String("namespace", namespace),
		)

		// Create enrichment with error explanation
		errorMessage := fmt.Sprintf("⚠️ Pod information unavailable for %s/%s\n\nReason: %v", namespace, podName, err)
		enrichment := a.createCrashInfoEnrichment([]ContainerCrashInfo{}, errorMessage)

		return &actions_interfaces.ActionResult{
			Success: true,
			Data: map[string]interface{}{
				"pod_name":  podName,
				"namespace": namespace,
				"error":     err.Error(),
			},
			Enrichments: []issue.Enrichment{*enrichment},
			Metadata: map[string]interface{}{
				"action_type": "pod_info",
				"timestamp":   time.Now(),
			},
		}, nil
	}

	// podObj is already *corev1.Pod type
	pod := podObj
	if pod == nil {
		err := fmt.Errorf("pod object is nil")
		a.logger.Error("Pod object is nil",
			zap.Error(err),
			zap.String("pod_name", podName),
			zap.String("namespace", namespace),
		)
		return a.CreateErrorResult(err, map[string]interface{}{
			"pod_name":  podName,
			"namespace": namespace,
		}), nil
	}

	// Extract crash info from pod status
	crashInfos := a.extractCrashInfo(pod)

	// Create enrichment with crash info
	enrichment := a.createCrashInfoEnrichment(crashInfos, "")

	result := &actions_interfaces.ActionResult{
		Success: true,
		Data: map[string]interface{}{
			"pod_name":          podName,
			"namespace":         namespace,
			"crash_count":       len(crashInfos),
			"include_previous":  a.config.IncludePreviousState,
			"include_init":      a.config.IncludeInitContainers,
			"min_restart_count": a.config.MinRestartCount,
		},
		Enrichments: []issue.Enrichment{*enrichment},
		Metadata: map[string]interface{}{
			"action_type": "pod_info",
			"timestamp":   time.Now(),
		},
	}

	a.logger.Info("Pod info action completed successfully",
		zap.String("action_name", a.GetName()),
		zap.String("pod_name", podName),
		zap.String("namespace", namespace),
		zap.Int("crash_count", len(crashInfos)),
	)

	return result, nil
}

// Validate checks if the action configuration is valid
func (a *PodInfoAction) Validate() error {
	if err := a.ValidateBasicConfig(); err != nil {
		return err
	}

	if a.kubeClient == nil {
		return fmt.Errorf("kubernetes client is required for PodInfoAction")
	}

	return nil
}

// extractPodInfo extracts pod name and namespace from the workflow event
func (a *PodInfoAction) extractPodInfo(workflowEvent event.WorkflowEvent) (string, string) {
	// Get namespace from event
	namespace := workflowEvent.GetNamespace()
	if namespace == "" {
		namespace = "default"
	}

	// Check if pod name is provided in action parameters
	if podName := a.GetStringParameter("pod_name", ""); podName != "" {
		return podName, namespace
	}

	// Check if config has pod name override
	if a.config.PodName != "" {
		return a.config.PodName, namespace
	}

	// Try to extract from alert labels if this is an AlertManager event
	if alertEvent, ok := workflowEvent.(*event.AlertManagerWorkflowEvent); ok {
		labels := alertEvent.GetAlertManagerEvent().GetLabels()

		a.logger.Debug("Extracting pod info from alert labels",
			zap.Any("labels", labels),
			zap.String("namespace", namespace),
		)

		// Check for pod label
		if podName, exists := labels["pod"]; exists && podName != "" {
			a.logger.Info("Pod info extracted from 'pod' label",
				zap.String("pod", podName),
				zap.String("namespace", namespace),
			)
			return podName, namespace
		}

		// Check for instance label (sometimes contains pod name)
		if instance, exists := labels["instance"]; exists && instance != "" {
			// Instance might be in format "pod-name:port" or just "pod-name"
			parts := make([]string, 0)
			for i, c := range instance {
				if c == ':' {
					parts = append(parts, instance[:i])
					break
				}
			}
			if len(parts) == 0 {
				parts = append(parts, instance)
			}

			if len(parts) > 0 && parts[0] != "" {
				a.logger.Info("Pod info extracted from 'instance' label",
					zap.String("instance", instance),
					zap.String("pod", parts[0]),
					zap.String("namespace", namespace),
				)
				return parts[0], namespace
			}
		}

		a.logger.Warn("No pod information found in alert labels",
			zap.Any("available_labels", labels),
			zap.String("namespace", namespace),
		)
	}

	return "", namespace
}

// extractCrashInfo extracts crash information from pod container statuses
func (a *PodInfoAction) extractCrashInfo(pod *corev1.Pod) []ContainerCrashInfo {
	var crashInfos []ContainerCrashInfo

	// Collect all container statuses to process
	var allStatuses []corev1.ContainerStatus
	allStatuses = append(allStatuses, pod.Status.ContainerStatuses...)
	if a.config.IncludeInitContainers {
		allStatuses = append(allStatuses, pod.Status.InitContainerStatuses...)
	}

	for _, containerStatus := range allStatuses {
		// Skip containers below minimum restart count
		if containerStatus.RestartCount < a.config.MinRestartCount {
			continue
		}

		// Skip if filtering by container name
		if a.config.Container != "" && containerStatus.Name != a.config.Container {
			continue
		}

		info := ContainerCrashInfo{
			Container:    containerStatus.Name,
			RestartCount: containerStatus.RestartCount,
		}

		// Extract current state
		if containerStatus.State.Waiting != nil {
			info.Status = "WAITING"
			info.Reason = containerStatus.State.Waiting.Reason
		} else if containerStatus.State.Terminated != nil {
			info.Status = "TERMINATED"
			info.Reason = containerStatus.State.Terminated.Reason
		} else if containerStatus.State.Running != nil {
			// Container is running but has restarts
			info.Status = "RUNNING"
			info.Reason = "Restarted"
		}

		// Extract previous state if requested
		if a.config.IncludePreviousState && containerStatus.LastTerminationState.Terminated != nil {
			info.LastStateInfo = &PreviousContainerInfo{
				Reason:     containerStatus.LastTerminationState.Terminated.Reason,
				ExitCode:   containerStatus.LastTerminationState.Terminated.ExitCode,
				StartedAt:  containerStatus.LastTerminationState.Terminated.StartedAt.Time,
				FinishedAt: containerStatus.LastTerminationState.Terminated.FinishedAt.Time,
			}
		}

		crashInfos = append(crashInfos, info)
	}

	return crashInfos
}

// createCrashInfoEnrichment creates an enrichment with crash information
func (a *PodInfoAction) createCrashInfoEnrichment(crashInfos []ContainerCrashInfo, errorMessage string) *issue.Enrichment {
	var blocks []issue.BaseBlock

	// If there's an error message, create a markdown block with it
	if errorMessage != "" {
		blocks = append(blocks, issue.NewMarkdownBlock(errorMessage))
	}

	// Create table blocks for each crashed container
	for _, info := range crashInfos {
		// Crash Info table
		crashRows := [][]string{
			{"Container", info.Container},
			{"Restarts", strconv.Itoa(int(info.RestartCount))},
			{"Status", info.Status},
			{"Reason", info.Reason},
		}

		crashTable := issue.NewTableBlock(
			[]string{"label", "value"},
			crashRows,
			"Crash Info",
			issue.TableBlockFormatVertical,
		)
		blocks = append(blocks, crashTable)

		// Previous Container table (if available and requested)
		if a.config.IncludePreviousState && info.LastStateInfo != nil {
			prevRows := [][]string{
				{"Status", "TERMINATED"},
				{"Reason", info.LastStateInfo.Reason},
				{"Exit Code", strconv.Itoa(int(info.LastStateInfo.ExitCode))},
				{"Started At", info.LastStateInfo.StartedAt.Format(time.RFC3339)},
				{"Finished At", info.LastStateInfo.FinishedAt.Format(time.RFC3339)},
			}

			prevTable := issue.NewTableBlock(
				[]string{"label", "value"},
				prevRows,
				"Previous Container",
				issue.TableBlockFormatVertical,
			)
			blocks = append(blocks, prevTable)
		}
	}

	// Create enrichment
	enrichment := issue.NewEnrichmentWithType(
		issue.EnrichmentTypeCrashInfo,
		"Container Crash Information",
	)

	for _, block := range blocks {
		enrichment.AddBlock(block)
	}

	return enrichment
}

// GetActionType returns the action type for registration
func (a *PodInfoAction) GetActionType() string {
	return "pod_info"
}
