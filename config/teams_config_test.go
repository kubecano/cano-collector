package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestFileTeamsLoader_Load_ValidConfig(t *testing.T) {
	tempDir := t.TempDir()
	configContent := `
teams:
  - name: devops
    destinations:
      - slack-dev
      - teams-dev
  - name: backend
    destinations:
      - slack-backend
`
	configPath := filepath.Join(tempDir, "teams.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	require.NoError(t, err)

	loader := NewFileTeamsLoader(configPath)
	cfg, err := loader.Load()

	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Len(t, cfg.Teams, 2)

	assert.Equal(t, "devops", cfg.Teams[0].Name)
	assert.ElementsMatch(t, []string{"slack-dev", "teams-dev"}, cfg.Teams[0].Destinations)

	assert.Equal(t, "backend", cfg.Teams[1].Name)
	assert.ElementsMatch(t, []string{"slack-backend"}, cfg.Teams[1].Destinations)
}

func TestFileTeamsLoader_Load_InvalidYAML(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "invalid.yaml")

	// Invalid YAML content
	err := os.WriteFile(configPath, []byte("teams: [invalid yaml"), 0o644)
	require.NoError(t, err)

	loader := NewFileTeamsLoader(configPath)
	cfg, err := loader.Load()

	require.Error(t, err)
	assert.Nil(t, cfg)
}

func TestFileTeamsLoader_Load_FileNotFound(t *testing.T) {
	loader := NewFileTeamsLoader("nonexistent.yaml")
	cfg, err := loader.Load()

	require.Error(t, err)
	assert.Nil(t, cfg)
}
