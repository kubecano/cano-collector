package slack

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	slackapi "github.com/slack-go/slack"
	"go.uber.org/zap"

	issuepkg "github.com/kubecano/cano-collector/pkg/core/issue"
	logger_interfaces "github.com/kubecano/cano-collector/pkg/logger/interfaces"
	sender_interfaces "github.com/kubecano/cano-collector/pkg/sender/interfaces"
	"github.com/kubecano/cano-collector/pkg/sender/slack/templates"
	"github.com/kubecano/cano-collector/pkg/util"
)

type SenderSlack struct {
	apiKey         string
	channel        string
	channelID      string // Cached channel ID after resolution
	logger         logger_interfaces.LoggerInterface
	unfurlLinks    bool
	slackClient    sender_interfaces.SlackClientInterface
	threadManager  sender_interfaces.SlackThreadManagerInterface
	templateLoader *templates.TemplateLoader // Template system for message formatting
	tableFormat    string
	maxTableRows   int

	// Prometheus metrics
	fileUploadsTotal          *prometheus.CounterVec
	fileUploadSizeBytes       *prometheus.HistogramVec
	tableConversionsTotal     *prometheus.CounterVec
	channelResolutionDuration prometheus.Histogram
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

	// Initialize template loader
	templateLoader, err := templates.NewTemplateLoader()
	if err != nil {
		logger.Error("Failed to load Slack message templates, falling back to hardcoded format", zap.Error(err))
		// Continue without templates - will use fallback formatting
	}

	s := &SenderSlack{
		apiKey:         apiKey,
		channel:        channel,
		logger:         logger,
		unfurlLinks:    unfurlLinks,
		slackClient:    slackClient,
		templateLoader: templateLoader,
		tableFormat:    "enhanced", // Default table format
		maxTableRows:   20,         // Default max rows before converting to file
	}

	// Initialize Prometheus metrics with error handling to avoid test panics
	s.fileUploadsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cano_slack_file_uploads_total",
			Help: "Total number of Slack file uploads",
		},
		[]string{"status", "channel"},
	)
	if err := prometheus.Register(s.fileUploadsTotal); err != nil {
		// Metric already registered, try to get existing one
		var are prometheus.AlreadyRegisteredError
		if errors.As(err, &are) {
			s.fileUploadsTotal = are.ExistingCollector.(*prometheus.CounterVec)
		}
	}

	s.fileUploadSizeBytes = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cano_slack_file_upload_size_bytes",
			Help:    "Size of uploaded files in bytes",
			Buckets: prometheus.ExponentialBuckets(1024, 2, 10), // 1KB to 512KB
		},
		[]string{"channel", "type"},
	)
	if err := prometheus.Register(s.fileUploadSizeBytes); err != nil {
		var are prometheus.AlreadyRegisteredError
		if errors.As(err, &are) {
			s.fileUploadSizeBytes = are.ExistingCollector.(*prometheus.HistogramVec)
		}
	}

	s.tableConversionsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cano_slack_table_conversions_total",
			Help: "Total number of table to CSV conversions",
		},
		[]string{"format"},
	)
	if err := prometheus.Register(s.tableConversionsTotal); err != nil {
		var are prometheus.AlreadyRegisteredError
		if errors.As(err, &are) {
			s.tableConversionsTotal = are.ExistingCollector.(*prometheus.CounterVec)
		}
	}

	s.channelResolutionDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "cano_slack_channel_resolution_duration_seconds",
			Help:    "Time taken to resolve channel ID",
			Buckets: prometheus.DefBuckets,
		},
	)
	if err := prometheus.Register(s.channelResolutionDuration); err != nil {
		var are prometheus.AlreadyRegisteredError
		if errors.As(err, &are) {
			s.channelResolutionDuration = are.ExistingCollector.(prometheus.Histogram)
		}
	}

	return s
}

func (s *SenderSlack) Send(ctx context.Context, issue *issuepkg.Issue) error {
	s.logger.Info("Sending Slack notification",
		zap.String("channel", s.channel),
		zap.String("status", issue.Status.String()),
	)

	// Preprocess enrichments: upload files and set FileInfo
	s.preprocessEnrichments(issue)

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

// deduplicateEnrichments removes duplicate enrichments based on composite key
func (s *SenderSlack) deduplicateEnrichments(enrichments []issuepkg.Enrichment) []issuepkg.Enrichment {
	seen := make(map[string]bool)
	unique := []issuepkg.Enrichment{}

	for _, e := range enrichments {
		// Generate unique key: type + title + first block identifier
		var key string
		key = e.Type.String()
		if e.Title != "" {
			key += ":" + e.Title
		}

		// Add first block identifier for better uniqueness
		if len(e.Blocks) > 0 {
			switch block := e.Blocks[0].(type) {
			case *issuepkg.FileBlock:
				key += ":" + block.Filename
			case *issuepkg.TableBlock:
				key += ":" + block.TableName
			case *issuepkg.MarkdownBlock:
				// Use first 50 chars of markdown as identifier
				if len(block.Text) > 50 {
					key += ":" + block.Text[:50]
				} else {
					key += ":" + block.Text
				}
			}
		}

		if !seen[key] {
			seen[key] = true
			unique = append(unique, e)
		} else {
			s.logger.Debug("Skipping duplicate enrichment",
				zap.String("key", key),
				zap.String("type", e.Type.String()),
			)
		}
	}

	return unique
}

// buildSlackBlocks creates the main message blocks
func (s *SenderSlack) buildSlackBlocks(issue *issuepkg.Issue) []slackapi.Block {
	// If template loader is not available, fall back to legacy format
	if s.templateLoader == nil {
		return s.buildSlackBlocksLegacy(issue)
	}

	// Build template context from issue
	context := s.buildMessageContext(issue)

	var blocks []slackapi.Block

	// 1. Render header from template
	headerBlocks, err := s.templateLoader.RenderToBlocks("header.tmpl", context)
	if err != nil {
		s.logger.Error("Failed to render header template, using fallback", zap.Error(err))
		headerBlocks = s.buildHeaderBlockFallback(issue)
	}
	blocks = append(blocks, headerBlocks...)

	// 2. Render context bar from template
	contextBlocks, err := s.templateLoader.RenderToBlocks("context_bar.tmpl", context)
	if err != nil {
		s.logger.Error("Failed to render context bar template", zap.Error(err))
	} else {
		blocks = append(blocks, contextBlocks...)
	}

	// 3. Add divider
	blocks = append(blocks, slackapi.NewDividerBlock())

	// 4. Render description if present
	if context.Description != "" {
		descBlocks, err := s.templateLoader.RenderToBlocks("description.tmpl", context)
		if err != nil {
			s.logger.Error("Failed to render description template", zap.Error(err))
		} else {
			blocks = append(blocks, descBlocks...)
		}
	}

	// 5. Render links if present
	if len(context.Links) > 0 {
		linksBlocks, err := s.templateLoader.RenderToBlocks("links.tmpl", context)
		if err != nil {
			s.logger.Error("Failed to render links template", zap.Error(err))
		} else {
			blocks = append(blocks, linksBlocks...)
		}
	}

	// 6. Render crash info if present (for pod alerts)
	if context.CrashInfo != nil {
		crashBlocks, err := s.templateLoader.RenderToBlocks("crash_info.tmpl", context)
		if err != nil {
			s.logger.Error("Failed to render crash info template", zap.Error(err))
		} else {
			blocks = append(blocks, crashBlocks...)
		}
	}

	// 7. Render file enrichments with permalinks (Robusta-style)
	for _, enrichment := range context.Enrichments {
		if enrichment.FileLink != "" {
			// This is a file enrichment with permalink
			fileBlocks, err := s.templateLoader.RenderToBlocks("file_enrichment.tmpl", enrichment)
			if err != nil {
				s.logger.Error("Failed to render file enrichment template", zap.Error(err))
			} else {
				blocks = append(blocks, fileBlocks...)
			}
		}
	}

	// 8. Add other enrichments (tables, markdown, etc.) using existing logic
	uniqueEnrichments := s.deduplicateEnrichments(issue.Enrichments)
	enrichmentBlocks := s.buildEnrichmentBlocks(uniqueEnrichments)
	blocks = append(blocks, enrichmentBlocks...)

	return blocks
}

// buildSlackBlocksLegacy is the old implementation (fallback if templates fail)
func (s *SenderSlack) buildSlackBlocksLegacy(issue *issuepkg.Issue) []slackapi.Block {
	var blocks []slackapi.Block

	// Header block with status, severity and title
	headerText := s.formatHeader(issue)
	headerBlock := slackapi.NewSectionBlock(
		slackapi.NewTextBlockObject("mrkdwn", headerText, false, false),
		nil, nil,
	)
	blocks = append(blocks, headerBlock)

	// For resolved alerts, only show minimal information
	if issue.IsResolved() {
		// Add source and cluster information
		sourceText := "📍 Source: " + strings.ToUpper(issue.Source.String())
		if issue.ClusterName != "" {
			sourceText += "\n🌐 Cluster: " + issue.ClusterName
		}
		sourceBlock := slackapi.NewSectionBlock(
			slackapi.NewTextBlockObject("mrkdwn", sourceText, false, false),
			nil, nil,
		)
		blocks = append(blocks, sourceBlock)

		// Add resolved timestamp if available
		if issue.EndsAt != nil {
			resolvedText := "🕐 Resolved: " + issue.EndsAt.UTC().Format("2006-01-02 15:04:05 UTC")
			resolvedBlock := slackapi.NewSectionBlock(
				slackapi.NewTextBlockObject("mrkdwn", resolvedText, false, false),
				nil, nil,
			)
			blocks = append(blocks, resolvedBlock)
		}

		// Return early for resolved alerts - skip all other details
		return blocks
	}

	// For firing alerts, show full details
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

	// Deduplicate and add enrichments
	uniqueEnrichments := s.deduplicateEnrichments(issue.Enrichments)
	enrichmentBlocks := s.buildEnrichmentBlocks(uniqueEnrichments)
	blocks = append(blocks, enrichmentBlocks...)

	// Add final divider if we have enrichments
	if len(issue.Enrichments) > 0 {
		blocks = append(blocks, slackapi.NewDividerBlock())
	}

	return blocks
}

// buildHeaderBlockFallback creates a simple header block if template rendering fails
func (s *SenderSlack) buildHeaderBlockFallback(issue *issuepkg.Issue) []slackapi.Block {
	headerText := s.formatHeader(issue)
	headerBlock := slackapi.NewSectionBlock(
		slackapi.NewTextBlockObject("mrkdwn", headerText, false, false),
		nil, nil,
	)
	return []slackapi.Block{headerBlock}
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
	if enrichment.Title != "" {
		// Small bold title instead of large header
		titleText := "*" + enrichment.Title + "*"

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

// tableToCSV converts a table block to CSV format
func (s *SenderSlack) tableToCSV(table *issuepkg.TableBlock) string {
	var csvContent strings.Builder

	// Add headers if they exist
	if len(table.Headers) > 0 {
		for i, header := range table.Headers {
			if i > 0 {
				csvContent.WriteString(",")
			}
			// Escape quotes and wrap in quotes if contains comma or quote
			escapedHeader := strings.ReplaceAll(header, "\"", "\"\"")
			if strings.Contains(header, ",") || strings.Contains(header, "\"") || strings.Contains(header, "\n") {
				csvContent.WriteString(fmt.Sprintf("\"%s\"", escapedHeader))
			} else {
				csvContent.WriteString(escapedHeader)
			}
		}
		csvContent.WriteString("\n")
	}

	// Add rows
	for _, row := range table.Rows {
		for i, cell := range row {
			if i > 0 {
				csvContent.WriteString(",")
			}
			// Escape quotes and wrap in quotes if contains comma or quote
			escapedCell := strings.ReplaceAll(cell, "\"", "\"\"")
			if strings.Contains(cell, ",") || strings.Contains(cell, "\"") || strings.Contains(cell, "\n") {
				csvContent.WriteString(fmt.Sprintf("\"%s\"", escapedCell))
			} else {
				csvContent.WriteString(escapedCell)
			}
		}
		csvContent.WriteString("\n")
	}

	return csvContent.String()
}

// convertLargeTableToFileBlock converts large tables to file attachment
func (s *SenderSlack) convertLargeTableToFileBlock(table *issuepkg.TableBlock) slackapi.Block {
	// Record table conversion metric (if metrics are initialized)
	if s.tableConversionsTotal != nil {
		s.tableConversionsTotal.WithLabelValues("csv").Inc()
	}

	rowCount := len(table.Rows)
	tableName := table.TableName
	if tableName == "" {
		tableName = "Large_Table"
	}

	// Convert table to CSV format
	csvContent := s.tableToCSV(table)

	// Generate filename with timestamp
	sanitizedName := strings.ReplaceAll(tableName, " ", "_")
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("%s-%s.csv", sanitizedName, timestamp)

	// Upload CSV file to Slack workspace storage
	csvBytes := []byte(csvContent)
	permalink, err := s.uploadFileToSlack(filename, csvBytes)
	if err != nil {
		s.logger.Warn("Table file upload failed, falling back to inline display",
			zap.Error(err),
			zap.String("table_name", tableName),
			zap.Int("rows", rowCount),
		)
		return s.createTableErrorBlock(table, err)
	}

	// Create block with successful file upload using permalink
	var textBuilder strings.Builder
	textBuilder.WriteString(fmt.Sprintf("📊 *%s* (%d rows)\n", tableName, rowCount))
	textBuilder.WriteString(fmt.Sprintf("Table converted to CSV file (limit: %d rows)\n", s.maxTableRows))
	textBuilder.WriteString(fmt.Sprintf("<%s|View CSV File>", permalink))

	return slackapi.NewSectionBlock(
		slackapi.NewTextBlockObject("mrkdwn", textBuilder.String(), false, false),
		nil, nil,
	)
}

// createTableErrorBlock creates a fallback block when table file upload fails
func (s *SenderSlack) createTableErrorBlock(table *issuepkg.TableBlock, err error) slackapi.Block {
	rowCount := len(table.Rows)
	tableName := table.TableName
	if tableName == "" {
		tableName = "Large Table"
	}

	var textBuilder strings.Builder
	textBuilder.WriteString(fmt.Sprintf("📊 *%s* (%d rows) - upload failed\n", tableName, rowCount))
	textBuilder.WriteString(fmt.Sprintf("Error: %s\n", err.Error()))
	textBuilder.WriteString("Showing simplified table view:\n\n")

	// Show first few rows as fallback
	maxRowsToShow := 5
	if len(table.Headers) > 0 {
		textBuilder.WriteString("*Headers:* " + strings.Join(table.Headers, " | ") + "\n")
	}

	rowsShown := 0
	for _, row := range table.Rows {
		if rowsShown >= maxRowsToShow {
			textBuilder.WriteString(fmt.Sprintf("... and %d more rows", len(table.Rows)-rowsShown))
			break
		}
		textBuilder.WriteString(fmt.Sprintf("• %s\n", strings.Join(row, " | ")))
		rowsShown++
	}

	return slackapi.NewSectionBlock(
		slackapi.NewTextBlockObject("mrkdwn", textBuilder.String(), false, false),
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

// resolveChannelID resolves channel name to channel ID if needed
// Returns cached channelID if available, otherwise resolves from Slack API
func (s *SenderSlack) resolveChannelID() (string, error) {
	// If we already have the channel ID cached, return it
	if s.channelID != "" {
		return s.channelID, nil
	}

	// If channel is already an ID (doesn't start with #), use it directly
	if !strings.HasPrefix(s.channel, "#") {
		s.channelID = s.channel
		return s.channelID, nil
	}

	// Measure channel resolution duration (if metrics are initialized)
	var timer *prometheus.Timer
	if s.channelResolutionDuration != nil {
		timer = prometheus.NewTimer(s.channelResolutionDuration)
		defer timer.ObserveDuration()
	}

	// Channel is a name (#channel-name), need to resolve to ID
	channelName := strings.TrimPrefix(s.channel, "#")
	s.logger.Debug("Resolving channel name to ID",
		zap.String("channel_name", channelName),
	)

	channels, _, err := s.slackClient.GetConversations(&slackapi.GetConversationsParameters{
		Types: []string{"public_channel", "private_channel"},
		Limit: 1000,
	})
	if err != nil {
		return "", fmt.Errorf("failed to list conversations: %w", err)
	}

	for _, ch := range channels {
		if ch.Name == channelName {
			s.channelID = ch.ID
			s.logger.Info("Channel resolved successfully",
				zap.String("channel_name", channelName),
				zap.String("channel_id", ch.ID),
			)
			return s.channelID, nil
		}
	}

	return "", fmt.Errorf("channel not found: %s", s.channel)
}

// preprocessEnrichments uploads files from FileBlocks and sets FileInfo for template rendering
func (s *SenderSlack) preprocessEnrichments(issue *issuepkg.Issue) {
	for i := range issue.Enrichments {
		enrichment := &issue.Enrichments[i]

		// Skip if enrichment already has FileInfo set
		if enrichment.FileInfo != nil && enrichment.FileInfo.Permalink != "" {
			continue
		}

		// Look for FileBlock in enrichment blocks
		for _, block := range enrichment.Blocks {
			if fileBlock, ok := block.(*issuepkg.FileBlock); ok {
				// Upload file to Slack
				permalink, err := s.uploadFileToSlack(fileBlock.Filename, fileBlock.Contents)
				if err != nil {
					s.logger.Warn("Failed to upload file for enrichment, skipping FileInfo",
						zap.Error(err),
						zap.String("filename", fileBlock.Filename),
						zap.String("enrichment_title", enrichment.Title),
					)
					continue
				}

				// Set FileInfo with permalink
				enrichment.FileInfo = &issuepkg.FileInfo{
					Permalink: permalink,
					Filename:  fileBlock.Filename,
					Size:      fileBlock.Size,
					MimeType:  fileBlock.MimeType,
				}

				// Set Content to file contents as string (for inline display if needed)
				enrichment.Content = string(fileBlock.Contents)

				s.logger.Debug("File uploaded and FileInfo set for enrichment",
					zap.String("filename", fileBlock.Filename),
					zap.String("permalink", permalink),
					zap.String("enrichment_title", enrichment.Title),
				)

				// Only process first FileBlock per enrichment
				break
			}
		}
	}
}

// uploadFileToSlack uploads a file to Slack workspace storage using files_upload_v2 API
// Returns the permalink to the uploaded file, which can be included in any message
func (s *SenderSlack) uploadFileToSlack(filename string, content []byte) (string, error) {
	// Skip empty files (Slack throws error on empty files)
	if len(content) == 0 {
		s.logger.Warn("Skipping empty file upload",
			zap.String("filename", filename),
		)
		return "", fmt.Errorf("file is empty")
	}

	// Determine file type for metrics
	fileType := "log"
	if strings.HasSuffix(filename, ".csv") {
		fileType = "csv"
	}

	// Observe file size (if metrics are initialized)
	if s.fileUploadSizeBytes != nil {
		s.fileUploadSizeBytes.WithLabelValues(s.channel, fileType).Observe(float64(len(content)))
	}

	// Try direct upload with bytes.Reader
	permalink, err := s.tryUploadDirect(filename, content)
	if err == nil {
		s.recordFileUploadSuccess("direct")
		s.logger.Info("File uploaded successfully",
			zap.String("method", "direct"),
			zap.String("filename", filename),
			zap.String("permalink", permalink),
			zap.Int("size", len(content)),
		)
		return permalink, nil
	}

	s.logger.Warn("Direct upload failed, trying temp file fallback",
		zap.Error(err),
		zap.String("filename", filename),
	)

	// Fallback: Upload via temporary file
	permalink, err = s.tryUploadViaTempFile(filename, content)
	if err != nil {
		s.recordFileUploadFailure()
		s.logger.Error("All upload strategies failed",
			zap.Error(err),
			zap.String("filename", filename),
			zap.Int("size", len(content)),
		)
		return "", fmt.Errorf("file upload failed: %w", err)
	}

	s.recordFileUploadSuccess("tempfile")
	s.logger.Info("File uploaded successfully",
		zap.String("method", "tempfile"),
		zap.String("filename", filename),
		zap.String("permalink", permalink),
		zap.Int("size", len(content)),
	)

	return permalink, nil
}

// tryUploadDirect attempts direct upload using bytes.Reader
func (s *SenderSlack) tryUploadDirect(filename string, content []byte) (string, error) {
	params := slackapi.UploadFileV2Parameters{
		Filename: filename,
		FileSize: len(content),
		Reader:   bytes.NewReader(content),
		Channel:  s.channel, // Share file to channel for visibility
	}

	return s.executeUpload(params)
}

// tryUploadViaTempFile attempts upload using temporary file as fallback strategy
func (s *SenderSlack) tryUploadViaTempFile(filename string, content []byte) (string, error) {
	tmpFile, err := os.CreateTemp("", "slack-upload-*")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err = tmpFile.Write(content); err != nil {
		return "", fmt.Errorf("write temp file: %w", err)
	}

	if _, err = tmpFile.Seek(0, 0); err != nil {
		return "", fmt.Errorf("seek temp file: %w", err)
	}

	params := slackapi.UploadFileV2Parameters{
		Filename: filename,
		FileSize: len(content),
		Reader:   tmpFile,
		Channel:  s.channel,
	}

	return s.executeUpload(params)
}

// executeUpload performs the actual upload and retrieves permalink
func (s *SenderSlack) executeUpload(params slackapi.UploadFileV2Parameters) (string, error) {
	fileSummary, err := s.slackClient.UploadFileV2(params)
	if err != nil {
		return "", fmt.Errorf("upload: %w", err)
	}

	fileInfo, _, _, err := s.slackClient.GetFileInfo(fileSummary.ID, 0, 0)
	if err != nil {
		return "", fmt.Errorf("get file info: %w", err)
	}

	return fileInfo.Permalink, nil
}

// recordFileUploadSuccess records successful file upload metric
func (s *SenderSlack) recordFileUploadSuccess(method string) {
	if s.fileUploadsTotal != nil {
		s.fileUploadsTotal.WithLabelValues("success", s.channel).Inc()
	}
}

// recordFileUploadFailure records failed file upload metric
func (s *SenderSlack) recordFileUploadFailure() {
	if s.fileUploadsTotal != nil {
		s.fileUploadsTotal.WithLabelValues("failure", s.channel).Inc()
	}
}

// convertFileBlockToSlack converts a file block to Slack section block with actual file upload
func (s *SenderSlack) convertFileBlockToSlack(file *issuepkg.FileBlock) slackapi.Block {
	// Attempt to upload file to Slack workspace storage
	permalink, err := s.uploadFileToSlack(file.Filename, file.Contents)
	if err != nil {
		s.logger.Warn("File upload failed, falling back to text display",
			zap.Error(err),
			zap.String("filename", file.Filename),
		)
		return s.createFileErrorBlock(file, err)
	}

	// Create block with successful file upload using permalink
	sizeKB := file.GetSizeKB()
	sizeText := fmt.Sprintf("%.1f KB", sizeKB)
	if sizeKB > 1024 {
		sizeMB := sizeKB / 1024
		sizeText = fmt.Sprintf("%.1f MB", sizeMB)
	}

	text := fmt.Sprintf("📎 *%s* (%s)", file.Filename, sizeText)
	if file.MimeType != "" {
		text += " - " + file.MimeType
	}
	text += fmt.Sprintf("\n<%s|View File>", permalink)

	return slackapi.NewSectionBlock(
		slackapi.NewTextBlockObject("mrkdwn", text, false, false),
		nil, nil,
	)
}

// createFileErrorBlock creates a fallback block when file upload fails
func (s *SenderSlack) createFileErrorBlock(file *issuepkg.FileBlock, err error) slackapi.Block {
	sizeKB := file.GetSizeKB()
	sizeText := fmt.Sprintf("%.1f KB", sizeKB)
	if sizeKB > 1024 {
		sizeMB := sizeKB / 1024
		sizeText = fmt.Sprintf("%.1f MB", sizeMB)
	}

	var textBuilder strings.Builder
	textBuilder.WriteString(fmt.Sprintf("📎 *File: %s* (upload failed)\n", file.Filename))
	textBuilder.WriteString("Size: " + sizeText)
	if file.MimeType != "" {
		textBuilder.WriteString("\nType: " + file.MimeType)
	}
	textBuilder.WriteString("\nError: " + err.Error())

	// Show preview of content if it's text and not too large
	if strings.HasPrefix(file.MimeType, "text/") && len(file.Contents) < 2000 {
		preview := string(file.Contents)
		if len(preview) > 500 {
			preview = preview[:500] + "..."
		}
		textBuilder.WriteString(fmt.Sprintf("\n\n*Content preview:*\n```\n%s\n```", preview))
	}

	return slackapi.NewSectionBlock(
		slackapi.NewTextBlockObject("mrkdwn", textBuilder.String(), false, false),
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

// getSeverityEmoji returns only the emoji for severity (for template use)
func (s *SenderSlack) getSeverityEmoji(severity issuepkg.Severity) string {
	switch severity {
	case issuepkg.SeverityHigh:
		return "🔴"
	case issuepkg.SeverityLow:
		return "🟡"
	case issuepkg.SeverityInfo:
		return "🟢"
	case issuepkg.SeverityDebug:
		return "🔵"
	default:
		return "🟢"
	}
}

// getSeverityName returns severity name without emoji (for template use)
func (s *SenderSlack) getSeverityName(severity issuepkg.Severity) string {
	switch severity {
	case issuepkg.SeverityHigh:
		return "High"
	case issuepkg.SeverityLow:
		return "Low"
	case issuepkg.SeverityInfo:
		return "Info"
	case issuepkg.SeverityDebug:
		return "Debug"
	default:
		return severity.String()
	}
}

// buildMessageContext prepares template context from Issue
func (s *SenderSlack) buildMessageContext(issue *issuepkg.Issue) *MessageContext {
	context := &MessageContext{
		Title:       issue.Title,
		Description: issue.Description,
		Cluster:     issue.ClusterName,
		Namespace:   issue.Subject.Namespace,
		PodName:     issue.Subject.Name,
		Source:      issue.Source.String(),
	}

	// Status
	if issue.IsResolved() {
		context.Status = "resolved"
		context.StatusEmoji = "✅"
		context.StatusText = "Alert resolved"
	} else {
		context.Status = "firing"
		context.StatusEmoji = "🔥"
		context.StatusText = "Alert firing"
	}

	// Severity
	context.Severity = s.getSeverityName(issue.Severity)
	context.SeverityEmoji = s.getSeverityEmoji(issue.Severity)

	// Alert type
	switch issue.Source {
	case issuepkg.SourcePrometheus:
		context.AlertType = "Prometheus Alert"
		context.AlertTypeEmoji = "📊"
	case issuepkg.SourceKubernetesAPIServer:
		context.AlertType = "K8s Event"
		context.AlertTypeEmoji = "👀"
	default:
		context.AlertType = "Notification"
		context.AlertTypeEmoji = "📬"
	}

	// Extract crash info from labels (for pod alerts)
	containerName := s.getIssueLabel(issue, "container")
	if containerName != "" {
		context.CrashInfo = &CrashInfo{
			Container: containerName,
			Restarts:  s.getIssueLabel(issue, "restarts"),
			Status:    s.getIssueLabel(issue, "status"),
			Reason:    s.getIssueLabel(issue, "reason"),
		}
	}

	// Links
	for _, link := range issue.Links {
		context.Links = append(context.Links, Link{
			Text: link.Text,
			URL:  link.URL,
		})
	}

	// Enrichments
	for _, enrichment := range issue.Enrichments {
		enrichData := EnrichmentData{
			Type:    enrichment.Type.String(),
			Title:   enrichment.Title,
			Content: enrichment.Content,
		}

		// Add file permalink if available
		if enrichment.FileInfo != nil && enrichment.FileInfo.Permalink != "" {
			enrichData.FileLink = enrichment.FileInfo.Permalink
		}

		context.Enrichments = append(context.Enrichments, enrichData)
	}

	return context
}

// getIssueLabel safely extracts label value from issue
func (s *SenderSlack) getIssueLabel(issue *issuepkg.Issue, key string) string {
	if issue.Subject.Labels == nil {
		return ""
	}
	return issue.Subject.Labels[key]
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
