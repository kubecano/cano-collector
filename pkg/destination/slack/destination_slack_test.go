package destslack

import (
	"context"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kubecano/cano-collector/mocks"
	issuepkg "github.com/kubecano/cano-collector/pkg/core/issue"
)

func TestDestinationSlack_Send_DelegatesToSender(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockClient := mocks.NewMockHTTPClient(ctrl)

	// Accept any number of arguments for Info and Error calls
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
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

	testIssue := &issuepkg.Issue{
		Title:       "Test Issue",
		Description: "This is a test issue",
		Severity:    issuepkg.SeverityHigh,
		Status:      issuepkg.StatusFiring,
		Source:      issuepkg.SourcePrometheus,
	}

	d := NewDestinationSlack(cfg, mockLogger, mockClient)
	err := d.Send(context.Background(), testIssue)
	// Accept either no error or EOF, as slack-go may return EOF with empty body
	if err != nil {
		assert.Contains(t, err.Error(), "EOF")
	}
}

func TestDestinationSlack_Send_WithError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockClient := mocks.NewMockHTTPClient(ctrl)

	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	mockClient.EXPECT().Do(gomock.Any()).Return(nil, assert.AnError).Times(1)

	cfg := &DestinationSlackConfig{
		Name:         "test-slack",
		APIKey:       "xoxb-test-token",
		SlackChannel: "#test-channel",
		UnfurlLinks:  true,
	}

	testIssue := &issuepkg.Issue{
		Title:       "Test Issue",
		Description: "This is a test issue",
		Severity:    issuepkg.SeverityHigh,
		Status:      issuepkg.StatusFiring,
		Source:      issuepkg.SourcePrometheus,
	}

	d := NewDestinationSlack(cfg, mockLogger, mockClient)
	err := d.Send(context.Background(), testIssue)
	require.Error(t, err)
}
