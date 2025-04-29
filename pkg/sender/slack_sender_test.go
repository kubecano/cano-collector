package sender

import (
	"errors"
	"fmt"
	"testing"

	"github.com/kubecano/cano-collector/pkg/logger"

	"github.com/kubecano/cano-collector/mocks"
	"github.com/kubecano/cano-collector/pkg/core/reporting"

	"github.com/golang/mock/gomock"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupSlackTest(t *testing.T) (*SlackSender, logger.LoggerInterface, SlackClientInterface) {
	t.Helper()
	ctrl := gomock.NewController(t)

	// Mock logger
	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().GetSlackLogger().Return(nil).AnyTimes()

	mockSlackClient := mocks.NewMockSlackClientInterface(ctrl)
	mockSlackClient.EXPECT().UpdateMessage("C12345678", "1234567890.123456", gomock.Any(), gomock.Any()).Return("C12345678", "1234567890.123457", "", nil).AnyTimes()

	// Tworzenie SlackSender z mockiem
	sender := &SlackSender{
		slackClient:     mockSlackClient,
		signingKey:      "test-signing-key",
		accountID:       "test-account",
		clusterName:     "test-cluster",
		channel:         "test-channel",
		channelNameToID: make(map[string]string),
		logger:          mockLogger,
	}

	return sender, mockLogger, mockSlackClient
}

func TestSlackSender_Send_Success(t *testing.T) {
	slackSender, _, mockSlackClient := setupSlackTest(t)

	mockSlackClient.(*mocks.MockSlackClientInterface).EXPECT().PostMessage(gomock.Any(), gomock.Any()).Return("channel", "timestamp", nil).AnyTimes()
	msg := SlackMessage{
		Channel: "test-channel",
		Text:    "Test Alert",
		Blocks:  []slack.Block{},
	}

	err := slackSender.Send(msg)
	assert.NoError(t, err)
}

func TestSlackSender_Send_Error(t *testing.T) {
	slackSender, _, mockSlackClient := setupSlackTest(t)

	mockSlackClient.(*mocks.MockSlackClientInterface).EXPECT().PostMessage(gomock.Any(), gomock.Any()).Return("", "", fmt.Errorf("failed to send to Slack")).AnyTimes()

	msg := SlackMessage{
		Channel: "test-channel",
		Text:    "Error Alert",
		Blocks:  []slack.Block{},
	}

	err := slackSender.Send(msg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send to Slack")
}

func TestSlackSender_UpdateMessage(t *testing.T) {
	slackSender, _, _ := setupSlackTest(t)

	// Dodaj mapowanie nazwy kanału na ID
	slackSender.channelNameToID["test-channel"] = "C12345678"

	blocks := []slack.Block{
		slack.NewHeaderBlock(slack.NewTextBlockObject("plain_text", "Test Header", false, false)),
	}

	ts, err := slackSender.UpdateMessage("test-channel", "1234567890.123456", "Updated message", blocks)

	require.NoError(t, err)
	assert.Equal(t, "1234567890.123457", ts)
}

func TestSlackSender_UpdateMessage_ChannelNotFound(t *testing.T) {
	slackSender, _, _ := setupSlackTest(t)

	blocks := []slack.Block{
		slack.NewHeaderBlock(slack.NewTextBlockObject("plain_text", "Test Header", false, false)),
	}

	_, err := slackSender.UpdateMessage("nonexistent-channel", "1234567890.123456", "Updated message", blocks)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "channel ID for nonexistent-channel could not be determined")
}

func TestSlackSender_FormatMessage(t *testing.T) {
	slackSender, _, _ := setupSlackTest(t)

	details := reporting.AlertDetails{
		Title:       "Test Alert",
		Description: "This is a test description",
		Severity:    "critical",
		Links: []reporting.LinkProp{
			{Text: "Link 1", URL: "http://example.com/1"},
		},
		Metadata: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}

	result := slackSender.FormatMessage(details)
	slackMsg, ok := result.(SlackMessage)

	assert.True(t, ok, "Rezultat FormatMessage powinien być typu SlackMessage")
	assert.Equal(t, "Test Alert", slackMsg.Text)
	assert.Equal(t, slackSender.channel, slackMsg.Channel)
	assert.NotEmpty(t, slackMsg.Blocks)

	// Sprawdź czy nagłówek zawiera severity
	headerBlock, ok := slackMsg.Blocks[0].(*slack.HeaderBlock)
	assert.True(t, ok, "Pierwszy blok powinien być typu HeaderBlock")
	assert.Contains(t, headerBlock.Text.Text, "[critical]")
}

func TestSlackClientAuthTest(t *testing.T) {
	_, mockLogger, mockSlackClient := setupSlackTest(t)

	mockSlackClient.(*mocks.MockSlackClientInterface).EXPECT().AuthTest().Return(&slack.AuthTestResponse{}, nil).AnyTimes()

	sender, err := createSlackSender("valid-token", "account-id", "cluster-name", "signing-key", "channel", mockSlackClient, mockLogger)
	require.NoError(t, err)
	assert.NotNil(t, sender)
}

func TestSlackClientAuthTestFailed(t *testing.T) {
	_, mockLogger, mockSlackClient := setupSlackTest(t)

	mockSlackClient.(*mocks.MockSlackClientInterface).EXPECT().AuthTest().Return(nil, errors.New(" invalid_auth")).AnyTimes()

	sender, err := createSlackSender("invalid-token", "account-id", "cluster-name", "signing-key", "channel", mockSlackClient, mockLogger)
	require.Error(t, err)
	assert.Contains(t, err.Error(), " invalid_auth")
	assert.Nil(t, sender)
}
