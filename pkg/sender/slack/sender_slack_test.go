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

	slackapi "github.com/slack-go/slack"

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
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockSlackClient := mocks.NewMockSlackClientInterface(ctrl)

	// Mock channel resolution for #test-channel (used by many tests)
	testChannel := slackapi.Channel{}
	testChannel.ID = "C123TEST"
	testChannel.Name = "test-channel"
	mockSlackClient.EXPECT().GetConversations(gomock.Any()).Return(
		[]slackapi.Channel{testChannel},
		"",
		nil,
	).AnyTimes()

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
				Title: "Alert Labels",
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
	_, ok := blocks[0].(*slackapi.SectionBlock)
	assert.True(t, ok, "First block should be section block (header)")

	// Should find runbook URL block
	found := false
	for _, block := range blocks {
		if sectionBlock, ok := block.(*slackapi.SectionBlock); ok {
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
				Title: "Alert Labels",
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
	headerBlock, ok := blocks[0].(*slackapi.SectionBlock)
	assert.True(t, ok, "First block should be section block (header)")
	assert.Contains(t, headerBlock.Text.Text, "Alert resolved")
	assert.Contains(t, headerBlock.Text.Text, "✅")

	// Should contain source information
	foundSource := false
	foundResolved := false
	for _, block := range blocks {
		if sectionBlock, ok := block.(*slackapi.SectionBlock); ok {
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
		if sectionBlock, ok := block.(*slackapi.SectionBlock); ok {
			assert.NotContains(t, sectionBlock.Text.Text, "Alert:")
			assert.NotContains(t, sectionBlock.Text.Text, "Runbook URL:")
		}
		// Should not have action blocks (links)
		_, isActionBlock := block.(*slackapi.ActionBlock)
		assert.False(t, isActionBlock, "Resolved alerts should not have action blocks")
	}
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
		if sectionBlock, ok := block.(*slackapi.SectionBlock); ok {
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

	attachments := slackSender.buildSlackAttachments(issue, issue.Enrichments)

	// Should have one attachment (metadata, since no alert labels enrichment)
	assert.Len(t, attachments, 1)

	attachment := attachments[0]

	// Metadata attachment is always yellow
	assert.Equal(t, "#FFCC00", attachment.Color)

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

	attachments := slackSender.buildSlackAttachments(issue, issue.Enrichments)

	// Should have one attachment
	assert.Len(t, attachments, 1)

	attachment := attachments[0]

	// Should have blocks with cluster information
	assert.NotEmpty(t, attachment.Blocks.BlockSet)

	// Convert blocks to find cluster info
	found := false
	for _, block := range attachment.Blocks.BlockSet {
		if sectionBlock, ok := block.(*slackapi.SectionBlock); ok {
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

	attachments := slackSender.buildSlackAttachments(issue, issue.Enrichments)

	// Should have one attachment with time information
	assert.Len(t, attachments, 1)

	attachment := attachments[0]
	assert.NotEmpty(t, attachment.Blocks.BlockSet)

	// Find the time block and verify UTC formatting
	foundTimeBlock := false
	for _, block := range attachment.Blocks.BlockSet {
		if sectionBlock, ok := block.(*slackapi.SectionBlock); ok {
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

	attachments := slackSender.buildSlackAttachments(issue, issue.Enrichments)

	// Should have one attachment but without time information
	assert.Len(t, attachments, 1)

	attachment := attachments[0]
	assert.NotEmpty(t, attachment.Blocks.BlockSet)

	// Verify that no time block exists
	for _, block := range attachment.Blocks.BlockSet {
		if sectionBlock, ok := block.(*slackapi.SectionBlock); ok {
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
	mockSlackClient.EXPECT().UploadFileV2(gomock.Any()).Return(&slackapi.FileSummary{
		ID:    "F123456789",
		Title: "test-file.csv",
	}, nil).AnyTimes()

	// Add mock for GetFileInfo to support permalink retrieval
	mockSlackClient.EXPECT().GetFileInfo("F123456789", 0, 0).Return(&slackapi.File{
		ID:        "F123456789",
		Name:      "test-file.csv",
		Permalink: "https://files.slackapi.com/files-pri/T123/F123456789/test-file.csv",
	}, nil, nil, nil).AnyTimes()

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

		// Should have 4 blocks: 2 enrichments * (header + section) = 4 blocks (dividers removed)
		assert.Len(t, enrichmentBlocks, 4)

		// Verify first enrichment blocks (labels)
		sectionBlock1, ok := enrichmentBlocks[0].(*slackapi.SectionBlock)
		assert.True(t, ok, "First block should be section block with title")
		assert.Equal(t, "*Alert Labels*", sectionBlock1.Text.Text)

		// Verify second enrichment blocks (annotations)
		sectionBlock2, ok := enrichmentBlocks[2].(*slackapi.SectionBlock)
		assert.True(t, ok, "Third block should be section block with title")
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

		// Should have 2 blocks: header + section (dividers removed in Phase 2)
		assert.Len(t, enrichmentBlocks, 2)

		// Verify title block
		titleBlock, ok := enrichmentBlocks[0].(*slackapi.SectionBlock)
		assert.True(t, ok, "First block should be section block with title")
		assert.Equal(t, "*Alert Labels (JSON)*", titleBlock.Text.Text)

		// Verify the JSON block was converted to a section block
		contentBlock := enrichmentBlocks[1]
		sectionBlock, ok := contentBlock.(*slackapi.SectionBlock)
		assert.True(t, ok, "Second block should be section block for JSON content")
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

		sectionBlock, ok := slackBlock.(*slackapi.SectionBlock)
		assert.True(t, ok, "Expected section block")

		text := sectionBlock.Text.Text
		assert.Contains(t, text, "*Test Table*")
		assert.Contains(t, text, "● key1  `value1`")
		assert.Contains(t, text, "● key2  `value2`")
	})

	t.Run("formats json blocks correctly", func(t *testing.T) {
		jsonBlock := &issuepkg.JsonBlock{
			Data: map[string]string{
				"test": "value",
			},
		}

		slackBlock := slackSender.convertJsonBlockToSlack(jsonBlock)

		sectionBlock, ok := slackBlock.(*slackapi.SectionBlock)
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

		headerSlackBlock, ok := slackBlock.(*slackapi.HeaderBlock)
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

		sectionBlock, ok := slackBlock.(*slackapi.SectionBlock)
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

		sectionBlock, ok := slackBlock.(*slackapi.SectionBlock)
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

		actionBlock, ok := slackBlock.(*slackapi.ActionBlock)
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
		mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
		mockSlackClient := mocks.NewMockSlackClientInterface(ctrl)

		// Mock channel resolution
		testChannel := slackapi.Channel{}
		testChannel.ID = "C123TEST"
		testChannel.Name = "test-channel"
		mockSlackClient.EXPECT().GetConversations(gomock.Any()).Return(
			[]slackapi.Channel{testChannel},
			"",
			nil,
		).AnyTimes()
		// Mock upload failure for this test to avoid complexity
		uploadError := fmt.Errorf("test upload error")
		mockSlackClient.EXPECT().UploadFileV2(gomock.Any()).Return(nil, uploadError).Times(2)

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

		sectionBlock, ok := slackBlock.(*slackapi.SectionBlock)
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

		_, ok := slackBlock.(*slackapi.DividerBlock)
		assert.True(t, ok, "Expected divider block")
	})

	t.Run("handles unknown block types gracefully", func(t *testing.T) {
		// Create a mock unknown block type
		unknownBlock := &mockUnsupportedBlock{blockType: "unsupported-test-type"}

		slackBlock := slackSender.convertBlockToSlack(unknownBlock)

		sectionBlock, ok := slackBlock.(*slackapi.SectionBlock)
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

		sectionBlock, ok := slackBlock.(*slackapi.SectionBlock)
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

		sectionBlock, ok := slackBlock.(*slackapi.SectionBlock)
		assert.True(t, ok, "Expected section block")

		text := sectionBlock.Text.Text
		assert.Contains(t, text, "*Enhanced Table*")
		assert.Contains(t, text, "● key1  `value1`")
		assert.Contains(t, text, "● key2  `value2`")
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

		sectionBlock, ok := slackBlock.(*slackapi.SectionBlock)
		assert.True(t, ok, "Expected section block")

		text := sectionBlock.Text.Text
		assert.Contains(t, text, "📊 *Attachment Table*")
		assert.Contains(t, text, "● key1  `value1`")
		assert.Contains(t, text, "● key2  `value2`")
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

		sectionBlock, ok := slackBlock.(*slackapi.SectionBlock)
		assert.True(t, ok, "Expected section block")

		text := sectionBlock.Text.Text
		assert.Contains(t, text, "📊 *Large Table* (3 rows)")
		assert.Contains(t, text, "Table converted to CSV file (limit: 2 rows)")
		assert.Contains(t, text, "View CSV File")
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

		sectionBlock, ok := slackBlock.(*slackapi.SectionBlock)
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

		sectionBlock, ok := slackBlock.(*slackapi.SectionBlock)
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

	sectionBlock, ok := slackBlock.(*slackapi.SectionBlock)
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

	sectionBlock, ok := slackBlock.(*slackapi.SectionBlock)
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
	sectionBlock, ok := slackBlock.(*slackapi.SectionBlock)
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

			sectionBlock, ok := slackBlock.(*slackapi.SectionBlock)
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

	// Should have 2 blocks: header + section (dividers removed in Phase 2)
	assert.Len(t, enrichmentBlocks, 2)

	// Verify title block
	titleBlock, ok := enrichmentBlocks[0].(*slackapi.SectionBlock)
	assert.True(t, ok, "First block should be section block with title")
	assert.Equal(t, "*AI Analysis*", titleBlock.Text.Text)

	// Verify the markdown block was converted to a section block
	contentBlock := enrichmentBlocks[1]
	sectionBlock, ok := contentBlock.(*slackapi.SectionBlock)
	assert.True(t, ok, "Second block should be section block for markdown content")
	assert.Contains(t, sectionBlock.Text.Text, "## Analysis Result")
	assert.Contains(t, sectionBlock.Text.Text, "**high CPU usage**")
	assert.Equal(t, "mrkdwn", sectionBlock.Text.Type)
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

func TestSenderSlack_CreateTableErrorBlock(t *testing.T) {
	slackSender, _, _ := setupSenderSlackTest(t)

	table := &issuepkg.TableBlock{
		TableName: "Test Table",
		Headers:   []string{"Name", "Status"},
		Rows: [][]string{
			{"pod1", "ready"},
			{"pod2", "pending"},
			{"pod3", "failed"},
		},
	}

	err := fmt.Errorf("file upload failed: network timeout")
	block := slackSender.createTableErrorBlock(table, err)

	sectionBlock, ok := block.(*slackapi.SectionBlock)
	assert.True(t, ok, "Expected section block")

	text := sectionBlock.Text.Text
	assert.Contains(t, text, "📊 *Test Table* (3 rows) - upload failed")
	assert.Contains(t, text, "Error: file upload failed: network timeout")
	assert.Contains(t, text, "Showing simplified table view:")
	assert.Contains(t, text, "*Headers:* Name | Status")
	assert.Contains(t, text, "• pod1 | ready")
	assert.Contains(t, text, "• pod2 | pending")
}

func TestSenderSlack_CreateTableErrorBlock_WithoutTableName(t *testing.T) {
	slackSender, _, _ := setupSenderSlackTest(t)

	table := &issuepkg.TableBlock{
		Headers: []string{"Key", "Value"},
		Rows: [][]string{
			{"key1", "value1"},
			{"key2", "value2"},
		},
	}

	err := fmt.Errorf("upload error")
	block := slackSender.createTableErrorBlock(table, err)

	sectionBlock, ok := block.(*slackapi.SectionBlock)
	assert.True(t, ok)

	text := sectionBlock.Text.Text
	assert.Contains(t, text, "📊 *Large Table* (2 rows) - upload failed")
}

func TestSenderSlack_CreateTableErrorBlock_ManyRows(t *testing.T) {
	slackSender, _, _ := setupSenderSlackTest(t)

	// Create table with more than 5 rows to test truncation
	rows := make([][]string, 10)
	for i := 0; i < 10; i++ {
		rows[i] = []string{fmt.Sprintf("row%d", i+1), "value"}
	}

	table := &issuepkg.TableBlock{
		TableName: "Large Table",
		Headers:   []string{"Name", "Value"},
		Rows:      rows,
	}

	err := fmt.Errorf("table too large")
	block := slackSender.createTableErrorBlock(table, err)

	sectionBlock, ok := block.(*slackapi.SectionBlock)
	assert.True(t, ok)

	text := sectionBlock.Text.Text
	assert.Contains(t, text, "... and 5 more rows") // Should show truncation
	assert.Contains(t, text, "• row1 | value")      // First row shown
	assert.Contains(t, text, "• row5 | value")      // Last row before truncation
}

func TestSenderSlack_ConvertFileBlockToSlack_ErrorPath(t *testing.T) {
	// Setup dedicated mock environment for error testing
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
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	mockSlackClient := mocks.NewMockSlackClientInterface(ctrl)
	// Mock channel resolution
	mockChannel := slackapi.Channel{}
	mockChannel.ID = "C123TEST"
	mockChannel.Name = "test-channel"
	mockSlackClient.EXPECT().GetConversations(gomock.Any()).Return(
		[]slackapi.Channel{mockChannel},
		"",
		nil,
	).AnyTimes()
	// Mock file upload failure
	uploadError := fmt.Errorf("file too large")
	mockSlackClient.EXPECT().UploadFileV2(gomock.Any()).Return(nil, uploadError).Times(2)

	slackSender := &SenderSlack{
		apiKey:       "xoxb-test-token",
		channel:      "#test-channel",
		logger:       mockLogger,
		unfurlLinks:  true,
		slackClient:  mockSlackClient,
		tableFormat:  "enhanced",
		maxTableRows: 20,
	}

	// Create file block
	fileBlock := issuepkg.NewFileBlock("large-file.log", []byte("very large content"), "text/plain")

	// Convert to Slack block (should use error fallback)
	slackBlock := slackSender.convertFileBlockToSlack(fileBlock)

	// Verify error fallback behavior
	sectionBlock, ok := slackBlock.(*slackapi.SectionBlock)
	assert.True(t, ok)

	text := sectionBlock.Text.Text
	assert.Contains(t, text, "📎 *File: large-file.log* (upload failed)")
	assert.Contains(t, text, "Error: file upload failed: upload: file too large")
	assert.Contains(t, text, "Content preview:")
	assert.Contains(t, text, "very large content")
}

func TestSenderSlack_ConvertFileBlockToSlack_ErrorPath_BinaryFile(t *testing.T) {
	// Setup dedicated mock environment for error testing
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
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	mockSlackClient := mocks.NewMockSlackClientInterface(ctrl)
	// Mock channel resolution
	mockChannel := slackapi.Channel{}
	mockChannel.ID = "C123TEST"
	mockChannel.Name = "test-channel"
	mockSlackClient.EXPECT().GetConversations(gomock.Any()).Return(
		[]slackapi.Channel{mockChannel},
		"",
		nil,
	).AnyTimes()
	// Mock file upload failure
	uploadError := fmt.Errorf("binary file not supported")
	mockSlackClient.EXPECT().UploadFileV2(gomock.Any()).Return(nil, uploadError).Times(2)

	slackSender := &SenderSlack{
		apiKey:       "xoxb-test-token",
		channel:      "#test-channel",
		logger:       mockLogger,
		unfurlLinks:  true,
		slackClient:  mockSlackClient,
		tableFormat:  "enhanced",
		maxTableRows: 20,
	}

	// Create binary file block (no preview should be shown)
	fileBlock := issuepkg.NewFileBlock("image.png", []byte("binary data"), "image/png")

	slackBlock := slackSender.convertFileBlockToSlack(fileBlock)

	sectionBlock, ok := slackBlock.(*slackapi.SectionBlock)
	assert.True(t, ok)

	text := sectionBlock.Text.Text
	assert.Contains(t, text, "📎 *File: image.png* (upload failed)")
	assert.Contains(t, text, "Type: image/png")
	assert.NotContains(t, text, "Content preview:") // No preview for non-text files
}

func TestSenderSlack_ConvertFileBlockToSlack_SuccessPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()

	mockSlackClient := mocks.NewMockSlackClientInterface(ctrl)

	// Mock channel resolution
	mockChannel := slackapi.Channel{}
	mockChannel.ID = "C123TEST"
	mockChannel.Name = "test-channel"
	mockSlackClient.EXPECT().GetConversations(gomock.Any()).Return(
		[]slackapi.Channel{mockChannel},
		"",
		nil,
	).AnyTimes()

	// Mock successful file upload
	mockSlackClient.EXPECT().UploadFileV2(gomock.Any()).Return(&slackapi.FileSummary{
		ID:    "F123TEST",
		Title: "test.log",
	}, nil)

	// Mock GetFileInfo to return permalink
	mockSlackClient.EXPECT().GetFileInfo("F123TEST", 0, 0).Return(&slackapi.File{
		ID:        "F123TEST",
		Name:      "test.log",
		Permalink: "https://files.slackapi.com/files-pri/T123/F123TEST/test.log",
	}, nil, nil, nil)

	slackSender := &SenderSlack{
		apiKey:      "xoxb-test-token",
		channel:     "#test-channel",
		logger:      mockLogger,
		slackClient: mockSlackClient,
	}

	// Create file block
	fileBlock := issuepkg.NewFileBlock("test.log", []byte("log content here"), "text/plain")

	slackBlock := slackSender.convertFileBlockToSlack(fileBlock)

	sectionBlock, ok := slackBlock.(*slackapi.SectionBlock)
	assert.True(t, ok)

	text := sectionBlock.Text.Text
	assert.Contains(t, text, "📎 *test.log*")
	assert.Contains(t, text, "KB")
	assert.Contains(t, text, "text/plain")
	assert.Contains(t, text, "https://files.slackapi.com/files-pri/T123/F123TEST/test.log")
	assert.Contains(t, text, "View File")
}

func TestSenderSlack_ConvertFileBlockToSlack_SuccessPath_LargeFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()

	mockSlackClient := mocks.NewMockSlackClientInterface(ctrl)

	// Mock channel resolution
	mockChannel := slackapi.Channel{}
	mockChannel.ID = "C123TEST"
	mockChannel.Name = "test-channel"
	mockSlackClient.EXPECT().GetConversations(gomock.Any()).Return(
		[]slackapi.Channel{mockChannel},
		"",
		nil,
	).AnyTimes()

	mockSlackClient.EXPECT().UploadFileV2(gomock.Any()).Return(&slackapi.FileSummary{
		ID:    "F456TEST",
		Title: "large.csv",
	}, nil)

	mockSlackClient.EXPECT().GetFileInfo("F456TEST", 0, 0).Return(&slackapi.File{
		ID:        "F456TEST",
		Name:      "large.csv",
		Permalink: "https://files.slackapi.com/files-pri/T123/F456TEST/large.csv",
	}, nil, nil, nil)

	slackSender := &SenderSlack{
		apiKey:      "xoxb-test-token",
		channel:     "#test-channel",
		logger:      mockLogger,
		slackClient: mockSlackClient,
	}

	// Create large file block (>1MB)
	largeContent := make([]byte, 2*1024*1024) // 2MB
	fileBlock := issuepkg.NewFileBlock("large.csv", largeContent, "text/csv")

	slackBlock := slackSender.convertFileBlockToSlack(fileBlock)

	sectionBlock, ok := slackBlock.(*slackapi.SectionBlock)
	assert.True(t, ok)

	text := sectionBlock.Text.Text
	assert.Contains(t, text, "📎 *large.csv*")
	assert.Contains(t, text, "MB") // Should display in MB, not KB
	assert.Contains(t, text, "text/csv")
	assert.Contains(t, text, "View File")
}

func TestSenderSlack_ConvertFileBlockToSlack_GetFileInfoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()

	mockSlackClient := mocks.NewMockSlackClientInterface(ctrl)

	// Mock channel resolution
	mockChannel := slackapi.Channel{}
	mockChannel.ID = "C123TEST"
	mockChannel.Name = "test-channel"
	mockSlackClient.EXPECT().GetConversations(gomock.Any()).Return(
		[]slackapi.Channel{mockChannel},
		"",
		nil,
	).AnyTimes()

	// Mock successful upload but GetFileInfo fails (both direct and fallback attempts)
	mockSlackClient.EXPECT().UploadFileV2(gomock.Any()).Return(&slackapi.FileSummary{
		ID:    "F789TEST",
		Title: "test.log",
	}, nil).Times(2)

	mockSlackClient.EXPECT().GetFileInfo("F789TEST", 0, 0).Return(
		nil, nil, nil, fmt.Errorf("file not found"),
	).Times(2)

	slackSender := &SenderSlack{
		apiKey:      "xoxb-test-token",
		channel:     "#test-channel",
		logger:      mockLogger,
		slackClient: mockSlackClient,
	}

	fileBlock := issuepkg.NewFileBlock("test.log", []byte("content"), "text/plain")

	slackBlock := slackSender.convertFileBlockToSlack(fileBlock)

	// Should fall back to error display
	sectionBlock, ok := slackBlock.(*slackapi.SectionBlock)
	assert.True(t, ok)

	text := sectionBlock.Text.Text
	assert.Contains(t, text, "📎 *File: test.log* (upload failed)")
	assert.Contains(t, text, "Error: file upload failed: get file info: file not found")
}

// ============================================================================
// Helper Function Tests
// ============================================================================

func TestSenderSlack_GetLinkEmoji(t *testing.T) {
	slackSender, _, _ := setupSenderSlackTest(t)

	tests := []struct {
		name     string
		linkType issuepkg.LinkType
		expected string
	}{
		{"Investigate link", issuepkg.LinkTypeInvestigate, "🔍"},
		{"Silence link", issuepkg.LinkTypeSilence, "🔕"},
		{"Prometheus link", issuepkg.LinkTypePrometheusGenerator, "📊"},
		{"General link", issuepkg.LinkTypeGeneral, "🔗"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := slackSender.getLinkEmoji(tt.linkType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSenderSlack_GetLinkButtonStyle(t *testing.T) {
	slackSender, _, _ := setupSenderSlackTest(t)

	tests := []struct {
		name     string
		linkType issuepkg.LinkType
		expected string
	}{
		{"Investigate button", issuepkg.LinkTypeInvestigate, "primary"},
		{"Silence button", issuepkg.LinkTypeSilence, "danger"},
		{"General button", issuepkg.LinkTypeGeneral, ""},
		{"Prometheus button", issuepkg.LinkTypePrometheusGenerator, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := slackSender.getLinkButtonStyle(tt.linkType)
			assert.Equal(t, tt.expected, string(result))
		})
	}
}

func TestSenderSlack_GetSeverityText(t *testing.T) {
	slackSender, _, _ := setupSenderSlackTest(t)

	tests := []struct {
		name     string
		severity issuepkg.Severity
		expected string
	}{
		{"High severity", issuepkg.SeverityHigh, "🔴 High"},
		{"Low severity", issuepkg.SeverityLow, "🟡 Low"},
		{"Info severity", issuepkg.SeverityInfo, "🟢 Info"},
		{"Debug severity", issuepkg.SeverityDebug, "🔵 Debug"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := slackSender.getSeverityText(tt.severity)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ============================================================================
// Deduplication Tests
// ============================================================================

func TestSenderSlack_DeduplicateEnrichments(t *testing.T) {
	slackSender, _, _ := setupSenderSlackTest(t)

	t.Run("removes duplicate enrichments with same type and title", func(t *testing.T) {
		enrichments := []issuepkg.Enrichment{
			*issuepkg.NewEnrichmentWithType(issuepkg.EnrichmentTypeAlertLabels, "Alert Labels"),
			*issuepkg.NewEnrichmentWithType(issuepkg.EnrichmentTypeAlertLabels, "Alert Labels"),
			*issuepkg.NewEnrichmentWithType(issuepkg.EnrichmentTypeAlertAnnotations, "Annotations"),
		}

		unique := slackSender.deduplicateEnrichments(enrichments)

		// Should have 2 unique items (first Alert Labels + Annotations)
		assert.Len(t, unique, 2)
		assert.Equal(t, issuepkg.EnrichmentTypeAlertLabels, unique[0].Type)
		assert.Equal(t, issuepkg.EnrichmentTypeAlertAnnotations, unique[1].Type)
	})

	t.Run("keeps enrichments with same type but different titles", func(t *testing.T) {
		enrichments := []issuepkg.Enrichment{
			*issuepkg.NewEnrichmentWithType(issuepkg.EnrichmentTypeTextFile, "Container Logs - app"),
			*issuepkg.NewEnrichmentWithType(issuepkg.EnrichmentTypeTextFile, "Container Logs - sidecar"),
		}

		unique := slackSender.deduplicateEnrichments(enrichments)

		// Should keep both (different titles)
		assert.Len(t, unique, 2)
		assert.Equal(t, "Container Logs - app", unique[0].Title)
		assert.Equal(t, "Container Logs - sidecar", unique[1].Title)
	})

	t.Run("uses first block identifier for uniqueness - file blocks", func(t *testing.T) {
		enrichment1 := issuepkg.NewEnrichmentWithType(issuepkg.EnrichmentTypeTextFile, "Pod Logs")
		enrichment1.AddBlock(issuepkg.NewFileBlock("pod-logs-app.log", []byte("log content"), "text/plain"))

		enrichment2 := issuepkg.NewEnrichmentWithType(issuepkg.EnrichmentTypeTextFile, "Pod Logs")
		enrichment2.AddBlock(issuepkg.NewFileBlock("pod-logs-app.log", []byte("log content"), "text/plain"))

		enrichment3 := issuepkg.NewEnrichmentWithType(issuepkg.EnrichmentTypeTextFile, "Pod Logs")
		enrichment3.AddBlock(issuepkg.NewFileBlock("pod-logs-sidecar.log", []byte("different logs"), "text/plain"))

		enrichments := []issuepkg.Enrichment{*enrichment1, *enrichment2, *enrichment3}
		unique := slackSender.deduplicateEnrichments(enrichments)

		// Should have 2 unique items (duplicate files removed)
		assert.Len(t, unique, 2)

		// Verify filenames are different
		file1, ok1 := unique[0].Blocks[0].(*issuepkg.FileBlock)
		file2, ok2 := unique[1].Blocks[0].(*issuepkg.FileBlock)
		assert.True(t, ok1 && ok2)
		assert.NotEqual(t, file1.Filename, file2.Filename)
	})

	t.Run("uses first block identifier for uniqueness - table blocks", func(t *testing.T) {
		enrichment1 := issuepkg.NewEnrichmentWithType(issuepkg.EnrichmentTypeAlertLabels, "Labels")
		enrichment1.AddBlock(&issuepkg.TableBlock{
			TableName: "Alert Labels",
			Headers:   []string{"Key", "Value"},
			Rows:      [][]string{{"severity", "high"}},
		})

		enrichment2 := issuepkg.NewEnrichmentWithType(issuepkg.EnrichmentTypeAlertLabels, "Labels")
		enrichment2.AddBlock(&issuepkg.TableBlock{
			TableName: "Alert Labels",
			Headers:   []string{"Key", "Value"},
			Rows:      [][]string{{"severity", "high"}},
		})

		enrichment3 := issuepkg.NewEnrichmentWithType(issuepkg.EnrichmentTypeAlertLabels, "Labels")
		enrichment3.AddBlock(&issuepkg.TableBlock{
			TableName: "Alert Annotations",
			Headers:   []string{"Key", "Value"},
			Rows:      [][]string{{"summary", "test"}},
		})

		enrichments := []issuepkg.Enrichment{*enrichment1, *enrichment2, *enrichment3}
		unique := slackSender.deduplicateEnrichments(enrichments)

		// Should have 2 unique items (duplicate table names removed)
		assert.Len(t, unique, 2)

		// Verify table names are different
		table1, ok1 := unique[0].Blocks[0].(*issuepkg.TableBlock)
		table2, ok2 := unique[1].Blocks[0].(*issuepkg.TableBlock)
		assert.True(t, ok1 && ok2)
		assert.NotEqual(t, table1.TableName, table2.TableName)
	})

	t.Run("uses first block identifier for uniqueness - markdown blocks", func(t *testing.T) {
		enrichment1 := issuepkg.NewEnrichmentWithType(issuepkg.EnrichmentTypeAIAnalysis, "Analysis")
		enrichment1.AddBlock(&issuepkg.MarkdownBlock{
			Text: "This is a very long markdown text that will be used for deduplication based on first 50 characters",
		})

		enrichment2 := issuepkg.NewEnrichmentWithType(issuepkg.EnrichmentTypeAIAnalysis, "Analysis")
		enrichment2.AddBlock(&issuepkg.MarkdownBlock{
			Text: "This is a very long markdown text that will be used for deduplication based on first 50 characters",
		})

		enrichment3 := issuepkg.NewEnrichmentWithType(issuepkg.EnrichmentTypeAIAnalysis, "Analysis")
		enrichment3.AddBlock(&issuepkg.MarkdownBlock{
			Text: "Different markdown text with different content",
		})

		enrichments := []issuepkg.Enrichment{*enrichment1, *enrichment2, *enrichment3}
		unique := slackSender.deduplicateEnrichments(enrichments)

		// Should have 2 unique items (duplicate markdown removed)
		assert.Len(t, unique, 2)

		// Verify markdown text is different
		md1, ok1 := unique[0].Blocks[0].(*issuepkg.MarkdownBlock)
		md2, ok2 := unique[1].Blocks[0].(*issuepkg.MarkdownBlock)
		assert.True(t, ok1 && ok2)
		assert.NotEqual(t, md1.Text, md2.Text)
	})

	t.Run("handles enrichments with no blocks", func(t *testing.T) {
		enrichments := []issuepkg.Enrichment{
			*issuepkg.NewEnrichmentWithType(issuepkg.EnrichmentTypeAlertLabels, "Empty 1"),
			*issuepkg.NewEnrichmentWithType(issuepkg.EnrichmentTypeAlertLabels, "Empty 1"),
			*issuepkg.NewEnrichmentWithType(issuepkg.EnrichmentTypeAlertAnnotations, "Empty 2"),
		}

		unique := slackSender.deduplicateEnrichments(enrichments)

		// Should have 2 unique items based on type+title
		assert.Len(t, unique, 2)
	})

	t.Run("handles enrichments without type", func(t *testing.T) {
		enrichment1 := issuepkg.NewEnrichment()
		enrichment1.Title = "Custom Enrichment 1"

		enrichment2 := issuepkg.NewEnrichment()
		enrichment2.Title = "Custom Enrichment 1"

		enrichment3 := issuepkg.NewEnrichment()
		enrichment3.Title = "Custom Enrichment 2"

		enrichments := []issuepkg.Enrichment{*enrichment1, *enrichment2, *enrichment3}
		unique := slackSender.deduplicateEnrichments(enrichments)

		// Should have 2 unique items (duplicate title removed)
		assert.Len(t, unique, 2)
		assert.Equal(t, "Custom Enrichment 1", unique[0].Title)
		assert.Equal(t, "Custom Enrichment 2", unique[1].Title)
	})

	t.Run("handles enrichments without title", func(t *testing.T) {
		enrichment1 := issuepkg.NewEnrichmentWithType(issuepkg.EnrichmentTypeAlertLabels, "")
		enrichment1.Title = ""

		enrichment2 := issuepkg.NewEnrichmentWithType(issuepkg.EnrichmentTypeAlertLabels, "")
		enrichment2.Title = ""

		enrichment3 := issuepkg.NewEnrichmentWithType(issuepkg.EnrichmentTypeAlertAnnotations, "")
		enrichment3.Title = ""

		enrichments := []issuepkg.Enrichment{*enrichment1, *enrichment2, *enrichment3}
		unique := slackSender.deduplicateEnrichments(enrichments)

		// Should have 2 unique items (based on type only)
		assert.Len(t, unique, 2)
	})

	t.Run("handles empty enrichments list", func(t *testing.T) {
		enrichments := []issuepkg.Enrichment{}
		unique := slackSender.deduplicateEnrichments(enrichments)

		assert.Empty(t, unique)
	})

	t.Run("preserves order of first occurrence", func(t *testing.T) {
		enrichments := []issuepkg.Enrichment{
			*issuepkg.NewEnrichmentWithType(issuepkg.EnrichmentTypeAlertAnnotations, "Annotations"),
			*issuepkg.NewEnrichmentWithType(issuepkg.EnrichmentTypeAlertLabels, "Labels"),
			*issuepkg.NewEnrichmentWithType(issuepkg.EnrichmentTypeAlertAnnotations, "Annotations"),
			*issuepkg.NewEnrichmentWithType(issuepkg.EnrichmentTypeTextFile, "Logs"),
		}

		unique := slackSender.deduplicateEnrichments(enrichments)

		// Should preserve order: Annotations, Labels, Logs
		assert.Len(t, unique, 3)
		assert.Equal(t, issuepkg.EnrichmentTypeAlertAnnotations, unique[0].Type)
		assert.Equal(t, issuepkg.EnrichmentTypeAlertLabels, unique[1].Type)
		assert.Equal(t, issuepkg.EnrichmentTypeTextFile, unique[2].Type)
	})

	t.Run("handles markdown blocks shorter than 50 chars", func(t *testing.T) {
		enrichment1 := issuepkg.NewEnrichmentWithType(issuepkg.EnrichmentTypeAIAnalysis, "Short")
		enrichment1.AddBlock(&issuepkg.MarkdownBlock{Text: "Short text"})

		enrichment2 := issuepkg.NewEnrichmentWithType(issuepkg.EnrichmentTypeAIAnalysis, "Short")
		enrichment2.AddBlock(&issuepkg.MarkdownBlock{Text: "Short text"})

		enrichment3 := issuepkg.NewEnrichmentWithType(issuepkg.EnrichmentTypeAIAnalysis, "Short")
		enrichment3.AddBlock(&issuepkg.MarkdownBlock{Text: "Different"})

		enrichments := []issuepkg.Enrichment{*enrichment1, *enrichment2, *enrichment3}
		unique := slackSender.deduplicateEnrichments(enrichments)

		// Should have 2 unique items
		assert.Len(t, unique, 2)
	})
}

// ============================================================================
// Channel Resolution Tests
// ============================================================================

func TestSenderSlack_ResolveChannelID(t *testing.T) {
	t.Run("returns cached channel ID when available", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLoggerInterface(ctrl)
		mockSlackClient := mocks.NewMockSlackClientInterface(ctrl)
		// No API calls expected since channelID is cached

		slackSender := &SenderSlack{
			apiKey:      "xoxb-test-token",
			channel:     "#test-channel",
			channelID:   "C123CACHED", // Cached ID
			logger:      mockLogger,
			slackClient: mockSlackClient,
		}

		channelID, err := slackSender.resolveChannelID()

		require.NoError(t, err)
		assert.Equal(t, "C123CACHED", channelID)
	})

	t.Run("uses channel directly when it's already an ID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLoggerInterface(ctrl)
		mockSlackClient := mocks.NewMockSlackClientInterface(ctrl)
		// No API calls expected since channel is already an ID

		slackSender := &SenderSlack{
			apiKey:      "xoxb-test-token",
			channel:     "C12345678", // Already an ID (no # prefix)
			logger:      mockLogger,
			slackClient: mockSlackClient,
		}

		channelID, err := slackSender.resolveChannelID()

		require.NoError(t, err)
		assert.Equal(t, "C12345678", channelID)
		assert.Equal(t, "C12345678", slackSender.channelID) // Should cache it
	})

	t.Run("resolves channel name to ID successfully", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLoggerInterface(ctrl)
		mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()

		mockSlackClient := mocks.NewMockSlackClientInterface(ctrl)

		// Mock API response with matching channel
		testChannel1 := slackapi.Channel{}
		testChannel1.ID = "C111111"
		testChannel1.Name = "general"

		testChannel2 := slackapi.Channel{}
		testChannel2.ID = "C222222"
		testChannel2.Name = "alerts"

		testChannel3 := slackapi.Channel{}
		testChannel3.ID = "C333333"
		testChannel3.Name = "monitoring"

		mockSlackClient.EXPECT().GetConversations(gomock.Any()).Return(
			[]slackapi.Channel{testChannel1, testChannel2, testChannel3},
			"",
			nil,
		)

		slackSender := &SenderSlack{
			apiKey:      "xoxb-test-token",
			channel:     "#alerts", // Channel name with # prefix
			logger:      mockLogger,
			slackClient: mockSlackClient,
		}

		channelID, err := slackSender.resolveChannelID()

		require.NoError(t, err)
		assert.Equal(t, "C222222", channelID)
		assert.Equal(t, "C222222", slackSender.channelID) // Should cache it
	})

	t.Run("returns error when channel name not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLoggerInterface(ctrl)
		mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()

		mockSlackClient := mocks.NewMockSlackClientInterface(ctrl)

		// Mock API response without matching channel
		testChannel1 := slackapi.Channel{}
		testChannel1.ID = "C111111"
		testChannel1.Name = "general"

		mockSlackClient.EXPECT().GetConversations(gomock.Any()).Return(
			[]slackapi.Channel{testChannel1},
			"",
			nil,
		)

		slackSender := &SenderSlack{
			apiKey:      "xoxb-test-token",
			channel:     "#nonexistent", // Channel that doesn't exist
			logger:      mockLogger,
			slackClient: mockSlackClient,
		}

		channelID, err := slackSender.resolveChannelID()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "channel not found: #nonexistent")
		assert.Empty(t, channelID)
		assert.Empty(t, slackSender.channelID) // Should not cache error
	})

	t.Run("returns error when API call fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLoggerInterface(ctrl)
		mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()

		mockSlackClient := mocks.NewMockSlackClientInterface(ctrl)

		// Mock API error
		apiError := fmt.Errorf("slack API error: rate limited")
		mockSlackClient.EXPECT().GetConversations(gomock.Any()).Return(
			nil,
			"",
			apiError,
		)

		slackSender := &SenderSlack{
			apiKey:      "xoxb-test-token",
			channel:     "#alerts",
			logger:      mockLogger,
			slackClient: mockSlackClient,
		}

		channelID, err := slackSender.resolveChannelID()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list conversations")
		assert.Empty(t, channelID)
		assert.Empty(t, slackSender.channelID)
	})

	t.Run("strips # prefix when resolving channel name", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLoggerInterface(ctrl)
		mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()

		mockSlackClient := mocks.NewMockSlackClientInterface(ctrl)

		testChannel := slackapi.Channel{}
		testChannel.ID = "C999999"
		testChannel.Name = "team-alerts" // Name without # in API

		mockSlackClient.EXPECT().GetConversations(gomock.Any()).Return(
			[]slackapi.Channel{testChannel},
			"",
			nil,
		)

		slackSender := &SenderSlack{
			apiKey:      "xoxb-test-token",
			channel:     "#team-alerts", // With # prefix
			logger:      mockLogger,
			slackClient: mockSlackClient,
		}

		channelID, err := slackSender.resolveChannelID()

		require.NoError(t, err)
		assert.Equal(t, "C999999", channelID)
	})

	t.Run("handles empty channel list from API", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLoggerInterface(ctrl)
		mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()

		mockSlackClient := mocks.NewMockSlackClientInterface(ctrl)

		// Empty channel list
		mockSlackClient.EXPECT().GetConversations(gomock.Any()).Return(
			[]slackapi.Channel{},
			"",
			nil,
		)

		slackSender := &SenderSlack{
			apiKey:      "xoxb-test-token",
			channel:     "#alerts",
			logger:      mockLogger,
			slackClient: mockSlackClient,
		}

		channelID, err := slackSender.resolveChannelID()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "channel not found: #alerts")
		assert.Empty(t, channelID)
	})

	t.Run("case-sensitive channel name matching", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLoggerInterface(ctrl)
		mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()

		mockSlackClient := mocks.NewMockSlackClientInterface(ctrl)

		testChannel := slackapi.Channel{}
		testChannel.ID = "C123456"
		testChannel.Name = "Alerts" // Capital A

		mockSlackClient.EXPECT().GetConversations(gomock.Any()).Return(
			[]slackapi.Channel{testChannel},
			"",
			nil,
		)

		slackSender := &SenderSlack{
			apiKey:      "xoxb-test-token",
			channel:     "#alerts", // Lowercase
			logger:      mockLogger,
			slackClient: mockSlackClient,
		}

		_, err := slackSender.resolveChannelID()

		// Should not match due to case sensitivity
		require.Error(t, err)
		assert.Contains(t, err.Error(), "channel not found: #alerts")
	})

	t.Run("finds correct channel among many results", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := mocks.NewMockLoggerInterface(ctrl)
		mockLogger.EXPECT().Debug(gomock.Any(), gomock.Any()).AnyTimes()
		mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()

		mockSlackClient := mocks.NewMockSlackClientInterface(ctrl)

		// Create many channels
		channels := make([]slackapi.Channel, 100)
		for i := 0; i < 100; i++ {
			channels[i] = slackapi.Channel{
				GroupConversation: slackapi.GroupConversation{
					Conversation: slackapi.Conversation{
						ID: fmt.Sprintf("C%d", i),
					},
					Name: fmt.Sprintf("channel-%d", i),
				},
			}
		}
		// Add target channel in the middle
		targetChannel := slackapi.Channel{}
		targetChannel.ID = "CTARGET"
		targetChannel.Name = "target-channel"
		channels[50] = targetChannel

		mockSlackClient.EXPECT().GetConversations(gomock.Any()).Return(
			channels,
			"",
			nil,
		)

		slackSender := &SenderSlack{
			apiKey:      "xoxb-test-token",
			channel:     "#target-channel",
			logger:      mockLogger,
			slackClient: mockSlackClient,
		}

		channelID, err := slackSender.resolveChannelID()

		require.NoError(t, err)
		assert.Equal(t, "CTARGET", channelID)
	})
}

func TestSenderSlack_GetSeverityEmoji(t *testing.T) {
	sender := &SenderSlack{}

	t.Run("high severity", func(t *testing.T) {
		result := sender.getSeverityEmoji(issuepkg.SeverityHigh)
		assert.Equal(t, "🔴", result)
	})

	t.Run("low severity", func(t *testing.T) {
		result := sender.getSeverityEmoji(issuepkg.SeverityLow)
		assert.Equal(t, "🟡", result)
	})

	t.Run("info severity", func(t *testing.T) {
		result := sender.getSeverityEmoji(issuepkg.SeverityInfo)
		assert.Equal(t, "🟢", result)
	})

	t.Run("debug severity", func(t *testing.T) {
		result := sender.getSeverityEmoji(issuepkg.SeverityDebug)
		assert.Equal(t, "🔵", result)
	})
}

func TestSenderSlack_GetSeverityName(t *testing.T) {
	sender := &SenderSlack{}

	t.Run("high severity", func(t *testing.T) {
		result := sender.getSeverityName(issuepkg.SeverityHigh)
		assert.Equal(t, "High", result)
	})

	t.Run("low severity", func(t *testing.T) {
		result := sender.getSeverityName(issuepkg.SeverityLow)
		assert.Equal(t, "Low", result)
	})

	t.Run("info severity", func(t *testing.T) {
		result := sender.getSeverityName(issuepkg.SeverityInfo)
		assert.Equal(t, "Info", result)
	})

	t.Run("debug severity", func(t *testing.T) {
		result := sender.getSeverityName(issuepkg.SeverityDebug)
		assert.Equal(t, "Debug", result)
	})
}

func TestSenderSlack_BuildMessageContext(t *testing.T) {
	sender := &SenderSlack{}

	t.Run("builds context for firing alert", func(t *testing.T) {
		issue := issuepkg.NewIssue("Test Alert", "test-key")
		issue.Description = "Test description"
		issue.ClusterName = "test-cluster"
		issue.Subject.Namespace = "default"
		issue.Subject.Name = "test-pod"
		issue.Source = issuepkg.SourcePrometheus
		issue.Severity = issuepkg.SeverityHigh
		issue.Status = issuepkg.StatusFiring

		context := sender.buildMessageContext(issue)

		require.NotNil(t, context)
		assert.Equal(t, "Test Alert", context.Title)
		assert.Equal(t, "Test description", context.Description)
		assert.Equal(t, "test-cluster", context.Cluster)
		assert.Equal(t, "default", context.Namespace)
		assert.Equal(t, "test-pod", context.PodName)
		assert.Equal(t, "PROMETHEUS", context.Source)
		assert.Equal(t, "firing", context.Status)
		assert.Equal(t, "🔥", context.StatusEmoji)
		assert.Equal(t, "High", context.Severity)
		assert.Equal(t, "🔴", context.SeverityEmoji)
	})

	t.Run("builds context for resolved alert", func(t *testing.T) {
		issue := issuepkg.NewIssue("Test Alert", "test-key")
		issue.Status = issuepkg.StatusResolved
		issue.Severity = issuepkg.SeverityLow

		context := sender.buildMessageContext(issue)

		require.NotNil(t, context)
		assert.Equal(t, "resolved", context.Status)
		assert.Equal(t, "✅", context.StatusEmoji)
		assert.Equal(t, "Low", context.Severity)
		assert.Equal(t, "🟡", context.SeverityEmoji)
	})
}

func TestSenderSlack_GetIssueLabel(t *testing.T) {
	sender := &SenderSlack{}

	t.Run("returns label from issue", func(t *testing.T) {
		issue := issuepkg.NewIssue("Test", "test-key")
		issue.Subject.Labels = map[string]string{
			"app":  "myapp",
			"tier": "backend",
		}

		result := sender.getIssueLabel(issue, "app")
		assert.Equal(t, "myapp", result)

		result = sender.getIssueLabel(issue, "tier")
		assert.Equal(t, "backend", result)
	})

	t.Run("returns empty for missing label", func(t *testing.T) {
		issue := issuepkg.NewIssue("Test", "test-key")
		issue.Subject.Labels = map[string]string{"app": "myapp"}

		result := sender.getIssueLabel(issue, "nonexistent")
		assert.Empty(t, result)
	})

	t.Run("returns empty when labels is nil", func(t *testing.T) {
		issue := issuepkg.NewIssue("Test", "test-key")
		issue.Subject.Labels = nil

		result := sender.getIssueLabel(issue, "app")
		assert.Empty(t, result)
	})
}

// TestSenderSlack_RemoveTimestampFromFilename tests timestamp removal from filenames
func TestSenderSlack_RemoveTimestampFromFilename(t *testing.T) {
	sender, _, _ := setupSenderSlackTest(t)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "filename with timestamp",
			input:    "pod-logs-namespace-pod-20251103-001242.log",
			expected: "pod-logs-namespace-pod.log",
		},
		{
			name:     "filename without timestamp",
			input:    "pod-logs-namespace-pod.log",
			expected: "pod-logs-namespace-pod.log",
		},
		{
			name:     "filename with timestamp and txt extension",
			input:    "error-report-20251103-123456.txt",
			expected: "error-report.txt",
		},
		{
			name:     "filename with timestamp and csv extension",
			input:    "metrics-20251103-235959.csv",
			expected: "metrics.csv",
		},
		{
			name:     "filename with partial timestamp should not match",
			input:    "file-2025.log",
			expected: "file-2025.log",
		},
		{
			name:     "filename with different pattern should not match",
			input:    "file-123456.log",
			expected: "file-123456.log",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sender.removeTimestampFromFilename(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestSenderSlack_DeduplicateEnrichments_WithFileBlocks tests deduplication with file blocks
func TestSenderSlack_DeduplicateEnrichments_WithFileBlocks(t *testing.T) {
	sender, _, _ := setupSenderSlackTest(t)

	t.Run("deduplicates file blocks with different timestamps", func(t *testing.T) {
		enrichments := []issuepkg.Enrichment{
			{
				Type:  issuepkg.EnrichmentTypeLogs,
				Title: "Pod Logs",
				Blocks: []issuepkg.BaseBlock{
					&issuepkg.FileBlock{
						Filename: "pod-logs-namespace-pod-20251103-001242.log",
					},
				},
			},
			{
				Type:  issuepkg.EnrichmentTypeLogs,
				Title: "Pod Logs",
				Blocks: []issuepkg.BaseBlock{
					&issuepkg.FileBlock{
						Filename: "pod-logs-namespace-pod-20251103-001300.log", // different timestamp
					},
				},
			},
		}

		result := sender.deduplicateEnrichments(enrichments)

		// Should deduplicate because base filename is the same
		assert.Len(t, result, 1)
		assert.Equal(t, "Pod Logs", result[0].Title)
	})

	t.Run("keeps file blocks with different base names", func(t *testing.T) {
		enrichments := []issuepkg.Enrichment{
			{
				Type:  issuepkg.EnrichmentTypeLogs,
				Title: "Pod Logs 1",
				Blocks: []issuepkg.BaseBlock{
					&issuepkg.FileBlock{
						Filename: "pod-logs-pod1-20251103-001242.log",
					},
				},
			},
			{
				Type:  issuepkg.EnrichmentTypeLogs,
				Title: "Pod Logs 2",
				Blocks: []issuepkg.BaseBlock{
					&issuepkg.FileBlock{
						Filename: "pod-logs-pod2-20251103-001242.log", // different pod
					},
				},
			},
		}

		result := sender.deduplicateEnrichments(enrichments)

		// Should keep both because base filenames are different
		assert.Len(t, result, 2)
	})

	t.Run("deduplicates table blocks by table name", func(t *testing.T) {
		enrichments := []issuepkg.Enrichment{
			{
				Type:  issuepkg.EnrichmentTypeAlertLabels,
				Title: "Labels",
				Blocks: []issuepkg.BaseBlock{
					&issuepkg.TableBlock{
						TableName: "alert-labels",
						Headers:   []string{"Label", "Value"},
						Rows:      [][]string{{"severity", "high"}},
					},
				},
			},
			{
				Type:  issuepkg.EnrichmentTypeAlertLabels,
				Title: "Labels",
				Blocks: []issuepkg.BaseBlock{
					&issuepkg.TableBlock{
						TableName: "alert-labels",
						Headers:   []string{"Label", "Value"},
						Rows:      [][]string{{"severity", "high"}},
					},
				},
			},
		}

		result := sender.deduplicateEnrichments(enrichments)

		// Should deduplicate tables with same name
		assert.Len(t, result, 1)
	})

	t.Run("deduplicates markdown blocks by content prefix", func(t *testing.T) {
		enrichments := []issuepkg.Enrichment{
			{
				Type:  issuepkg.EnrichmentTypeTextFile,
				Title: "Description",
				Blocks: []issuepkg.BaseBlock{
					&issuepkg.MarkdownBlock{
						Text: "This is a long description that should be deduplicated based on first 50 characters",
					},
				},
			},
			{
				Type:  issuepkg.EnrichmentTypeTextFile,
				Title: "Description",
				Blocks: []issuepkg.BaseBlock{
					&issuepkg.MarkdownBlock{
						Text: "This is a long description that should be deduplicated based on first 50 characters - different end",
					},
				},
			},
		}

		result := sender.deduplicateEnrichments(enrichments)

		// Should deduplicate based on first 50 chars
		assert.Len(t, result, 1)
	})
}

// TestSenderSlack_BuildSlackBlocks_WithFileEnrichments tests file enrichment handling
func TestSenderSlack_BuildSlackBlocks_WithFileEnrichments(t *testing.T) {
	sender, _, _ := setupSenderSlackTest(t)

	t.Run("renders file enrichments separately", func(t *testing.T) {
		issue := issuepkg.NewIssue("Test Issue", "test-key")
		issue.Severity = issuepkg.SeverityHigh
		issue.Status = issuepkg.StatusFiring
		issue.Source = issuepkg.SourcePrometheus

		// Add file enrichment with permalink
		fileEnrichment := issuepkg.NewEnrichmentWithType(issuepkg.EnrichmentTypeLogs, "Pod Logs")
		fileEnrichment.FileInfo = &issuepkg.FileInfo{
			Permalink: "https://files.slackapi.com/files-pri/T123/F456/pod-logs.log",
			Filename:  "pod-logs-namespace-pod-20251103-001242.log",
			Size:      1024,
		}
		issue.AddEnrichment(*fileEnrichment)

		// Add regular enrichment
		tableEnrichment := issuepkg.NewEnrichmentWithType(issuepkg.EnrichmentTypeAlertLabels, "Alert Labels")
		tableEnrichment.AddBlock(&issuepkg.TableBlock{
			TableName: "labels",
			Headers:   []string{"Label", "Value"},
			Rows:      [][]string{{"severity", "high"}},
		})
		issue.AddEnrichment(*tableEnrichment)

		blocks := sender.buildSlackBlocks(issue)

		// Should have blocks: header, context, file enrichment, table enrichment, divider
		assert.GreaterOrEqual(t, len(blocks), 4)
	})

	t.Run("adds divider at end when enrichments exist", func(t *testing.T) {
		issue := issuepkg.NewIssue("Test Issue", "test-key")
		issue.Severity = issuepkg.SeverityHigh
		issue.Status = issuepkg.StatusFiring

		// Add file enrichment
		fileEnrichment := issuepkg.NewEnrichmentWithType(issuepkg.EnrichmentTypeLogs, "Pod Logs")
		fileEnrichment.FileInfo = &issuepkg.FileInfo{
			Permalink: "https://files.slackapi.com/files-pri/T123/F456/file.log",
			Filename:  "file.log",
			Size:      1024,
		}
		issue.AddEnrichment(*fileEnrichment)

		blocks := sender.buildSlackBlocks(issue)

		// Should have divider at the end (because enrichments exist)
		lastBlock := blocks[len(blocks)-1]
		_, isDivider := lastBlock.(*slackapi.DividerBlock)
		assert.True(t, isDivider, "Last block should be divider when enrichments exist")
	})

	t.Run("no divider when no enrichments", func(t *testing.T) {
		issue := issuepkg.NewIssue("Test Issue", "test-key")
		issue.Severity = issuepkg.SeverityHigh
		issue.Status = issuepkg.StatusFiring

		blocks := sender.buildSlackBlocks(issue)

		// Should NOT have divider at the end when no enrichments
		lastBlock := blocks[len(blocks)-1]
		_, isDivider := lastBlock.(*slackapi.DividerBlock)
		assert.False(t, isDivider, "Should not have divider when no enrichments")
	})
}

// TestSenderSlack_BuildHeaderBlockFallback tests fallback header generation
func TestSenderSlack_BuildHeaderBlockFallback(t *testing.T) {
	sender, _, _ := setupSenderSlackTest(t)

	t.Run("creates fallback header for firing issue", func(t *testing.T) {
		issue := issuepkg.NewIssue("Test Alert", "test-key")
		issue.Severity = issuepkg.SeverityHigh
		issue.Status = issuepkg.StatusFiring

		blocks := sender.buildHeaderBlockFallback(issue)

		assert.GreaterOrEqual(t, len(blocks), 1)

		// First block should be section with title
		sectionBlock, ok := blocks[0].(*slackapi.SectionBlock)
		assert.True(t, ok, "First block should be section block")
		assert.Contains(t, sectionBlock.Text.Text, "Test Alert")
	})

	t.Run("creates fallback header for resolved issue", func(t *testing.T) {
		issue := issuepkg.NewIssue("Test Alert", "test-key")
		issue.Severity = issuepkg.SeverityHigh
		issue.Status = issuepkg.StatusResolved

		blocks := sender.buildHeaderBlockFallback(issue)

		assert.GreaterOrEqual(t, len(blocks), 1)

		sectionBlock, ok := blocks[0].(*slackapi.SectionBlock)
		assert.True(t, ok)
		// Should contain resolved indicator (text contains Resolved or ✅)
		text := sectionBlock.Text.Text
		hasResolved := false
		if len(text) > 0 {
			hasResolved = containsSubstring(text, "Resolved") || containsSubstring(text, "✅")
		}
		assert.True(t, hasResolved, "Should indicate resolved status")
	})
}

// TestSenderSlack_PreprocessEnrichments tests enrichment preprocessing
func TestSenderSlack_PreprocessEnrichments(t *testing.T) {
	sender, mockClient, _ := setupSenderSlackTest(t)

	t.Run("uploads file blocks and sets FileInfo", func(t *testing.T) {
		issue := issuepkg.NewIssue("Test", "test-key")

		// Create enrichment with FileBlock
		fileBlock := &issuepkg.FileBlock{
			Filename: "test-file.log",
			Contents: []byte("test content"),
			Size:     12,
			MimeType: "text/plain",
		}

		enrichment := issuepkg.NewEnrichmentWithType(issuepkg.EnrichmentTypeLogs, "Test File")
		enrichment.AddBlock(fileBlock)
		issue.AddEnrichment(*enrichment)

		// Mock successful file upload - returns FileSummary with ID
		mockClient.EXPECT().
			UploadFileV2(gomock.Any()).
			Return(&slackapi.FileSummary{
				ID: "F123456",
			}, nil)

		// Mock GetFileInfo to return file details with permalink
		mockClient.EXPECT().
			GetFileInfo("F123456", 0, 0).
			Return(&slackapi.File{
				Permalink: "https://files.slackapi.com/files-pri/T123/F456/test-file.log",
			}, nil, nil, nil)

		// Call preprocessing
		sender.preprocessEnrichments(issue)

		// Should have set FileInfo on enrichment
		assert.NotNil(t, issue.Enrichments[0].FileInfo)
		assert.Equal(t, "https://files.slackapi.com/files-pri/T123/F456/test-file.log", issue.Enrichments[0].FileInfo.Permalink)
		assert.Equal(t, "test-file.log", issue.Enrichments[0].FileInfo.Filename)
		assert.Equal(t, int64(12), issue.Enrichments[0].FileInfo.Size)
	})

	t.Run("skips enrichments without FileBlocks", func(t *testing.T) {
		issue := issuepkg.NewIssue("Test", "test-key")

		// Create enrichment with TableBlock (not FileBlock)
		tableBlock := &issuepkg.TableBlock{
			TableName: "test-table",
			Headers:   []string{"Key", "Value"},
			Rows:      [][]string{{"key1", "value1"}},
		}

		enrichment := issuepkg.NewEnrichmentWithType(issuepkg.EnrichmentTypeAlertLabels, "Test Table")
		enrichment.AddBlock(tableBlock)
		issue.AddEnrichment(*enrichment)

		// Call preprocessing (should not interact with Slack client)
		sender.preprocessEnrichments(issue)

		// Should not set FileInfo
		assert.Nil(t, issue.Enrichments[0].FileInfo)
	})

	t.Run("skips enrichments with existing FileInfo", func(t *testing.T) {
		issue := issuepkg.NewIssue("Test", "test-key")

		// Create enrichment with FileInfo already set
		enrichment := issuepkg.NewEnrichmentWithType(issuepkg.EnrichmentTypeLogs, "Existing File")
		enrichment.FileInfo = &issuepkg.FileInfo{
			Permalink: "https://existing.com/file.log",
			Filename:  "file.log",
		}
		issue.AddEnrichment(*enrichment)

		// Call preprocessing (should not upload anything)
		sender.preprocessEnrichments(issue)

		// FileInfo should remain unchanged
		assert.Equal(t, "https://existing.com/file.log", issue.Enrichments[0].FileInfo.Permalink)
	})
}

// Helper function for substring check
func containsSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
