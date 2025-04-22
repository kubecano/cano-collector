package sender

import (
	"testing"

	"github.com/kubecano/cano-collector/mocks"
	"github.com/kubecano/cano-collector/pkg/core/reporting"

	"github.com/golang/mock/gomock"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//// MockSlackClient to zastępstwo prawdziwego klienta Slack dla testów
//type MockSlackClient struct {
//	postMessageFn func(channelID string, options ...slack.MsgOption) (string, string, error)
//	authTestFn    func() (*slack.AuthTestResponse, error)
//}
//
//func (m *MockSlackClient) PostMessage(channelID string, options ...slack.MsgOption) (string, string, error) {
//	if m.postMessageFn != nil {
//		return m.postMessageFn(channelID, options...)
//	}
//	return "channel", "timestamp", nil
//}
//
//func (m *MockSlackClient) AuthTest() (*slack.AuthTestResponse, error) {
//	if m.authTestFn != nil {
//		return m.authTestFn()
//	}
//	return &slack.AuthTestResponse{}, nil
//}

func setupSlackTest(t *testing.T) *SlackSender {
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

	mock := &slack.Client{}

	// Tworzenie SlackSender z mockiem
	sender := &SlackSender{
		slackClient:     mock,
		signingKey:      "test-signing-key",
		accountID:       "test-account",
		clusterName:     "test-cluster",
		channel:         "test-channel",
		channelNameToID: make(map[string]string),
		logger:          mockLogger,
	}

	return sender
}

func TestSlackSender_Send_Success(t *testing.T) {
	//mockClient := &MockSlackClient{
	//	postMessageFn: func(channelID string, options ...slack.MsgOption) (string, string, error) {
	//		return "channel", "timestamp", nil
	//	},
	//}

	slackSender := setupSlackTest(t)

	msg := SlackMessage{
		Channel: "test-channel",
		Text:    "Test Alert",
		Blocks:  []slack.Block{},
	}

	err := slackSender.Send(msg)
	assert.NoError(t, err)
}

func TestSlackSender_Send_Error(t *testing.T) {
	//mockClient := &MockSlackClient{
	//	postMessageFn: func(channelID string, options ...slack.MsgOption) (string, string, error) {
	//		return "", "", fmt.Errorf("failed to send to Slack")
	//	},
	//}

	slackSender := setupSlackTest(t)

	msg := SlackMessage{
		Channel: "test-channel",
		Text:    "Error Alert",
		Blocks:  []slack.Block{},
	}

	err := slackSender.Send(msg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send to Slack")
}

func TestSlackSender_FormatMessage(t *testing.T) {
	// mockClient := &MockSlackClient{}
	slackSender := setupSlackTest(t)

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
}

func TestNewSlackSender(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockLogger.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().GetSlackLogger().Return(nil).AnyTimes()

	// Test sukcesu
	sender, err := NewSlackSender("valid-token", "account-id", "cluster-name", "signing-key", "channel", "", mockLogger)
	require.NoError(t, err)
	assert.NotNil(t, sender)

	// Test niepowodzenia weryfikacji tokena można dodać w przyszłości
}
