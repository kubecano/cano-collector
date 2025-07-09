package issue

import (
	"encoding/json"
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
	// Determine number of columns - consider both headers and all rows
	numColumns := len(tb.Headers)

	// Always check all rows to find the maximum number of columns
	for _, row := range tb.Rows {
		if len(row) > numColumns {
			numColumns = len(row)
		}
	}

	// If no columns at all, return empty string
	if numColumns == 0 {
		return ""
	}

	// Render headers if present
	if len(tb.Headers) > 0 {
		// Headers
		builder.WriteString("|")
		for i := 0; i < numColumns; i++ {
			var header string
			if i < len(tb.Headers) {
				header = tb.Headers[i]
			} else {
				header = "" // Pad missing headers with empty string
			}
			builder.WriteString(fmt.Sprintf(" %s |", header))
		}
		builder.WriteString("\n")

		// Separator - use numColumns, not len(tb.Headers)
		builder.WriteString("|")
		for i := 0; i < numColumns; i++ {
			builder.WriteString(" --- |")
		}
		builder.WriteString("\n")
	}

	// Rows
	for _, row := range tb.Rows {
		builder.WriteString("|")
		for i := 0; i < numColumns; i++ {
			var cell string
			if i < len(row) {
				cell = row[i]
			} else {
				cell = "" // Pad missing cells with empty string
			}
			builder.WriteString(fmt.Sprintf(" %s |", cell))
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
	// Use the same logic as ToMarkdown() method
	numColumns := len(tb.Headers)

	// Always check all rows to find the maximum number of columns
	for _, row := range tb.Rows {
		if len(row) > numColumns {
			numColumns = len(row)
		}
	}
	return numColumns
}

// JsonBlock represents a block of JSON content
type JsonBlock struct {
	Data interface{} `json:"data"`
}

// BlockType returns the type of this block
func (jb *JsonBlock) BlockType() string {
	return "json"
}

// NewJsonBlock creates a new JsonBlock
func NewJsonBlock(data interface{}) *JsonBlock {
	return &JsonBlock{
		Data: data,
	}
}

// ToJson converts the data to JSON string
func (jb *JsonBlock) ToJson() string {
	jsonBytes, err := json.MarshalIndent(jb.Data, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error marshaling JSON: %v", err)
	}
	return string(jsonBytes)
}

// HeaderBlock represents a header text block
type HeaderBlock struct {
	Text string `json:"text"`
}

// BlockType returns the type of this block
func (hb *HeaderBlock) BlockType() string {
	return "header"
}

// NewHeaderBlock creates a new HeaderBlock
func NewHeaderBlock(text string) *HeaderBlock {
	return &HeaderBlock{
		Text: text,
	}
}

// ListBlock represents a list of items
type ListBlock struct {
	Items    []string `json:"items"`
	Ordered  bool     `json:"ordered"` // true for numbered lists, false for bullet points
	ListName string   `json:"list_name,omitempty"`
}

// BlockType returns the type of this block
func (lb *ListBlock) BlockType() string {
	return "list"
}

// NewListBlock creates a new ListBlock
func NewListBlock(items []string, ordered bool, listName string) *ListBlock {
	return &ListBlock{
		Items:    items,
		Ordered:  ordered,
		ListName: listName,
	}
}

// AddItem adds an item to the list
func (lb *ListBlock) AddItem(item string) {
	lb.Items = append(lb.Items, item)
}

// ToMarkdown converts the list to markdown format
func (lb *ListBlock) ToMarkdown() string {
	if len(lb.Items) == 0 {
		return ""
	}

	var builder strings.Builder

	// Add list name if present
	if lb.ListName != "" {
		builder.WriteString(fmt.Sprintf("**%s**\n\n", lb.ListName))
	}

	// Generate list items
	for i, item := range lb.Items {
		if lb.Ordered {
			builder.WriteString(fmt.Sprintf("%d. %s\n", i+1, item))
		} else {
			builder.WriteString(fmt.Sprintf("- %s\n", item))
		}
	}

	return builder.String()
}

// LinksBlock represents a collection of links
type LinksBlock struct {
	Links     []Link `json:"links"`
	BlockName string `json:"block_name,omitempty"`
}

// BlockType returns the type of this block
func (lb *LinksBlock) BlockType() string {
	return "links"
}

// NewLinksBlock creates a new LinksBlock
func NewLinksBlock(links []Link, blockName string) *LinksBlock {
	return &LinksBlock{
		Links:     links,
		BlockName: blockName,
	}
}

// AddLink adds a link to the block
func (lb *LinksBlock) AddLink(link Link) {
	lb.Links = append(lb.Links, link)
}

// FileBlock represents a file attachment
type FileBlock struct {
	Filename string `json:"filename"`
	Contents []byte `json:"contents"`
	MimeType string `json:"mime_type,omitempty"`
	Size     int64  `json:"size"`
}

// BlockType returns the type of this block
func (fb *FileBlock) BlockType() string {
	return "file"
}

// NewFileBlock creates a new FileBlock
func NewFileBlock(filename string, contents []byte, mimeType string) *FileBlock {
	return &FileBlock{
		Filename: filename,
		Contents: contents,
		MimeType: mimeType,
		Size:     int64(len(contents)),
	}
}

// GetSizeKB returns file size in kilobytes
func (fb *FileBlock) GetSizeKB() float64 {
	return float64(fb.Size) / 1024.0
}

// DividerBlock represents a visual separator
type DividerBlock struct{}

// BlockType returns the type of this block
func (db *DividerBlock) BlockType() string {
	return "divider"
}

// NewDividerBlock creates a new DividerBlock
func NewDividerBlock() *DividerBlock {
	return &DividerBlock{}
}
