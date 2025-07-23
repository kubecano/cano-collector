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

// ResourceStatusActionConfig contains configuration for ResourceStatusAction
type ResourceStatusActionConfig struct {
	actions_interfaces.ActionConfig `yaml:",inline"`

	// ResourceTypes types of resources to check status for
	ResourceTypes []string `yaml:"resource_types" json:"resource_types"`

	// IncludeEvents whether to include related Kubernetes events
	IncludeEvents bool `yaml:"include_events" json:"include_events"`

	// MaxEvents maximum number of events to include
	MaxEvents int `yaml:"max_events" json:"max_events"`

	// IncludeLabels whether to include resource labels
	IncludeLabels bool `yaml:"include_labels" json:"include_labels"`

	// IncludeAnnotations whether to include resource annotations
	IncludeAnnotations bool `yaml:"include_annotations" json:"include_annotations"`
}

// ResourceInfo represents information about a Kubernetes resource
type ResourceInfo struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Kind        string            `json:"kind"`
	Status      string            `json:"status"`
	Ready       string            `json:"ready"`
	Age         string            `json:"age"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Events      []EventInfo       `json:"events,omitempty"`
}

// EventInfo represents a Kubernetes event
type EventInfo struct {
	Type      string `json:"type"`
	Reason    string `json:"reason"`
	Message   string `json:"message"`
	Age       string `json:"age"`
	Component string `json:"component"`
}

// KubernetesResourceClient represents a simplified kubernetes resource client interface
type KubernetesResourceClient interface {
	GetResourceStatus(ctx context.Context, namespace, resourceType, resourceName string) (*ResourceInfo, error)
	GetRelatedEvents(ctx context.Context, namespace, resourceType, resourceName string, maxEvents int) ([]EventInfo, error)
}

// ResourceStatusAction checks the status of Kubernetes resources
type ResourceStatusAction struct {
	*BaseAction
	config     ResourceStatusActionConfig
	kubeClient KubernetesResourceClient
}

// NewResourceStatusAction creates a new ResourceStatusAction
func NewResourceStatusAction(
	config ResourceStatusActionConfig,
	logger logger_interfaces.LoggerInterface,
	metrics metric_interfaces.MetricsInterface,
	kubeClient KubernetesResourceClient,
) *ResourceStatusAction {
	baseAction := NewBaseAction(config.ActionConfig, logger, metrics)

	return &ResourceStatusAction{
		BaseAction: baseAction,
		config:     config,
		kubeClient: kubeClient,
	}
}

// Execute checks the status of resources related to the alert
func (a *ResourceStatusAction) Execute(ctx context.Context, event event.WorkflowEvent) (*actions_interfaces.ActionResult, error) {
	a.logger.Info("Starting resource status action execution",
		zap.String("action_name", a.GetName()),
		zap.String("action_type", a.GetType()),
		zap.String("event_id", event.GetID().String()),
		zap.String("event_type", string(event.GetType())),
	)

	// Extract resource information from the event
	resourceName, namespace, resourceType := a.extractResourceInfo(event)
	if resourceName == "" {
		msg := "no resource information found in event"
		a.logger.Info(msg,
			zap.String("action_name", a.GetName()),
			zap.String("event_id", event.GetID().String()),
			zap.String("namespace", namespace),
		)
		return a.CreateSuccessResult(msg), nil
	}

	a.logger.Info("Checking status for resource",
		zap.String("action_name", a.GetName()),
		zap.String("resource_name", resourceName),
		zap.String("namespace", namespace),
		zap.String("resource_type", resourceType),
	)

	// Get resource status
	resourceInfo, err := a.getResourceStatus(ctx, resourceName, namespace, resourceType)
	if err != nil {
		a.logger.Error("Failed to retrieve resource status",
			zap.Error(err),
			zap.String("action_name", a.GetName()),
			zap.String("resource_name", resourceName),
			zap.String("namespace", namespace),
			zap.String("resource_type", resourceType),
		)
		return a.CreateErrorResult(err, nil), nil
	}

	// Create enrichment with resource status
	enrichment := a.createResourceStatusEnrichment(resourceInfo)

	result := &actions_interfaces.ActionResult{
		Success: true,
		Data: map[string]interface{}{
			"resource_name": resourceName,
			"namespace":     namespace,
			"resource_type": resourceType,
			"status":        resourceInfo.Status,
			"ready":         resourceInfo.Ready,
		},
		Enrichments: []issue.Enrichment{*enrichment},
		Metadata: map[string]interface{}{
			"action_type": "resource_status",
			"timestamp":   time.Now(),
		},
	}

	a.logger.Info("Resource status action completed successfully",
		zap.String("action_name", a.GetName()),
		zap.String("resource_name", resourceName),
		zap.String("namespace", namespace),
		zap.String("resource_type", resourceType),
		zap.String("status", resourceInfo.Status),
	)

	return result, nil
}

// Validate checks if the action configuration is valid
func (a *ResourceStatusAction) Validate() error {
	if err := a.ValidateBasicConfig(); err != nil {
		return err
	}

	if a.kubeClient == nil {
		return fmt.Errorf("kubernetes client is required for ResourceStatusAction")
	}

	if a.config.MaxEvents < 0 {
		return fmt.Errorf("max_events must be non-negative")
	}

	// Set default resource types if none specified
	if len(a.config.ResourceTypes) == 0 {
		a.config.ResourceTypes = []string{"pod", "deployment", "service"}
	}

	return nil
}

// extractResourceInfo extracts resource information from the workflow event
func (a *ResourceStatusAction) extractResourceInfo(event event.WorkflowEvent) (string, string, string) {
	// Get namespace from event
	namespace := event.GetNamespace()
	if namespace == "" {
		namespace = "default"
	}

	// Check if resource name and type are provided in action parameters
	resourceName := a.GetStringParameter("resource_name", "")
	resourceType := a.GetStringParameter("resource_type", "")

	if resourceName != "" && resourceType != "" {
		return resourceName, namespace, resourceType
	}

	// Try to extract from alert name (simplified approach)
	alertName := event.GetAlertName()

	// Try to determine resource type from alert name patterns
	alertNameLower := strings.ToLower(alertName)

	for _, resourceTypeCandidate := range a.config.ResourceTypes {
		if strings.Contains(alertNameLower, resourceTypeCandidate) {
			// If resource name is not specified, try to get it from parameters
			if resourceName == "" {
				resourceName = a.GetStringParameter("resource_name", "")
			}
			return resourceName, namespace, resourceTypeCandidate
		}
	}

	// Default to pod if no specific resource type found
	if resourceName == "" {
		resourceName = a.GetStringParameter("resource_name", "")
	}

	return resourceName, namespace, "pod"
}

// getResourceStatus retrieves status information for the specified resource
func (a *ResourceStatusAction) getResourceStatus(ctx context.Context, resourceName, namespace, resourceType string) (*ResourceInfo, error) {
	// Get basic resource status
	resourceInfo, err := a.kubeClient.GetResourceStatus(ctx, namespace, resourceType, resourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource status: %w", err)
	}

	// Get related events if enabled
	if a.config.IncludeEvents {
		maxEvents := a.config.MaxEvents
		if maxEvents <= 0 {
			maxEvents = 10 // Default
		}

		events, err := a.kubeClient.GetRelatedEvents(ctx, namespace, resourceType, resourceName, maxEvents)
		if err != nil {
			a.logger.Error("Failed to retrieve related events",
				zap.Error(err),
				zap.String("resource_name", resourceName),
				zap.String("namespace", namespace),
				zap.String("resource_type", resourceType),
			)
			// Don't fail the action, just log the error
		} else {
			resourceInfo.Events = events
		}
	}

	// Clear labels and annotations if not requested
	if !a.config.IncludeLabels {
		resourceInfo.Labels = nil
	}

	if !a.config.IncludeAnnotations {
		resourceInfo.Annotations = nil
	}

	return resourceInfo, nil
}

// createResourceStatusEnrichment creates an enrichment with the resource status information
func (a *ResourceStatusAction) createResourceStatusEnrichment(resourceInfo *ResourceInfo) *issue.Enrichment {
	// Create enrichment with resource status table
	enrichment := issue.NewEnrichmentWithType(
		issue.EnrichmentTypeContainerInfo,
		fmt.Sprintf("Resource Status: %s/%s (%s)", resourceInfo.Namespace, resourceInfo.Name, resourceInfo.Kind),
	)

	// Create basic status table
	statusTable := issue.NewTableBlock(
		[]string{"Property", "Value"},
		[][]string{
			{"Name", resourceInfo.Name},
			{"Namespace", resourceInfo.Namespace},
			{"Kind", resourceInfo.Kind},
			{"Status", resourceInfo.Status},
			{"Ready", resourceInfo.Ready},
			{"Age", resourceInfo.Age},
		},
		"Resource Information",
		issue.TableBlockFormatVertical,
	)
	enrichment.AddBlock(statusTable)

	// Add labels table if present
	if len(resourceInfo.Labels) > 0 {
		labelsRows := make([][]string, 0, len(resourceInfo.Labels))
		for key, value := range resourceInfo.Labels {
			labelsRows = append(labelsRows, []string{key, value})
		}

		labelsTable := issue.NewTableBlock(
			[]string{"Label", "Value"},
			labelsRows,
			"Labels",
			issue.TableBlockFormatVertical,
		)
		enrichment.AddBlock(labelsTable)
	}

	// Add annotations table if present
	if len(resourceInfo.Annotations) > 0 {
		annotationsRows := make([][]string, 0, len(resourceInfo.Annotations))
		for key, value := range resourceInfo.Annotations {
			// Truncate long annotation values
			if len(value) > 100 {
				value = value[:97] + "..."
			}
			annotationsRows = append(annotationsRows, []string{key, value})
		}

		annotationsTable := issue.NewTableBlock(
			[]string{"Annotation", "Value"},
			annotationsRows,
			"Annotations",
			issue.TableBlockFormatVertical,
		)
		enrichment.AddBlock(annotationsTable)
	}

	// Add events table if present
	if len(resourceInfo.Events) > 0 {
		eventsRows := make([][]string, 0, len(resourceInfo.Events))
		for _, event := range resourceInfo.Events {
			eventsRows = append(eventsRows, []string{
				event.Type,
				event.Reason,
				event.Age,
				event.Component,
				event.Message,
			})
		}

		eventsTable := issue.NewTableBlock(
			[]string{"Type", "Reason", "Age", "From", "Message"},
			eventsRows,
			"Related Events",
			issue.TableBlockFormatHorizontal,
		)
		enrichment.AddBlock(eventsTable)
	}

	return enrichment
}

// GetActionType returns the action type for registration
func (a *ResourceStatusAction) GetActionType() string {
	return "resource_status"
}
