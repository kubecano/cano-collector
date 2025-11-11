package templates

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTemplateLoader(t *testing.T) {
	loader, err := NewTemplateLoader()

	require.NoError(t, err)
	require.NotNil(t, loader)
	assert.NotNil(t, loader.templates)

	// Verify all expected templates are loaded
	expectedTemplates := []string{
		"header.tmpl",
		"context_bar.tmpl",
		"crash_info.tmpl",
		"description.tmpl",
		"links.tmpl",
		"file_enrichment.tmpl",
	}

	for _, tmpl := range expectedTemplates {
		assert.Contains(t, loader.templates, tmpl, "Template %s should be loaded", tmpl)
	}
}

func TestTemplateLoader_RenderToBlocks_HeaderTemplate(t *testing.T) {
	loader, err := NewTemplateLoader()
	require.NoError(t, err)

	context := map[string]interface{}{
		"StatusEmoji":   "🔥",
		"StatusText":    "Alert firing",
		"SeverityEmoji": "🔴",
		"Severity":      "High",
		"Title":         "Test Alert",
	}

	blocks, err := loader.RenderToBlocks("header.tmpl", context)

	require.NoError(t, err)
	require.NotNil(t, blocks)
	assert.NotEmpty(t, blocks, "Should return at least one block")
}

func TestTemplateLoader_RenderToBlocks_ContextBarTemplate(t *testing.T) {
	loader, err := NewTemplateLoader()
	require.NoError(t, err)

	context := map[string]interface{}{
		"Source":    "prometheus",
		"Cluster":   "prod-cluster",
		"Namespace": "default",
		"Timestamp": "2024-01-01 12:00:00",
	}

	blocks, err := loader.RenderToBlocks("context_bar.tmpl", context)

	require.NoError(t, err)
	require.NotNil(t, blocks)
	assert.NotEmpty(t, blocks, "Should return at least one block")
}

func TestTemplateLoader_RenderToBlocks_DescriptionTemplate(t *testing.T) {
	loader, err := NewTemplateLoader()
	require.NoError(t, err)

	context := map[string]interface{}{
		"Description": "This is a test alert description",
	}

	blocks, err := loader.RenderToBlocks("description.tmpl", context)

	require.NoError(t, err)
	require.NotNil(t, blocks)
	assert.NotEmpty(t, blocks, "Should return at least one block")
}

func TestTemplateLoader_RenderToBlocks_LinksTemplate(t *testing.T) {
	loader, err := NewTemplateLoader()
	require.NoError(t, err)

	context := map[string]interface{}{
		"Links": []map[string]string{
			{"Text": "Grafana", "URL": "https://grafana.example.com"},
			{"Text": "Logs", "URL": "https://logs.example.com"},
		},
	}

	blocks, err := loader.RenderToBlocks("links.tmpl", context)

	require.NoError(t, err)
	require.NotNil(t, blocks)
}

func TestTemplateLoader_RenderToBlocks_TemplateNotFound(t *testing.T) {
	loader, err := NewTemplateLoader()
	require.NoError(t, err)

	context := map[string]interface{}{}

	blocks, err := loader.RenderToBlocks("nonexistent.tmpl", context)

	require.Error(t, err)
	assert.Nil(t, blocks)
	assert.Contains(t, err.Error(), "template not found")
}

func TestConvertToSlackBlocks_WithSectionBlock(t *testing.T) {
	rawBlocks := []map[string]interface{}{
		{
			"type": "section",
			"text": map[string]interface{}{
				"type": "mrkdwn",
				"text": "Test section text",
			},
		},
	}

	blocks := convertToSlackBlocks(rawBlocks)

	require.NotNil(t, blocks)
	assert.Len(t, blocks, 1)
}

func TestConvertToSlackBlocks_WithContextBlock(t *testing.T) {
	rawBlocks := []map[string]interface{}{
		{
			"type": "context",
			"elements": []interface{}{
				map[string]interface{}{
					"type": "mrkdwn",
					"text": "Context element 1",
				},
				map[string]interface{}{
					"type": "plain_text",
					"text": "Context element 2",
				},
			},
		},
	}

	blocks := convertToSlackBlocks(rawBlocks)

	require.NotNil(t, blocks)
	assert.Len(t, blocks, 1)
}

func TestConvertToSlackBlocks_WithDividerBlock(t *testing.T) {
	rawBlocks := []map[string]interface{}{
		{
			"type": "divider",
		},
	}

	blocks := convertToSlackBlocks(rawBlocks)

	require.NotNil(t, blocks)
	assert.Len(t, blocks, 1)
}

func TestConvertToSlackBlocks_WithUnknownBlockType(t *testing.T) {
	rawBlocks := []map[string]interface{}{
		{
			"type": "unknown_type",
			"data": "some data",
		},
	}

	blocks := convertToSlackBlocks(rawBlocks)

	require.NotNil(t, blocks)
	assert.Empty(t, blocks, "Unknown block types should be skipped")
}

func TestConvertToSlackBlocks_WithMissingType(t *testing.T) {
	rawBlocks := []map[string]interface{}{
		{
			"text": "Block without type field",
		},
	}

	blocks := convertToSlackBlocks(rawBlocks)

	require.NotNil(t, blocks)
	assert.Empty(t, blocks, "Blocks without type should be skipped")
}

func TestParseSectionBlock_Valid(t *testing.T) {
	raw := map[string]interface{}{
		"text": map[string]interface{}{
			"type": "mrkdwn",
			"text": "Section text content",
		},
	}

	block := parseSectionBlock(raw)

	require.NotNil(t, block)
}

func TestParseSectionBlock_MissingText(t *testing.T) {
	raw := map[string]interface{}{
		"type": "section",
	}

	block := parseSectionBlock(raw)

	assert.Nil(t, block, "Should return nil when text is missing")
}

func TestParseSectionBlock_EmptyText(t *testing.T) {
	raw := map[string]interface{}{
		"text": map[string]interface{}{
			"type": "mrkdwn",
			"text": "",
		},
	}

	block := parseSectionBlock(raw)

	assert.Nil(t, block, "Should return nil when text is empty")
}

func TestParseContextBlock_Valid(t *testing.T) {
	raw := map[string]interface{}{
		"elements": []interface{}{
			map[string]interface{}{
				"type": "mrkdwn",
				"text": "Element 1",
			},
			map[string]interface{}{
				"type": "plain_text",
				"text": "Element 2",
			},
		},
	}

	block := parseContextBlock(raw)

	require.NotNil(t, block)
}

func TestParseContextBlock_MissingElements(t *testing.T) {
	raw := map[string]interface{}{
		"type": "context",
	}

	block := parseContextBlock(raw)

	assert.Nil(t, block, "Should return nil when elements are missing")
}

func TestParseContextBlock_EmptyElements(t *testing.T) {
	raw := map[string]interface{}{
		"elements": []interface{}{},
	}

	block := parseContextBlock(raw)

	assert.Nil(t, block, "Should return nil when elements array is empty")
}

func TestParseContextBlock_ElementsWithEmptyText(t *testing.T) {
	raw := map[string]interface{}{
		"elements": []interface{}{
			map[string]interface{}{
				"type": "mrkdwn",
				"text": "",
			},
			map[string]interface{}{
				"type": "plain_text",
				"text": "",
			},
		},
	}

	block := parseContextBlock(raw)

	assert.Nil(t, block, "Should return nil when all elements have empty text")
}

func TestParseContextBlock_MixedValidAndInvalidElements(t *testing.T) {
	raw := map[string]interface{}{
		"elements": []interface{}{
			map[string]interface{}{
				"type": "mrkdwn",
				"text": "",
			},
			map[string]interface{}{
				"type": "plain_text",
				"text": "Valid element",
			},
		},
	}

	block := parseContextBlock(raw)

	require.NotNil(t, block, "Should return block when at least one element is valid")
}
