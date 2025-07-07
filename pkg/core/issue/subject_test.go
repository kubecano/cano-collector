package issue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubject_FormatWithEmoji(t *testing.T) {
	// Test with namespace
	subjectWithNS := &Subject{
		Name:        "test-pod",
		Namespace:   "default",
		SubjectType: SubjectTypePod,
	}

	result := subjectWithNS.FormatWithEmoji()
	expected := "🎯 Subject: default/test-pod (POD)"
	assert.Equal(t, expected, result)

	// Test without namespace
	subjectWithoutNS := &Subject{
		Name:        "test-node",
		SubjectType: SubjectTypeNode,
	}

	result = subjectWithoutNS.FormatWithEmoji()
	expected = "🎯 Subject: test-node (NODE)"
	assert.Equal(t, expected, result)
}
