package sender

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/kubecano/cano-collector/mocks"
	issuepkg "github.com/kubecano/cano-collector/pkg/core/issue"
)

func setupSenderSlackTest(t *testing.T) (*SenderSlack, *mocks.MockSlackClientInterface) {
	t.Helper()
	ctrl := gomock.NewController(t)

	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockSlackClient := mocks.NewMockSlackClientInterface(ctrl)

	sender := &SenderSlack{
		apiKey:      "xoxb-test-token",
		channel:     "#test-channel",
		logger:      mockLogger,
		unfurlLinks: true,
		slackClient: mockSlackClient,
	}
	return sender, mockSlackClient
}

func TestSenderSlack_Send_Success(t *testing.T) {
	slackSender, mockSlackClient := setupSenderSlackTest(t)

	ctx := context.Background()
	testIssue := &issuepkg.Issue{
		Title:       "Test Issue",
		Description: "This is a test issue",
		Severity:    issuepkg.SeverityHigh,
		Status:      issuepkg.StatusFiring,
		Source:      issuepkg.SourcePrometheus,
	}

	mockSlackClient.EXPECT().PostMessage(
		"#test-channel",
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
	).Return("channel", "timestamp", nil)

	err := slackSender.Send(ctx, testIssue)
	assert.NoError(t, err)
}

func TestSenderSlack_Config(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockClient := mocks.NewMockHTTPClient(ctrl)
	slackSender := NewSenderSlack("xoxb-test-token", "#test-channel", false, mockLogger, mockClient)
	assert.Equal(t, "xoxb-test-token", slackSender.apiKey)
	assert.Equal(t, "#test-channel", slackSender.channel)
	assert.False(t, slackSender.unfurlLinks)
}

func TestSlackSender_SetUnfurlLinks(t *testing.T) {
	slackSender, _ := setupSenderSlackTest(t)

	slackSender.SetUnfurlLinks(false)
	assert.False(t, slackSender.unfurlLinks)

	slackSender.SetUnfurlLinks(true)
	assert.True(t, slackSender.unfurlLinks)
}

// Test funkcji formatowania

func TestSenderSlack_GetSeverityColor(t *testing.T) {
	slackSender, _ := setupSenderSlackTest(t)

	tests := []struct {
		severity issuepkg.Severity
		expected string
	}{
		{issuepkg.SeverityHigh, "#EF311F"},  // Red
		{issuepkg.SeverityLow, "#FFCC00"},   // Yellow
		{issuepkg.SeverityInfo, "#00B302"},  // Green
		{issuepkg.SeverityDebug, "#36a64f"}, // Gray/Green
	}

	for _, test := range tests {
		color := slackSender.getSeverityColor(test.severity)
		assert.Equal(t, test.expected, color, "Incorrect color for severity %s", test.severity.String())
	}
}

func TestSenderSlack_FormatHeader(t *testing.T) {
	slackSender, _ := setupSenderSlackTest(t)

	// Test firing alert
	firingIssue := &issuepkg.Issue{
		Title:    "Test Alert",
		Severity: issuepkg.SeverityHigh,
		Status:   issuepkg.StatusFiring,
	}

	header := slackSender.formatHeader(firingIssue)
	assert.Contains(t, header, "🔥") // Should contain fire emoji
	assert.Contains(t, header, "🔴") // Should contain red circle
	assert.Contains(t, header, "Prometheus Alert Firing")
	assert.Contains(t, header, "Test Alert")

	// Test resolved alert
	resolvedIssue := &issuepkg.Issue{
		Title:    "Resolved Alert",
		Severity: issuepkg.SeverityInfo,
		Status:   issuepkg.StatusResolved,
	}

	header = slackSender.formatHeader(resolvedIssue)
	assert.Contains(t, header, "✅")  // Should contain checkmark
	assert.Contains(t, header, "⚪️") // Should contain white circle
	assert.Contains(t, header, "Prometheus resolved")
	assert.Contains(t, header, "Resolved Alert")
}

func TestSenderSlack_FormatLabels(t *testing.T) {
	slackSender, _ := setupSenderSlackTest(t)

	// Test with labels
	labels := map[string]string{
		"alertname": "HighCPUUsage",
		"severity":  "critical",
		"instance":  "node-1",
	}

	result := slackSender.formatLabels(labels)
	assert.Contains(t, result, "*Alert labels*")
	assert.Contains(t, result, "alertname `HighCPUUsage`")
	assert.Contains(t, result, "severity `critical`")
	assert.Contains(t, result, "instance `node-1`")

	// Test with empty labels
	emptyResult := slackSender.formatLabels(map[string]string{})
	assert.Empty(t, emptyResult)
}
