package sender

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/kubecano/cano-collector/mocks"
	issuepkg "github.com/kubecano/cano-collector/pkg/core/issue"
	"github.com/slack-go/slack"
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

// Formatting tests

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

func TestSenderSlack_BuildSlackBlocks(t *testing.T) {
	slackSender, _ := setupSenderSlackTest(t)

	// Test issue with links
	issue := &issuepkg.Issue{
		Title:       "Test Issue",
		Description: "Test description",
		Severity:    issuepkg.SeverityHigh,
		Status:      issuepkg.StatusFiring,
		Source:      issuepkg.SourcePrometheus,
		Links: []issuepkg.Link{
			{Text: "Dashboard", URL: "https://example.com/dashboard"},
			{Text: "Runbook", URL: "https://example.com/runbook"},
		},
	}

	blocks := slackSender.buildSlackBlocks(issue)

	// Should have at least header block and action block (for links)
	assert.GreaterOrEqual(t, len(blocks), 2)

	// First block should be section block (header)
	assert.Equal(t, "section", string(blocks[0].BlockType()))

	// Second block should be actions block (links)
	if len(blocks) > 1 {
		assert.Equal(t, "actions", string(blocks[1].BlockType()))
	}
}

func TestSenderSlack_BuildSlackAttachments(t *testing.T) {
	slackSender, _ := setupSenderSlackTest(t)

	// Test issue with subject and labels
	subject := issuepkg.NewSubject("test-pod", issuepkg.SubjectTypePod)
	subject.Namespace = "default"
	subject.Labels = map[string]string{
		"app": "test-app",
		"env": "production",
	}

	issue := &issuepkg.Issue{
		Title:       "Test Issue",
		Description: "Test description with details",
		Severity:    issuepkg.SeverityHigh,
		Status:      issuepkg.StatusFiring,
		Source:      issuepkg.SourcePrometheus,
		Subject:     subject,
	}

	attachments := slackSender.buildSlackAttachments(issue)

	// Should have one attachment
	assert.Len(t, attachments, 1)

	attachment := attachments[0]

	// Should have correct color based on severity
	assert.Equal(t, "#EF311F", attachment.Color) // High severity = red

	// Should have blocks with content
	assert.NotEmpty(t, attachment.Blocks.BlockSet)
}

func TestSenderSlack_FormatIssueToString(t *testing.T) {
	slackSender, _ := setupSenderSlackTest(t)

	// Test firing issue
	firingIssue := &issuepkg.Issue{
		Title:       "High CPU Usage Alert",
		Description: "CPU usage exceeded 80%",
		Severity:    issuepkg.SeverityHigh,
		Status:      issuepkg.StatusFiring,
		Source:      issuepkg.SourcePrometheus,
		Links: []issuepkg.Link{
			{Text: "Grafana Dashboard", URL: "https://grafana.example.com"},
		},
	}

	result := slackSender.formatIssueToString(firingIssue)

	assert.Contains(t, result, "*High CPU Usage Alert*")
	assert.Contains(t, result, "📝 CPU usage exceeded 80%")
	assert.Contains(t, result, "🔥 Severity: HIGH")
	assert.Contains(t, result, "📍 Source: PROMETHEUS")
	assert.Contains(t, result, "🔗 Links:")
	assert.Contains(t, result, "Grafana Dashboard")
	assert.NotContains(t, result, "[RESOLVED]") // Should not contain resolved prefix

	// Test resolved issue
	resolvedIssue := &issuepkg.Issue{
		Title:       "Resolved Alert",
		Description: "Issue has been resolved",
		Severity:    issuepkg.SeverityInfo,
		Status:      issuepkg.StatusResolved,
		Source:      issuepkg.SourcePrometheus,
	}

	resolvedResult := slackSender.formatIssueToString(resolvedIssue)

	assert.Contains(t, resolvedResult, "[RESOLVED] *Resolved Alert*")
	assert.Contains(t, resolvedResult, "📝 Issue has been resolved")
	assert.Contains(t, resolvedResult, "🔥 Severity: INFO")
}

func TestSenderSlack_BuildLinkButtons(t *testing.T) {
	slackSender, _ := setupSenderSlackTest(t)

	links := []issuepkg.Link{
		{Text: "Dashboard", URL: "https://example.com/dashboard"},
		{Text: "Runbook", URL: "https://example.com/runbook"},
		{Text: "Logs", URL: "https://example.com/logs"},
	}

	buttons := slackSender.buildLinkButtons(links)

	// Should create button for each link
	assert.Len(t, buttons, 3)

	// Each button should be a button element
	for _, button := range buttons {
		assert.NotNil(t, button)
		// Note: We can't easily verify the ID due to the slack library's internal structure,
		// but we can verify that buttons were created
	}
}

func TestSenderSlack_BuildLinkButtons_LimitToFive(t *testing.T) {
	slackSender, _ := setupSenderSlackTest(t)

	// Create more than 5 links
	links := make([]issuepkg.Link, 7)
	for i := range 7 {
		links[i] = issuepkg.Link{
			Text: fmt.Sprintf("Link %d", i+1),
			URL:  fmt.Sprintf("https://example.com/link%d", i+1),
		}
	}

	buttons := slackSender.buildLinkButtons(links)

	// Should limit to 5 buttons to avoid Slack limits
	assert.Len(t, buttons, 5)
}

func TestSenderSlack_EnrichmentSupport(t *testing.T) {
	slackSender, mockClient := setupSenderSlackTest(t)

	t.Run("builds enrichment attachments for table blocks", func(t *testing.T) {
		// Create issue with table enrichments
		issue := &issuepkg.Issue{
			Title:       "Test Alert with Enrichments",
			Description: "Alert with label and annotation enrichments",
			Severity:    issuepkg.SeverityHigh,
			Status:      issuepkg.StatusFiring,
			Source:      issuepkg.SourcePrometheus,
		}

		// Add label enrichment
		labelEnrichment := issuepkg.NewEnrichmentWithType(issuepkg.EnrichmentTypeAlertLabels, "Alert Labels")
		tableBlock := &issuepkg.TableBlock{
			Headers: []string{"Label", "Value"},
			Rows: [][]string{
				{"alertname", "TestAlert"},
				{"severity", "critical"},
				{"namespace", "default"},
			},
		}
		labelEnrichment.AddBlock(tableBlock)
		issue.AddEnrichment(*labelEnrichment)

		// Add annotation enrichment
		annotationEnrichment := issuepkg.NewEnrichmentWithType(issuepkg.EnrichmentTypeAlertAnnotations, "Alert Annotations")
		annotationTableBlock := &issuepkg.TableBlock{
			Headers: []string{"Annotation", "Value"},
			Rows: [][]string{
				{"summary", "Test alert summary"},
				{"description", "Test alert description"},
			},
		}
		annotationEnrichment.AddBlock(annotationTableBlock)
		issue.AddEnrichment(*annotationEnrichment)

		// Mock successful Slack API call
		mockClient.EXPECT().PostMessage(
			"#test-channel",
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
			gomock.Any(),
		).Return("", "", nil)

		// Send the issue
		err := slackSender.Send(context.Background(), issue)

		// Verify
		assert.NoError(t, err)

		// Test attachment building directly
		attachments := slackSender.buildSlackAttachments(issue)

		// Should have main attachment + 2 enrichment attachments
		assert.Len(t, attachments, 3)

		// Verify enrichment attachments have the right colors
		assert.Equal(t, "#17A2B8", attachments[1].Color) // Blue for labels
		assert.Equal(t, "#6610F2", attachments[2].Color) // Purple for annotations
	})

	t.Run("builds enrichment attachments for json blocks", func(t *testing.T) {
		issue := &issuepkg.Issue{
			Title:    "Test Alert with JSON Enrichment",
			Severity: issuepkg.SeverityLow,
			Status:   issuepkg.StatusFiring,
			Source:   issuepkg.SourcePrometheus,
		}

		// Add JSON enrichment
		jsonEnrichment := issuepkg.NewEnrichmentWithType(issuepkg.EnrichmentTypeAlertLabels, "Alert Labels (JSON)")
		jsonBlock := &issuepkg.JsonBlock{
			Data: map[string]string{
				"alertname": "TestAlert",
				"severity":  "warning",
			},
		}
		jsonEnrichment.AddBlock(jsonBlock)
		issue.AddEnrichment(*jsonEnrichment)

		// Test attachment building
		attachments := slackSender.buildSlackAttachments(issue)

		// Should have main attachment + 1 enrichment attachment
		assert.Len(t, attachments, 2)

		// Verify the JSON enrichment attachment
		jsonAttachment := attachments[1]
		assert.Equal(t, "#17A2B8", jsonAttachment.Color) // Blue for labels
		assert.Len(t, jsonAttachment.Blocks.BlockSet, 2) // Title + JSON content

		// Verify that the JSON block was converted to a section block
		contentBlock := jsonAttachment.Blocks.BlockSet[1]
		sectionBlock, ok := contentBlock.(*slack.SectionBlock)
		assert.True(t, ok, "Expected section block for JSON content")
		assert.Contains(t, sectionBlock.Text.Text, "```") // Should be wrapped in code block
	})

	t.Run("handles empty enrichments gracefully", func(t *testing.T) {
		issue := &issuepkg.Issue{
			Title:    "Test Alert with Empty Enrichments",
			Severity: issuepkg.SeverityInfo,
			Status:   issuepkg.StatusFiring,
			Source:   issuepkg.SourcePrometheus,
		}

		// Add enrichment with no blocks
		emptyEnrichment := issuepkg.NewEnrichment()
		issue.AddEnrichment(*emptyEnrichment)

		// Test attachment building
		attachments := slackSender.buildSlackAttachments(issue)

		// Should only have main attachment (empty enrichment should be skipped)
		assert.Len(t, attachments, 1)
	})

	t.Run("formats table blocks correctly", func(t *testing.T) {
		tableBlock := &issuepkg.TableBlock{
			Headers:   []string{"Label", "Value"},
			TableName: "Test Table",
			Rows: [][]string{
				{"key1", "value1"},
				{"key2", "value2"},
			},
		}

		slackBlock := slackSender.convertTableBlockToSlack(tableBlock)

		sectionBlock, ok := slackBlock.(*slack.SectionBlock)
		assert.True(t, ok, "Expected section block")

		text := sectionBlock.Text.Text
		assert.Contains(t, text, "*Test Table*")
		assert.Contains(t, text, "• key1: `value1`")
		assert.Contains(t, text, "• key2: `value2`")
	})

	t.Run("formats json blocks correctly", func(t *testing.T) {
		jsonBlock := &issuepkg.JsonBlock{
			Data: map[string]string{
				"test": "value",
			},
		}

		slackBlock := slackSender.convertJsonBlockToSlack(jsonBlock)

		sectionBlock, ok := slackBlock.(*slack.SectionBlock)
		assert.True(t, ok, "Expected section block")

		text := sectionBlock.Text.Text
		assert.Contains(t, text, "```")
		assert.Contains(t, text, "\"test\": \"value\"")
	})

	t.Run("returns correct colors for enrichment types", func(t *testing.T) {
		tests := []struct {
			enrichmentType issuepkg.EnrichmentType
			expectedColor  string
		}{
			{issuepkg.EnrichmentTypeAlertLabels, "#17A2B8"},
			{issuepkg.EnrichmentTypeAlertAnnotations, "#6610F2"},
			{issuepkg.EnrichmentTypeGraph, "#28A745"},
			{issuepkg.EnrichmentTypeAIAnalysis, "#FD7E14"},
		}

		for _, test := range tests {
			color := slackSender.getEnrichmentColor(&test.enrichmentType)
			assert.Equal(t, test.expectedColor, color)
		}

		// Test nil enrichment type
		color := slackSender.getEnrichmentColor(nil)
		assert.Equal(t, "#E8E8E8", color)
	})
}
