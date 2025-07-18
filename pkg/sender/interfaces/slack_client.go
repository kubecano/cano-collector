package interfaces

import (
	"github.com/slack-go/slack"
)

// SlackClientInterface defines the interface for Slack client
//
//go:generate mockgen -source=slack_client.go -destination=../../../mocks/slack_client_mock.go -package=mocks
type SlackClientInterface interface {
	PostMessage(channelID string, options ...slack.MsgOption) (string, string, error)
	AuthTest() (*slack.AuthTestResponse, error)
	UpdateMessage(channelID, timestamp string, options ...slack.MsgOption) (string, string, string, error)
	UploadFileV2(params slack.UploadFileV2Parameters) (*slack.FileSummary, error)
	GetConversationHistory(params *slack.GetConversationHistoryParameters) (*slack.GetConversationHistoryResponse, error)
}
