package sender

import (
	"context"

	"github.com/slack-go/slack"

	"github.com/kubecano/cano-collector/pkg/logger"
	"github.com/kubecano/cano-collector/pkg/util"
)

// SlackClientInterface defines the interface for Slack client
//
//go:generate mockgen -destination=../../mocks/slack_client_mock.go -package=mocks github.com/kubecano/cano-collector/pkg/sender SlackClientInterface
type SlackClientInterface interface {
	PostMessage(channelID string, options ...slack.MsgOption) (string, string, error)
	AuthTest() (*slack.AuthTestResponse, error)
	UpdateMessage(channelID, timestamp string, options ...slack.MsgOption) (string, string, string, error)
	UploadFileV2(params slack.UploadFileV2Parameters) (*slack.FileSummary, error)
}

type SenderSlack struct {
	apiKey      string
	channel     string
	logger      logger.LoggerInterface
	unfurlLinks bool
	slackClient SlackClientInterface
}

func NewSenderSlack(apiKey, channel string, unfurlLinks bool, logger logger.LoggerInterface, client util.HTTPClient) *SenderSlack {
	var slackClient SlackClientInterface

	if client != nil {
		// Use custom HTTP client with slack-go
		slackClient = slack.New(apiKey, slack.OptionHTTPClient(client))
	} else {
		// Use default HTTP client from slack-go
		slackClient = slack.New(apiKey)
	}

	return &SenderSlack{
		apiKey:      apiKey,
		channel:     channel,
		logger:      logger,
		unfurlLinks: unfurlLinks,
		slackClient: slackClient,
	}
}

func (s *SenderSlack) Send(ctx context.Context, message string) error {
	s.logger.Info("Sending Slack notification", "channel", s.channel)

	params := slack.PostMessageParameters{
		UnfurlLinks: s.unfurlLinks,
		UnfurlMedia: s.unfurlLinks,
	}

	_, _, err := s.slackClient.PostMessage(
		s.channel,
		slack.MsgOptionText(message, false),
		slack.MsgOptionPostMessageParameters(params),
	)
	if err != nil {
		s.logger.Error("Failed to send Slack message", "error", err, "channel", s.channel)
		return err
	}

	s.logger.Info("Slack message sent successfully", "channel", s.channel, "message", message)
	return nil
}

func (s *SenderSlack) SetLogger(logger logger.LoggerInterface) {
	s.logger = logger
}

func (s *SenderSlack) SetUnfurlLinks(unfurl bool) {
	s.unfurlLinks = unfurl
}
