package sender

import (
	"context"

	"github.com/kubecano/cano-collector/pkg/logger"
	"github.com/kubecano/cano-collector/pkg/util"
)

// SlackSender sends notifications to Slack
type SlackSender struct {
	apiKey      string
	channel     string
	logger      logger.LoggerInterface
	httpClient  util.HTTPClient
	unfurlLinks bool
}

// NewSlackSender creates a new SlackSender using Slack API key
func NewSlackSender(apiKey, channel string, logger logger.LoggerInterface) *SlackSender {
	return &SlackSender{
		apiKey:      apiKey,
		channel:     channel,
		logger:      logger,
		httpClient:  util.DefaultHTTPClient(),
		unfurlLinks: true, // Default to true
	}
}

// Send sends a notification to Slack
func (s *SlackSender) Send(ctx context.Context, message string) error {
	s.logger.Info("Sending Slack notification", "channel", s.channel)

	// TODO: Implement using slack-go library
	// For now, just log the message
	s.logger.Info("Slack message would be sent",
		"channel", s.channel,
		"message", message,
		"unfurl_links", s.unfurlLinks)

	return nil
}

// SetLogger sets the logger for this sender
func (s *SlackSender) SetLogger(logger logger.LoggerInterface) {
	s.logger = logger
}

// SetClient sets the HTTP client for this sender
func (s *SlackSender) SetClient(client util.HTTPClient) {
	s.httpClient = client
}

// SetUnfurlLinks sets whether links should be unfurled
func (s *SlackSender) SetUnfurlLinks(unfurl bool) {
	s.unfurlLinks = unfurl
}
