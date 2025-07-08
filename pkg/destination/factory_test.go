package destination

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	destination_config "github.com/kubecano/cano-collector/config/destination"
	"github.com/kubecano/cano-collector/mocks"
)

func setupTest(t *testing.T) (*DestinationFactory, *gomock.Controller) {
	t.Helper()
	ctrl := gomock.NewController(t)

	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockClient := mocks.NewMockHTTPClient(ctrl)

	// Accept logger calls that may happen during destination creation
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

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
	require.NoError(t, err)
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

func TestDestinationFactory_CreateDestinationSlack_WithThreadingAndEnrichments(t *testing.T) {
	factory, ctrl := setupTest(t)
	defer ctrl.Finish()

	unfurlLinks := true
	dest := destination_config.DestinationSlack{
		Name:         "test-slack-advanced",
		APIKey:       "xoxb-test-token",
		SlackChannel: "#test-channel",
		UnfurlLinks:  &unfurlLinks,
		Threading: &destination_config.SlackThreadingConfig{
			Enabled:               true,
			CacheTTL:              "15m",
			SearchLimit:           50,
			SearchWindow:          "12h",
			FingerprintInMetadata: &[]bool{true}[0],
		},
		Enrichments: &destination_config.SlackEnrichmentsConfig{
			FormatAsBlocks:      &[]bool{true}[0],
			ColorCoding:         &[]bool{true}[0],
			TableFormatting:     "enhanced",
			MaxTableRows:        25,
			AttachmentThreshold: 2000,
		},
	}

	d, err := factory.CreateDestination(dest)
	require.NoError(t, err)
	assert.NotNil(t, d)

	// Verify that it created a Slack destination
	// Note: We can't easily test the internal configuration conversion without exposing internal fields
	// This test mainly ensures the factory doesn't crash with extended configuration
}
