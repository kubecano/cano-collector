package slack

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/slack-go/slack"

	"github.com/kubecano/cano-collector/mocks"
	issuepkg "github.com/kubecano/cano-collector/pkg/core/issue"
)

// mockUnsupportedBlock is a test mock for unsupported block type
type mockUnsupportedBlock struct {
	blockType string
}

func (m *mockUnsupportedBlock) BlockType() string {
	return m.blockType
}

func setupSenderSlackTest(t *testing.T) (*SenderSlack, *mocks.MockSlackClientInterface, *mocks.MockLoggerInterface) {
	t.Helper()
	ctrl := gomock.NewController(t)

	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockSlackClient := mocks.NewMockSlackClientInterface(ctrl)

	sender := &SenderSlack{
		apiKey:       "xoxb-test-token",
		channel:      "#test-channel",
		logger:       mockLogger,
		unfurlLinks:  true,
		slackClient:  mockSlackClient,
		tableFormat:  "enhanced",
		maxTableRows: 20,
	}
	return sender, mockSlackClient, mockLogger
}

func TestSenderSlack_Send_Success(t *testing.T) {
	slackSender, mockSlackClient, _ := setupSenderSlackTest(t)

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
	require.NoError(t, err)
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
	slackSender, _, _ := setupSenderSlackTest(t)

	slackSender.SetUnfurlLinks(false)
	assert.False(t, slackSender.unfurlLinks)

	slackSender.SetUnfurlLinks(true)
	assert.True(t, slackSender.unfurlLinks)
}

// Formatting tests

func TestSenderSlack_GetSeverityColor(t *testing.T) {
	slackSender, _, _ := setupSenderSlackTest(t)

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
	slackSender, _, _ := setupSenderSlackTest(t)

	// Test firing alert
	firingIssue := &issuepkg.Issue{
		Title:    "Test Alert",
		Severity: issuepkg.SeverityHigh,
		Status:   issuepkg.StatusFiring,
	}

	header := slackSender.formatHeader(firingIssue)
	assert.Contains(t, header, "🔥") // Should contain fire emoji
	assert.Contains(t, header, "🔴") // Should contain red circle
	assert.Contains(t, header, "Alert firing")
	assert.Contains(t, header, "Test Alert")

	// Test resolved alert
	resolvedIssue := &issuepkg.Issue{
		Title:    "Resolved Alert",
		Severity: issuepkg.SeverityInfo,
		Status:   issuepkg.StatusResolved,
	}

	header = slackSender.formatHeader(resolvedIssue)
	assert.Contains(t, header, "✅")              // Should contain checkmark
	assert.Contains(t, header, "🟢")              // Should contain green circle
	assert.Contains(t, header, "Alert resolved") // Simplified text
	assert.Contains(t, header, "Resolved Alert")
}

func TestSenderSlack_FormatLabels(t *testing.T) {
	slackSender, _, _ := setupSenderSlackTest(t)

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
	slackSender, _, _ := setupSenderSlackTest(t)

	// Create an issue with all components including runbook link
	firingIssue := &issuepkg.Issue{
		Title:       "Test Issue",
		Description: "Test description",
		Severity:    issuepkg.SeverityHigh,
		Status:      issuepkg.StatusFiring,
		Source:      issuepkg.SourcePrometheus,
		Subject: &issuepkg.Subject{
			Annotations: map[string]string{
				"description": "Pod is crash looping",
			},
		},
		Links: []issuepkg.Link{
			{Text: "Generator URL", URL: "https://example.com/graph", Type: issuepkg.LinkTypePrometheusGenerator},
			{Text: "Runbook", URL: "https://runbooks.prometheus-operator.dev/runbooks/kubernetes/kubepodcrashlooping", Type: issuepkg.LinkTypeRunbook},
		},
		Enrichments: []issuepkg.Enrichment{
			{
				Title: stringPtr("Alert Labels"),
				Blocks: []issuepkg.BaseBlock{
					&issuepkg.TableBlock{
						TableName: "Labels",
						Headers:   []string{"Label", "Value"},
						Rows:      [][]string{{"severity", "high"}},
					},
				},
			},
		},
	}

	blocks := slackSender.buildSlackBlocks(firingIssue)

	// Should have: header, links, alert description, runbook, enrichments, divider
	assert.GreaterOrEqual(t, len(blocks), 5)

	// First block should be header
	_, ok := blocks[0].(*slack.SectionBlock)
	assert.True(t, ok, "First block should be section block (header)")

	// Should find runbook URL block
	found := false
	for _, block := range blocks {
		if sectionBlock, ok := block.(*slack.SectionBlock); ok {
			if strings.Contains(sectionBlock.Text.Text, "Runbook URL:") &&
				strings.Contains(sectionBlock.Text.Text, "runbooks.prometheus-operator.dev") {
				found = true
				break
			}
		}
	}
	assert.True(t, found, "Should contain runbook URL as text block")
}

func TestSenderSlack_BuildSlackBlocks_ResolvedAlert(t *testing.T) {
	slackSender, _, _ := setupSenderSlackTest(t)

	// Create a resolved issue with timestamp
	resolvedTime := time.Now().Add(-time.Minute)
	resolvedIssue := &issuepkg.Issue{
		Title:       "Resolved Alert",
		Description: "This alert has been resolved",
		Severity:    issuepkg.SeverityInfo,
		Status:      issuepkg.StatusResolved,
		Source:      issuepkg.SourcePrometheus,
		ClusterName: "test-cluster",
		EndsAt:      &resolvedTime,
		Links: []issuepkg.Link{
			{Text: "Generator URL", URL: "https://example.com/graph", Type: issuepkg.LinkTypePrometheusGenerator},
			{Text: "Runbook", URL: "https://runbooks.prometheus-operator.dev/runbooks/kubernetes/kubepodcrashlooping", Type: issuepkg.LinkTypeRunbook},
		},
		Enrichments: []issuepkg.Enrichment{
			{
				Title: stringPtr("Alert Labels"),
				Blocks: []issuepkg.BaseBlock{
					&issuepkg.TableBlock{
						TableName: "Labels",
						Headers:   []string{"Label", "Value"},
						Rows:      [][]string{{"severity", "low"}},
					},
				},
			},
		},
	}

	blocks := slackSender.buildSlackBlocks(resolvedIssue)

	// For resolved alerts, should only have: header, source info, resolved timestamp (3 blocks max)
	assert.LessOrEqual(t, len(blocks), 3)

	// First block should be header
	headerBlock, ok := blocks[0].(*slack.SectionBlock)
	assert.True(t, ok, "First block should be section block (header)")
	assert.Contains(t, headerBlock.Text.Text, "Alert resolved")
	assert.Contains(t, headerBlock.Text.Text, "✅")

	// Should contain source information
	foundSource := false
	foundResolved := false
	for _, block := range blocks {
		if sectionBlock, ok := block.(*slack.SectionBlock); ok {
			if strings.Contains(sectionBlock.Text.Text, "Source: PROMETHEUS") &&
				strings.Contains(sectionBlock.Text.Text, "Cluster: test-cluster") {
				foundSource = true
			}
			if strings.Contains(sectionBlock.Text.Text, "Resolved:") {
				foundResolved = true
			}
		}
	}
	assert.True(t, foundSource, "Should contain source and cluster information")
	assert.True(t, foundResolved, "Should contain resolved timestamp")

	// Should NOT contain enrichments, links, or description blocks
	for _, block := range blocks {
		if sectionBlock, ok := block.(*slack.SectionBlock); ok {
			assert.NotContains(t, sectionBlock.Text.Text, "Alert:")
			assert.NotContains(t, sectionBlock.Text.Text, "Runbook URL:")
		}
		// Should not have action blocks (links)
		_, isActionBlock := block.(*slack.ActionBlock)
		assert.False(t, isActionBlock, "Resolved alerts should not have action blocks")
	}
}

// Helper function for string pointer
func stringPtr(s string) *string {
	return &s
}

func TestSenderSlack_BuildSlackBlocks_WithoutRunbook(t *testing.T) {
	slackSender, _, _ := setupSenderSlackTest(t)

	// Create an issue without runbook link
	firingIssue := &issuepkg.Issue{
		Title:       "Test Issue",
		Description: "Test description",
		Severity:    issuepkg.SeverityHigh,
		Status:      issuepkg.StatusFiring,
		Source:      issuepkg.SourcePrometheus,
		Subject:     &issuepkg.Subject{Annotations: map[string]string{"description": "Some other annotation"}},
		Links: []issuepkg.Link{
			{Text: "Generator URL", URL: "https://example.com/graph", Type: issuepkg.LinkTypePrometheusGenerator},
		},
	}

	blocks := slackSender.buildSlackBlocks(firingIssue)

	// Should NOT find runbook URL block
	found := false
	for _, block := range blocks {
		if sectionBlock, ok := block.(*slack.SectionBlock); ok {
			if strings.Contains(sectionBlock.Text.Text, "Runbook URL:") {
				found = true
				break
			}
		}
	}
	assert.False(t, found, "Should NOT contain runbook URL when not present in links")
}

func TestSenderSlack_BuildSlackAttachments(t *testing.T) {
	slackSender, _, _ := setupSenderSlackTest(t)

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

func TestSenderSlack_BuildSlackAttachments_WithClusterName(t *testing.T) {
	slackSender, _, _ := setupSenderSlackTest(t)

	// Test issue with cluster name
	issue := &issuepkg.Issue{
		Title:       "Test Issue",
		Description: "Test description",
		Severity:    issuepkg.SeverityHigh,
		Status:      issuepkg.StatusFiring,
		Source:      issuepkg.SourcePrometheus,
		ClusterName: "production-cluster",
		Subject:     issuepkg.NewSubject("test-pod", issuepkg.SubjectTypePod),
	}

	attachments := slackSender.buildSlackAttachments(issue)

	// Should have one attachment
	assert.Len(t, attachments, 1)

	attachment := attachments[0]

	// Should have blocks with cluster information
	assert.NotEmpty(t, attachment.Blocks.BlockSet)

	// Convert blocks to find cluster info
	found := false
	for _, block := range attachment.Blocks.BlockSet {
		if sectionBlock, ok := block.(*slack.SectionBlock); ok {
			if sectionBlock.Text != nil && strings.Contains(sectionBlock.Text.Text, "🌐 *Cluster:* `production-cluster`") {
				found = true
				break
			}
		}
	}
	assert.True(t, found, "Should contain cluster name in attachment")
}

func TestSenderSlack_BuildSlackAttachments_WithTimeFormatting(t *testing.T) {
	slackSender, _, _ := setupSenderSlackTest(t)

	// Create a specific time in a non-UTC timezone for testing
	// This will be converted to UTC in the formatting
	testTime := time.Date(2024, 1, 15, 14, 30, 45, 0, time.FixedZone("CET", 1*60*60)) // CET +1 hour

	subject := issuepkg.NewSubject("test-pod", issuepkg.SubjectTypePod)
	subject.Namespace = "default"

	issue := &issuepkg.Issue{
		Title:       "Test Issue with Time",
		Description: "Test description",
		Severity:    issuepkg.SeverityHigh,
		Status:      issuepkg.StatusFiring,
		Source:      issuepkg.SourcePrometheus,
		Subject:     subject,
		StartsAt:    testTime,
	}

	attachments := slackSender.buildSlackAttachments(issue)

	// Should have one attachment with time information
	assert.Len(t, attachments, 1)

	attachment := attachments[0]
	assert.NotEmpty(t, attachment.Blocks.BlockSet)

	// Find the time block and verify UTC formatting
	foundTimeBlock := false
	for _, block := range attachment.Blocks.BlockSet {
		if sectionBlock, ok := block.(*slack.SectionBlock); ok {
			if sectionBlock.Text != nil && sectionBlock.Text.Text != "" {
				text := sectionBlock.Text.Text
				// Check if this is the time block
				if strings.Contains(text, "⏰ *Started:*") {
					foundTimeBlock = true
					// Verify that the time is properly formatted in UTC
					// Original time: 2024-01-15 14:30:45 CET (+1)
					// Expected UTC: 2024-01-15 13:30:45 UTC
					assert.Contains(t, text, "2024-01-15 13:30:45 UTC", "Time should be formatted in UTC")
					assert.Contains(t, text, "⏰ *Started:*", "Should contain time label")
					break
				}
			}
		}
	}

	assert.True(t, foundTimeBlock, "Should find time block in attachment")
}

func TestSenderSlack_BuildSlackAttachments_WithoutTime(t *testing.T) {
	slackSender, _, _ := setupSenderSlackTest(t)

	// Test issue without StartsAt time (zero time)
	subject := issuepkg.NewSubject("test-pod", issuepkg.SubjectTypePod)
	subject.Namespace = "default"

	issue := &issuepkg.Issue{
		Title:       "Test Issue No Time",
		Description: "Test description",
		Severity:    issuepkg.SeverityLow,
		Status:      issuepkg.StatusFiring,
		Source:      issuepkg.SourcePrometheus,
		Subject:     subject,
		// StartsAt is zero value - should not be displayed
	}

	attachments := slackSender.buildSlackAttachments(issue)

	// Should have one attachment but without time information
	assert.Len(t, attachments, 1)

	attachment := attachments[0]
	assert.NotEmpty(t, attachment.Blocks.BlockSet)

	// Verify that no time block exists
	for _, block := range attachment.Blocks.BlockSet {
		if sectionBlock, ok := block.(*slack.SectionBlock); ok {
			if sectionBlock.Text != nil && sectionBlock.Text.Text != "" {
				text := sectionBlock.Text.Text
				// Should not contain time information
				assert.NotContains(t, text, "⏰ *Started:*", "Should not contain time label for zero time")
			}
		}
	}
}

func TestSenderSlack_FormatIssueToString(t *testing.T) {
	slackSender, _, _ := setupSenderSlackTest(t)

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
	slackSender, _, _ := setupSenderSlackTest(t)

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
	slackSender, _, _ := setupSenderSlackTest(t)

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
	slackSender, mockSlackClient, _ := setupSenderSlackTest(t)

	// Add mock for file upload to support large table tests
	mockSlackClient.EXPECT().UploadFileV2(gomock.Any()).Return(&slack.FileSummary{
		ID:    "F123456789",
		Title: "test-file.csv",
	}, nil).AnyTimes()

	t.Run("builds enrichment blocks for table blocks", func(t *testing.T) {
		issue := &issuepkg.Issue{
			Title:    "Test Alert with Table Enrichments",
			Severity: issuepkg.SeverityHigh,
			Status:   issuepkg.StatusFiring,
			Source:   issuepkg.SourcePrometheus,
		}

		// Add labels enrichment with table block
		labelsEnrichment := issuepkg.NewEnrichmentWithType(issuepkg.EnrichmentTypeAlertLabels, "Alert Labels")
		labelsTable := &issuepkg.TableBlock{
			Headers:   []string{"Label", "Value"},
			TableName: "Alert Labels",
			Rows: [][]string{
				{"alertname", "TestAlert"},
				{"severity", "warning"},
			},
		}
		labelsEnrichment.AddBlock(labelsTable)
		issue.AddEnrichment(*labelsEnrichment)

		// Add annotations enrichment with table block
		annotationsEnrichment := issuepkg.NewEnrichmentWithType(issuepkg.EnrichmentTypeAlertAnnotations, "Alert Annotations")
		annotationsTable := &issuepkg.TableBlock{
			Headers:   []string{"Annotation", "Value"},
			TableName: "Alert Annotations",
			Rows: [][]string{
				{"summary", "Test summary"},
				{"description", "Test description"},
			},
		}
		annotationsEnrichment.AddBlock(annotationsTable)
		issue.AddEnrichment(*annotationsEnrichment)

		// Test block building directly
		blocks := slackSender.buildSlackBlocks(issue)

		// Should have: header + enrichment blocks + links
		// Each enrichment: header + section + divider = 3 blocks per enrichment
		// Total: 1 (header) + 6 (2 enrichments * 3) + 0 (no links) = 7 blocks
		assert.GreaterOrEqual(t, len(blocks), 6, "Should have header block + enrichment blocks")

		// Test enrichment blocks building directly
		enrichmentBlocks := slackSender.buildEnrichmentBlocks(issue.Enrichments)

		// Should have 6 blocks: 2 enrichments * (header + section + divider) = 6 blocks
		assert.Len(t, enrichmentBlocks, 6)

		// Verify first enrichment blocks (labels)
		sectionBlock1, ok := enrichmentBlocks[0].(*slack.SectionBlock)
		assert.True(t, ok, "First block should be section block with title")
		assert.Equal(t, "*Alert Labels*", sectionBlock1.Text.Text)

		// Verify second enrichment blocks (annotations)
		sectionBlock2, ok := enrichmentBlocks[3].(*slack.SectionBlock)
		assert.True(t, ok, "Fourth block should be section block with title")
		assert.Equal(t, "*Alert Annotations*", sectionBlock2.Text.Text)
	})

	t.Run("builds enrichment blocks for json blocks", func(t *testing.T) {
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

		// Test enrichment blocks building
		enrichmentBlocks := slackSender.buildEnrichmentBlocks(issue.Enrichments)

		// Should have 3 blocks: header + section + divider
		assert.Len(t, enrichmentBlocks, 3)

		// Verify title block
		titleBlock, ok := enrichmentBlocks[0].(*slack.SectionBlock)
		assert.True(t, ok, "First block should be section block with title")
		assert.Equal(t, "*Alert Labels (JSON)*", titleBlock.Text.Text)

		// Verify the JSON block was converted to a section block
		contentBlock := enrichmentBlocks[1]
		sectionBlock, ok := contentBlock.(*slack.SectionBlock)
		assert.True(t, ok, "Second block should be section block for JSON content")
		assert.Contains(t, sectionBlock.Text.Text, "```") // Should be wrapped in code block

		// Verify divider block
		_, ok = enrichmentBlocks[2].(*slack.DividerBlock)
		assert.True(t, ok, "Third block should be divider")
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

		// Test enrichment blocks building
		enrichmentBlocks := slackSender.buildEnrichmentBlocks(issue.Enrichments)

		// Should have no blocks (empty enrichment should be skipped)
		assert.Empty(t, enrichmentBlocks)

		// Test that main blocks building also handles this correctly
		blocks := slackSender.buildSlackBlocks(issue)
		// Should have just the header block (description, timing sections added by default)
		assert.GreaterOrEqual(t, len(blocks), 1)
		assert.LessOrEqual(t, len(blocks), 3) // header + description + timing but no enrichments/links
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
		assert.Contains(t, text, "▸ *key1*: `value1`")
		assert.Contains(t, text, "▸ *key2*: `value2`")
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

	t.Run("formats header blocks correctly", func(t *testing.T) {
		headerBlock := &issuepkg.HeaderBlock{
			Text: "Test Header",
		}

		slackBlock := slackSender.convertHeaderBlockToSlack(headerBlock)

		headerSlackBlock, ok := slackBlock.(*slack.HeaderBlock)
		assert.True(t, ok, "Expected header block")
		assert.Equal(t, "Test Header", headerSlackBlock.Text.Text)
	})

	t.Run("formats list blocks correctly", func(t *testing.T) {
		listBlock := &issuepkg.ListBlock{
			Items:    []string{"Item 1", "Item 2", "Item 3"},
			Ordered:  false,
			ListName: "Test List",
		}

		slackBlock := slackSender.convertListBlockToSlack(listBlock)

		sectionBlock, ok := slackBlock.(*slack.SectionBlock)
		assert.True(t, ok, "Expected section block")

		text := sectionBlock.Text.Text
		assert.Contains(t, text, "*Test List*")
		assert.Contains(t, text, "• Item 1")
		assert.Contains(t, text, "• Item 2")
		assert.Contains(t, text, "• Item 3")
	})

	t.Run("formats ordered list blocks correctly", func(t *testing.T) {
		listBlock := &issuepkg.ListBlock{
			Items:   []string{"First", "Second", "Third"},
			Ordered: true,
		}

		slackBlock := slackSender.convertListBlockToSlack(listBlock)

		sectionBlock, ok := slackBlock.(*slack.SectionBlock)
		assert.True(t, ok, "Expected section block")

		text := sectionBlock.Text.Text
		assert.Contains(t, text, "1. First")
		assert.Contains(t, text, "2. Second")
		assert.Contains(t, text, "3. Third")
	})

	t.Run("formats links blocks as buttons correctly", func(t *testing.T) {
		links := []issuepkg.Link{
			{Text: "Dashboard", URL: "https://example.com/dashboard", Type: issuepkg.LinkTypeGeneral},
			{Text: "Logs", URL: "https://example.com/logs", Type: issuepkg.LinkTypeGeneral},
		}
		linksBlock := &issuepkg.LinksBlock{
			Links:     links,
			BlockName: "Related Links",
		}

		slackBlock := slackSender.convertLinksBlockToSlack(linksBlock)

		actionBlock, ok := slackBlock.(*slack.ActionBlock)
		assert.True(t, ok, "Expected action block")
		assert.Equal(t, "links_related_links", actionBlock.BlockID)
		assert.Len(t, actionBlock.Elements.ElementSet, 2)
	})

	t.Run("formats file blocks correctly", func(t *testing.T) {
		// Setup a separate mock environment for this test since it needs upload mock
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLoggerInterface(ctrl)
		mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
		mockSlackClient := mocks.NewMockSlackClientInterface(ctrl)

		// Mock upload failure for this test to avoid complexity
		uploadError := fmt.Errorf("test upload error")
		mockSlackClient.EXPECT().UploadFileV2(gomock.Any()).Return(nil, uploadError)

		testSender := &SenderSlack{
			apiKey:       "xoxb-test-token",
			channel:      "#test-channel",
			logger:       mockLogger,
			unfurlLinks:  true,
			slackClient:  mockSlackClient,
			tableFormat:  "enhanced",
			maxTableRows: 20,
		}

		fileContent := []byte("test file content")
		fileBlock := &issuepkg.FileBlock{
			Filename: "test.txt",
			Contents: fileContent,
			MimeType: "text/plain",
			Size:     int64(len(fileContent)),
		}

		slackBlock := testSender.convertFileBlockToSlack(fileBlock)

		sectionBlock, ok := slackBlock.(*slack.SectionBlock)
		assert.True(t, ok, "Expected section block")

		text := sectionBlock.Text.Text
		assert.Contains(t, text, "📎 *File: test.txt* (upload failed)")
		assert.Contains(t, text, "Size:")
		assert.Contains(t, text, "Type: text/plain")
		assert.Contains(t, text, "Content preview:")
	})

	t.Run("formats divider blocks correctly", func(t *testing.T) {
		dividerBlock := &issuepkg.DividerBlock{}

		slackBlock := slackSender.convertBlockToSlack(dividerBlock)

		_, ok := slackBlock.(*slack.DividerBlock)
		assert.True(t, ok, "Expected divider block")
	})

	t.Run("handles unknown block types gracefully", func(t *testing.T) {
		// Create a mock unknown block type
		unknownBlock := &mockUnsupportedBlock{blockType: "unsupported-test-type"}

		slackBlock := slackSender.convertBlockToSlack(unknownBlock)

		sectionBlock, ok := slackBlock.(*slack.SectionBlock)
		assert.True(t, ok, "Expected section block for unknown type")
		assert.Contains(t, sectionBlock.Text.Text, "Unknown block type: unsupported-test-type")
	})

	t.Run("adaptive formatting - simple table format", func(t *testing.T) {
		// Set table formatting parameters for simple formatting
		slackSender.SetTableFormat("simple")
		slackSender.SetMaxTableRows(20)

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
		assert.Contains(t, text, "• key1 `value1`")
		assert.Contains(t, text, "• key2 `value2`")
	})

	t.Run("adaptive formatting - enhanced table format", func(t *testing.T) {
		// Set table formatting parameters for enhanced formatting
		slackSender.SetTableFormat("enhanced")
		slackSender.SetMaxTableRows(20)

		tableBlock := &issuepkg.TableBlock{
			Headers:   []string{"Label", "Value"},
			TableName: "Enhanced Table",
			Rows: [][]string{
				{"key1", "value1"},
				{"key2", "value2"},
			},
		}

		slackBlock := slackSender.convertTableBlockToSlack(tableBlock)

		sectionBlock, ok := slackBlock.(*slack.SectionBlock)
		assert.True(t, ok, "Expected section block")

		text := sectionBlock.Text.Text
		assert.Contains(t, text, "*Enhanced Table*")
		assert.Contains(t, text, "▸ *key1*: `value1`")
		assert.Contains(t, text, "▸ *key2*: `value2`")
	})

	t.Run("adaptive formatting - attachment table format", func(t *testing.T) {
		// Set table formatting parameters for attachment formatting
		slackSender.SetTableFormat("attachment")
		slackSender.SetMaxTableRows(20)

		tableBlock := &issuepkg.TableBlock{
			Headers:   []string{"Label", "Value"},
			TableName: "Attachment Table",
			Rows: [][]string{
				{"key1", "value1"},
				{"key2", "value2"},
			},
		}

		slackBlock := slackSender.convertTableBlockToSlack(tableBlock)

		sectionBlock, ok := slackBlock.(*slack.SectionBlock)
		assert.True(t, ok, "Expected section block")

		text := sectionBlock.Text.Text
		assert.Contains(t, text, "📊 *Attachment Table*")
		assert.Contains(t, text, "└ key1: `value1`")
		assert.Contains(t, text, "└ key2: `value2`")
	})

	t.Run("adaptive formatting - large table exceeding row limit", func(t *testing.T) {
		// Set table formatting parameters with low row limit
		slackSender.SetTableFormat("enhanced")
		slackSender.SetMaxTableRows(2) // Set low limit to trigger file conversion

		// Create table with more rows than the limit
		tableBlock := &issuepkg.TableBlock{
			Headers:   []string{"Label", "Value"},
			TableName: "Large Table",
			Rows: [][]string{
				{"row1", "value1"},
				{"row2", "value2"},
				{"row3", "value3"}, // This exceeds the limit
			},
		}

		slackBlock := slackSender.convertTableBlockToSlack(tableBlock)

		sectionBlock, ok := slackBlock.(*slack.SectionBlock)
		assert.True(t, ok, "Expected section block")

		text := sectionBlock.Text.Text
		assert.Contains(t, text, "📊 *Large Table* (3 rows)")
		assert.Contains(t, text, "Table converted to CSV file (limit: 2 rows)")
		assert.Contains(t, text, "Download CSV file")
	})

	t.Run("adaptive formatting - enhanced multi-column table", func(t *testing.T) {
		// Set table formatting parameters for enhanced formatting
		slackSender.SetTableFormat("enhanced")
		slackSender.SetMaxTableRows(20)

		tableBlock := &issuepkg.TableBlock{
			Headers:   []string{"Name", "Status", "CPU", "Memory"},
			TableName: "Pod Status",
			Rows: [][]string{
				{"pod-1", "Running", "50m", "128Mi"},
				{"pod-2", "Pending", "0", "0"},
			},
		}

		slackBlock := slackSender.convertTableBlockToSlack(tableBlock)

		sectionBlock, ok := slackBlock.(*slack.SectionBlock)
		assert.True(t, ok, "Expected section block")

		text := sectionBlock.Text.Text
		assert.Contains(t, text, "*Pod Status*")
		assert.Contains(t, text, "```")
		assert.Contains(t, text, "Name")
		assert.Contains(t, text, "Status")
		assert.Contains(t, text, "CPU")
		assert.Contains(t, text, "Memory")
		assert.Contains(t, text, "pod-1")
		assert.Contains(t, text, "Running")
	})

	t.Run("no table formatting config uses default simple format", func(t *testing.T) {
		// Clear table formatting parameters (use defaults)
		slackSender.SetTableFormat("")
		slackSender.SetMaxTableRows(0)

		tableBlock := &issuepkg.TableBlock{
			Headers:   []string{"Label", "Value"},
			TableName: "Default Table",
			Rows: [][]string{
				{"key1", "value1"},
				{"key2", "value2"},
			},
		}

		slackBlock := slackSender.convertTableBlockToSlack(tableBlock)

		sectionBlock, ok := slackBlock.(*slack.SectionBlock)
		assert.True(t, ok, "Expected section block")

		text := sectionBlock.Text.Text
		assert.Contains(t, text, "*Default Table*")
		assert.Contains(t, text, "• key1 `value1`") // Simple format
		assert.Contains(t, text, "• key2 `value2`")
	})
}

// Test threading functionality

func TestSenderSlack_GenerateFingerprint(t *testing.T) {
	slackSender, _, _ := setupSenderSlackTest(t)

	// Test with issue that has existing fingerprint
	issueWithFingerprint := &issuepkg.Issue{
		Title:       "Test Issue",
		Fingerprint: "existing-fingerprint-123",
	}

	fingerprint := slackSender.generateFingerprint(issueWithFingerprint)
	assert.Equal(t, "existing-fingerprint-123", fingerprint)

	// Test with issue without fingerprint - create new sender with additional Debug expectation
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	testSender := &SenderSlack{
		logger: mockLogger,
	}

	subject := issuepkg.NewSubject("test-pod", issuepkg.SubjectTypePod)
	subject.Namespace = "default"

	issueWithoutFingerprint := &issuepkg.Issue{
		Title:    "Test Issue Without Fingerprint",
		Source:   issuepkg.SourcePrometheus,
		Subject:  subject,
		StartsAt: time.Unix(1640995200, 0), // Fixed timestamp for test
	}

	fingerprint = testSender.generateFingerprint(issueWithoutFingerprint)
	assert.NotEmpty(t, fingerprint)
	assert.Contains(t, fingerprint, "alert:")
}

func TestSenderSlack_SetThreadManager(t *testing.T) {
	slackSender, _, _ := setupSenderSlackTest(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockThreadManager := mocks.NewMockSlackThreadManagerInterface(ctrl)
	slackSender.SetThreadManager(mockThreadManager)

	assert.Equal(t, mockThreadManager, slackSender.threadManager)
}

func TestSenderSlack_EnableThreading(t *testing.T) {
	slackSender, _, _ := setupSenderSlackTest(t)

	// Enable threading with valid configuration
	cacheTTL := 10 * time.Minute
	searchLimit := 50
	searchWindow := 24 * time.Hour

	slackSender.EnableThreading(cacheTTL, searchLimit, searchWindow)

	// Verify that threadManager was set (we can't check internal details)
	assert.NotNil(t, slackSender.threadManager)
}

func TestSenderSlack_SetLogger(t *testing.T) {
	slackSender, _, _ := setupSenderSlackTest(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	newMockLogger := mocks.NewMockLoggerInterface(ctrl)
	slackSender.SetLogger(newMockLogger)

	assert.Equal(t, newMockLogger, slackSender.logger)
}

func TestSenderSlack_ConvertMarkdownBlockToSlack(t *testing.T) {
	slackSender, _, _ := setupSenderSlackTest(t)

	markdownBlock := &issuepkg.MarkdownBlock{
		Text: "**This is bold text** and *this is italic*",
	}

	slackBlock := slackSender.convertMarkdownBlockToSlack(markdownBlock)

	sectionBlock, ok := slackBlock.(*slack.SectionBlock)
	assert.True(t, ok, "Expected section block")
	assert.Equal(t, "**This is bold text** and *this is italic*", sectionBlock.Text.Text)
	assert.Equal(t, "mrkdwn", sectionBlock.Text.Type)
}

func TestSenderSlack_ConvertBlockToSlack_WithMarkdown(t *testing.T) {
	slackSender, _, _ := setupSenderSlackTest(t)

	// Test with markdown block (this should improve coverage of convertBlockToSlack)
	markdownBlock := &issuepkg.MarkdownBlock{
		Text: "# Header\n\nSome **bold** text",
	}

	slackBlock := slackSender.convertBlockToSlack(markdownBlock)

	sectionBlock, ok := slackBlock.(*slack.SectionBlock)
	assert.True(t, ok, "Expected section block")
	assert.Equal(t, "# Header\n\nSome **bold** text", sectionBlock.Text.Text)
	assert.Equal(t, "mrkdwn", sectionBlock.Text.Type)
}

func TestSenderSlack_ConvertBlockToSlack_WithUnsupportedBlock(t *testing.T) {
	slackSender, _, _ := setupSenderSlackTest(t)

	// Test with unsupported block type (should return fallback section block)
	mockBlock := &mockUnsupportedBlock{blockType: "unsupported-test-type"}
	slackBlock := slackSender.convertBlockToSlack(mockBlock)

	// Should return a section block with fallback text
	assert.NotNil(t, slackBlock, "Expected non-nil block for unsupported type")
	sectionBlock, ok := slackBlock.(*slack.SectionBlock)
	assert.True(t, ok, "Expected section block for unsupported type")
	assert.Contains(t, sectionBlock.Text.Text, "Unknown block type: unsupported-test-type")
}

func TestSenderSlack_ConvertBlockToSlack_AllTypes(t *testing.T) {
	slackSender, _, _ := setupSenderSlackTest(t)

	// Test all supported block types
	tests := []struct {
		name  string
		block issuepkg.BaseBlock
	}{
		{
			name: "TableBlock",
			block: &issuepkg.TableBlock{
				Headers:   []string{"Key", "Value"},
				TableName: "Test Table",
				Rows: [][]string{
					{"key1", "value1"},
				},
			},
		},
		{
			name: "JsonBlock",
			block: &issuepkg.JsonBlock{
				Data: map[string]string{"test": "value"},
			},
		},
		{
			name: "MarkdownBlock",
			block: &issuepkg.MarkdownBlock{
				Text: "**Bold text**",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slackBlock := slackSender.convertBlockToSlack(tt.block)
			assert.NotNil(t, slackBlock, "Expected non-nil block for %s", tt.name)

			sectionBlock, ok := slackBlock.(*slack.SectionBlock)
			assert.True(t, ok, "Expected section block for %s", tt.name)
			assert.NotEmpty(t, sectionBlock.Text.Text, "Expected non-empty text for %s", tt.name)
		})
	}
}

func TestSenderSlack_EnrichmentBlocks_WithMarkdown(t *testing.T) {
	slackSender, _, _ := setupSenderSlackTest(t)

	issue := &issuepkg.Issue{
		Title:    "Test Alert with Markdown Enrichment",
		Severity: issuepkg.SeverityInfo,
		Status:   issuepkg.StatusFiring,
		Source:   issuepkg.SourcePrometheus,
	}

	// Add markdown enrichment
	markdownEnrichment := issuepkg.NewEnrichmentWithType(issuepkg.EnrichmentTypeAIAnalysis, "AI Analysis")
	markdownBlock := &issuepkg.MarkdownBlock{
		Text: "## Analysis Result\n\nThis alert indicates a **high CPU usage** on the pod.",
	}
	markdownEnrichment.AddBlock(markdownBlock)
	issue.AddEnrichment(*markdownEnrichment)

	// Test enrichment blocks building
	enrichmentBlocks := slackSender.buildEnrichmentBlocks(issue.Enrichments)

	// Should have 3 blocks: header + section + divider
	assert.Len(t, enrichmentBlocks, 3)

	// Verify title block
	titleBlock, ok := enrichmentBlocks[0].(*slack.SectionBlock)
	assert.True(t, ok, "First block should be section block with title")
	assert.Equal(t, "*AI Analysis*", titleBlock.Text.Text)

	// Verify the markdown block was converted to a section block
	contentBlock := enrichmentBlocks[1]
	sectionBlock, ok := contentBlock.(*slack.SectionBlock)
	assert.True(t, ok, "Second block should be section block for markdown content")
	assert.Contains(t, sectionBlock.Text.Text, "## Analysis Result")
	assert.Contains(t, sectionBlock.Text.Text, "**high CPU usage**")
	assert.Equal(t, "mrkdwn", sectionBlock.Text.Type)

	// Verify divider block
	_, ok = enrichmentBlocks[2].(*slack.DividerBlock)
	assert.True(t, ok, "Third block should be divider")
}

func TestSenderSlack_TableToCSV(t *testing.T) {
	slackSender, _, _ := setupSenderSlackTest(t)

	// Create table with headers and data
	table := &issuepkg.TableBlock{
		TableName: "Test Table",
		Headers:   []string{"Name", "Value", "Description"},
		Rows: [][]string{
			{"pod1", "ready", "Pod is running"},
			{"pod2", "pending", "Pod with, comma"},
			{"pod3", "failed", "Pod with \"quotes\""},
		},
	}

	// Convert to CSV
	csvContent := slackSender.tableToCSV(table)

	// Verify CSV format
	lines := strings.Split(csvContent, "\n")
	assert.GreaterOrEqual(t, len(lines), 4) // headers + 3 rows + empty line

	// Verify headers
	assert.Equal(t, "Name,Value,Description", lines[0])

	// Verify data rows
	assert.Equal(t, "pod1,ready,Pod is running", lines[1])
	assert.Equal(t, "pod2,pending,\"Pod with, comma\"", lines[2])        // CSV escaping
	assert.Equal(t, "pod3,failed,\"Pod with \"\"quotes\"\"\"", lines[3]) // Quote escaping
}
