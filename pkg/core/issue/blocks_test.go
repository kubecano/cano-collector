package issue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarkdownBlock_NewMarkdownBlock(t *testing.T) {
	text := "# Test Header\n\nThis is a test."
	block := NewMarkdownBlock(text)

	assert.Equal(t, text, block.Text)
	assert.Equal(t, "markdown", block.BlockType())
}

func TestMarkdownBlock_BlockType(t *testing.T) {
	block := &MarkdownBlock{Text: "test"}
	assert.Equal(t, "markdown", block.BlockType())
}

func TestTableBlock_NewTableBlock(t *testing.T) {
	headers := []string{"Name", "Age"}
	rows := [][]string{
		{"John", "25"},
		{"Jane", "30"},
	}
	tableName := "Users"
	format := TableBlockFormatHorizontal

	block := NewTableBlock(headers, rows, tableName, format)

	assert.Equal(t, headers, block.Headers)
	assert.Equal(t, rows, block.Rows)
	assert.Equal(t, tableName, block.TableName)
	assert.Equal(t, format, block.TableFormat)
	assert.Equal(t, "table", block.BlockType())
}

func TestTableBlock_ToMarkdown_Horizontal(t *testing.T) {
	headers := []string{"Name", "Age"}
	rows := [][]string{
		{"John", "25"},
		{"Jane", "30"},
	}
	tableName := "Users"

	block := NewTableBlock(headers, rows, tableName, TableBlockFormatHorizontal)
	markdown := block.ToMarkdown()

	expected := `**Users**

| Name | Age |
| --- | --- |
| John | 25 |
| Jane | 30 |
`
	assert.Equal(t, expected, markdown)
}

func TestTableBlock_ToMarkdown_Vertical(t *testing.T) {
	headers := []string{"Key", "Value"}
	rows := [][]string{
		{"Name", "John"},
		{"Age", "25"},
		{"City", "New York"},
	}
	tableName := "User Details"

	block := NewTableBlock(headers, rows, tableName, TableBlockFormatVertical)
	markdown := block.ToMarkdown()

	expected := `**User Details**

| Name | John |
| Age | 25 |
| City | New York |
`
	assert.Equal(t, expected, markdown)
}

func TestTableBlock_ToMarkdown_EmptyTable(t *testing.T) {
	block := NewTableBlock([]string{}, [][]string{}, "", TableBlockFormatHorizontal)
	markdown := block.ToMarkdown()

	assert.Empty(t, markdown)
}

func TestTableBlock_ToMarkdown_NoTableName(t *testing.T) {
	headers := []string{"Name", "Age"}
	rows := [][]string{
		{"John", "25"},
	}

	block := NewTableBlock(headers, rows, "", TableBlockFormatHorizontal)
	markdown := block.ToMarkdown()

	expected := `| Name | Age |
| --- | --- |
| John | 25 |
`
	assert.Equal(t, expected, markdown)
}

func TestTableBlock_AddRow(t *testing.T) {
	headers := []string{"Name", "Age"}
	rows := [][]string{
		{"John", "25"},
	}

	block := NewTableBlock(headers, rows, "", TableBlockFormatHorizontal)
	assert.Equal(t, 1, block.GetRowCount())

	block.AddRow([]string{"Jane", "30"})
	assert.Equal(t, 2, block.GetRowCount())
	assert.Equal(t, []string{"Jane", "30"}, block.Rows[1])
}

func TestTableBlock_GetRowCount(t *testing.T) {
	headers := []string{"Name", "Age"}
	rows := [][]string{
		{"John", "25"},
		{"Jane", "30"},
		{"Bob", "35"},
	}

	block := NewTableBlock(headers, rows, "", TableBlockFormatHorizontal)
	assert.Equal(t, 3, block.GetRowCount())
}

func TestTableBlock_GetColumnCount(t *testing.T) {
	headers := []string{"Name", "Age", "City"}
	rows := [][]string{
		{"John", "25", "NYC"},
	}

	block := NewTableBlock(headers, rows, "", TableBlockFormatHorizontal)
	assert.Equal(t, 3, block.GetColumnCount())
}

func TestTableBlockFormat_String(t *testing.T) {
	tests := []struct {
		format   TableBlockFormat
		expected string
	}{
		{TableBlockFormatHorizontal, "horizontal"},
		{TableBlockFormatVertical, "vertical"},
		{TableBlockFormat(999), "horizontal"}, // default case
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, test.format.String())
	}
}
