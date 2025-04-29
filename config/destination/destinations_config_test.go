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
      token: "xoxb-1234567890-0987654321-ABCDEF"
      channel: "#alerts"
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
	assert.Equal(t, "incident-alerts", cfg.Destinations.Slack[0].Name)
	assert.Equal(t, "xoxb-1234567890-0987654321-ABCDEF", cfg.Destinations.Slack[0].Token)
	assert.Equal(t, "#alerts", cfg.Destinations.Slack[0].Channel)

	// Validate Teams destination
	assert.Len(t, cfg.Destinations.Teams, 1)
	assert.Equal(t, "infra-team", cfg.Destinations.Teams[0].Name)
	assert.Equal(t, "https://outlook.office.com/webhook/YYY", cfg.Destinations.Teams[0].WebhookURL)
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
