package sender

import (
	"context"
	"testing"

	"github.com/kubecano/cano-collector/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func setupSenderSlackTest(t *testing.T) *SenderSlack {
	t.Helper()
	ctrl := gomock.NewController(t)

	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockClient := mocks.NewMockHTTPClient(ctrl)

	cfg := SenderSlackConfig{
		APIKey:      "xoxb-test-token",
		Channel:     "#test-channel",
		UnfurlLinks: true,
	}

	return NewSenderSlack(cfg, mockLogger, mockClient)
}

func TestSenderSlack_Send_Success(t *testing.T) {
	slackSender := setupSenderSlackTest(t)

	ctx := context.Background()
	message := "This is a test message"

	err := slackSender.Send(ctx, message)
	assert.NoError(t, err)
}

func TestSenderSlack_Config(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockClient := mocks.NewMockHTTPClient(ctrl)
	cfg := SenderSlackConfig{
		APIKey:      "xoxb-test-token",
		Channel:     "#test-channel",
		UnfurlLinks: false,
	}
	slackSender := NewSenderSlack(cfg, mockLogger, mockClient)

	assert.NotNil(t, slackSender)
	assert.Equal(t, "xoxb-test-token", slackSender.apiKey)
	assert.Equal(t, "#test-channel", slackSender.channel)
	assert.True(t, slackSender.unfurlLinks) // Default value
}

func TestSlackSender_SetUnfurlLinks(t *testing.T) {
	slackSender := setupSenderSlackTest(t)

	slackSender.SetUnfurlLinks(false)
	assert.False(t, slackSender.unfurlLinks)

	slackSender.SetUnfurlLinks(true)
	assert.True(t, slackSender.unfurlLinks)
}
