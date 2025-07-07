package issue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIssue_GenerateFingerprint(t *testing.T) {
	tests := []struct {
		name     string
		issue    *Issue
		expected string // We'll check that fingerprint is generated and consistent
	}{
		{
			name: "Basic fingerprint generation",
			issue: &Issue{
				AggregationKey: "test-alert",
				Source:         SourcePrometheus,
				Subject: &Subject{
					Name:        "test-pod",
					SubjectType: SubjectTypePod,
					Namespace:   "default",
					Node:        "node1",
				},
			},
		},
		{
			name: "Fingerprint with empty subject",
			issue: &Issue{
				AggregationKey: "test-alert",
				Source:         SourcePrometheus,
				Subject:        nil,
			},
		},
		{
			name: "Fingerprint with partial subject",
			issue: &Issue{
				AggregationKey: "test-alert",
				Source:         SourcePrometheus,
				Subject: &Subject{
					Name:        "test-deployment",
					SubjectType: SubjectTypeDeployment,
					Namespace:   "kube-system",
					// No node specified
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fingerprint := tt.issue.generateFingerprint()

			// Check that fingerprint is generated and not empty
			assert.NotEmpty(t, fingerprint)

			// Check that fingerprint is consistent (same input produces same output)
			fingerprint2 := tt.issue.generateFingerprint()
			assert.Equal(t, fingerprint, fingerprint2)

			// Check that fingerprint is a valid SHA256 hash (64 hex characters)
			assert.Len(t, fingerprint, 64)
			assert.Regexp(t, "^[a-f0-9]+$", fingerprint)
		})
	}
}

func TestIssue_SetFingerprint(t *testing.T) {
	issue := NewIssue("Test Issue", "test-alert")
	originalFingerprint := issue.Fingerprint

	// Test setting a custom fingerprint
	customFingerprint := "custom-fingerprint-12345"
	issue.SetFingerprint(customFingerprint)

	assert.Equal(t, customFingerprint, issue.Fingerprint)
	assert.NotEqual(t, originalFingerprint, issue.Fingerprint)

	// Test setting empty fingerprint (should not change)
	issue.SetFingerprint("")
	assert.Equal(t, customFingerprint, issue.Fingerprint)
}

func TestNewIssue_GeneratesFingerprint(t *testing.T) {
	issue := NewIssue("Test Issue", "test-alert")

	// Check that fingerprint is generated automatically
	assert.NotEmpty(t, issue.Fingerprint)
	assert.Len(t, issue.Fingerprint, 64)
	assert.Regexp(t, "^[a-f0-9]+$", issue.Fingerprint)
}

func TestIssue_SetSubject_RegeneratesFingerprint(t *testing.T) {
	issue := NewIssue("Test Issue", "test-alert")
	originalFingerprint := issue.Fingerprint

	// Set a new subject
	subject := NewSubject("test-pod", SubjectTypePod)
	subject.Namespace = "default"
	issue.SetSubject(subject)

	// Fingerprint should be regenerated
	assert.NotEqual(t, originalFingerprint, issue.Fingerprint)
	assert.NotEmpty(t, issue.Fingerprint)

	// Setting the same subject should not change fingerprint
	currentFingerprint := issue.Fingerprint
	issue.SetSubject(subject)
	assert.Equal(t, currentFingerprint, issue.Fingerprint)
}

func TestIssue_FingerprintConsistency(t *testing.T) {
	// Test that same parameters produce same fingerprint
	issue1 := NewIssue("Test Issue", "test-alert")
	subject1 := NewSubject("test-pod", SubjectTypePod)
	subject1.Namespace = "default"
	subject1.Node = "node1"
	issue1.SetSubject(subject1)
	issue1.Source = SourcePrometheus

	issue2 := NewIssue("Test Issue", "test-alert")
	subject2 := NewSubject("test-pod", SubjectTypePod)
	subject2.Namespace = "default"
	subject2.Node = "node1"
	issue2.SetSubject(subject2)
	issue2.Source = SourcePrometheus

	assert.Equal(t, issue1.Fingerprint, issue2.Fingerprint)
}

func TestIssue_FingerprintDifferences(t *testing.T) {
	// Test that different parameters produce different fingerprints
	baseIssue := NewIssue("Test Issue", "test-alert")
	baseSubject := NewSubject("test-pod", SubjectTypePod)
	baseSubject.Namespace = "default"
	baseIssue.SetSubject(baseSubject)
	baseIssue.Source = SourcePrometheus

	// Different aggregation key
	issue1 := NewIssue("Test Issue", "different-alert")
	issue1.SetSubject(baseSubject)
	issue1.Source = SourcePrometheus
	assert.NotEqual(t, baseIssue.Fingerprint, issue1.Fingerprint)

	// Different subject name
	issue2 := NewIssue("Test Issue", "test-alert")
	subject2 := NewSubject("different-pod", SubjectTypePod)
	subject2.Namespace = "default"
	issue2.SetSubject(subject2)
	issue2.Source = SourcePrometheus
	assert.NotEqual(t, baseIssue.Fingerprint, issue2.Fingerprint)

	// Different namespace
	issue3 := NewIssue("Test Issue", "test-alert")
	subject3 := NewSubject("test-pod", SubjectTypePod)
	subject3.Namespace = "different-namespace"
	issue3.SetSubject(subject3)
	issue3.Source = SourcePrometheus
	assert.NotEqual(t, baseIssue.Fingerprint, issue3.Fingerprint)
}
