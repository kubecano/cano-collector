package sender

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/kubecano/cano-collector/mocks"
	"github.com/kubecano/cano-collector/pkg/core/issue"
)

func setupSenderSlackTest(t *testing.T) (*SenderSlack, *mocks.MockSlackClientInterface) {
	t.Helper()
	ctrl := gomock.NewController(t)

	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockSlackClient := mocks.NewMockSlackClientInterface(ctrl)

	sender := &SenderSlack{
		apiKey:      "xoxb-test-token",
		channel:     "#test-channel",
		logger:      mockLogger,
		unfurlLinks: true,
		slackClient: mockSlackClient,
	}
	return sender, mockSlackClient
}

func TestSenderSlack_Send_Success(t *testing.T) {
	slackSender, mockSlackClient := setupSenderSlackTest(t)

	ctx := context.Background()
	testIssue := &issue.Issue{
		Title:       "Test Issue",
		Description: "This is a test issue",
		Severity:    issue.SeverityHigh,
		Status:      issue.StatusFiring,
		Source:      issue.SourcePrometheus,
	}

	mockSlackClient.EXPECT().PostMessage(
		"#test-channel",
		gomock.Any(),
		gomock.Any(),
	).Return("channel", "timestamp", nil)

	err := slackSender.Send(ctx, testIssue)
	assert.NoError(t, err)
}

func TestSenderSlack_Config(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockClient := mocks.NewMockHTTPClient(ctrl)
	slackSender := NewSenderSlack("xoxb-test-token", "#test-channel", false, mockLogger, mockClient)
	assert.Equal(t, "xoxb-test-token", slackSender.apiKey)
	assert.Equal(t, "#test-channel", slackSender.channel)
	assert.False(t, slackSender.unfurlLinks)
}

func TestSlackSender_SetUnfurlLinks(t *testing.T) {
	slackSender, _ := setupSenderSlackTest(t)

	slackSender.SetUnfurlLinks(false)
	assert.False(t, slackSender.unfurlLinks)

	slackSender.SetUnfurlLinks(true)
	assert.True(t, slackSender.unfurlLinks)
}
