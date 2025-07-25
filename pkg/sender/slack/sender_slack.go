package slack

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	slackapi "github.com/slack-go/slack"
	"go.uber.org/zap"

	issuepkg "github.com/kubecano/cano-collector/pkg/core/issue"
	logger_interfaces "github.com/kubecano/cano-collector/pkg/logger/interfaces"
	sender_interfaces "github.com/kubecano/cano-collector/pkg/sender/interfaces"
	"github.com/kubecano/cano-collector/pkg/util"
)

type SenderSlack struct {
	apiKey      string
	channel     string
	logger      logger_interfaces.LoggerInterface
	unfurlLinks bool
	slackClient sender_interfaces.SlackClientInterface
	// Threading configuration - will be added in next step
	threadManager sender_interfaces.SlackThreadManagerInterface
	// Table formatting parameters (instead of full config dependency)
	tableFormat  string
	maxTableRows int
}

func NewSenderSlack(apiKey, channel string, unfurlLinks bool, logger logger_interfaces.LoggerInterface, client util.HTTPClient) *SenderSlack {
	var slackClient sender_interfaces.SlackClientInterface

	if client != nil {
		// Use custom HTTP client with slack-go
		slackClient = slackapi.New(apiKey, slackapi.OptionHTTPClient(client))
	} else {
		// Use default HTTP client from slack-go
		slackClient = slackapi.New(apiKey)
	}

	return &SenderSlack{
		apiKey:      apiKey,
		channel:     channel,
		logger:      logger,
		unfurlLinks: unfurlLinks,
		slackClient: slackClient,
	}
}

func (s *SenderSlack) Send(ctx context.Context, issue *issuepkg.Issue) error {
	s.logger.Info("Sending Slack notification",
		zap.String("channel", s.channel),
		zap.String("status", issue.Status.String()),
	)

	// Create formatted blocks
	blocks := s.buildSlackBlocks(issue)
	attachments := s.buildSlackAttachments(issue)

	// Fallback text for notifications
	fallbackText := s.formatIssueToString(issue)

	params := slackapi.PostMessageParameters{
		UnfurlLinks: s.unfurlLinks,
		UnfurlMedia: s.unfurlLinks,
	}

	var msgOptions []slackapi.MsgOption
	msgOptions = append(msgOptions, slackapi.MsgOptionText(fallbackText, false))
	msgOptions = append(msgOptions, slackapi.MsgOptionBlocks(blocks...))
	msgOptions = append(msgOptions, slackapi.MsgOptionAttachments(attachments...))
	msgOptions = append(msgOptions, slackapi.MsgOptionPostMessageParameters(params))

	// Threading logic: check if we should post as thread reply
	var threadTS string
	if s.threadManager != nil {
		fingerprint := s.generateFingerprint(issue)

		if issue.Status == issuepkg.StatusResolved {
			// For resolved alerts, try to find existing thread
			ts, err := s.threadManager.GetThreadTS(ctx, fingerprint)
			if err != nil {
				s.logger.Warn("Failed to get thread timestamp",
					zap.Error(err),
					zap.String("fingerprint", fingerprint))
			} else if ts != "" {
				threadTS = ts
				s.logger.Debug("Posting resolved alert as thread reply",
					zap.String("threadTS", threadTS),
					zap.String("fingerprint", fingerprint))
			}
		}

		// Optionally add fingerprint to message metadata for searching
		// This adds the fingerprint as invisible metadata that can be searched later
		// We add it as a code block in the message for simple searching
		fingerprintBlock := slackapi.NewSectionBlock(
			slackapi.NewTextBlockObject("mrkdwn", fmt.Sprintf("```%s```", fingerprint), false, false),
			nil, nil,
		)
		// Add fingerprint block to the end (it will be small and relatively unobtrusive)
		blocks = append(blocks, fingerprintBlock)

		// Update msgOptions with new blocks that include fingerprint
		// Find and replace the blocks option instead of recreating the entire slice
		for i := range msgOptions {
			// Check if this is the blocks option by trying to replace it
			if i == 1 { // blocks option is typically the second one
				msgOptions[i] = slackapi.MsgOptionBlocks(blocks...)
				break
			}
		}

		// Add thread timestamp if this is a thread reply
		if threadTS != "" {
			msgOptions = append(msgOptions, slackapi.MsgOptionTS(threadTS))
		}
	}

	// Send the message
	_, timestamp, err := s.slackClient.PostMessage(s.channel, msgOptions...)
	if err != nil {
		s.logger.Error("Failed to send Slack message",
			zap.Error(err),
			zap.String("channel", s.channel),
		)
		return err
	}

	// For firing alerts, cache the thread timestamp for future resolved alerts
	if s.threadManager != nil && issue.Status == issuepkg.StatusFiring {
		fingerprint := s.generateFingerprint(issue)
		s.threadManager.SetThreadTS(fingerprint, timestamp)
		s.logger.Debug("Cached thread timestamp for firing alert",
			zap.String("fingerprint", fingerprint),
			zap.String("timestamp", timestamp))
	}

	s.logger.Info("Slack message sent successfully",
		zap.String("channel", s.channel),
		zap.String("timestamp", timestamp),
	)
	return nil
}

// buildSlackBlocks creates the main message blocks
func (s *SenderSlack) buildSlackBlocks(issue *issuepkg.Issue) []slackapi.Block {
	var blocks []slackapi.Block

	// Header block with status, severity and title
	headerText := s.formatHeader(issue)
	headerBlock := slackapi.NewSectionBlock(
		slackapi.NewTextBlockObject("mrkdwn", headerText, false, false),
		nil, nil,
	)
	blocks = append(blocks, headerBlock)

	// Separate runbook links (display as text) from other links (display as buttons)
	var runbookLinks []issuepkg.Link
	var otherLinks []issuepkg.Link
	for _, link := range issue.Links {
		if link.Type == issuepkg.LinkTypeRunbook {
			runbookLinks = append(runbookLinks, link)
		} else {
			otherLinks = append(otherLinks, link)
		}
	}

	// Action buttons for non-runbook links right after header
	if len(otherLinks) > 0 {
		linkButtons := s.buildLinkButtons(otherLinks)
		if len(linkButtons) > 0 {
			actionBlock := slackapi.NewActionBlock("links", linkButtons...)
			blocks = append(blocks, actionBlock)
		}
	}

	// Alert description
	if issue.Description != "" {
		alertText := "🚨 *Alert:* " + issue.Description
		alertBlock := slackapi.NewSectionBlock(
			slackapi.NewTextBlockObject("mrkdwn", alertText, false, false),
			nil, nil,
		)
		blocks = append(blocks, alertBlock)
	}

	// Runbook URLs displayed as plain text for Slack auto-preview
	for _, runbookLink := range runbookLinks {
		runbookText := "📖 *Runbook URL:* " + runbookLink.URL
		runbookBlock := slackapi.NewSectionBlock(
			slackapi.NewTextBlockObject("mrkdwn", runbookText, false, false),
			nil, nil,
		)
		blocks = append(blocks, runbookBlock)
	}

	// Add enrichments directly
	enrichmentBlocks := s.buildEnrichmentBlocks(issue.Enrichments)
	blocks = append(blocks, enrichmentBlocks...)

	// Add final divider if we have enrichments
	if len(issue.Enrichments) > 0 {
		blocks = append(blocks, slackapi.NewDividerBlock())
	}

	return blocks
}

// buildSlackAttachments creates colored attachment with secondary issue details
func (s *SenderSlack) buildSlackAttachments(issue *issuepkg.Issue) []slackapi.Attachment {
	var attachments []slackapi.Attachment

	// Only create attachment if we have secondary information to show
	var attachmentBlocks []slackapi.Block

	// Source information (less critical, keep in attachment)
	if issue.Source != issuepkg.SourceUnknown {
		sourceText := fmt.Sprintf("📍 *Source:* `%s`", issue.Source.String())
		sourceBlock := slackapi.NewSectionBlock(
			slackapi.NewTextBlockObject("mrkdwn", sourceText, false, false),
			nil, nil,
		)
		attachmentBlocks = append(attachmentBlocks, sourceBlock)
	}

	// Cluster information
	if issue.ClusterName != "" {
		clusterText := fmt.Sprintf("🌐 *Cluster:* `%s`", issue.ClusterName)
		clusterBlock := slackapi.NewSectionBlock(
			slackapi.NewTextBlockObject("mrkdwn", clusterText, false, false),
			nil, nil,
		)
		attachmentBlocks = append(attachmentBlocks, clusterBlock)
	}

	// Add any additional metadata that might be useful but not critical
	if issue.Subject != nil && issue.Subject.Namespace != "" {
		namespaceText := fmt.Sprintf("🏷️ *Namespace:* `%s`", issue.Subject.Namespace)
		namespaceBlock := slackapi.NewSectionBlock(
			slackapi.NewTextBlockObject("mrkdwn", namespaceText, false, false),
			nil, nil,
		)
		attachmentBlocks = append(attachmentBlocks, namespaceBlock)
	}

	// Add timing information if available
	if !issue.StartsAt.IsZero() {
		timeText := "⏰ *Started:* " + issue.StartsAt.UTC().Format("2006-01-02 15:04:05 UTC")
		timeBlock := slackapi.NewSectionBlock(
			slackapi.NewTextBlockObject("mrkdwn", timeText, false, false),
			nil, nil,
		)
		attachmentBlocks = append(attachmentBlocks, timeBlock)
	}

	// Only create attachment if we have content
	if len(attachmentBlocks) > 0 {
		color := s.getSeverityColor(issue.Severity)
		attachment := slackapi.Attachment{
			Color:  color,
			Blocks: slackapi.Blocks{BlockSet: attachmentBlocks},
		}
		attachments = append(attachments, attachment)
	}

	return attachments
}

// buildEnrichmentBlocks creates blocks for enrichments in the main message
func (s *SenderSlack) buildEnrichmentBlocks(enrichments []issuepkg.Enrichment) []slackapi.Block {
	var blocks []slackapi.Block

	for _, enrichment := range enrichments {
		enrichmentBlocks := s.convertEnrichmentToBlocks(enrichment)
		blocks = append(blocks, enrichmentBlocks...)
	}

	return blocks
}

// convertEnrichmentToBlocks converts a single enrichment to Slack blocks
func (s *SenderSlack) convertEnrichmentToBlocks(enrichment issuepkg.Enrichment) []slackapi.Block {
	var blocks []slackapi.Block

	if len(enrichment.Blocks) == 0 {
		return blocks
	}

	// Add enrichment title as smaller formatted text if available
	if enrichment.Title != nil && *enrichment.Title != "" {
		// Small bold title instead of large header
		titleText := "*" + *enrichment.Title + "*"

		titleBlock := slackapi.NewSectionBlock(
			slackapi.NewTextBlockObject("mrkdwn", titleText, false, false),
			nil, nil,
		)
		blocks = append(blocks, titleBlock)
	}

	// Process each block in the enrichment
	for _, block := range enrichment.Blocks {
		slackBlock := s.convertBlockToSlack(block)
		blocks = append(blocks, slackBlock)
	}

	// Add divider for visual separation between enrichments
	if len(blocks) > 0 {
		dividerBlock := slackapi.NewDividerBlock()
		blocks = append(blocks, dividerBlock)
	}

	return blocks
}

// convertBlockToSlack converts an issue block to a Slack block
func (s *SenderSlack) convertBlockToSlack(block issuepkg.BaseBlock) slackapi.Block {
	switch b := block.(type) {
	case *issuepkg.TableBlock:
		return s.convertTableBlockToSlack(b)
	case *issuepkg.JsonBlock:
		return s.convertJsonBlockToSlack(b)
	case *issuepkg.MarkdownBlock:
		return s.convertMarkdownBlockToSlack(b)
	case *issuepkg.HeaderBlock:
		return s.convertHeaderBlockToSlack(b)
	case *issuepkg.ListBlock:
		return s.convertListBlockToSlack(b)
	case *issuepkg.LinksBlock:
		return s.convertLinksBlockToSlack(b)
	case *issuepkg.FileBlock:
		return s.convertFileBlockToSlack(b)
	case *issuepkg.DividerBlock:
		return slackapi.NewDividerBlock()
	default:
		// Fallback - convert unknown block to text
		return slackapi.NewSectionBlock(
			slackapi.NewTextBlockObject("mrkdwn", "Unknown block type: "+block.BlockType(), false, false),
			nil, nil,
		)
	}
}

// convertTableBlockToSlack converts a table block to Slack section block with adaptive formatting
func (s *SenderSlack) convertTableBlockToSlack(table *issuepkg.TableBlock) slackapi.Block {
	// Check if table exceeds row limit and should be converted to file
	if s.maxTableRows > 0 && len(table.Rows) > s.maxTableRows {
		return s.convertLargeTableToFileBlock(table)
	}

	// Check table formatting preference
	switch s.tableFormat {
	case "attachment":
		// For attachment format, still render as block but note it could be an attachment
		return s.convertTableToAttachmentStyleBlock(table)
	case "enhanced":
		return s.convertTableToEnhancedBlock(table)
	case "simple":
		return s.convertTableToSimpleBlock(table)
	default:
		// Default behavior - simple format
		return s.convertTableToSimpleBlock(table)
	}
}

// convertTableToSimpleBlock converts table to simple key-value format
func (s *SenderSlack) convertTableToSimpleBlock(table *issuepkg.TableBlock) slackapi.Block {
	var text string
	if table.TableName != "" {
		text = fmt.Sprintf("*%s*\n", table.TableName)
	}

	// Add rows as key-value pairs
	for _, row := range table.Rows {
		if len(row) >= 2 {
			text += fmt.Sprintf("• %s `%s`\n", row[0], row[1])
		}
	}

	return slackapi.NewSectionBlock(
		slackapi.NewTextBlockObject("mrkdwn", text, false, false),
		nil, nil,
	)
}

// convertTableToEnhancedBlock converts table to enhanced format
func (s *SenderSlack) convertTableToEnhancedBlock(table *issuepkg.TableBlock) slackapi.Block {
	var text string
	if table.TableName != "" {
		text = fmt.Sprintf("*%s*\n", table.TableName)
	}

	// For two-column tables, use enhanced key-value format
	if len(table.Headers) == 2 {
		for _, row := range table.Rows {
			if len(row) >= 2 {
				text += fmt.Sprintf("▸ *%s*: `%s`\n", row[0], row[1])
			}
		}
	} else if len(table.Headers) > 0 {
		// For multi-column tables, create a more structured format
		text += "\n```\n"
		// Add headers
		headerLine := ""
		for i, header := range table.Headers {
			if i > 0 {
				headerLine += " | "
			}
			headerLine += fmt.Sprintf("%-15s", header)
		}
		text += headerLine + "\n"
		text += strings.Repeat("-", len(headerLine)) + "\n"

		// Add rows
		for _, row := range table.Rows {
			rowLine := ""
			for i := 0; i < len(table.Headers); i++ {
				if i > 0 {
					rowLine += " | "
				}
				cellValue := ""
				if i < len(row) {
					cellValue = row[i]
				}
				rowLine += fmt.Sprintf("%-15s", cellValue)
			}
			text += rowLine + "\n"
		}
		text += "```"
	} else {
		// For headerless tables, use enhanced key-value format similar to two-column tables
		for _, row := range table.Rows {
			if len(row) >= 2 {
				text += fmt.Sprintf("▸ *%s*: `%s`\n", row[0], row[1])
			} else if len(row) == 1 {
				text += fmt.Sprintf("▸ %s\n", row[0])
			}
		}
	}

	return slackapi.NewSectionBlock(
		slackapi.NewTextBlockObject("mrkdwn", text, false, false),
		nil, nil,
	)
}

// convertTableToAttachmentStyleBlock formats table for attachment-style display
func (s *SenderSlack) convertTableToAttachmentStyleBlock(table *issuepkg.TableBlock) slackapi.Block {
	var text string
	if table.TableName != "" {
		text = fmt.Sprintf("📊 *%s*\n", table.TableName)
	}

	// More compact format suitable for attachments
	for _, row := range table.Rows {
		if len(row) >= 2 {
			text += fmt.Sprintf("└ %s: `%s`\n", row[0], row[1])
		}
	}

	return slackapi.NewSectionBlock(
		slackapi.NewTextBlockObject("mrkdwn", text, false, false),
		nil, nil,
	)
}

// convertLargeTableToFileBlock converts large tables to file placeholder
// TODO: Implement actual file upload with Slack files_upload_v2 API
func (s *SenderSlack) convertLargeTableToFileBlock(table *issuepkg.TableBlock) slackapi.Block {
	rowCount := len(table.Rows)
	tableName := table.TableName
	if tableName == "" {
		tableName = "Large Table"
	}

	text := fmt.Sprintf("📊 *%s* (%d rows)\n", tableName, rowCount)
	text += fmt.Sprintf("Table too large for inline display (limit: %d rows)\n", s.maxTableRows)
	text += "_Would be converted to file attachment in full implementation_"

	// TODO: Convert table to CSV/text format and upload using Slack files_upload_v2 API
	// For now, show placeholder

	return slackapi.NewSectionBlock(
		slackapi.NewTextBlockObject("mrkdwn", text, false, false),
		nil, nil,
	)
}

// convertJsonBlockToSlack converts a JSON block to Slack section block
func (s *SenderSlack) convertJsonBlockToSlack(jsonBlock *issuepkg.JsonBlock) slackapi.Block {
	// Convert JSON to formatted string
	jsonStr := jsonBlock.ToJson()

	// Wrap in code block
	text := fmt.Sprintf("```\n%s\n```", jsonStr)

	return slackapi.NewSectionBlock(
		slackapi.NewTextBlockObject("mrkdwn", text, false, false),
		nil, nil,
	)
}

// convertMarkdownBlockToSlack converts a markdown block to Slack section block
func (s *SenderSlack) convertMarkdownBlockToSlack(markdown *issuepkg.MarkdownBlock) slackapi.Block {
	return slackapi.NewSectionBlock(
		slackapi.NewTextBlockObject("mrkdwn", markdown.Text, false, false),
		nil, nil,
	)
}

// convertHeaderBlockToSlack converts a header block to Slack header block
func (s *SenderSlack) convertHeaderBlockToSlack(header *issuepkg.HeaderBlock) slackapi.Block {
	return slackapi.NewHeaderBlock(
		slackapi.NewTextBlockObject("plain_text", header.Text, false, false),
	)
}

// convertListBlockToSlack converts a list block to Slack section block
func (s *SenderSlack) convertListBlockToSlack(list *issuepkg.ListBlock) slackapi.Block {
	var text string

	// Add list name if available
	if list.ListName != "" {
		text = fmt.Sprintf("*%s*\n", list.ListName)
	}

	// Add list items
	for i, item := range list.Items {
		if list.Ordered {
			text += fmt.Sprintf("%d. %s\n", i+1, item)
		} else {
			text += fmt.Sprintf("• %s\n", item)
		}
	}

	return slackapi.NewSectionBlock(
		slackapi.NewTextBlockObject("mrkdwn", text, false, false),
		nil, nil,
	)
}

// convertLinksBlockToSlack converts a links block to Slack actions block
func (s *SenderSlack) convertLinksBlockToSlack(links *issuepkg.LinksBlock) slackapi.Block {
	var buttons []slackapi.BlockElement

	// Convert links to buttons (limit to 5 for Slack constraints)
	for i, link := range links.Links {
		if i >= 5 {
			break
		}

		button := slackapi.NewButtonBlockElement(
			fmt.Sprintf("link_%d", i),
			link.URL,
			slackapi.NewTextBlockObject("plain_text", link.Text, false, false),
		)
		button.URL = link.URL
		buttons = append(buttons, button)
	}

	// If no buttons, create a section block with link text
	if len(buttons) == 0 {
		text := ""
		if links.BlockName != "" {
			text = fmt.Sprintf("*%s*\n", links.BlockName)
		}
		for _, link := range links.Links {
			text += fmt.Sprintf("• <%s|%s>\n", link.URL, link.Text)
		}

		return slackapi.NewSectionBlock(
			slackapi.NewTextBlockObject("mrkdwn", text, false, false),
			nil, nil,
		)
	}

	// Create action block with buttons
	blockID := "links"
	if links.BlockName != "" {
		blockID = "links_" + strings.ReplaceAll(strings.ToLower(links.BlockName), " ", "_")
	}

	return slackapi.NewActionBlock(blockID, buttons...)
}

// convertFileBlockToSlack converts a file block to Slack section block
// For now, we'll display file info as text. File upload implementation would require files_upload_v2 API
func (s *SenderSlack) convertFileBlockToSlack(file *issuepkg.FileBlock) slackapi.Block {
	sizeKB := file.GetSizeKB()
	sizeText := fmt.Sprintf("%.1f KB", sizeKB)
	if sizeKB > 1024 {
		sizeMB := sizeKB / 1024
		sizeText = fmt.Sprintf("%.1f MB", sizeMB)
	}

	text := fmt.Sprintf("📎 *File: %s*\n", file.Filename)
	text += "Size: " + sizeText
	if file.MimeType != "" {
		text += "\nType: " + file.MimeType
	}

	// TODO: Implement actual file upload using Slack files_upload_v2 API
	// For now, display as attachment info
	text += "\n_File content not uploaded - upload functionality to be implemented_"

	return slackapi.NewSectionBlock(
		slackapi.NewTextBlockObject("mrkdwn", text, false, false),
		nil, nil,
	)
}

// formatHeader creates the header text with status and severity
func (s *SenderSlack) formatHeader(issue *issuepkg.Issue) string {
	var statusText string
	var statusEmoji string
	if issue.IsResolved() {
		statusText = "Alert resolved"
		statusEmoji = "✅"
	} else {
		statusText = "Alert firing"
		statusEmoji = "🔥"
	}

	severityText := s.getSeverityText(issue.Severity)

	// Simple, clean format
	return fmt.Sprintf("%s %s %s\n%s",
		statusEmoji, statusText, severityText, issue.Title)
}

// formatLabels formats subject labels
func (s *SenderSlack) formatLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}

	text := "*Alert labels*\n"
	for key, value := range labels {
		text += fmt.Sprintf("• %s `%s`\n", key, value)
	}
	return text
}

// buildLinkButtons creates action buttons for links with emojis based on link type
func (s *SenderSlack) buildLinkButtons(links []issuepkg.Link) []slackapi.BlockElement {
	var buttons []slackapi.BlockElement

	for i, link := range links {
		// Limit to first 5 links to avoid Slack limits
		if i >= 5 {
			break
		}

		// Add emoji based on link type
		emoji := s.getLinkEmoji(link.Type)
		buttonText := fmt.Sprintf("%s %s", emoji, link.Text)

		button := slackapi.NewButtonBlockElement(
			fmt.Sprintf("link_%d", i),
			link.URL,
			slackapi.NewTextBlockObject("plain_text", buttonText, false, false),
		)
		button.URL = link.URL

		// Add styling based on link type
		button.Style = s.getLinkButtonStyle(link.Type)

		buttons = append(buttons, button)
	}

	return buttons
}

// getLinkEmoji returns appropriate emoji for link type
func (s *SenderSlack) getLinkEmoji(linkType issuepkg.LinkType) string {
	switch linkType {
	case issuepkg.LinkTypeInvestigate:
		return "🔍" // Investigate
	case issuepkg.LinkTypeSilence:
		return "🔕" // Silence
	case issuepkg.LinkTypePrometheusGenerator:
		return "📊" // Prometheus/Graphs
	case issuepkg.LinkTypeGeneral:
		return "🔗" // General link
	default:
		return "🔗" // Default
	}
}

// getLinkButtonStyle returns appropriate Slack button style for link type
func (s *SenderSlack) getLinkButtonStyle(linkType issuepkg.LinkType) slackapi.Style {
	switch linkType {
	case issuepkg.LinkTypeInvestigate:
		return slackapi.StylePrimary // Blue button for investigate
	case issuepkg.LinkTypeSilence:
		return slackapi.StyleDanger // Red button for silence
	default:
		return slackapi.StyleDefault // No special styling
	}
}

// getSeverityText returns formatted severity text with emoji
func (s *SenderSlack) getSeverityText(severity issuepkg.Severity) string {
	switch severity {
	case issuepkg.SeverityHigh:
		return "🔴 High"
	case issuepkg.SeverityLow:
		return "🟡 Low"
	case issuepkg.SeverityInfo:
		return "🟢 Info"
	case issuepkg.SeverityDebug:
		return "🔵 Debug"
	default:
		return "🟢 " + severity.String()
	}
}

// getSeverityColor returns Slack color for severity
func (s *SenderSlack) getSeverityColor(severity issuepkg.Severity) string {
	switch severity {
	case issuepkg.SeverityHigh:
		return "#EF311F" // Red
	case issuepkg.SeverityLow:
		return "#FFCC00" // Yellow
	case issuepkg.SeverityInfo:
		return "#00B302" // Green
	case issuepkg.SeverityDebug:
		return "#36a64f" // Gray/Green
	default:
		return "#00B302" // Default green
	}
}

// formatIssueToString converts an Issue to a formatted string message (fallback)
func (s *SenderSlack) formatIssueToString(issue *issuepkg.Issue) string {
	statusPrefix := ""
	if issue.IsResolved() {
		statusPrefix = "[RESOLVED] "
	}

	message := fmt.Sprintf("%s*%s*\n", statusPrefix, issue.Title)

	if issue.Description != "" {
		message += fmt.Sprintf("📝 %s\n", issue.Description)
	}

	message += fmt.Sprintf("🔥 Severity: %s\n", issue.Severity.String())
	message += fmt.Sprintf("📍 Source: %s\n", issue.Source.String())

	if issue.Subject != nil && issue.Subject.Name != "" {
		message += issue.Subject.FormatWithEmoji() + "\n"
	}

	if len(issue.Links) > 0 {
		message += "🔗 Links:\n"
		for _, link := range issue.Links {
			message += fmt.Sprintf("• <%s|%s>\n", link.URL, link.Text)
		}
	}

	return message
}

func (s *SenderSlack) SetLogger(logger logger_interfaces.LoggerInterface) {
	s.logger = logger
}

func (s *SenderSlack) SetUnfurlLinks(unfurl bool) {
	s.unfurlLinks = unfurl
}

// SetThreadManager sets the thread manager for handling threading
func (s *SenderSlack) SetThreadManager(threadManager sender_interfaces.SlackThreadManagerInterface) {
	s.threadManager = threadManager
}

// SetTableFormat sets the table formatting parameters
func (s *SenderSlack) SetTableFormat(tableFormat string) {
	s.tableFormat = tableFormat
}

// SetMaxTableRows sets the maximum number of rows for table formatting
func (s *SenderSlack) SetMaxTableRows(maxTableRows int) {
	s.maxTableRows = maxTableRows
}

// EnableThreading enables threading support by creating and configuring a ThreadManager
func (s *SenderSlack) EnableThreading(cacheTTL time.Duration, searchLimit int, searchWindow time.Duration) {
	s.threadManager = NewThreadManager(
		s.slackClient,
		s.channel,
		s.logger,
		cacheTTL,
		searchLimit,
		searchWindow,
	)

	s.logger.Info("Thread management enabled",
		zap.String("channel", s.channel),
		zap.String("cacheTTL", cacheTTL.String()),
		zap.String("searchWindow", searchWindow.String()),
		zap.Int("searchLimit", searchLimit),
	)
}

// generateFingerprint creates a unique fingerprint for the issue to identify related alerts
func (s *SenderSlack) generateFingerprint(issue *issuepkg.Issue) string {
	// Use existing fingerprint if available
	if issue.Fingerprint != "" {
		return issue.Fingerprint
	}

	// Create a stable identifier based on issue characteristics
	// This allows firing and resolved alerts to be grouped together

	var parts []string

	// Add issue title (normalized)
	if issue.Title != "" {
		parts = append(parts, issue.Title)
	}

	// Add source
	if issue.Source != issuepkg.SourceUnknown {
		parts = append(parts, issue.Source.String())
	}

	// Add subject information if available
	if issue.Subject != nil {
		if issue.Subject.Name != "" {
			parts = append(parts, issue.Subject.Name)
		}
		if issue.Subject.Namespace != "" {
			parts = append(parts, issue.Subject.Namespace)
		}
		if issue.Subject.SubjectType != issuepkg.SubjectTypeNone {
			parts = append(parts, issue.Subject.SubjectType.String())
		}
	}

	// Join all parts with a separator
	if len(parts) == 0 {
		return fmt.Sprintf("issue-%d", issue.StartsAt.Unix())
	}

	// Create a simple but stable fingerprint by hashing the joined parts
	joinedParts := strings.Join(parts, "|")
	hash := sha256.Sum256([]byte(joinedParts))
	fingerprint := "alert:" + hex.EncodeToString(hash[:8]) // Use first 8 bytes for shorter fingerprint

	s.logger.Debug("Generated fingerprint for issue",
		zap.String("fingerprint", fingerprint),
		zap.String("title", issue.Title),
		zap.String("status", issue.Status.String()),
	)

	return fingerprint
}
