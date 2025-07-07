package sender

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"

	"github.com/kubecano/cano-collector/pkg/core/issue"
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

func (s *SenderSlack) Send(ctx context.Context, issue *issue.Issue) error {
	s.logger.Info("Sending Slack notification", "channel", s.channel)

	// Convert Issue to message string
	message := s.formatIssueToString(issue)

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

// formatIssueToString converts an Issue to a formatted string message
// This is temporary - in the future this will be replaced with Slack Block Kit formatting
func (s *SenderSlack) formatIssueToString(issue *issue.Issue) string {
	statusPrefix := ""
	if issue.IsResolved() {
		statusPrefix = "[RESOLVED] "
	}

	message := fmt.Sprintf("%s*%s*\n", statusPrefix, issue.Title)

	if issue.Description != "" {
		message += fmt.Sprintf("📝 %s\n", issue.Description)
	}

	message += fmt.Sprintf("🔥 Severity: %s\n", issue.Severity.String())
	message += fmt.Sprintf("📍 Source: %s\n", issue.Source.String())

	if issue.Subject != nil && issue.Subject.Name != "" {
		if issue.Subject.Namespace != "" {
			message += fmt.Sprintf("🎯 Subject: %s/%s (%s)\n",
				issue.Subject.Namespace, issue.Subject.Name, issue.Subject.SubjectType.String())
		} else {
			message += fmt.Sprintf("🎯 Subject: %s (%s)\n",
				issue.Subject.Name, issue.Subject.SubjectType.String())
		}
	}

	if len(issue.Links) > 0 {
		message += "🔗 Links:\n"
		for _, link := range issue.Links {
			message += fmt.Sprintf("• <%s|%s>\n", link.URL, link.Text)
		}
	}

	return message
}

func (s *SenderSlack) SetLogger(logger logger.LoggerInterface) {
	s.logger = logger
}

func (s *SenderSlack) SetUnfurlLinks(unfurl bool) {
	s.unfurlLinks = unfurl
}
