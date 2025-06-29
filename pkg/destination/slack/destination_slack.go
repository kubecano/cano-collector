package destslack

import (
	"context"

	"github.com/kubecano/cano-collector/pkg/logger"
	"github.com/kubecano/cano-collector/pkg/sender"
	"github.com/kubecano/cano-collector/pkg/util"
)

type DestinationSlackConfig struct {
	Name             string
	APIKey           string
	SlackChannel     string
	GroupingInterval int
	UnfurlLinks      bool
}

type DestinationSlack struct {
	sender *sender.SenderSlack
	cfg    *DestinationSlackConfig
}

func NewDestinationSlack(cfg *DestinationSlackConfig, logger logger.LoggerInterface, client util.HTTPClient) *DestinationSlack {
	s := sender.NewSenderSlack(cfg.APIKey, cfg.SlackChannel, cfg.UnfurlLinks, logger, client)
	return &DestinationSlack{sender: s, cfg: cfg}
}

func (d *DestinationSlack) Send(ctx context.Context, message string) error {
	return d.sender.Send(ctx, message)
}
