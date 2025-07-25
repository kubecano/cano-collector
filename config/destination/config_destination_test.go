package config_destination

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestLoadDestinationsConfig_Success(t *testing.T) {
	// Sample YAML config content
	yamlContent := `
destinations:
  slack:
    - name: "incident-alerts"
      api_key: "xoxb-slack-token"
      slack_channel: "#incident-alerts"
      grouping_interval: 30
      unfurl_links: true
  teams:
    - name: "infra-team"
      webhookURL: "https://outlook.office.com/webhook/YYY"
`

	// Create temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "destinations.yaml")
	err := os.WriteFile(tmpFile, []byte(yamlContent), 0o644)
	require.NoError(t, err)

	loader := NewFileDestinationsLoader(tmpFile)
	cfg, err := loader.Load()
	require.NoError(t, err)
	assert.NotNil(t, cfg)

	// Validate Slack destination
	assert.Len(t, cfg.Destinations.Slack, 1)
	slackDest := cfg.Destinations.Slack[0]
	assert.Equal(t, "incident-alerts", slackDest.Name)
	assert.Equal(t, "xoxb-slack-token", slackDest.APIKey)
	assert.Equal(t, "#incident-alerts", slackDest.SlackChannel)
	assert.Equal(t, 30, slackDest.GroupingInterval)
	assert.True(t, *slackDest.UnfurlLinks)
}

func TestLoadDestinationsConfig_FileNotFound(t *testing.T) {
	loader := NewFileDestinationsLoader("non-existent.yaml")
	cfg, err := loader.Load()
	require.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoadDestinationsConfig_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "invalid.yaml")
	err := os.WriteFile(tmpFile, []byte("this is not: valid: yaml"), 0o644)
	require.NoError(t, err)

	loader := NewFileDestinationsLoader(tmpFile)
	cfg, err := loader.Load()
	require.Error(t, err)
	assert.Nil(t, cfg)
}

func TestValidateSlackDestination_Success(t *testing.T) {
	dest := DestinationSlack{
		Name:         "test",
		APIKey:       "xoxb-token",
		SlackChannel: "#test",
	}
	err := validateSlackDestination(dest)
	assert.NoError(t, err)
}

func TestValidateSlackDestination_MissingName(t *testing.T) {
	dest := DestinationSlack{
		APIKey:       "xoxb-token",
		SlackChannel: "#test",
	}
	err := validateSlackDestination(dest)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

func TestValidateSlackDestination_MissingChannel(t *testing.T) {
	dest := DestinationSlack{
		Name:   "test",
		APIKey: "xoxb-token",
	}
	err := validateSlackDestination(dest)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "slack_channel is required")
}

func TestValidateSlackDestination_MissingAPIKey(t *testing.T) {
	dest := DestinationSlack{
		Name:         "test",
		SlackChannel: "#test",
	}
	err := validateSlackDestination(dest)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "api_key is required")
}

func TestValidateSlackDestination_NegativeGroupingInterval(t *testing.T) {
	dest := DestinationSlack{
		Name:             "test",
		APIKey:           "xoxb-token",
		SlackChannel:     "#test",
		GroupingInterval: -1,
	}
	err := validateSlackDestination(dest)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "grouping_interval must be non-negative")
}

func TestValidateSlackDestination_PlaceholderAPIKey(t *testing.T) {
	dest := DestinationSlack{
		Name:         "test",
		APIKey:       "${SLACK_API_KEY_TEST}",
		SlackChannel: "#test",
	}
	err := validateSlackDestination(dest)
	assert.NoError(t, err, "placeholder API keys should pass validation")
}

func TestLoadDestinationsConfig_EnvironmentVariableSubstitution(t *testing.T) {
	// Set up environment variable
	envVar := "SLACK_API_KEY_PROD"
	expectedToken := "xoxb-env-token"
	err := os.Setenv(envVar, expectedToken)
	require.NoError(t, err)
	defer os.Unsetenv(envVar)

	// Sample YAML config content with placeholder
	yamlContent := `
destinations:
  slack:
    - name: "prod-alerts"
      api_key: "${SLACK_API_KEY_PROD}"
      slack_channel: "#prod-alerts"
`

	// Create temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "destinations.yaml")
	err = os.WriteFile(tmpFile, []byte(yamlContent), 0o644)
	require.NoError(t, err)

	loader := NewFileDestinationsLoader(tmpFile)
	cfg, err := loader.Load()
	require.NoError(t, err)
	assert.NotNil(t, cfg)

	// Validate that placeholder was replaced with environment variable value
	assert.Len(t, cfg.Destinations.Slack, 1)
	slackDest := cfg.Destinations.Slack[0]
	assert.Equal(t, "prod-alerts", slackDest.Name)
	assert.Equal(t, expectedToken, slackDest.APIKey)
	assert.Equal(t, "#prod-alerts", slackDest.SlackChannel)
}

func TestLoadDestinationsConfig_MissingEnvironmentVariable(t *testing.T) {
	// Sample YAML config content with placeholder for non-existent env var
	yamlContent := `
destinations:
  slack:
    - name: "test-alerts"
      api_key: "${SLACK_API_KEY_NONEXISTENT}"
      slack_channel: "#test-alerts"
`

	// Create temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "destinations.yaml")
	err := os.WriteFile(tmpFile, []byte(yamlContent), 0o644)
	require.NoError(t, err)

	loader := NewFileDestinationsLoader(tmpFile)
	cfg, err := loader.Load()
	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "missing required env SLACK_API_KEY_NONEXISTENT for slack destination test-alerts")
}

func TestParseDestinationsYAML_MixedPlaceholderAndDirectAPIKeys(t *testing.T) {
	// Set up environment variable for one destination
	envVar := "SLACK_API_KEY_PROD"
	expectedToken := "xoxb-prod-token"
	err := os.Setenv(envVar, expectedToken)
	require.NoError(t, err)
	defer os.Unsetenv(envVar)

	// Sample YAML config content with both placeholder and direct API key
	yamlContent := `
destinations:
  slack:
    - name: "dev-alerts"
      api_key: "xoxb-dev-token"
      slack_channel: "#dev-alerts"
    - name: "prod-alerts"
      api_key: "${SLACK_API_KEY_PROD}"
      slack_channel: "#prod-alerts"
`

	cfg, err := parseDestinationsYAML(strings.NewReader(yamlContent))
	require.NoError(t, err)
	assert.NotNil(t, cfg)

	// Validate both destinations
	assert.Len(t, cfg.Destinations.Slack, 2)

	// Dev destination should have direct API key
	devDest := cfg.Destinations.Slack[0]
	assert.Equal(t, "dev-alerts", devDest.Name)
	assert.Equal(t, "xoxb-dev-token", devDest.APIKey)
	assert.Equal(t, "#dev-alerts", devDest.SlackChannel)

	// Prod destination should have substituted API key
	prodDest := cfg.Destinations.Slack[1]
	assert.Equal(t, "prod-alerts", prodDest.Name)
	assert.Equal(t, expectedToken, prodDest.APIKey)
	assert.Equal(t, "#prod-alerts", prodDest.SlackChannel)
}

func TestLoadDestinationsConfig_WithThreadingAndEnrichments(t *testing.T) {
	// Sample YAML config with threading and enrichments configuration
	yamlContent := `
destinations:
  slack:
    - name: "enhanced-alerts"
      api_key: "xoxb-enhanced-token"
      slack_channel: "#enhanced-alerts"
      threading:
        enabled: true
        cache_ttl: "15m"
        search_limit: 150
        search_window: "48h"
        fingerprint_in_metadata: true
      enrichments:
        format_as_blocks: true
        color_coding: true
        table_formatting: "enhanced"
        max_table_rows: 25
        attachment_threshold: 1500
`

	// Create temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "destinations.yaml")
	err := os.WriteFile(tmpFile, []byte(yamlContent), 0o644)
	require.NoError(t, err)

	loader := NewFileDestinationsLoader(tmpFile)
	cfg, err := loader.Load()
	require.NoError(t, err)
	assert.NotNil(t, cfg)

	// Validate Slack destination with enhanced config
	assert.Len(t, cfg.Destinations.Slack, 1)
	slackDest := cfg.Destinations.Slack[0]
	assert.Equal(t, "enhanced-alerts", slackDest.Name)
	assert.Equal(t, "xoxb-enhanced-token", slackDest.APIKey)
	assert.Equal(t, "#enhanced-alerts", slackDest.SlackChannel)

	// Validate threading configuration
	assert.NotNil(t, slackDest.Threading)
	assert.True(t, slackDest.Threading.Enabled)
	assert.Equal(t, "15m", slackDest.Threading.CacheTTL)
	assert.Equal(t, 150, slackDest.Threading.SearchLimit)
	assert.Equal(t, "48h", slackDest.Threading.SearchWindow)
	assert.NotNil(t, slackDest.Threading.FingerprintInMetadata)
	assert.True(t, *slackDest.Threading.FingerprintInMetadata)

	// Validate enrichments configuration
	assert.NotNil(t, slackDest.Enrichments)
	assert.NotNil(t, slackDest.Enrichments.FormatAsBlocks)
	assert.True(t, *slackDest.Enrichments.FormatAsBlocks)
	assert.NotNil(t, slackDest.Enrichments.ColorCoding)
	assert.True(t, *slackDest.Enrichments.ColorCoding)
	assert.Equal(t, "enhanced", slackDest.Enrichments.TableFormatting)
	assert.Equal(t, 25, slackDest.Enrichments.MaxTableRows)
	assert.Equal(t, 1500, slackDest.Enrichments.AttachmentThreshold)
}

func TestLoadDestinationsConfig_WithDefaults(t *testing.T) {
	// Sample YAML config with minimal threading/enrichments configuration
	yamlContent := `
destinations:
  slack:
    - name: "default-alerts"
      api_key: "xoxb-default-token"
      slack_channel: "#default-alerts"
      threading:
        enabled: true
      enrichments: {}
`

	// Create temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "destinations.yaml")
	err := os.WriteFile(tmpFile, []byte(yamlContent), 0o644)
	require.NoError(t, err)

	loader := NewFileDestinationsLoader(tmpFile)
	cfg, err := loader.Load()
	require.NoError(t, err)
	assert.NotNil(t, cfg)

	// Validate defaults were applied
	slackDest := cfg.Destinations.Slack[0]

	// Threading defaults
	assert.Equal(t, "10m", slackDest.Threading.CacheTTL)
	assert.Equal(t, 100, slackDest.Threading.SearchLimit)
	assert.Equal(t, "24h", slackDest.Threading.SearchWindow)
	assert.NotNil(t, slackDest.Threading.FingerprintInMetadata)
	assert.True(t, *slackDest.Threading.FingerprintInMetadata)

	// Enrichments defaults
	assert.NotNil(t, slackDest.Enrichments.FormatAsBlocks)
	assert.True(t, *slackDest.Enrichments.FormatAsBlocks)
	assert.NotNil(t, slackDest.Enrichments.ColorCoding)
	assert.True(t, *slackDest.Enrichments.ColorCoding)
	assert.Equal(t, "enhanced", slackDest.Enrichments.TableFormatting)
	assert.Equal(t, 20, slackDest.Enrichments.MaxTableRows)
	assert.Equal(t, 1000, slackDest.Enrichments.AttachmentThreshold)
}

func TestValidateThreadingConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  SlackThreadingConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: SlackThreadingConfig{
				Enabled:               true,
				CacheTTL:              "10m",
				SearchLimit:           100,
				SearchWindow:          "24h",
				FingerprintInMetadata: &[]bool{true}[0],
			},
			wantErr: false,
		},
		{
			name: "valid complex duration formats",
			config: SlackThreadingConfig{
				CacheTTL:     "1h30m",
				SearchWindow: "2h45m30s",
			},
			wantErr: false,
		},
		{
			name: "valid day suffix",
			config: SlackThreadingConfig{
				SearchWindow: "7d",
			},
			wantErr: false,
		},
		{
			name: "negative search limit",
			config: SlackThreadingConfig{
				SearchLimit: -1,
			},
			wantErr: true,
			errMsg:  "search_limit must be non-negative",
		},
		{
			name: "search limit too high",
			config: SlackThreadingConfig{
				SearchLimit: 1001,
			},
			wantErr: true,
			errMsg:  "search_limit must not exceed 1000",
		},
		{
			name: "invalid cache_ttl format",
			config: SlackThreadingConfig{
				CacheTTL: "invalid",
			},
			wantErr: true,
			errMsg:  "cache_ttl must be a valid duration",
		},
		{
			name: "invalid cache_ttl with prefix text",
			config: SlackThreadingConfig{
				CacheTTL: "abc10m",
			},
			wantErr: true,
			errMsg:  "cache_ttl must be a valid duration",
		},
		{
			name: "invalid search_window format",
			config: SlackThreadingConfig{
				SearchWindow: "xyz",
			},
			wantErr: true,
			errMsg:  "search_window validation failed",
		},
		{
			name: "invalid search_window with prefix text",
			config: SlackThreadingConfig{
				SearchWindow: "test24h",
			},
			wantErr: true,
			errMsg:  "search_window validation failed",
		},
		{
			name: "invalid day suffix without number",
			config: SlackThreadingConfig{
				SearchWindow: "d",
			},
			wantErr: true,
			errMsg:  "search_window validation failed",
		},
		{
			name: "invalid day suffix with non-numeric prefix",
			config: SlackThreadingConfig{
				SearchWindow: "abcd",
			},
			wantErr: true,
			errMsg:  "search_window validation failed",
		},
		{
			name: "negative search_window hours",
			config: SlackThreadingConfig{
				SearchWindow: "-24h",
			},
			wantErr: true,
			errMsg:  "search_window must be positive",
		},
		{
			name: "negative search_window days",
			config: SlackThreadingConfig{
				SearchWindow: "-7d",
			},
			wantErr: true,
			errMsg:  "search_window must be positive",
		},
		{
			name: "negative search_window complex",
			config: SlackThreadingConfig{
				SearchWindow: "-1h30m",
			},
			wantErr: true,
			errMsg:  "search_window must be positive",
		},
		{
			name: "zero search_window hours",
			config: SlackThreadingConfig{
				SearchWindow: "0h",
			},
			wantErr: true,
			errMsg:  "search_window must be positive",
		},
		{
			name: "zero search_window days",
			config: SlackThreadingConfig{
				SearchWindow: "0d",
			},
			wantErr: true,
			errMsg:  "search_window must be positive",
		},
		{
			name: "zero search_window seconds",
			config: SlackThreadingConfig{
				SearchWindow: "0s",
			},
			wantErr: true,
			errMsg:  "search_window must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateThreadingConfig(tt.config)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestParseAndValidateSearchWindow(t *testing.T) {
	tests := []struct {
		name             string
		searchWindow     string
		expectedDuration string // Expected duration as string (for easier comparison)
		wantErr          bool
		errMsg           string
	}{
		{
			name:             "valid hours",
			searchWindow:     "24h",
			expectedDuration: "24h0m0s",
			wantErr:          false,
		},
		{
			name:             "valid complex duration",
			searchWindow:     "1h30m45s",
			expectedDuration: "1h30m45s",
			wantErr:          false,
		},
		{
			name:             "valid single day converts to 24 hours",
			searchWindow:     "1d",
			expectedDuration: "24h0m0s",
			wantErr:          false,
		},
		{
			name:             "valid multiple days converts correctly",
			searchWindow:     "7d",
			expectedDuration: "168h0m0s", // 7 * 24 = 168 hours
			wantErr:          false,
		},
		{
			name:             "valid fractional day converts correctly",
			searchWindow:     "0.5d",
			expectedDuration: "12h0m0s", // 0.5 * 24 = 12 hours
			wantErr:          false,
		},
		{
			name:             "negative hours - parsed but not validated",
			searchWindow:     "-24h",
			expectedDuration: "-24h0m0s",
			wantErr:          false, // parseAndValidateSearchWindow doesn't check positivity
		},
		{
			name:             "negative days - parsed but not validated",
			searchWindow:     "-7d",
			expectedDuration: "-168h0m0s", // -7 * 24 = -168 hours
			wantErr:          false,       // parseAndValidateSearchWindow doesn't check positivity
		},
		{
			name:             "zero hours - parsed but not validated",
			searchWindow:     "0h",
			expectedDuration: "0s",
			wantErr:          false, // parseAndValidateSearchWindow doesn't check positivity
		},
		{
			name:             "zero days - parsed but not validated",
			searchWindow:     "0d",
			expectedDuration: "0s",  // 0 * 24 = 0 hours
			wantErr:          false, // parseAndValidateSearchWindow doesn't check positivity
		},
		{
			name:         "empty day suffix",
			searchWindow: "d",
			wantErr:      true,
			errMsg:       "must be a valid duration", // Falls back to standard parsing
		},
		{
			name:         "invalid format",
			searchWindow: "invalid",
			wantErr:      true,
			errMsg:       "must be a valid duration",
		},
		{
			name:         "invalid day suffix with non-numeric",
			searchWindow: "abcd",
			wantErr:      true,
			errMsg:       "must be a valid duration", // Falls back to standard parsing
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			duration, err := parseAndValidateSearchWindow(tt.searchWindow)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedDuration, duration.String())
			}
		})
	}
}

func TestValidateEnrichmentsConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  SlackEnrichmentsConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: SlackEnrichmentsConfig{
				FormatAsBlocks:      &[]bool{true}[0],
				ColorCoding:         &[]bool{true}[0],
				TableFormatting:     "enhanced",
				MaxTableRows:        20,
				AttachmentThreshold: 1000,
			},
			wantErr: false,
		},
		{
			name: "invalid table formatting",
			config: SlackEnrichmentsConfig{
				TableFormatting: "invalid",
			},
			wantErr: true,
			errMsg:  "table_formatting must be one of: simple, enhanced, attachment",
		},
		{
			name: "negative max table rows",
			config: SlackEnrichmentsConfig{
				MaxTableRows: -1,
			},
			wantErr: true,
			errMsg:  "max_table_rows must be non-negative",
		},
		{
			name: "negative attachment threshold",
			config: SlackEnrichmentsConfig{
				AttachmentThreshold: -1,
			},
			wantErr: true,
			errMsg:  "attachment_threshold must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEnrichmentsConfig(tt.config)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestLoadDestinationsConfig_ExplicitFalseValues(t *testing.T) {
	// Test case to verify that explicitly set false values are respected, not overridden by defaults
	yamlContent := `
destinations:
  slack:
    - name: "explicit-false-config"
      api_key: "xoxb-test-token"
      slack_channel: "#test"
      threading:
        enabled: true
        fingerprint_in_metadata: false
      enrichments:
        format_as_blocks: false
        color_coding: false
        table_formatting: "simple"
`

	// Create temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "destinations.yaml")
	err := os.WriteFile(tmpFile, []byte(yamlContent), 0o644)
	require.NoError(t, err)

	loader := NewFileDestinationsLoader(tmpFile)
	cfg, err := loader.Load()
	require.NoError(t, err)
	assert.NotNil(t, cfg)

	// Validate that explicitly set false values are preserved
	slackDest := cfg.Destinations.Slack[0]

	// Threading configuration - fingerprint_in_metadata explicitly set to false should remain false
	assert.NotNil(t, slackDest.Threading)
	assert.True(t, slackDest.Threading.Enabled)
	assert.NotNil(t, slackDest.Threading.FingerprintInMetadata)
	assert.False(t, *slackDest.Threading.FingerprintInMetadata, "Explicitly set false value should be preserved")

	// Enrichments configuration - both fields explicitly set to false should remain false
	assert.NotNil(t, slackDest.Enrichments)
	assert.NotNil(t, slackDest.Enrichments.FormatAsBlocks)
	assert.False(t, *slackDest.Enrichments.FormatAsBlocks, "Explicitly set false value should be preserved")
	assert.NotNil(t, slackDest.Enrichments.ColorCoding)
	assert.False(t, *slackDest.Enrichments.ColorCoding, "Explicitly set false value should be preserved")

	// Other fields should still get defaults
	assert.Equal(t, "simple", slackDest.Enrichments.TableFormatting)
	assert.Equal(t, 20, slackDest.Enrichments.MaxTableRows)
	assert.Equal(t, 1000, slackDest.Enrichments.AttachmentThreshold)
}
