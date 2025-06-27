package sender

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/kubecano/cano-collector/config/destination"
	"github.com/kubecano/cano-collector/mocks"
)

func setupTest(t *testing.T) (*SenderFactory, *gomock.Controller) {
	t.Helper()
	ctrl := gomock.NewController(t)

	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockClient := mocks.NewMockHTTPClient(ctrl)

	return NewSenderFactory(mockLogger, mockClient), ctrl
}

func TestSenderFactory_CreateSender_Slack(t *testing.T) {
	factory, ctrl := setupTest(t)
	defer ctrl.Finish()

	dest := destination.SlackDestination{
		Name:         "test-slack",
		APIKey:       "xoxb-test-token",
		SlackChannel: "#test-channel",
	}

	sender, err := factory.CreateSender(dest)
	require.NoError(t, err)
	assert.NotNil(t, sender)
}

func TestSenderFactory_CreateSender_Slack_MissingAPIKey(t *testing.T) {
	factory, ctrl := setupTest(t)
	defer ctrl.Finish()

	dest := destination.SlackDestination{
		Name:         "test-slack",
		SlackChannel: "#test-channel",
		// Missing APIKey
	}

	sender, err := factory.CreateSender(dest)
	require.Error(t, err)
	assert.Nil(t, sender)
	assert.Contains(t, err.Error(), "must have api_key")
}

func TestSenderFactory_CreateSender_Teams(t *testing.T) {
	factory, ctrl := setupTest(t)
	defer ctrl.Finish()

	dest := destination.TeamsDestination{
		Name:       "teams",
		WebhookURL: "https://outlook.office.com/webhook/XXXX",
	}

	sender, err := factory.CreateSender(dest)
	require.NoError(t, err)
	assert.IsType(t, &MSTeamsSender{}, sender)
}

func TestSenderFactory_CreateSender_Teams_NoWebhookURL(t *testing.T) {
	factory, ctrl := setupTest(t)
	defer ctrl.Finish()

	dest := destination.TeamsDestination{
		Name: "teams",
	}

	sender, err := factory.CreateSender(dest)
	assert.Nil(t, sender)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must have webhookURL")
}

func TestSenderFactory_CreateSender_UnsupportedType(t *testing.T) {
	factory, ctrl := setupTest(t)
	defer ctrl.Finish()

	unsupportedDest := "not a destination"

	sender, err := factory.CreateSender(unsupportedDest)
	require.Error(t, err)
	assert.Nil(t, sender)
	assert.Contains(t, err.Error(), "unsupported destination type")
}
