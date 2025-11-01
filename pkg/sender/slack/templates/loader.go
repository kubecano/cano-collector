package templates

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"text/template"

	slackapi "github.com/slack-go/slack"
)

//go:embed *.tmpl
var templateFS embed.FS

// TemplateLoader manages Slack message templates
type TemplateLoader struct {
	templates map[string]*template.Template
}

// NewTemplateLoader creates a new template loader and parses all embedded templates
func NewTemplateLoader() (*TemplateLoader, error) {
	loader := &TemplateLoader{
		templates: make(map[string]*template.Template),
	}

	// Load all template files
	files := []string{
		"header.tmpl",
		"context_bar.tmpl",
		"crash_info.tmpl",
		"description.tmpl",
		"links.tmpl",
		"file_enrichment.tmpl",
	}

	for _, file := range files {
		tmpl, err := template.ParseFS(templateFS, file)
		if err != nil {
			return nil, fmt.Errorf("failed to parse template %s: %w", file, err)
		}
		loader.templates[file] = tmpl
	}

	return loader, nil
}

// RenderToBlocks renders a template with context and returns Slack blocks
func (l *TemplateLoader) RenderToBlocks(templateName string, context interface{}) ([]slackapi.Block, error) {
	tmpl, exists := l.templates[templateName]
	if !exists {
		return nil, fmt.Errorf("template not found: %s", templateName)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, context); err != nil {
		return nil, fmt.Errorf("failed to execute template %s: %w", templateName, err)
	}

	// Parse JSON array of blocks
	var rawBlocks []map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &rawBlocks); err != nil {
		return nil, fmt.Errorf("failed to parse template output as JSON for %s: %w", templateName, err)
	}

	// Convert to Slack blocks
	blocks, err := convertToSlackBlocks(rawBlocks)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to Slack blocks for %s: %w", templateName, err)
	}

	return blocks, nil
}

// convertToSlackBlocks converts generic JSON blocks to slack.Block types
func convertToSlackBlocks(rawBlocks []map[string]interface{}) ([]slackapi.Block, error) {
	blocks := make([]slackapi.Block, 0, len(rawBlocks))

	for _, raw := range rawBlocks {
		blockType, ok := raw["type"].(string)
		if !ok {
			continue // Skip blocks without type
		}

		switch blockType {
		case "section":
			block := parseSectionBlock(raw)
			if block != nil {
				blocks = append(blocks, block)
			}
		case "context":
			block := parseContextBlock(raw)
			if block != nil {
				blocks = append(blocks, block)
			}
		case "divider":
			blocks = append(blocks, slackapi.NewDividerBlock())
		default:
			// Unknown block type, skip
			continue
		}
	}

	return blocks, nil
}

// parseSectionBlock parses a section block from raw JSON
func parseSectionBlock(raw map[string]interface{}) *slackapi.SectionBlock {
	textObj, ok := raw["text"].(map[string]interface{})
	if !ok {
		return nil
	}

	textType, _ := textObj["type"].(string)
	textContent, _ := textObj["text"].(string)

	if textContent == "" {
		return nil
	}

	textBlockObj := slackapi.NewTextBlockObject(textType, textContent, false, false)
	return slackapi.NewSectionBlock(textBlockObj, nil, nil)
}

// parseContextBlock parses a context block from raw JSON
func parseContextBlock(raw map[string]interface{}) *slackapi.ContextBlock {
	elementsRaw, ok := raw["elements"].([]interface{})
	if !ok || len(elementsRaw) == 0 {
		return nil
	}

	elements := make([]slackapi.MixedElement, 0, len(elementsRaw))
	for _, elemRaw := range elementsRaw {
		elemMap, ok := elemRaw.(map[string]interface{})
		if !ok {
			continue
		}

		elemType, _ := elemMap["type"].(string)
		elemText, _ := elemMap["text"].(string)

		if elemText == "" {
			continue
		}

		textObj := slackapi.NewTextBlockObject(elemType, elemText, false, false)
		elements = append(elements, textObj)
	}

	if len(elements) == 0 {
		return nil
	}

	return slackapi.NewContextBlock("", elements...)
}
