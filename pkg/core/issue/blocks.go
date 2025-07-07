package issue

import (
	"fmt"
	"strings"
)

// MarkdownBlock represents a block of Markdown content
type MarkdownBlock struct {
	Text string `json:"text"`
}

// BlockType returns the type of this block
func (mb *MarkdownBlock) BlockType() string {
	return "markdown"
}

// NewMarkdownBlock creates a new MarkdownBlock
func NewMarkdownBlock(text string) *MarkdownBlock {
	return &MarkdownBlock{
		Text: text,
	}
}

// TableBlockFormat represents the format of a table block
type TableBlockFormat int

const (
	TableBlockFormatHorizontal TableBlockFormat = iota
	TableBlockFormatVertical
)

// String returns the string representation of the table block format
func (tbf TableBlockFormat) String() string {
	switch tbf {
	case TableBlockFormatHorizontal:
		return "horizontal"
	case TableBlockFormatVertical:
		return "vertical"
	default:
		return "horizontal"
	}
}

// TableBlock represents a table with headers and rows
type TableBlock struct {
	Headers     []string         `json:"headers"`
	Rows        [][]string       `json:"rows"`
	TableName   string           `json:"table_name,omitempty"`
	TableFormat TableBlockFormat `json:"table_format"`
}

// BlockType returns the type of this block
func (tb *TableBlock) BlockType() string {
	return "table"
}

// NewTableBlock creates a new TableBlock
func NewTableBlock(headers []string, rows [][]string, tableName string, format TableBlockFormat) *TableBlock {
	return &TableBlock{
		Headers:     headers,
		Rows:        rows,
		TableName:   tableName,
		TableFormat: format,
	}
}

// ToMarkdown converts the table to markdown format
func (tb *TableBlock) ToMarkdown() string {
	if len(tb.Headers) == 0 && len(tb.Rows) == 0 {
		return ""
	}

	var builder strings.Builder

	// Add table name if present
	if tb.TableName != "" {
		builder.WriteString(fmt.Sprintf("**%s**\n\n", tb.TableName))
	}

	// Handle vertical format (key-value pairs)
	if tb.TableFormat == TableBlockFormatVertical {
		for _, row := range tb.Rows {
			if len(row) >= 2 {
				builder.WriteString(fmt.Sprintf("| %s | %s |\n", row[0], row[1]))
			}
		}
		return builder.String()
	}

	// Handle horizontal format (traditional table)
	if len(tb.Headers) > 0 {
		// Headers
		builder.WriteString("|")
		for _, header := range tb.Headers {
			builder.WriteString(fmt.Sprintf(" %s |", header))
		}
		builder.WriteString("\n")

		// Separator
		builder.WriteString("|")
		for range tb.Headers {
			builder.WriteString(" --- |")
		}
		builder.WriteString("\n")
	}

	// Rows
	for _, row := range tb.Rows {
		builder.WriteString("|")
		for i, cell := range row {
			if i < len(tb.Headers) {
				builder.WriteString(fmt.Sprintf(" %s |", cell))
			}
		}
		builder.WriteString("\n")
	}

	return builder.String()
}

// AddRow adds a new row to the table
func (tb *TableBlock) AddRow(row []string) {
	tb.Rows = append(tb.Rows, row)
}

// GetRowCount returns the number of rows in the table
func (tb *TableBlock) GetRowCount() int {
	return len(tb.Rows)
}

// GetColumnCount returns the number of columns in the table
func (tb *TableBlock) GetColumnCount() int {
	return len(tb.Headers)
}
