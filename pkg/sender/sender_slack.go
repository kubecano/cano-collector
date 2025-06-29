package sender

import (
	"context"

	"github.com/kubecano/cano-collector/pkg/logger"
	"github.com/kubecano/cano-collector/pkg/util"
)

//go:generate mockgen -destination=../../mocks/sender_slack_mock.go -package=mocks github.com/kubecano/cano-collector/pkg/sender SenderSlack
type SenderSlack struct {
	apiKey      string
	channel     string
	logger      logger.LoggerInterface
	httpClient  util.HTTPClient
	unfurlLinks bool
}

func NewSenderSlack(apiKey, channel string, unfurlLinks bool, logger logger.LoggerInterface, client util.HTTPClient) *SenderSlack {
	if client == nil {
		client = util.DefaultHTTPClient()
	}
	return &SenderSlack{
		apiKey:      apiKey,
		channel:     channel,
		logger:      logger,
		httpClient:  client,
		unfurlLinks: unfurlLinks,
	}
}

func (s *SenderSlack) Send(ctx context.Context, message string) error {
	s.logger.Info("Sending Slack notification", "channel", s.channel)
	// TODO: Implementacja slack-go
	s.logger.Info("Slack message would be sent", "channel", s.channel, "message", message, "unfurl_links", s.unfurlLinks)
	return nil
}

func (s *SenderSlack) SetLogger(logger logger.LoggerInterface) {
	s.logger = logger
}

func (s *SenderSlack) SetHTTPClient(client util.HTTPClient) {
	s.httpClient = client
}

func (s *SenderSlack) SetUnfurlLinks(unfurl bool) {
	s.unfurlLinks = unfurl
}
