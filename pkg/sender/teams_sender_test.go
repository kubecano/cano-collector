package sender

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/kubecano/cano-collector/pkg/core/reporting"

	"github.com/stretchr/testify/require"

	"github.com/kubecano/cano-collector/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

// Helper function to set up test environment for TeamsSender
func setupMSTeamsTest(t *testing.T, statusCode int, responseBody string) *TeamsSender {
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

	// Create TeamsSender with mock logger and client
	msTeamsSender, _ := NewTeamsSender("http://example.com", mockLogger, WithHTTPClient(mockClient))

	return msTeamsSender
}

func TestMSTeamsSender_Send_Success(t *testing.T) {
	msTeamsSender := setupMSTeamsTest(t, http.StatusOK, "ok")

	alert := reporting.AlertDetails{
		Title:       "Test Alert",
		Description: "This is a test alert message",
	}

	err := msTeamsSender.Send(alert)
	assert.NoError(t, err)
}

func TestMSTeamsSender_Send_Error(t *testing.T) {
	msTeamsSender := setupMSTeamsTest(t, http.StatusInternalServerError, "error")

	alert := reporting.AlertDetails{
		Title:       "Error Alert",
		Description: "This is a failing test",
	}

	err := msTeamsSender.Send(alert)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send to Microsoft Teams")
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

	msTeamsSender, _ := NewTeamsSender("http://example.com", mockLogger, WithHTTPClient(mockClient))

	alert := reporting.AlertDetails{
		Title:       "Request Error Alert",
		Description: "This alert will fail to send",
	}

	err := msTeamsSender.Send(alert)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send alert to MS Teams")
}
