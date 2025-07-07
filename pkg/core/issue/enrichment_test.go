package issue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEnrichment(t *testing.T) {
	enrichment := NewEnrichment()

	assert.NotNil(t, enrichment)
	assert.NotNil(t, enrichment.Blocks)
	assert.NotNil(t, enrichment.Annotations)
	assert.Empty(t, enrichment.Blocks)
	assert.Empty(t, enrichment.Annotations)
	assert.Nil(t, enrichment.EnrichmentType)
	assert.Nil(t, enrichment.Title)
}

func TestNewEnrichmentWithType(t *testing.T) {
	enrichmentType := EnrichmentTypeAlertLabels
	title := "Test Enrichment"

	enrichment := NewEnrichmentWithType(enrichmentType, title)

	assert.NotNil(t, enrichment)
	assert.NotNil(t, enrichment.Blocks)
	assert.NotNil(t, enrichment.Annotations)
	assert.Empty(t, enrichment.Blocks)
	assert.Empty(t, enrichment.Annotations)
	assert.NotNil(t, enrichment.EnrichmentType)
	assert.Equal(t, enrichmentType, *enrichment.EnrichmentType)
	assert.NotNil(t, enrichment.Title)
	assert.Equal(t, title, *enrichment.Title)
}

func TestEnrichment_AddBlock(t *testing.T) {
	enrichment := NewEnrichment()
	block := NewMarkdownBlock("Test markdown")

	enrichment.AddBlock(block)

	assert.Len(t, enrichment.Blocks, 1)
	assert.Equal(t, block, enrichment.Blocks[0])
}

func TestEnrichment_AddAnnotation(t *testing.T) {
	enrichment := NewEnrichment()
	key := "test_key"
	value := "test_value"

	enrichment.AddAnnotation(key, value)

	assert.Len(t, enrichment.Annotations, 1)
	assert.Equal(t, value, enrichment.Annotations[key])
}

func TestEnrichment_AddAnnotation_NilAnnotations(t *testing.T) {
	enrichment := &Enrichment{
		Blocks:      make([]BaseBlock, 0),
		Annotations: nil,
	}
	key := "test_key"
	value := "test_value"

	enrichment.AddAnnotation(key, value)

	assert.NotNil(t, enrichment.Annotations)
	assert.Len(t, enrichment.Annotations, 1)
	assert.Equal(t, value, enrichment.Annotations[key])
}

func TestEnrichment_MultipleBlocks(t *testing.T) {
	enrichment := NewEnrichment()
	block1 := NewMarkdownBlock("First block")
	block2 := NewTableBlock([]string{"Name"}, [][]string{{"John"}}, "Test Table", TableBlockFormatHorizontal)

	enrichment.AddBlock(block1)
	enrichment.AddBlock(block2)

	assert.Len(t, enrichment.Blocks, 2)
	assert.Equal(t, block1, enrichment.Blocks[0])
	assert.Equal(t, block2, enrichment.Blocks[1])
}

func TestEnrichment_MultipleAnnotations(t *testing.T) {
	enrichment := NewEnrichment()

	enrichment.AddAnnotation("key1", "value1")
	enrichment.AddAnnotation("key2", "value2")

	assert.Len(t, enrichment.Annotations, 2)
	assert.Equal(t, "value1", enrichment.Annotations["key1"])
	assert.Equal(t, "value2", enrichment.Annotations["key2"])
}

func TestNewLink(t *testing.T) {
	text := "Test Link"
	url := "https://example.com"
	linkType := LinkTypeGeneral

	link := NewLink(text, url, linkType)

	assert.Equal(t, text, link.Text)
	assert.Equal(t, url, link.URL)
	assert.Equal(t, linkType, link.Type)
}

func TestLinkType_String(t *testing.T) {
	tests := []struct {
		linkType LinkType
		expected string
	}{
		{LinkTypeGeneral, "GENERAL"},
		{LinkTypePrometheusGenerator, "PROMETHEUS_GENERATOR"},
		{LinkTypeInvestigate, "INVESTIGATE"},
		{LinkTypeSilence, "SILENCE"},
		{LinkType(999), "UNKNOWN"}, // default case
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, test.linkType.String())
	}
}
