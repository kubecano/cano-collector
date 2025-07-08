package config

import (
	"os"
	"testing"

	"github.com/golang/mock/gomock"

	config_destination "github.com/kubecano/cano-collector/config/destination"
	config_team "github.com/kubecano/cano-collector/config/team"
	"github.com/kubecano/cano-collector/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestLoader(t *testing.T) (Config, error) {
	t.Helper()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	destinationsConfig := config_destination.DestinationsConfig{
		Destinations: struct {
			Slack []config_destination.DestinationSlack `yaml:"slack"`
		}{
			Slack: []config_destination.DestinationSlack{
				{
					Name:         "alerts",
					APIKey:       "xoxb-slack-token",
					SlackChannel: "#alerts",
				},
			},
		},
	}
	mockDestinations := mocks.NewMockDestinationsLoader(ctrl)
	mockDestinations.EXPECT().Load().AnyTimes().Return(&destinationsConfig, nil)

	teamsConfig := config_team.TeamsConfig{
		Teams: []config_team.Team{
			{Name: "devops", Destinations: []string{"alerts"}},
		},
	}
	mockTeams := mocks.NewMockTeamsLoader(ctrl)
	mockTeams.EXPECT().Load().AnyTimes().Return(&teamsConfig, nil)

	mockLoader := mocks.NewMockFullConfigLoader(ctrl)
	mockLoader.EXPECT().Load().AnyTimes().Return(destinationsConfig, teamsConfig, nil)

	return LoadConfigWithLoader(mockLoader)
}

func TestLoadConfigWithLoader(t *testing.T) {
	_ = os.Setenv("APP_NAME", "test-app")
	_ = os.Setenv("LOG_LEVEL", "debug")
	_ = os.Setenv("SENTRY_DSN", "https://example@sentry.io/123")
	_ = os.Setenv("ENABLE_TELEMETRY", "true")

	t.Cleanup(func() {
		_ = os.Unsetenv("APP_NAME")
		_ = os.Unsetenv("LOG_LEVEL")
		_ = os.Unsetenv("SENTRY_DSN")
		_ = os.Unsetenv("ENABLE_TELEMETRY")
	})

	cfg, err := setupTestLoader(t)
	require.NoError(t, err)

	assert.Equal(t, "test-app", cfg.AppName)
	assert.Equal(t, "debug", cfg.LogLevel)
	assert.Equal(t, "https://example@sentry.io/123", cfg.SentryDSN)
	assert.True(t, cfg.SentryEnabled)
	assert.Len(t, cfg.Destinations.Destinations.Slack, 1)
	assert.Equal(t, "alerts", cfg.Destinations.Destinations.Slack[0].Name)
	assert.Len(t, cfg.Teams.Teams, 1)
	assert.Equal(t, "devops", cfg.Teams.Teams[0].Name)
}

func TestGetEnvString(t *testing.T) {
	_ = os.Setenv("TEST_STRING", "value1")
	t.Cleanup(func() {
		_ = os.Unsetenv("TEST_STRING")
	})

	assert.Equal(t, "value1", getEnvString("TEST_STRING", "default"))
	assert.Equal(t, "default", getEnvString("NON_EXISTENT_STRING", "default"))
}

func TestGetEnvBool(t *testing.T) {
	_ = os.Setenv("TEST_BOOL_TRUE", "true")
	_ = os.Setenv("TEST_BOOL_FALSE", "false")
	_ = os.Setenv("TEST_BOOL_INVALID", "invalid")

	t.Cleanup(func() {
		_ = os.Unsetenv("TEST_BOOL_TRUE")
		_ = os.Unsetenv("TEST_BOOL_FALSE")
		_ = os.Unsetenv("TEST_BOOL_INVALID")
	})

	assert.True(t, getEnvBool("TEST_BOOL_TRUE", false))
	assert.False(t, getEnvBool("TEST_BOOL_FALSE", true))
	assert.True(t, getEnvBool("NON_EXISTENT_BOOL", true))
	assert.False(t, getEnvBool("TEST_BOOL_INVALID", false))
}

func TestGetEnvEnum(t *testing.T) {
	allowed := []string{"disabled", "local", "remote"}

	_ = os.Setenv("TEST_ENUM_VALID", "local")
	_ = os.Setenv("TEST_ENUM_INVALID", "xxx")

	t.Cleanup(func() {
		_ = os.Unsetenv("TEST_ENUM_VALID")
		_ = os.Unsetenv("TEST_ENUM_INVALID")
	})

	assert.Equal(t, "local", getEnvEnum("TEST_ENUM_VALID", allowed, "disabled"))
	assert.Equal(t, "disabled", getEnvEnum("TEST_ENUM_INVALID", allowed, "disabled"))
	assert.Equal(t, "disabled", getEnvEnum("NON_EXISTENT_ENUM", allowed, "disabled"))
}

func TestLoadConfigWithLoader_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockErr := assert.AnError

	mockLoader := mocks.NewMockFullConfigLoader(ctrl)
	mockLoader.EXPECT().Load().AnyTimes().Return(config_destination.DestinationsConfig{}, config_team.TeamsConfig{}, mockErr)

	cfg, err := LoadConfigWithLoader(mockLoader)

	require.Error(t, err, "Expected error when loader fails")
	assert.Equal(t, Config{}, cfg, "Expected empty config on loader failure")
}

func TestGetEnvStringSlice(t *testing.T) {
	_ = os.Setenv("TEST_STRING_SLICE_COMMA", "value1,value2,value3")
	_ = os.Setenv("TEST_STRING_SLICE_SINGLE", "single")
	_ = os.Setenv("TEST_STRING_SLICE_EMPTY", "")

	t.Cleanup(func() {
		_ = os.Unsetenv("TEST_STRING_SLICE_COMMA")
		_ = os.Unsetenv("TEST_STRING_SLICE_SINGLE")
		_ = os.Unsetenv("TEST_STRING_SLICE_EMPTY")
	})

	assert.Equal(t, []string{"value1", "value2", "value3"}, getEnvStringSlice("TEST_STRING_SLICE_COMMA", []string{}))
	assert.Equal(t, []string{"single"}, getEnvStringSlice("TEST_STRING_SLICE_SINGLE", []string{}))
	assert.Equal(t, []string{"default"}, getEnvStringSlice("TEST_STRING_SLICE_EMPTY", []string{"default"}))
	assert.Equal(t, []string{"default"}, getEnvStringSlice("NON_EXISTENT_SLICE", []string{"default"}))
}

func TestLoadEnrichmentConfig(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected EnrichmentConfig
	}{
		{
			name:    "default configuration",
			envVars: map[string]string{},
			expected: EnrichmentConfig{
				Labels: LabelEnrichmentConfig{
					Enabled:       true,
					DisplayFormat: "table",
					IncludeLabels: []string{},
					ExcludeLabels: []string{
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
					},
				},
				Annotations: AnnotationEnrichmentConfig{
					Enabled:            true,
					DisplayFormat:      "table",
					IncludeAnnotations: []string{},
					ExcludeAnnotations: []string{
						"kubectl.kubernetes.io/last-applied-configuration", "deployment.kubernetes.io/revision",
						"control-plane.alpha.kubernetes.io/leader", "prometheus.io/scrape",
						"prometheus.io/port", "prometheus.io/path",
					},
				},
			},
		},
		{
			name: "custom configuration - labels disabled",
			envVars: map[string]string{
				"ENRICHMENT_LABELS_ENABLED":             "false",
				"ENRICHMENT_LABELS_DISPLAY_FORMAT":      "json",
				"ENRICHMENT_ANNOTATIONS_ENABLED":        "true",
				"ENRICHMENT_ANNOTATIONS_DISPLAY_FORMAT": "table",
			},
			expected: EnrichmentConfig{
				Labels: LabelEnrichmentConfig{
					Enabled:       false,
					DisplayFormat: "json",
					IncludeLabels: []string{},
					ExcludeLabels: []string{
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
					},
				},
				Annotations: AnnotationEnrichmentConfig{
					Enabled:            true,
					DisplayFormat:      "table",
					IncludeAnnotations: []string{},
					ExcludeAnnotations: []string{
						"kubectl.kubernetes.io/last-applied-configuration", "deployment.kubernetes.io/revision",
						"control-plane.alpha.kubernetes.io/leader", "prometheus.io/scrape",
						"prometheus.io/port", "prometheus.io/path",
					},
				},
			},
		},
		{
			name: "custom configuration - json format with custom filters",
			envVars: map[string]string{
				"ENRICHMENT_LABELS_ENABLED":             "true",
				"ENRICHMENT_LABELS_DISPLAY_FORMAT":      "json",
				"ENRICHMENT_LABELS_INCLUDE":             "alertname,severity",
				"ENRICHMENT_LABELS_EXCLUDE":             "job,instance",
				"ENRICHMENT_ANNOTATIONS_ENABLED":        "true",
				"ENRICHMENT_ANNOTATIONS_DISPLAY_FORMAT": "json",
				"ENRICHMENT_ANNOTATIONS_INCLUDE":        "summary,description",
				"ENRICHMENT_ANNOTATIONS_EXCLUDE":        "prometheus.io/scrape",
			},
			expected: EnrichmentConfig{
				Labels: LabelEnrichmentConfig{
					Enabled:       true,
					DisplayFormat: "json",
					IncludeLabels: []string{"alertname", "severity"},
					ExcludeLabels: []string{"job", "instance"},
				},
				Annotations: AnnotationEnrichmentConfig{
					Enabled:            true,
					DisplayFormat:      "json",
					IncludeAnnotations: []string{"summary", "description"},
					ExcludeAnnotations: []string{"prometheus.io/scrape"},
				},
			},
		},
		{
			name: "both disabled",
			envVars: map[string]string{
				"ENRICHMENT_LABELS_ENABLED":      "false",
				"ENRICHMENT_ANNOTATIONS_ENABLED": "false",
			},
			expected: EnrichmentConfig{
				Labels: LabelEnrichmentConfig{
					Enabled:       false,
					DisplayFormat: "table",
					IncludeLabels: []string{},
					ExcludeLabels: []string{
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
					},
				},
				Annotations: AnnotationEnrichmentConfig{
					Enabled:            false,
					DisplayFormat:      "table",
					IncludeAnnotations: []string{},
					ExcludeAnnotations: []string{
						"kubectl.kubernetes.io/last-applied-configuration", "deployment.kubernetes.io/revision",
						"control-plane.alpha.kubernetes.io/leader", "prometheus.io/scrape",
						"prometheus.io/port", "prometheus.io/path",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				_ = os.Setenv(key, value)
			}

			t.Cleanup(func() {
				for key := range tt.envVars {
					_ = os.Unsetenv(key)
				}
			})

			config := loadEnrichmentConfig()

			assert.Equal(t, tt.expected.Labels.Enabled, config.Labels.Enabled)
			assert.Equal(t, tt.expected.Labels.DisplayFormat, config.Labels.DisplayFormat)
			assert.Equal(t, tt.expected.Labels.IncludeLabels, config.Labels.IncludeLabels)
			assert.Equal(t, tt.expected.Labels.ExcludeLabels, config.Labels.ExcludeLabels)

			assert.Equal(t, tt.expected.Annotations.Enabled, config.Annotations.Enabled)
			assert.Equal(t, tt.expected.Annotations.DisplayFormat, config.Annotations.DisplayFormat)
			assert.Equal(t, tt.expected.Annotations.IncludeAnnotations, config.Annotations.IncludeAnnotations)
			assert.Equal(t, tt.expected.Annotations.ExcludeAnnotations, config.Annotations.ExcludeAnnotations)
		})
	}
}

func TestLoadConfigWithLoader_WithEnrichmentConfig(t *testing.T) {
	_ = os.Setenv("ENRICHMENT_LABELS_ENABLED", "false")
	_ = os.Setenv("ENRICHMENT_LABELS_DISPLAY_FORMAT", "json")
	_ = os.Setenv("ENRICHMENT_ANNOTATIONS_ENABLED", "true")
	_ = os.Setenv("ENRICHMENT_ANNOTATIONS_DISPLAY_FORMAT", "table")

	t.Cleanup(func() {
		_ = os.Unsetenv("ENRICHMENT_LABELS_ENABLED")
		_ = os.Unsetenv("ENRICHMENT_LABELS_DISPLAY_FORMAT")
		_ = os.Unsetenv("ENRICHMENT_ANNOTATIONS_ENABLED")
		_ = os.Unsetenv("ENRICHMENT_ANNOTATIONS_DISPLAY_FORMAT")
	})

	cfg, err := setupTestLoader(t)
	require.NoError(t, err)

	// Test enrichment configuration
	assert.False(t, cfg.Enrichment.Labels.Enabled)
	assert.Equal(t, "json", cfg.Enrichment.Labels.DisplayFormat)
	assert.True(t, cfg.Enrichment.Annotations.Enabled)
	assert.Equal(t, "table", cfg.Enrichment.Annotations.DisplayFormat)
}
