package issue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSeverity_ToEmoji(t *testing.T) {
	tests := []struct {
		severity Severity
		expected string
	}{
		{SeverityDebug, "🔵"}, // Blue
		{SeverityInfo, "⚪️"}, // White
		{SeverityLow, "🟡"},   // Yellow
		{SeverityHigh, "🔴"},  // Red
	}

	for _, test := range tests {
		emoji := test.severity.ToEmoji()
		assert.Equal(t, test.expected, emoji, "Incorrect emoji for severity %s", test.severity.String())
	}
}
