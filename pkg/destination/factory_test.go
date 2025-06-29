package destination

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/kubecano/cano-collector/mocks"
	destination_config "github.com/kubecano/cano-collector/pkg/config/destination"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T) (*DestinationFactory, *gomock.Controller) {
	t.Helper()
	ctrl := gomock.NewController(t)

	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockClient := mocks.NewMockHTTPClient(ctrl)

	return NewDestinationFactory(mockLogger, mockClient), ctrl
}

func TestDestinationFactory_CreateDestinationSlack(t *testing.T) {
	factory, ctrl := setupTest(t)
	defer ctrl.Finish()

	dest := destination_config.DestinationSlack{
		Name:         "test-slack",
		APIKey:       "xoxb-test-token",
		SlackChannel: "#test-channel",
	}

	d, err := factory.CreateDestination(dest)
	assert.NoError(t, err)
	assert.NotNil(t, d)
}

func TestDestinationFactory_CreateDestinationSlack_MissingAPIKey(t *testing.T) {
	factory, ctrl := setupTest(t)
	defer ctrl.Finish()

	dest := destination_config.DestinationSlack{
		Name:         "test-slack",
		SlackChannel: "#test-channel",
		// Missing APIKey
	}

	d, err := factory.CreateDestination(dest)
	require.Error(t, err)
	assert.Nil(t, d)
	assert.Contains(t, err.Error(), "must have api_key")
}

func TestDestinationFactory_CreateDestination_UnsupportedType(t *testing.T) {
	factory, ctrl := setupTest(t)
	defer ctrl.Finish()

	unsupportedDest := "not a destination"

	d, err := factory.CreateDestination(unsupportedDest)
	require.Error(t, err)
	assert.Nil(t, d)
	assert.Contains(t, err.Error(), "unsupported destination type")
}
