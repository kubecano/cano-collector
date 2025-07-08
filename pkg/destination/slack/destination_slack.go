package destslack

import (
	"context"

	"go.uber.org/zap"

	issuepkg "github.com/kubecano/cano-collector/pkg/core/issue"
	logger_interfaces "github.com/kubecano/cano-collector/pkg/logger/interfaces"
	"github.com/kubecano/cano-collector/pkg/sender"
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
	sender *sender.SenderSlack
	cfg    *DestinationSlackConfig
	logger logger_interfaces.LoggerInterface
}

func NewDestinationSlack(cfg *DestinationSlackConfig, logger logger_interfaces.LoggerInterface, client util.HTTPClient) *DestinationSlack {
	s := sender.NewSenderSlack(cfg.APIKey, cfg.SlackChannel, cfg.UnfurlLinks, logger, client)
	return &DestinationSlack{sender: s, cfg: cfg, logger: logger}
}

// Send implements the destination interface
func (d *DestinationSlack) Send(ctx context.Context, issue *issuepkg.Issue) error {
	d.logger.Info("Sending to Slack destination", zap.String("destination", d.cfg.Name))

	// Send issue directly using sender
	return d.sender.Send(ctx, issue)
}
