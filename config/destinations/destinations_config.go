package destinations

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type DestinationsConfig struct {
	Destinations struct {
		Slack []Destination `yaml:"slack"`
		Teams []Destination `yaml:"teams"`
	} `yaml:"destinations"`
}

type Destination struct {
	Name       string `yaml:"name"`
	WebhookURL string `yaml:"webhookURL"`
}

//go:generate mockgen -destination=../../mocks/destinations_loader_mock.go -package=mocks github.com/kubecano/cano-collector/config/destinations DestinationsLoader
type DestinationsLoader interface {
	Load() (*DestinationsConfig, error)
}

// FileDestinationsLoader loads destinations from a file or secret (ConfigMap/Secret mount)
type FileDestinationsLoader struct {
	Path string
}

func NewFileDestinationsLoader(path string) *FileDestinationsLoader {
	return &FileDestinationsLoader{Path: path}
}

func (f *FileDestinationsLoader) Load() (*DestinationsConfig, error) {
	file, err := os.Open(f.Path)
	if err != nil {
		return nil, fmt.Errorf("cannot open destination config: %w", err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	return parseDestinationsYAML(file)
}

func parseDestinationsYAML(r io.Reader) (*DestinationsConfig, error) {
	var config DestinationsConfig
	decoder := yaml.NewDecoder(r)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode destinations YAML: %w", err)
	}

	// Optional: Add basic validation
	for _, d := range append(config.Destinations.Slack, config.Destinations.Teams...) {
		if d.Name == "" || d.WebhookURL == "" {
			return nil, fmt.Errorf("invalid destination entry: name and webhookURL must be set")
		}
	}

	return &config, nil
}
