package teams

import (
	"os"

	"gopkg.in/yaml.v3"
)

// TeamsLoader defines the interface for loading team configuration
//
//go:generate mockgen -destination=../../mocks/teams_loader_mock.go -package=mocks github.com/kubecano/cano-collector/config/teams TeamsLoader
type TeamsLoader interface {
	Load() (*TeamsConfig, error)
}

// FileTeamsLoader loads team config from a YAML file
type FileTeamsLoader struct {
	configPath string
}

func NewFileTeamsLoader(configPath string) *FileTeamsLoader {
	return &FileTeamsLoader{configPath: configPath}
}

func (f *FileTeamsLoader) Load() (*TeamsConfig, error) {
	var config TeamsConfig

	data, err := os.ReadFile(f.configPath)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// TeamsConfig represents the top-level YAML structure
type TeamsConfig struct {
	Teams []Team `yaml:"teams"`
}

type Team struct {
	Name         string   `yaml:"name"`
	Destinations []string `yaml:"destinations"`
}
