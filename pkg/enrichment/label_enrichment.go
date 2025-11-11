package enrichment

import (
	"fmt"
	"sort"
	"strings"

	"go.uber.org/zap"

	"github.com/kubecano/cano-collector/pkg/core/issue"
	logger_interfaces "github.com/kubecano/cano-collector/pkg/logger/interfaces"
)

// LabelEnrichmentConfig contains configuration for label enrichment
type LabelEnrichmentConfig struct {
	EnableLabels            bool     `json:"enable_labels" yaml:"enable_labels"`
	EnableAnnotations       bool     `json:"enable_annotations" yaml:"enable_annotations"`
	ExcludeLabels           []string `json:"exclude_labels" yaml:"exclude_labels"`
	IncludeLabels           []string `json:"include_labels" yaml:"include_labels"`
	IncludeAnnotations      []string `json:"include_annotations" yaml:"include_annotations"`
	ExcludeAnnotations      []string `json:"exclude_annotations" yaml:"exclude_annotations"`
	DisplayFormat           string   `json:"display_format" yaml:"display_format"`                       // "table" or "json" for labels
	AnnotationDisplayFormat string   `json:"annotation_display_format" yaml:"annotation_display_format"` // "table" or "json" for annotations
}

// DefaultLabelEnrichmentConfig returns default configuration
func DefaultLabelEnrichmentConfig() *LabelEnrichmentConfig {
	return &LabelEnrichmentConfig{
		EnableLabels:      true,
		EnableAnnotations: true,
		ExcludeLabels: []string{
			// Internal Prometheus labels
			"__name__",
			"job",
			"instance",
			"endpoint",
			"prometheus", // Prometheus instance identifier
			"service",    // kube-state-metrics service name
			"uid",        // Kubernetes UID (long and not useful in alerts)
			// Kubernetes internal metadata
			"__meta_kubernetes_pod_uid",
			"__meta_kubernetes_pod_container_id",
		},
		IncludeLabels:      []string{}, // empty means include all (except excluded)
		IncludeAnnotations: []string{}, // empty means include all (except excluded)
		ExcludeAnnotations: []string{
			// Kubernetes internal annotations
			"kubectl.kubernetes.io/last-applied-configuration",
			"deployment.kubernetes.io/revision",
			"control-plane.alpha.kubernetes.io/leader",
			// Redundant annotations (already shown elsewhere)
			"runbook_url", // Already displayed in Links section
		},
		DisplayFormat:           "table",
		AnnotationDisplayFormat: "table",
	}
}

// LabelEnrichment handles enriching issues with Prometheus labels and annotations
type LabelEnrichment struct {
	logger logger_interfaces.LoggerInterface
	config *LabelEnrichmentConfig
}

// NewLabelEnrichment creates a new label enrichment handler
func NewLabelEnrichment(logger logger_interfaces.LoggerInterface, config *LabelEnrichmentConfig) *LabelEnrichment {
	if config == nil {
		config = DefaultLabelEnrichmentConfig()
	}

	return &LabelEnrichment{
		logger: logger,
		config: config,
	}
}

// EnrichIssue adds label and annotation enrichments to the issue
func (le *LabelEnrichment) EnrichIssue(iss *issue.Issue) error {
	if iss == nil {
		return fmt.Errorf("issue is nil")
	}

	if iss.Subject == nil {
		le.logger.Debug("Issue has no subject, skipping label enrichment")
		return nil
	}

	// Create labels enrichment
	if le.config.EnableLabels && len(iss.Subject.Labels) > 0 {
		if err := le.addLabelsEnrichment(iss); err != nil {
			le.logger.Error("Failed to add labels enrichment", zap.Error(err))
			return err
		}
	}

	// Create annotations enrichment
	if le.config.EnableAnnotations && len(iss.Subject.Annotations) > 0 {
		if err := le.addAnnotationsEnrichment(iss); err != nil {
			le.logger.Error("Failed to add annotations enrichment", zap.Error(err))
			return err
		}
	}

	return nil
}

// addLabelsEnrichment adds labels as enrichment blocks
func (le *LabelEnrichment) addLabelsEnrichment(iss *issue.Issue) error {
	filteredLabels := le.filterLabels(iss.Subject.Labels)
	if len(filteredLabels) == 0 {
		return nil
	}

	var blocks []issue.BaseBlock

	switch le.config.DisplayFormat {
	case "table":
		tableBlock := le.createLabelsTableBlock(filteredLabels)
		blocks = append(blocks, tableBlock)
	case "json":
		jsonBlock := le.createLabelsJsonBlock(filteredLabels)
		blocks = append(blocks, jsonBlock)
	default:
		tableBlock := le.createLabelsTableBlock(filteredLabels)
		blocks = append(blocks, tableBlock)
	}

	iss.AddEnrichmentWithType(blocks, issue.EnrichmentTypeAlertLabels, "Alert Labels")

	le.logger.Debug("Added labels enrichment", zap.Int("labels_count", len(filteredLabels)))
	return nil
}

// addAnnotationsEnrichment adds annotations as enrichment blocks
func (le *LabelEnrichment) addAnnotationsEnrichment(iss *issue.Issue) error {
	filteredAnnotations := le.filterAnnotations(iss.Subject.Annotations)
	if len(filteredAnnotations) == 0 {
		return nil
	}

	var blocks []issue.BaseBlock

	annotationFormat := le.config.AnnotationDisplayFormat
	if annotationFormat == "" {
		annotationFormat = le.config.DisplayFormat // fallback to main display format
	}

	switch annotationFormat {
	case "table":
		tableBlock := le.createAnnotationsTableBlock(filteredAnnotations)
		blocks = append(blocks, tableBlock)
	case "json":
		jsonBlock := le.createAnnotationsJsonBlock(filteredAnnotations)
		blocks = append(blocks, jsonBlock)
	default:
		tableBlock := le.createAnnotationsTableBlock(filteredAnnotations)
		blocks = append(blocks, tableBlock)
	}

	iss.AddEnrichmentWithType(blocks, issue.EnrichmentTypeAlertAnnotations, "Alert Annotations")

	le.logger.Debug("Added annotations enrichment", zap.Int("annotations_count", len(filteredAnnotations)))
	return nil
}

// filterLabels filters labels based on include/exclude lists
func (le *LabelEnrichment) filterLabels(labels map[string]string) map[string]string {
	filtered := make(map[string]string)

	for key, value := range labels {
		// Skip if in exclude list
		if le.isExcluded(key) {
			continue
		}

		// Include if include list is empty or key is in include list
		if len(le.config.IncludeLabels) == 0 || le.isIncluded(key) {
			filtered[key] = value
		}
	}

	return filtered
}

// filterAnnotations filters annotations based on include/exclude lists
func (le *LabelEnrichment) filterAnnotations(annotations map[string]string) map[string]string {
	filtered := make(map[string]string)

	for key, value := range annotations {
		// Skip if in exclude list
		if le.isAnnotationExcluded(key) {
			continue
		}

		// Include if include list is empty or key is in include list
		if len(le.config.IncludeAnnotations) == 0 || le.isAnnotationIncluded(key) {
			filtered[key] = value
		}
	}

	return filtered
}

// isExcluded checks if a label key should be excluded
func (le *LabelEnrichment) isExcluded(key string) bool {
	for _, exclude := range le.config.ExcludeLabels {
		if strings.Contains(key, exclude) {
			return true
		}
	}
	return false
}

// isIncluded checks if a label key should be included
func (le *LabelEnrichment) isIncluded(key string) bool {
	for _, include := range le.config.IncludeLabels {
		if strings.Contains(key, include) {
			return true
		}
	}
	return false
}

// isAnnotationExcluded checks if an annotation key should be excluded
func (le *LabelEnrichment) isAnnotationExcluded(key string) bool {
	for _, exclude := range le.config.ExcludeAnnotations {
		if strings.Contains(key, exclude) {
			return true
		}
	}
	return false
}

// isAnnotationIncluded checks if an annotation key should be included
func (le *LabelEnrichment) isAnnotationIncluded(key string) bool {
	for _, include := range le.config.IncludeAnnotations {
		if strings.Contains(key, include) {
			return true
		}
	}
	return false
}

// createLabelsTableBlock creates a table block for labels
func (le *LabelEnrichment) createLabelsTableBlock(labels map[string]string) *issue.TableBlock {
	// Sort labels by key for consistent display
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Create table data
	rows := make([][]string, 0, len(labels))
	for _, key := range keys {
		rows = append(rows, []string{key, labels[key]})
	}

	return &issue.TableBlock{
		Headers: []string{"Label", "Value"},
		Rows:    rows,
	}
}

// createAnnotationsTableBlock creates a table block for annotations
func (le *LabelEnrichment) createAnnotationsTableBlock(annotations map[string]string) *issue.TableBlock {
	// Sort annotations by key for consistent display
	keys := make([]string, 0, len(annotations))
	for k := range annotations {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Create table data
	rows := make([][]string, 0, len(annotations))
	for _, key := range keys {
		rows = append(rows, []string{key, annotations[key]})
	}

	return &issue.TableBlock{
		Headers: []string{"Annotation", "Value"},
		Rows:    rows,
	}
}

// createLabelsJsonBlock creates a JSON block for labels
func (le *LabelEnrichment) createLabelsJsonBlock(labels map[string]string) *issue.JsonBlock {
	return &issue.JsonBlock{
		Data: labels,
	}
}

// createAnnotationsJsonBlock creates a JSON block for annotations
func (le *LabelEnrichment) createAnnotationsJsonBlock(annotations map[string]string) *issue.JsonBlock {
	return &issue.JsonBlock{
		Data: annotations,
	}
}
