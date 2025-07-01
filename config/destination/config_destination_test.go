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

func TestLoadDestinationsConfig_WithPlaceholders(t *testing.T) {
	// Set environment variable for testing
	envVarName := "SLACK_API_KEY_TEST"
	envVarValue := "xoxb-env-token"
	os.Setenv(envVarName, envVarValue)
	defer os.Unsetenv(envVarName)

	// Sample YAML config with placeholder
	yamlContent := `
destinations:
  slack:
    - name: "incident-alerts"
      api_key: "${SLACK_API_KEY_TEST}"
      slack_channel: "#incident-alerts"
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

	// Validate that placeholder was replaced with env var value
	assert.Len(t, cfg.Destinations.Slack, 1)
	slackDest := cfg.Destinations.Slack[0]
	assert.Equal(t, "incident-alerts", slackDest.Name)
	assert.Equal(t, envVarValue, slackDest.APIKey)
	assert.Equal(t, "#incident-alerts", slackDest.SlackChannel)
}

func TestLoadDestinationsConfig_MissingEnvVar(t *testing.T) {
	// Use a placeholder for an env var that doesn't exist
	nonExistentEnvVar := "SLACK_API_KEY_NONEXISTENT_" + t.Name()

	// Make sure the env var doesn't exist
	os.Unsetenv(nonExistentEnvVar)

	yamlContent := `
destinations:
  slack:
    - name: "incident-alerts"
      api_key: "${` + nonExistentEnvVar + `}"
      slack_channel: "#incident-alerts"
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
	assert.True(t, strings.Contains(err.Error(), "missing required env"))
	assert.True(t, strings.Contains(err.Error(), nonExistentEnvVar))
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

func TestValidateSlackDestination_WithPlaceholder(t *testing.T) {
	dest := DestinationSlack{
		Name:         "test",
		APIKey:       "${SLACK_API_KEY_TEST}",
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
