package destslack

import (
	"context"
	"time"

	"go.uber.org/zap"

	issuepkg "github.com/kubecano/cano-collector/pkg/core/issue"
	logger_interfaces "github.com/kubecano/cano-collector/pkg/logger/interfaces"
	slacksender "github.com/kubecano/cano-collector/pkg/sender/slack"
	"github.com/kubecano/cano-collector/pkg/util"
)

type DestinationSlackConfig struct {
	Name             string
	APIKey           string
	SlackChannel     string
	GroupingInterval int
	UnfurlLinks      bool
	// Threading configuration
	Threading *SlackThreadingConfig
	// Enrichments configuration
	Enrichments *SlackEnrichmentsConfig
}

// SlackThreadingConfig contains threading-specific configuration
type SlackThreadingConfig struct {
	Enabled               bool
	CacheTTL              string
	SearchLimit           int
	SearchWindow          string
	FingerprintInMetadata bool
}

// SlackEnrichmentsConfig contains enrichments formatting configuration
type SlackEnrichmentsConfig struct {
	FormatAsBlocks      bool
	ColorCoding         bool
	TableFormatting     string // "simple", "enhanced", or "attachment"
	MaxTableRows        int
	AttachmentThreshold int
}

type DestinationSlack struct {
	sender *slacksender.SenderSlack
	cfg    *DestinationSlackConfig
	logger logger_interfaces.LoggerInterface
}

func NewDestinationSlack(cfg *DestinationSlackConfig, logger logger_interfaces.LoggerInterface, client util.HTTPClient) *DestinationSlack {
	// Create basic sender
	s := slacksender.NewSenderSlack(cfg.APIKey, cfg.SlackChannel, cfg.UnfurlLinks, logger, client)

	destination := &DestinationSlack{
		sender: s,
		cfg:    cfg,
		logger: logger,
	}

	// Enable threading if configured
	if cfg.Threading != nil && cfg.Threading.Enabled {
		destination.enableThreading()
	}

	// Configure enrichments formatting if present
	if cfg.Enrichments != nil {
		destination.configureEnrichments()
	}

	return destination
}

// enableThreading sets up thread management by delegating to SenderSlack
func (d *DestinationSlack) enableThreading() {
	threadingConfig := d.cfg.Threading

	// Parse duration strings with error handling
	cacheTTL, err := time.ParseDuration(threadingConfig.CacheTTL)
	if err != nil {
		d.logger.Warn("Invalid cache TTL, using default",
			zap.String("cacheTTL", threadingConfig.CacheTTL),
			zap.Error(err))
		cacheTTL = 10 * time.Minute // default
	}

	searchWindow, err := time.ParseDuration(threadingConfig.SearchWindow)
	if err != nil {
		d.logger.Warn("Invalid search window, using default",
			zap.String("searchWindow", threadingConfig.SearchWindow),
			zap.Error(err))
		searchWindow = 24 * time.Hour // default
	}

	// Enable threading on the sender - much cleaner!
	d.sender.EnableThreading(cacheTTL, threadingConfig.SearchLimit, searchWindow)
}

// configureEnrichments sets up enrichments formatting by passing parameters to sender
func (d *DestinationSlack) configureEnrichments() {
	enrichmentsConfig := d.cfg.Enrichments

	// Set table formatting parameters on sender (not the whole config!)
	if enrichmentsConfig.TableFormatting != "" {
		d.sender.SetTableFormat(enrichmentsConfig.TableFormatting)
	}

	if enrichmentsConfig.MaxTableRows > 0 {
		d.sender.SetMaxTableRows(enrichmentsConfig.MaxTableRows)
	}

	// Log enrichments configuration
	d.logger.Info("Enrichments configuration applied to sender",
		zap.Bool("formatAsBlocks", enrichmentsConfig.FormatAsBlocks),
		zap.Bool("colorCoding", enrichmentsConfig.ColorCoding),
		zap.String("tableFormatting", enrichmentsConfig.TableFormatting),
		zap.Int("maxTableRows", enrichmentsConfig.MaxTableRows),
		zap.Int("attachmentThreshold", enrichmentsConfig.AttachmentThreshold),
	)
}

// Send implements the destination interface
func (d *DestinationSlack) Send(ctx context.Context, issue *issuepkg.Issue) error {
	d.logger.Info("Sending to Slack destination", zap.String("destination", d.cfg.Name))

	// Send issue directly using sender
	return d.sender.Send(ctx, issue)
}
