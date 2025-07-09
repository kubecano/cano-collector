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

func TestTableBlock_ToMarkdown_NoHeaders(t *testing.T) {
	// Test case for bug fix: table without headers should render row data
	rows := [][]string{
		{"John", "25", "NYC"},
		{"Jane", "30", "LA"},
	}

	block := NewTableBlock([]string{}, rows, "", TableBlockFormatHorizontal)
	markdown := block.ToMarkdown()

	expected := `| John | 25 | NYC |
| Jane | 30 | LA |
`
	assert.Equal(t, expected, markdown)
}

func TestTableBlock_ToMarkdown_UnevenRows(t *testing.T) {
	// Test case for bug fix: rows with fewer cells than headers should be padded
	headers := []string{"Name", "Age", "City", "Country"}
	rows := [][]string{
		{"John", "25"},              // Missing City and Country
		{"Jane", "30", "LA"},        // Missing Country
		{"Bob", "35", "NYC", "USA"}, // All fields present
	}

	block := NewTableBlock(headers, rows, "", TableBlockFormatHorizontal)
	markdown := block.ToMarkdown()

	expected := `| Name | Age | City | Country |
| --- | --- | --- | --- |
| John | 25 |  |  |
| Jane | 30 | LA |  |
| Bob | 35 | NYC | USA |
`
	assert.Equal(t, expected, markdown)
}

func TestTableBlock_ToMarkdown_RowsWithMoreColumnsThanHeaders(t *testing.T) {
	// Test case for bug fix: rows with more columns than headers should expand table
	headers := []string{"Name", "Age"}
	rows := [][]string{
		{"John", "25", "NYC", "USA"},
		{"Jane", "30", "LA"},
	}

	block := NewTableBlock(headers, rows, "", TableBlockFormatHorizontal)
	markdown := block.ToMarkdown()

	expected := `| Name | Age |  |  |
| --- | --- | --- | --- |
| John | 25 | NYC | USA |
| Jane | 30 | LA |  |
`
	assert.Equal(t, expected, markdown)
}

func TestTableBlock_GetColumnCount_WithUnevenRows(t *testing.T) {
	// Test GetColumnCount with rows that have different numbers of columns
	headers := []string{"Name", "Age"}
	rows := [][]string{
		{"John", "25", "NYC", "USA"}, // 4 columns
		{"Jane", "30"},               // 2 columns
	}

	block := NewTableBlock(headers, rows, "", TableBlockFormatHorizontal)
	assert.Equal(t, 4, block.GetColumnCount()) // Should return max columns from all rows
}

func TestTableBlock_GetColumnCount_NoHeaders(t *testing.T) {
	// Test GetColumnCount when there are no headers
	rows := [][]string{
		{"John", "25", "NYC"},
		{"Jane", "30"},
	}

	block := NewTableBlock([]string{}, rows, "", TableBlockFormatHorizontal)
	assert.Equal(t, 3, block.GetColumnCount()) // Should return max columns from rows
}

func TestHeaderBlock_NewHeaderBlock(t *testing.T) {
	headerBlock := NewHeaderBlock("Test Header")

	assert.Equal(t, "Test Header", headerBlock.Text)
	assert.Equal(t, "header", headerBlock.BlockType())
}

func TestListBlock_NewListBlock(t *testing.T) {
	items := []string{"Item 1", "Item 2", "Item 3"}
	listBlock := NewListBlock(items, false, "Test List")

	assert.Equal(t, items, listBlock.Items)
	assert.False(t, listBlock.Ordered)
	assert.Equal(t, "Test List", listBlock.ListName)
	assert.Equal(t, "list", listBlock.BlockType())
}

func TestListBlock_AddItem(t *testing.T) {
	listBlock := NewListBlock([]string{}, false, "")

	listBlock.AddItem("Item 1")
	listBlock.AddItem("Item 2")

	assert.Len(t, listBlock.Items, 2)
	assert.Equal(t, "Item 1", listBlock.Items[0])
	assert.Equal(t, "Item 2", listBlock.Items[1])
}

func TestListBlock_ToMarkdown_Unordered(t *testing.T) {
	items := []string{"First item", "Second item", "Third item"}
	listBlock := NewListBlock(items, false, "My List")

	markdown := listBlock.ToMarkdown()

	assert.Contains(t, markdown, "**My List**")
	assert.Contains(t, markdown, "- First item")
	assert.Contains(t, markdown, "- Second item")
	assert.Contains(t, markdown, "- Third item")
}

func TestListBlock_ToMarkdown_Ordered(t *testing.T) {
	items := []string{"First step", "Second step", "Third step"}
	listBlock := NewListBlock(items, true, "Steps")

	markdown := listBlock.ToMarkdown()

	assert.Contains(t, markdown, "**Steps**")
	assert.Contains(t, markdown, "1. First step")
	assert.Contains(t, markdown, "2. Second step")
	assert.Contains(t, markdown, "3. Third step")
}

func TestListBlock_ToMarkdown_Empty(t *testing.T) {
	listBlock := NewListBlock([]string{}, false, "Empty List")

	markdown := listBlock.ToMarkdown()

	assert.Empty(t, markdown)
}

func TestLinksBlock_NewLinksBlock(t *testing.T) {
	links := []Link{
		{Text: "Dashboard", URL: "https://example.com/dashboard", Type: LinkTypeGeneral},
		{Text: "Logs", URL: "https://example.com/logs", Type: LinkTypeGeneral},
	}
	linksBlock := NewLinksBlock(links, "Related Links")

	assert.Equal(t, links, linksBlock.Links)
	assert.Equal(t, "Related Links", linksBlock.BlockName)
	assert.Equal(t, "links", linksBlock.BlockType())
}

func TestLinksBlock_AddLink(t *testing.T) {
	linksBlock := NewLinksBlock([]Link{}, "")

	link1 := Link{Text: "Dashboard", URL: "https://example.com/dashboard", Type: LinkTypeGeneral}
	link2 := Link{Text: "Logs", URL: "https://example.com/logs", Type: LinkTypeGeneral}

	linksBlock.AddLink(link1)
	linksBlock.AddLink(link2)

	assert.Len(t, linksBlock.Links, 2)
	assert.Equal(t, link1, linksBlock.Links[0])
	assert.Equal(t, link2, linksBlock.Links[1])
}

func TestFileBlock_NewFileBlock(t *testing.T) {
	content := []byte("test file content")
	fileBlock := NewFileBlock("test.txt", content, "text/plain")

	assert.Equal(t, "test.txt", fileBlock.Filename)
	assert.Equal(t, content, fileBlock.Contents)
	assert.Equal(t, "text/plain", fileBlock.MimeType)
	assert.Equal(t, int64(len(content)), fileBlock.Size)
	assert.Equal(t, "file", fileBlock.BlockType())
}

func TestFileBlock_GetSizeKB(t *testing.T) {
	content := make([]byte, 2048) // 2KB
	fileBlock := NewFileBlock("test.bin", content, "application/octet-stream")

	sizeKB := fileBlock.GetSizeKB()
	assert.InDelta(t, 2.0, sizeKB, 0.01)
}

func TestDividerBlock_NewDividerBlock(t *testing.T) {
	dividerBlock := NewDividerBlock()

	assert.Equal(t, "divider", dividerBlock.BlockType())
}
