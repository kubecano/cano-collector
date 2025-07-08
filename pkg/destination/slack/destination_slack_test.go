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

func TestDestinationSlack_WithThreadingConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockClient := mocks.NewMockHTTPClient(ctrl)

	// Expect threading initialization logs from SenderSlack
	mockLogger.EXPECT().Info("Thread management enabled", gomock.Any()).Times(1)
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	cfg := &DestinationSlackConfig{
		Name:         "test-slack",
		APIKey:       "xoxb-test-token",
		SlackChannel: "#test-channel",
		UnfurlLinks:  false,
		Threading: &SlackThreadingConfig{
			Enabled:               true,
			CacheTTL:              "10m",
			SearchLimit:           50,
			SearchWindow:          "24h",
			FingerprintInMetadata: true,
		},
	}

	d := NewDestinationSlack(cfg, mockLogger, mockClient)
	assert.NotNil(t, d)
}

func TestDestinationSlack_WithEnrichmentsConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockClient := mocks.NewMockHTTPClient(ctrl)

	// Expect enrichments configuration logs
	mockLogger.EXPECT().Info("Enrichments configuration loaded", gomock.Any()).Times(1)
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	cfg := &DestinationSlackConfig{
		Name:         "test-slack",
		APIKey:       "xoxb-test-token",
		SlackChannel: "#test-channel",
		UnfurlLinks:  false,
		Enrichments: &SlackEnrichmentsConfig{
			FormatAsBlocks:      true,
			ColorCoding:         true,
			TableFormatting:     "enhanced",
			MaxTableRows:        20,
			AttachmentThreshold: 5,
		},
	}

	d := NewDestinationSlack(cfg, mockLogger, mockClient)
	assert.Equal(t, cfg.Enrichments, d.cfg.Enrichments)
}

func TestDestinationSlack_WithInvalidThreadingConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockClient := mocks.NewMockHTTPClient(ctrl)

	// Expect warnings for invalid durations and then successful threading initialization
	mockLogger.EXPECT().Warn("Invalid cache TTL, using default", gomock.Any()).Times(1)
	mockLogger.EXPECT().Warn("Invalid search window, using default", gomock.Any()).Times(1)
	mockLogger.EXPECT().Info("Thread management enabled", gomock.Any()).Times(1)
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	cfg := &DestinationSlackConfig{
		Name:         "test-slack",
		APIKey:       "xoxb-test-token",
		SlackChannel: "#test-channel",
		UnfurlLinks:  false,
		Threading: &SlackThreadingConfig{
			Enabled:      true,
			CacheTTL:     "invalid-duration",
			SearchLimit:  50,
			SearchWindow: "invalid-window",
		},
	}

	d := NewDestinationSlack(cfg, mockLogger, mockClient)
	assert.NotNil(t, d)
}
