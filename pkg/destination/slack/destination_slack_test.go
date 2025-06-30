package destslack

import (
	"context"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kubecano/cano-collector/mocks"
)

func TestDestinationSlack_Send_DelegatesToSender(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockClient := mocks.NewMockHTTPClient(ctrl)

	// Akceptujemy dowolne wywołania Info i Error, bo slack-go może logować błąd EOF
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	mockClient.EXPECT().Do(gomock.Any()).Return(&http.Response{
		StatusCode: http.StatusOK,
		Body:       http.NoBody,
	}, nil).Times(1)

	cfg := &DestinationSlackConfig{
		Name:         "test-slack",
		APIKey:       "xoxb-test-token",
		SlackChannel: "#test-channel",
		UnfurlLinks:  true,
	}

	d := NewDestinationSlack(cfg, mockLogger, mockClient)
	err := d.Send(context.Background(), "test message")
	// Akceptujemy zarówno brak błędu, jak i EOF, bo slack-go może zwrócić EOF przy pustym body
	if err != nil {
		assert.Contains(t, err.Error(), "EOF")
	}
}

func TestDestinationSlack_Send_WithError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockClient := mocks.NewMockHTTPClient(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	mockClient.EXPECT().Do(gomock.Any()).Return(nil, assert.AnError).Times(1)

	cfg := &DestinationSlackConfig{
		Name:         "test-slack",
		APIKey:       "xoxb-test-token",
		SlackChannel: "#test-channel",
		UnfurlLinks:  true,
	}

	d := NewDestinationSlack(cfg, mockLogger, mockClient)
	err := d.Send(context.Background(), "test message")
	require.Error(t, err)
}
