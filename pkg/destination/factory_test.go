package destination

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/kubecano/cano-collector/config/destination"
	"github.com/kubecano/cano-collector/mocks"
)

func setupTest(t *testing.T) (*DestinationFactory, *gomock.Controller) {
	t.Helper()
	ctrl := gomock.NewController(t)

	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockClient := mocks.NewMockHTTPClient(ctrl)

	return NewDestinationFactory(mockLogger, mockClient), ctrl
}

func TestDestinationFactory_CreateSlackDestination(t *testing.T) {
	factory, ctrl := setupTest(t)
	defer ctrl.Finish()

	config := destination.SlackDestinationConfig{
		BaseDestinationConfig: destination.BaseDestinationConfig{Name: "test-slack"},
		Token:                 "xoxb-test-token",
		Channel:               "test-channel",
		SigningKey:            "test-signing-key",
		AccountID:             "test-account-id",
		ClusterName:           "test-cluster",
	}

	dest, err := factory.CreateSlackDestination(config)
	require.NoError(t, err)
	assert.IsType(t, &SlackDestination{}, dest)
}

func TestDestinationFactory_CreateTeamsDestination(t *testing.T) {
	factory, ctrl := setupTest(t)
	defer ctrl.Finish()

	config := destination.TeamsDestinationConfig{
		BaseDestinationConfig: destination.BaseDestinationConfig{Name: "test-teams"},
		WebhookURL:            "https://outlook.office.com/webhook/XXXX",
	}

	dest, err := factory.CreateTeamsDestination(config)
	require.NoError(t, err)
	assert.IsType(t, &TeamsDestination{}, dest)
}

func TestDestinationFactory_CreateAllDestinations(t *testing.T) {
	factory, ctrl := setupTest(t)
	defer ctrl.Finish()

	config := destination.DestinationsConfig{
		Destinations: struct {
			Slack []destination.SlackDestinationConfig `yaml:"slack"`
			Teams []destination.TeamsDestinationConfig `yaml:"teams"`
		}{
			Slack: []destination.SlackDestinationConfig{
				{
					BaseDestinationConfig: destination.BaseDestinationConfig{Name: "test-slack"},
					Token:                 "xoxb-test-token",
					Channel:               "test-channel",
					SigningKey:            "test-signing-key",
					AccountID:             "test-account-id",
					ClusterName:           "test-cluster",
				},
			},
			Teams: []destination.TeamsDestinationConfig{
				{
					BaseDestinationConfig: destination.BaseDestinationConfig{Name: "test-teams"},
					WebhookURL:            "https://outlook.office.com/webhook/XXXX",
				},
			},
		},
	}

	destinations, err := factory.CreateAllDestinations(config)
	require.NoError(t, err)
	assert.Len(t, destinations, 2)
	assert.Contains(t, destinations, "test-slack")
	assert.Contains(t, destinations, "test-teams")
	assert.IsType(t, &SlackDestination{}, destinations["test-slack"])
	assert.IsType(t, &TeamsDestination{}, destinations["test-teams"])
}
