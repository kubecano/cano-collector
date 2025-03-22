package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"

	"github.com/kubecano/cano-collector/config"
)

func TestLoadDestinationsConfig_Success(t *testing.T) {
	// Sample YAML config content
	yamlContent := `
destinations:
  slack:
    - name: "incident-alerts"
      webhookURL: "https://hooks.slack.com/services/XXX"
  teams:
    - name: "infra-team"
      webhookURL: "https://outlook.office.com/webhook/YYY"
`

	// Create temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "destinations.yaml")
	err := os.WriteFile(tmpFile, []byte(yamlContent), 0o644)
	require.NoError(t, err)

	loader := config.NewFileDestinationsLoader(tmpFile)
	cfg, err := loader.Load()
	require.NoError(t, err)
	assert.NotNil(t, cfg)

	// Validate Slack destination
	assert.Len(t, cfg.Destinations.Slack, 1)
	assert.Equal(t, "incident-alerts", cfg.Destinations.Slack[0].Name)
	assert.Equal(t, "https://hooks.slack.com/services/XXX", cfg.Destinations.Slack[0].WebhookURL)

	// Validate Teams destination
	assert.Len(t, cfg.Destinations.Teams, 1)
	assert.Equal(t, "infra-team", cfg.Destinations.Teams[0].Name)
	assert.Equal(t, "https://outlook.office.com/webhook/YYY", cfg.Destinations.Teams[0].WebhookURL)
}

func TestLoadDestinationsConfig_FileNotFound(t *testing.T) {
	loader := config.NewFileDestinationsLoader("non-existent.yaml")
	cfg, err := loader.Load()
	require.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoadDestinationsConfig_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "invalid.yaml")
	err := os.WriteFile(tmpFile, []byte("this is not: valid: yaml"), 0o644)
	require.NoError(t, err)

	loader := config.NewFileDestinationsLoader(tmpFile)
	cfg, err := loader.Load()
	require.Error(t, err)
	assert.Nil(t, cfg)
}
