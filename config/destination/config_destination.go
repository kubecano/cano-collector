package config_destination

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type DestinationsConfig struct {
	Destinations struct {
		Slack []DestinationSlack `yaml:"slack"`
	} `yaml:"destinations"`
}

// SlackDestination represents a Slack notification destination
type DestinationSlack struct {
	Name             string                  `yaml:"name"`
	APIKey           string                  `yaml:"api_key"`
	SlackChannel     string                  `yaml:"slack_channel"`
	GroupingInterval int                     `yaml:"grouping_interval,omitempty"`
	UnfurlLinks      *bool                   `yaml:"unfurl_links,omitempty"`
	Threading        *SlackThreadingConfig   `yaml:"threading,omitempty"`
	Enrichments      *SlackEnrichmentsConfig `yaml:"enrichments,omitempty"`
}

// SlackThreadingConfig represents thread management settings for Slack
type SlackThreadingConfig struct {
	Enabled               bool   `yaml:"enabled"`
	CacheTTL              string `yaml:"cache_ttl,omitempty"`               // Duration string like "10m"
	SearchLimit           int    `yaml:"search_limit,omitempty"`            // Max messages to search in history
	SearchWindow          string `yaml:"search_window,omitempty"`           // Time window string like "24h"
	FingerprintInMetadata *bool  `yaml:"fingerprint_in_metadata,omitempty"` // Include fingerprint in message metadata
}

// SlackEnrichmentsConfig represents enrichment display settings for Slack
type SlackEnrichmentsConfig struct {
	FormatAsBlocks      *bool  `yaml:"format_as_blocks,omitempty"`     // Use Slack blocks instead of plain text
	ColorCoding         *bool  `yaml:"color_coding,omitempty"`         // Color-code enrichments by type
	TableFormatting     string `yaml:"table_formatting,omitempty"`     // "simple", "enhanced", or "attachment"
	MaxTableRows        int    `yaml:"max_table_rows,omitempty"`       // Convert large tables to files
	AttachmentThreshold int    `yaml:"attachment_threshold,omitempty"` // Characters threshold for file conversion
}

//go:generate mockgen -destination=../../mocks/destinations_loader_mock.go -package=mocks github.com/kubecano/cano-collector/config/destination DestinationsLoader
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

	// Replace environment variable placeholders and set defaults for Slack destinations
	for i, d := range config.Destinations.Slack {
		if strings.HasPrefix(d.APIKey, "${") && strings.HasSuffix(d.APIKey, "}") {
			envVar := strings.TrimSuffix(strings.TrimPrefix(d.APIKey, "${"), "}")
			val, ok := os.LookupEnv(envVar)
			if !ok {
				return nil, fmt.Errorf("missing required env %s for slack destination %s", envVar, d.Name)
			}
			config.Destinations.Slack[i].APIKey = val
		}

		// Set default values for new configuration options
		config.Destinations.Slack[i] = setSlackDefaults(config.Destinations.Slack[i])
	}

	// Validate Slack destinations after environment variables have been replaced and defaults set
	for _, d := range config.Destinations.Slack {
		if err := validateSlackDestination(d); err != nil {
			return nil, fmt.Errorf("invalid Slack destination '%s': %w", d.Name, err)
		}
	}

	return &config, nil
}

// setSlackDefaults sets default values for Slack destination configuration
func setSlackDefaults(d DestinationSlack) DestinationSlack {
	// Set threading defaults if threading config is present
	if d.Threading != nil {
		if d.Threading.CacheTTL == "" {
			d.Threading.CacheTTL = "10m"
		}
		if d.Threading.SearchLimit == 0 {
			d.Threading.SearchLimit = 100
		}
		if d.Threading.SearchWindow == "" {
			d.Threading.SearchWindow = "24h"
		}
		// FingerprintInMetadata defaults to true when threading is enabled
		if !d.Threading.Enabled {
			// If threading is explicitly disabled, ensure defaults don't override
		} else if d.Threading.FingerprintInMetadata == nil {
			// Only set default if not explicitly configured
			fingerprintDefault := true
			d.Threading.FingerprintInMetadata = &fingerprintDefault
		}
	}

	// Set enrichments defaults if enrichments config is present
	if d.Enrichments != nil {
		// FormatAsBlocks defaults to true if not explicitly set
		if d.Enrichments.FormatAsBlocks == nil {
			formatDefault := true
			d.Enrichments.FormatAsBlocks = &formatDefault
		}
		// ColorCoding defaults to true if not explicitly set
		if d.Enrichments.ColorCoding == nil {
			colorDefault := true
			d.Enrichments.ColorCoding = &colorDefault
		}
		if d.Enrichments.TableFormatting == "" {
			d.Enrichments.TableFormatting = "enhanced"
		}
		if d.Enrichments.MaxTableRows == 0 {
			d.Enrichments.MaxTableRows = 20
		}
		if d.Enrichments.AttachmentThreshold == 0 {
			d.Enrichments.AttachmentThreshold = 1000
		}
	}

	return d
}

func validateSlackDestination(d DestinationSlack) error {
	if d.Name == "" {
		return fmt.Errorf("name is required")
	}

	if d.SlackChannel == "" {
		return fmt.Errorf("slack_channel is required")
	}

	// Skip validation for placeholder values that will be resolved at runtime
	if strings.HasPrefix(d.APIKey, "${") && strings.HasSuffix(d.APIKey, "}") {
		return nil
	}

	if d.APIKey == "" {
		return fmt.Errorf("api_key is required")
	}

	// Validate grouping_interval if provided
	if d.GroupingInterval < 0 {
		return fmt.Errorf("grouping_interval must be non-negative")
	}

	// Validate threading configuration
	if d.Threading != nil {
		if err := validateThreadingConfig(*d.Threading); err != nil {
			return fmt.Errorf("invalid threading config: %w", err)
		}
	}

	// Validate enrichments configuration
	if d.Enrichments != nil {
		if err := validateEnrichmentsConfig(*d.Enrichments); err != nil {
			return fmt.Errorf("invalid enrichments config: %w", err)
		}
	}

	return nil
}

func validateThreadingConfig(c SlackThreadingConfig) error {
	if c.SearchLimit < 0 {
		return fmt.Errorf("search_limit must be non-negative")
	}
	if c.SearchLimit > 1000 {
		return fmt.Errorf("search_limit must not exceed 1000")
	}

	// Validate cache_ttl duration string format if provided
	if c.CacheTTL != "" {
		if _, err := time.ParseDuration(c.CacheTTL); err != nil {
			return fmt.Errorf("cache_ttl must be a valid duration (e.g., '10m', '1h30m'): %w", err)
		}
	}

	// Validate search_window duration string format if provided
	if c.SearchWindow != "" {
		// Handle "d" suffix which is not supported by time.ParseDuration
		if strings.HasSuffix(c.SearchWindow, "d") {
			// Extract numeric part and validate it's a valid number followed by 'd'
			durationStr := strings.TrimSuffix(c.SearchWindow, "d")
			if durationStr == "" {
				return fmt.Errorf("search_window with 'd' suffix must have a numeric value (e.g., '1d', '7d')")
			}
			// Try to parse as hours to validate the numeric part
			if _, err := time.ParseDuration(durationStr + "h"); err != nil {
				return fmt.Errorf("search_window with 'd' suffix must have a valid numeric value (e.g., '1d', '7d'): %w", err)
			}
		} else {
			// Use standard time.ParseDuration for other formats
			if _, err := time.ParseDuration(c.SearchWindow); err != nil {
				return fmt.Errorf("search_window must be a valid duration (e.g., '24h', '1d', '1h30m'): %w", err)
			}
		}
	}

	return nil
}

func validateEnrichmentsConfig(c SlackEnrichmentsConfig) error {
	if c.TableFormatting != "" {
		validFormats := []string{"simple", "enhanced", "attachment"}
		valid := false
		for _, format := range validFormats {
			if c.TableFormatting == format {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("table_formatting must be one of: %s", strings.Join(validFormats, ", "))
		}
	}

	if c.MaxTableRows < 0 {
		return fmt.Errorf("max_table_rows must be non-negative")
	}

	if c.AttachmentThreshold < 0 {
		return fmt.Errorf("attachment_threshold must be non-negative")
	}

	return nil
}
