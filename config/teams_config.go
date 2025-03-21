package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type TeamsConfig struct {
	Teams []Team `yaml:"teams"`
}

type Team struct {
	Name         string   `yaml:"name"`
	Destinations []string `yaml:"destinations"`
}

// LoadTeamsConfig loads the Teams configuration from the given file
func LoadTeamsConfig(configPath string) (*TeamsConfig, error) {
	var config TeamsConfig

	configData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
