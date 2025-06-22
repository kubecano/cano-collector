package issue

// BaseBlock is the interface for all content blocks in an enrichment.
type BaseBlock interface {
	// IsBlock is a marker method for the interface.
	IsBlock()
}

// MarkdownBlock represents a block of Markdown text.
type MarkdownBlock struct {
	Text string
}

func (m MarkdownBlock) IsBlock() {}

// TableBlock represents a table of data.
type TableBlock struct {
	Rows    [][]string
	Headers []string
	Name    string
}

func (t TableBlock) IsBlock() {}

// FileBlock represents a file to be attached to a report.
type FileBlock struct {
	Filename string
	Contents []byte
}

func (f FileBlock) IsBlock() {}

// ListBlock represents a list of items.
type ListBlock struct {
	Items []string
}

func (l ListBlock) IsBlock() {}

// HeaderBlock represents a simple header text.
type HeaderBlock struct {
	Text string
}

func (h HeaderBlock) IsBlock() {}

// DividerBlock represents a horizontal divider.
type DividerBlock struct{}

func (d DividerBlock) IsBlock() {}

// LinksBlock represents a set of clickable links.
type LinksBlock struct {
	Links []Link
}

func (l LinksBlock) IsBlock() {}
