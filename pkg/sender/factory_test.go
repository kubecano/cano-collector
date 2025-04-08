package sender

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/kubecano/cano-collector/config/destinations"
	"github.com/kubecano/cano-collector/mocks"
)

func setupTest(t *testing.T) *SenderFactory {
	t.Helper()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockClient := mocks.NewMockHTTPClient(ctrl)

	return NewSenderFactory(mockLogger, mockClient)
}

func TestSenderFactory_Create_Slack(t *testing.T) {
	factory := setupTest(t)

	dest := destinations.Destination{
		Name:       "slack",
		WebhookURL: "https://hooks.slack.com/services/XXXX/XXXX",
	}

	sender, err := factory.Create(dest)
	assert.NoError(t, err)
	assert.IsType(t, &SlackSender{}, sender)
}

func TestSenderFactory_Create_MSTeams(t *testing.T) {
	factory := setupTest(t)

	dest := destinations.Destination{
		Name:       "teams",
		WebhookURL: "https://outlook.office.com/webhook/XXXX",
	}

	sender, err := factory.Create(dest)
	assert.NoError(t, err)
	assert.IsType(t, &MSTeamsSender{}, sender)
}

func TestSenderFactory_Create_UnsupportedType(t *testing.T) {
	factory := setupTest(t)

	dest := destinations.Destination{
		Name:       "pagerduty",
		WebhookURL: "https://events.pagerduty.com/...",
	}

	sender, err := factory.Create(dest)
	assert.Nil(t, sender)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported destination type: pagerduty")
}
