package destslack

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/kubecano/cano-collector/mocks"
)

func TestDestinationSlack_Send_DelegatesToSender(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockClient := mocks.NewMockHTTPClient(ctrl)
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()

	// Tworzymy config dla destination
	cfg := &DestinationSlackConfig{
		Name:         "test-slack",
		APIKey:       "xoxb-test-token",
		SlackChannel: "#test-channel",
		UnfurlLinks:  true,
	}

	d := NewDestinationSlack(cfg, mockLogger, mockClient)
	err := d.Send(context.Background(), "test message")
	assert.NoError(t, err)
}
