package workflow

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kubecano/cano-collector/config/workflow"
	"github.com/kubecano/cano-collector/pkg/core/event"
	"github.com/kubecano/cano-collector/pkg/core/issue"
	"github.com/kubecano/cano-collector/pkg/logger"
)

func TestNewEnrichmentProcessor(t *testing.T) {
	log := logger.NewLogger("debug", "test")
	processor := NewEnrichmentProcessor(log)

	assert.NotNil(t, processor)
	assert.NotNil(t, processor.logger)
}

func TestProcessWorkflowEnrichments_NilWorkflowEngine(t *testing.T) {
	log := logger.NewLogger("debug", "test")
	processor := NewEnrichmentProcessor(log)

	ctx := context.Background()
	issues := []*issue.Issue{{}}
	alertEvent := createTestAlertManagerEvent()

	err := processor.ProcessWorkflowEnrichments(ctx, issues, nil, alertEvent)
	assert.NoError(t, err)
}

func TestProcessWorkflowEnrichments_NoIssues(t *testing.T) {
	log := logger.NewLogger("debug", "test")
	processor := NewEnrichmentProcessor(log)
	mockEngine := &mockWorkflowEngine{}

	ctx := context.Background()
	var issues []*issue.Issue
	alertEvent := createTestAlertManagerEvent()

	err := processor.ProcessWorkflowEnrichments(ctx, issues, mockEngine, alertEvent)
	assert.NoError(t, err)
}

func TestProcessWorkflowEnrichments_NoMatchingWorkflows(t *testing.T) {
	log := logger.NewLogger("debug", "test")
	processor := NewEnrichmentProcessor(log)
	mockEngine := &mockWorkflowEngine{
		selectWorkflowsResult: []*workflow.WorkflowDefinition{},
	}

	ctx := context.Background()
	issues := []*issue.Issue{createTestIssue()}
	alertEvent := createTestAlertManagerEvent()

	err := processor.ProcessWorkflowEnrichments(ctx, issues, mockEngine, alertEvent)
	assert.NoError(t, err)
}

func TestProcessWorkflowEnrichments_WithEnrichments(t *testing.T) {
	log := logger.NewLogger("debug", "test")
	processor := NewEnrichmentProcessor(log)

	enrichment := issue.NewEnrichmentWithType(issue.EnrichmentTypeTextFile, "Test Title")
	enrichment.AddBlock(issue.NewMarkdownBlock("Test Content"))
	mockEngine := &mockWorkflowEngine{
		selectWorkflowsResult: []*workflow.WorkflowDefinition{
			{Name: "test-workflow"},
		},
		executeWorkflowsWithEnrichmentsResult: []issue.Enrichment{*enrichment},
	}

	ctx := context.Background()
	testIssue := createTestIssue()
	issues := []*issue.Issue{testIssue}
	alertEvent := createTestAlertManagerEvent()

	err := processor.ProcessWorkflowEnrichments(ctx, issues, mockEngine, alertEvent)
	require.NoError(t, err)

	// Verify enrichments were added to the issue
	assert.Len(t, testIssue.Enrichments, 1)
	assert.Equal(t, "Test Title", *testIssue.Enrichments[0].Title)
}

func TestProcessWorkflowEnrichments_ExecutionError(t *testing.T) {
	log := logger.NewLogger("debug", "test")
	processor := NewEnrichmentProcessor(log)

	mockEngine := &mockWorkflowEngine{
		selectWorkflowsResult: []*workflow.WorkflowDefinition{
			{Name: "test-workflow"},
		},
		executeWorkflowsWithEnrichmentsError: assert.AnError,
	}

	ctx := context.Background()
	issues := []*issue.Issue{createTestIssue()}
	alertEvent := createTestAlertManagerEvent()

	err := processor.ProcessWorkflowEnrichments(ctx, issues, mockEngine, alertEvent)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "assert.AnError general error for testing")
}

// Test helper functions

func createTestAlertManagerEvent() *event.AlertManagerEvent {
	templateData := createTestTemplateData("firing", "TestAlert", "warning", "default")
	return event.NewAlertManagerEvent(templateData)
}

func createTestIssue() *issue.Issue {
	return issue.NewIssue("Test Issue", "test-key")
}

// Mock workflow engine for testing
type mockWorkflowEngine struct {
	selectWorkflowsResult                 []*workflow.WorkflowDefinition
	executeWorkflowsWithEnrichmentsResult []issue.Enrichment
	executeWorkflowsWithEnrichmentsError  error
}

func (m *mockWorkflowEngine) SelectWorkflows(event event.WorkflowEvent) []*workflow.WorkflowDefinition {
	return m.selectWorkflowsResult
}

func (m *mockWorkflowEngine) ExecuteWorkflowWithEnrichments(ctx context.Context, wf *workflow.WorkflowDefinition, event event.WorkflowEvent) ([]issue.Enrichment, error) {
	return m.executeWorkflowsWithEnrichmentsResult, m.executeWorkflowsWithEnrichmentsError
}

func (m *mockWorkflowEngine) ExecuteWorkflowsWithEnrichments(ctx context.Context, workflows []*workflow.WorkflowDefinition, event event.WorkflowEvent) ([]issue.Enrichment, error) {
	return m.executeWorkflowsWithEnrichmentsResult, m.executeWorkflowsWithEnrichmentsError
}
