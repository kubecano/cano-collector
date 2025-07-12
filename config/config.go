package config

import (
	"os"
	"strconv"
	"strings"

	config_destination "github.com/kubecano/cano-collector/config/destination"
	config_team "github.com/kubecano/cano-collector/config/team"
	config_workflow "github.com/kubecano/cano-collector/config/workflow"
)

type EnrichmentConfig struct {
	Labels      LabelEnrichmentConfig      `json:"labels"`
	Annotations AnnotationEnrichmentConfig `json:"annotations"`
}

type LabelEnrichmentConfig struct {
	Enabled       bool     `json:"enabled"`
	DisplayFormat string   `json:"displayFormat"`
	IncludeLabels []string `json:"includeLabels"`
	ExcludeLabels []string `json:"excludeLabels"`
}

type AnnotationEnrichmentConfig struct {
	Enabled            bool     `json:"enabled"`
	DisplayFormat      string   `json:"displayFormat"`
	IncludeAnnotations []string `json:"includeAnnotations"`
	ExcludeAnnotations []string `json:"excludeAnnotations"`
}

type Config struct {
	AppName         string
	AppVersion      string
	AppEnv          string
	LogLevel        string
	TracingMode     string
	TracingEndpoint string
	SentryDSN       string
	SentryEnabled   bool
	Destinations    config_destination.DestinationsConfig
	Teams           config_team.TeamsConfig
	Workflows       config_workflow.WorkflowConfig
	Enrichment      EnrichmentConfig
}

//go:generate mockgen -destination=../mocks/fullconfig_loader_mock.go -package=mocks github.com/kubecano/cano-collector/config FullConfigLoader
type FullConfigLoader interface {
	Load() (config_destination.DestinationsConfig, config_team.TeamsConfig, config_workflow.WorkflowConfig, error)
}

// LoadConfigWithLoader reads the Config from the provided loader
func LoadConfigWithLoader(loader FullConfigLoader) (Config, error) {
	destinations, teams, workflows, err := loader.Load()
	if err != nil {
		return Config{}, err
	}

	return Config{
		AppName:         getEnvString("APP_NAME", "cano-collector"),
		AppVersion:      getEnvString("APP_VERSION", "dev"),
		AppEnv:          getEnvString("APP_ENV", "production"),
		LogLevel:        getEnvEnum("LOG_LEVEL", []string{"debug", "info", "warn", "error"}, "info"),
		TracingMode:     getEnvEnum("TRACING_MODE", []string{"disabled", "local", "remote"}, "disabled"),
		TracingEndpoint: getEnvString("TRACING_ENDPOINT", "http://localhost:4317"),
		SentryDSN:       getEnvString("SENTRY_DSN", ""),
		SentryEnabled:   getEnvBool("ENABLE_TELEMETRY", true),
		Destinations:    destinations,
		Teams:           teams,
		Workflows:       workflows,
		Enrichment:      loadEnrichmentConfig(),
	}, nil
}

type fileConfigLoader struct {
	destinationsPath string
	teamsPath        string
	workflowsPath    string
}

func NewFileConfigLoader(destinationsPath, teamsPath, workflowsPath string) FullConfigLoader {
	return &fileConfigLoader{
		destinationsPath: destinationsPath,
		teamsPath:        teamsPath,
		workflowsPath:    workflowsPath,
	}
}

func (f *fileConfigLoader) Load() (config_destination.DestinationsConfig, config_team.TeamsConfig, config_workflow.WorkflowConfig, error) {
	destLoader := config_destination.NewFileDestinationsLoader(f.destinationsPath)
	teamLoader := config_team.NewFileTeamsLoader(f.teamsPath)
	workflowLoader := config_workflow.NewConfigLoader(f.workflowsPath)

	d, err := destLoader.Load()
	if err != nil {
		return config_destination.DestinationsConfig{}, config_team.TeamsConfig{}, config_workflow.WorkflowConfig{}, err
	}

	t, err := teamLoader.Load()
	if err != nil {
		return config_destination.DestinationsConfig{}, config_team.TeamsConfig{}, config_workflow.WorkflowConfig{}, err
	}

	w, err := workflowLoader.LoadConfig()
	if err != nil {
		return config_destination.DestinationsConfig{}, config_team.TeamsConfig{}, config_workflow.WorkflowConfig{}, err
	}

	return *d, *t, *w, nil
}

func LoadConfig() (Config, error) {
	loader := NewFileConfigLoader(
		"/etc/cano-collector/destinations/destinations.yaml",
		"/etc/cano-collector/teams/teams.yaml",
		"/etc/cano-collector/workflows/workflows.yaml",
	)
	return LoadConfigWithLoader(loader)
}

// Helpers
func getEnvString(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	if parsed, err := strconv.ParseBool(value); err == nil {
		return parsed
	}
	return defaultValue
}

func getEnvEnum(key string, allowedValues []string, defaultValue string) string {
	value := getEnvString(key, defaultValue)
	for _, allowed := range allowedValues {
		if value == allowed {
			return value
		}
	}
	return defaultValue
}

func getEnvStringSlice(key string, defaultValue []string) []string {
	value := getEnvString(key, "")
	if value == "" {
		return defaultValue
	}
	return strings.Split(value, ",")
}

func loadEnrichmentConfig() EnrichmentConfig {
	return EnrichmentConfig{
		Labels: LabelEnrichmentConfig{
			Enabled:       getEnvBool("ENRICHMENT_LABELS_ENABLED", true),
			DisplayFormat: getEnvEnum("ENRICHMENT_LABELS_DISPLAY_FORMAT", []string{"table", "json"}, "table"),
			IncludeLabels: getEnvStringSlice("ENRICHMENT_LABELS_INCLUDE", []string{}),
			ExcludeLabels: getEnvStringSlice("ENRICHMENT_LABELS_EXCLUDE", []string{
				"__name__", "job", "instance", "__meta_kubernetes_pod_container_port_name",
				"__meta_kubernetes_pod_container_port_number", "__meta_kubernetes_pod_container_port_protocol",
				"__meta_kubernetes_pod_ready", "__meta_kubernetes_pod_phase", "__meta_kubernetes_pod_ip",
				"__meta_kubernetes_pod_host_ip", "__meta_kubernetes_pod_node_name", "__meta_kubernetes_pod_uid",
				"__meta_kubernetes_namespace", "__meta_kubernetes_service_port_name",
				"__meta_kubernetes_service_port_number", "__meta_kubernetes_service_port_protocol",
				"__meta_kubernetes_service_cluster_ip", "__meta_kubernetes_service_external_name",
				"__meta_kubernetes_service_type", "__meta_kubernetes_ingress_scheme",
				"__meta_kubernetes_ingress_host", "__meta_kubernetes_ingress_path",
				"__meta_kubernetes_ingress_class_name",
			}),
		},
		Annotations: AnnotationEnrichmentConfig{
			Enabled:            getEnvBool("ENRICHMENT_ANNOTATIONS_ENABLED", true),
			DisplayFormat:      getEnvEnum("ENRICHMENT_ANNOTATIONS_DISPLAY_FORMAT", []string{"table", "json"}, "table"),
			IncludeAnnotations: getEnvStringSlice("ENRICHMENT_ANNOTATIONS_INCLUDE", []string{}),
			ExcludeAnnotations: getEnvStringSlice("ENRICHMENT_ANNOTATIONS_EXCLUDE", []string{
				"kubectl.kubernetes.io/last-applied-configuration", "deployment.kubernetes.io/revision",
				"control-plane.alpha.kubernetes.io/leader", "prometheus.io/scrape",
				"prometheus.io/port", "prometheus.io/path",
			}),
		},
	}
}
