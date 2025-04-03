package sender

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kubecano/cano-collector/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

// Helper function to set up test environment for MSTeamsSender
func setupMSTeamsTest(t *testing.T, statusCode int, responseBody string) *MSTeamsSender {
	t.Helper()
	ctrl := gomock.NewController(t)

	// Mock logger
	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	// Mock HTTP client
	mockClient := mocks.NewMockHTTPClient(ctrl)
	mockClient.EXPECT().
		Do(gomock.Any()).
		Return(&http.Response{
			StatusCode: statusCode,
			Body:       io.NopCloser(bytes.NewBufferString(responseBody)),
		}, nil).
		AnyTimes()

	// Create MSTeamsSender with mock logger and client
	msTeamsSender := NewMSTeamsSender("http://example.com", mockLogger, WithHTTPClient(mockClient))

	return msTeamsSender
}

func TestMSTeamsSender_Send_Success(t *testing.T) {
	msTeamsSender := setupMSTeamsTest(t, http.StatusOK, "ok")

	alert := Alert{
		Title:   "Test Alert",
		Message: "This is a test alert message",
	}

	err := msTeamsSender.Send(alert)
	assert.NoError(t, err)
}

func TestMSTeamsSender_Send_Error(t *testing.T) {
	msTeamsSender := setupMSTeamsTest(t, http.StatusInternalServerError, "error")

	alert := Alert{
		Title:   "Error Alert",
		Message: "This is a failing test",
	}

	err := msTeamsSender.Send(alert)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "MS Teams returned non-OK status")
}

func TestMSTeamsSender_Send_RequestError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockClient := mocks.NewMockHTTPClient(ctrl)

	mockClient.EXPECT().
		Do(gomock.Any()).
		Return(nil, fmt.Errorf("request error")).
		Times(1)

	msTeamsSender := NewMSTeamsSender("http://example.com", mockLogger, WithHTTPClient(mockClient))

	alert := Alert{
		Title:   "Request Error Alert",
		Message: "This alert will fail to send",
	}

	err := msTeamsSender.Send(alert)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send alert to MS Teams")
}
