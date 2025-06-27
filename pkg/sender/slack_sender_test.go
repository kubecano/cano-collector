package sender

import (
	"context"
	"testing"

	"github.com/kubecano/cano-collector/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func setupSlackTest(t *testing.T) *SlackSender {
	t.Helper()
	ctrl := gomock.NewController(t)

	// Mock logger
	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()

	// Create SlackSender with mock logger
	slackSender := NewSlackSender("xoxb-test-token", "#test-channel", mockLogger)

	return slackSender
}

func TestSlackSender_Send_Success(t *testing.T) {
	slackSender := setupSlackTest(t)

	ctx := context.Background()
	message := "This is a test message"

	err := slackSender.Send(ctx, message)
	assert.NoError(t, err)
}

func TestSlackSender_NewSlackSenderWithAPIKey(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLoggerInterface(ctrl)

	slackSender := NewSlackSender("xoxb-test-token", "#test-channel", mockLogger)

	assert.NotNil(t, slackSender)
	assert.Equal(t, "xoxb-test-token", slackSender.apiKey)
	assert.Equal(t, "#test-channel", slackSender.channel)
	assert.True(t, slackSender.unfurlLinks) // Default value
}

func TestSlackSender_SetUnfurlLinks(t *testing.T) {
	slackSender := setupSlackTest(t)

	slackSender.SetUnfurlLinks(false)
	assert.False(t, slackSender.unfurlLinks)

	slackSender.SetUnfurlLinks(true)
	assert.True(t, slackSender.unfurlLinks)
}
