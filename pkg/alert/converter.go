package alert

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/kubecano/cano-collector/config"
	"github.com/kubecano/cano-collector/pkg/core/event"
	"github.com/kubecano/cano-collector/pkg/core/issue"
	"github.com/kubecano/cano-collector/pkg/enrichment"
	logger_interfaces "github.com/kubecano/cano-collector/pkg/logger/interfaces"
)

// Converter handles conversion from AlertManagerEvent to Issues
type Converter struct {
	logger          logger_interfaces.LoggerInterface
	labelEnrichment *enrichment.LabelEnrichment
}

// NewConverter creates a new Converter
func NewConverter(logger logger_interfaces.LoggerInterface) *Converter {
	return &Converter{
		logger:          logger,
		labelEnrichment: enrichment.NewLabelEnrichment(logger, nil), // Use default config
	}
}

// NewConverterWithEnrichmentConfig creates a new Converter with custom enrichment configuration
func NewConverterWithEnrichmentConfig(logger logger_interfaces.LoggerInterface, enrichmentConfig *enrichment.LabelEnrichmentConfig) *Converter {
	return &Converter{
		logger:          logger,
		labelEnrichment: enrichment.NewLabelEnrichment(logger, enrichmentConfig),
	}
}

// NewConverterWithConfig creates a new Converter with enrichment configuration from Config
func NewConverterWithConfig(logger logger_interfaces.LoggerInterface, enrichmentConfig config.EnrichmentConfig) *Converter {
	// Convert config types to enrichment types
	enrichmentLabelConfig := &enrichment.LabelEnrichmentConfig{
		EnableLabels:            enrichmentConfig.Labels.Enabled,
		EnableAnnotations:       enrichmentConfig.Annotations.Enabled,
		DisplayFormat:           enrichmentConfig.Labels.DisplayFormat,
		AnnotationDisplayFormat: enrichmentConfig.Annotations.DisplayFormat,
		IncludeLabels:           enrichmentConfig.Labels.IncludeLabels,
		ExcludeLabels:           enrichmentConfig.Labels.ExcludeLabels,
		IncludeAnnotations:      enrichmentConfig.Annotations.IncludeAnnotations,
		ExcludeAnnotations:      enrichmentConfig.Annotations.ExcludeAnnotations,
	}

	return &Converter{
		logger:          logger,
		labelEnrichment: enrichment.NewLabelEnrichment(logger, enrichmentLabelConfig),
	}
}

// ConvertAlertManagerEventToIssues converts AlertManagerEvent to a slice of Issues
func (c *Converter) ConvertAlertManagerEventToIssues(event *event.AlertManagerEvent) ([]*issue.Issue, error) {
	if event == nil {
		return nil, fmt.Errorf("alertmanager event is nil")
	}

	if len(event.Alerts) == 0 {
		return nil, fmt.Errorf("no alerts in alertmanager event")
	}

	var issues []*issue.Issue
	for _, alert := range event.Alerts {
		iss, err := c.convertPrometheusAlertToIssue(alert)
		if err != nil {
			c.logger.Error("Failed to convert prometheus alert to issue",
				zap.Error(err),
				zap.String("fingerprint", alert.Fingerprint),
				zap.String("alertname", alert.Labels["alertname"]),
			)
			continue
		}
		issues = append(issues, iss)
	}

	if len(issues) == 0 {
		return nil, fmt.Errorf("no issues created from alertmanager event")
	}

	return issues, nil
}

// convertPrometheusAlertToIssue converts a single PrometheusAlert to Issue
func (c *Converter) convertPrometheusAlertToIssue(alert event.PrometheusAlert) (*issue.Issue, error) {
	// Extract alert name from labels
	alertName, exists := alert.Labels["alertname"]
	if !exists {
		return nil, fmt.Errorf("missing alertname label")
	}

	// Create title
	title := c.createTitle(alert)

	// Create description
	description := c.createDescription(alert)

	// Create new issue
	iss := issue.NewIssue(title, alertName)

	// Set basic fields
	iss.Description = description
	iss.Severity = issue.SeverityFromPrometheusLabel(alert.Labels["severity"])
	iss.Status = issue.StatusFromPrometheusStatus(alert.Status)
	iss.Source = issue.SourcePrometheus
	iss.StartsAt = alert.StartsAt

	// Set end time if available
	if !alert.EndsAt.IsZero() {
		iss.EndsAt = &alert.EndsAt
	}

	// Create subject from labels
	subject := c.createSubject(alert)
	iss.SetSubject(subject)

	// Set fingerprint - use alert fingerprint if available, otherwise use generated one
	if alert.Fingerprint != "" {
		iss.SetFingerprint(alert.Fingerprint)
	}
	// If alert.Fingerprint is empty, the issue will use its generated fingerprint

	// Add generator URL as link if available
	if alert.GeneratorURL != "" {
		link := issue.NewLink("Generator URL", alert.GeneratorURL, issue.LinkTypePrometheusGenerator)
		iss.AddLink(*link)
	}

	// Add runbook URL as link if available in annotations
	if runbookURL, exists := alert.Annotations["runbook_url"]; exists && runbookURL != "" {
		link := issue.NewLink("Runbook", runbookURL, issue.LinkTypeRunbook)
		iss.AddLink(*link)
	}

	// Apply label enrichment
	if err := c.labelEnrichment.EnrichIssue(iss); err != nil {
		c.logger.Warn("Failed to apply label enrichment", zap.Error(err))
		// Don't fail the conversion, just log the warning
	}

	return iss, nil
}

// createTitle creates a title for the issue
func (c *Converter) createTitle(alert event.PrometheusAlert) string {
	// Check for summary annotation first
	if summary, exists := alert.Annotations["summary"]; exists {
		return summary
	}

	// Fall back to alert name
	if alertName, exists := alert.Labels["alertname"]; exists {
		return alertName
	}

	return "Unknown Alert"
}

// createDescription creates a description for the issue
func (c *Converter) createDescription(alert event.PrometheusAlert) string {
	// Check for description annotation first
	if desc, exists := alert.Annotations["description"]; exists {
		return desc
	}

	// Fall back to summary
	if summary, exists := alert.Annotations["summary"]; exists {
		return summary
	}

	// Fall back to alert name
	if alertName, exists := alert.Labels["alertname"]; exists {
		return "Alert: " + alertName
	}

	return "No description available"
}

// createSubject creates a Subject from alert labels
func (c *Converter) createSubject(alert event.PrometheusAlert) *issue.Subject {
	// Determine subject type and name using helper functions
	subjectType := c.getSubjectTypeFromLabels(alert.Labels)
	subjectName := c.getSubjectNameFromLabels(alert.Labels, subjectType)

	// Create subject
	subject := issue.NewSubject(subjectName, subjectType)

	// Set namespace if available
	if namespace, exists := alert.Labels["namespace"]; exists {
		subject.Namespace = namespace
	}

	// Set node if available
	if node, exists := alert.Labels["node"]; exists {
		subject.Node = node
	}

	// Set container if available
	if container, exists := alert.Labels["container"]; exists {
		subject.Container = container
	}

	// Copy all labels
	for k, v := range alert.Labels {
		if subject.Labels == nil {
			subject.Labels = make(map[string]string)
		}
		subject.Labels[k] = v
	}

	// Copy all annotations
	for k, v := range alert.Annotations {
		if subject.Annotations == nil {
			subject.Annotations = make(map[string]string)
		}
		subject.Annotations[k] = v
	}

	return subject
}

// getSubjectTypeFromLabels determines the subject type from alert labels
func (c *Converter) getSubjectTypeFromLabels(labels map[string]string) issue.SubjectType {
	// Priority order for determining subject type
	priorities := []struct {
		label       string
		subjectType issue.SubjectType
	}{
		{"pod", issue.SubjectTypePod},
		{"deployment", issue.SubjectTypeDeployment},
		{"service", issue.SubjectTypeService},
		{"node", issue.SubjectTypeNode},
		{"instance", issue.SubjectTypeNode},
		{"job", issue.SubjectTypeJob},
		{"cronjob", issue.SubjectTypeCronJob},
		{"daemonset", issue.SubjectTypeDaemonSet},
		{"statefulset", issue.SubjectTypeStatefulSet},
		{"replicaset", issue.SubjectTypeReplicaSet},
		{"ingress", issue.SubjectTypeIngress},
		{"configmap", issue.SubjectTypeConfigMap},
		{"secret", issue.SubjectTypeSecret},
		{"persistentvolume", issue.SubjectTypePersistentVolume},
		{"persistentvolumeclaim", issue.SubjectTypePersistentVolumeClaim},
		{"hpa", issue.SubjectTypeHPA},
		{"namespace", issue.SubjectTypeNamespace},
	}

	for _, priority := range priorities {
		if _, exists := labels[priority.label]; exists {
			return priority.subjectType
		}
	}

	return issue.SubjectTypeNone
}

// getSubjectNameFromLabels gets the subject name based on the determined type
func (c *Converter) getSubjectNameFromLabels(labels map[string]string, subjectType issue.SubjectType) string {
	switch subjectType {
	case issue.SubjectTypePod:
		return labels["pod"]
	case issue.SubjectTypeDeployment:
		return labels["deployment"]
	case issue.SubjectTypeService:
		return labels["service"]
	case issue.SubjectTypeNode:
		if node, exists := labels["node"]; exists {
			return node
		}
		return labels["instance"]
	case issue.SubjectTypeJob:
		return labels["job"]
	case issue.SubjectTypeCronJob:
		return labels["cronjob"]
	case issue.SubjectTypeDaemonSet:
		return labels["daemonset"]
	case issue.SubjectTypeStatefulSet:
		return labels["statefulset"]
	case issue.SubjectTypeReplicaSet:
		return labels["replicaset"]
	case issue.SubjectTypeIngress:
		return labels["ingress"]
	case issue.SubjectTypeConfigMap:
		return labels["configmap"]
	case issue.SubjectTypeSecret:
		return labels["secret"]
	case issue.SubjectTypePersistentVolume:
		return labels["persistentvolume"]
	case issue.SubjectTypePersistentVolumeClaim:
		return labels["persistentvolumeclaim"]
	case issue.SubjectTypeHPA:
		return labels["hpa"]
	case issue.SubjectTypeNamespace:
		return labels["namespace"]
	default:
		return "Unknown"
	}
}
