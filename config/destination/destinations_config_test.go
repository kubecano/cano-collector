package destination

import (
	"os"
	"path/filepath"
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

	// Validate Teams destination
	assert.Len(t, cfg.Destinations.Teams, 1)
	teamsDest := cfg.Destinations.Teams[0]
	assert.Equal(t, "infra-team", teamsDest.Name)
	assert.Equal(t, "https://outlook.office.com/webhook/YYY", teamsDest.WebhookURL)
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
	dest := SlackDestination{
		Name:         "test",
		APIKey:       "xoxb-token",
		SlackChannel: "#test",
	}
	err := validateSlackDestination(dest)
	assert.NoError(t, err)
}

func TestValidateSlackDestination_MissingName(t *testing.T) {
	dest := SlackDestination{
		APIKey:       "xoxb-token",
		SlackChannel: "#test",
	}
	err := validateSlackDestination(dest)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

func TestValidateSlackDestination_MissingChannel(t *testing.T) {
	dest := SlackDestination{
		Name:   "test",
		APIKey: "xoxb-token",
	}
	err := validateSlackDestination(dest)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "slack_channel is required")
}

func TestValidateSlackDestination_MissingAPIKey(t *testing.T) {
	dest := SlackDestination{
		Name:         "test",
		SlackChannel: "#test",
	}
	err := validateSlackDestination(dest)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "api_key is required")
}

func TestValidateSlackDestination_NegativeGroupingInterval(t *testing.T) {
	dest := SlackDestination{
		Name:             "test",
		APIKey:           "xoxb-token",
		SlackChannel:     "#test",
		GroupingInterval: -1,
	}
	err := validateSlackDestination(dest)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "grouping_interval must be non-negative")
}

func TestValidateTeamsDestination_Success(t *testing.T) {
	dest := TeamsDestination{
		Name:       "test",
		WebhookURL: "https://outlook.office.com/webhook/XXX",
	}
	err := validateTeamsDestination(dest)
	assert.NoError(t, err)
}

func TestValidateTeamsDestination_MissingName(t *testing.T) {
	dest := TeamsDestination{
		WebhookURL: "https://outlook.office.com/webhook/XXX",
	}
	err := validateTeamsDestination(dest)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

func TestValidateTeamsDestination_MissingWebhookURL(t *testing.T) {
	dest := TeamsDestination{
		Name: "test",
	}
	err := validateTeamsDestination(dest)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "webhookURL is required")
}
