package issue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatus_ToEmoji(t *testing.T) {
	tests := []struct {
		status   Status
		expected string
	}{
		{StatusResolved, "✅"},
		{StatusFiring, "🔥"},
	}

	for _, test := range tests {
		emoji := test.status.ToEmoji()
		assert.Equal(t, test.expected, emoji, "Incorrect emoji for status %s", test.status.String())
	}
}
