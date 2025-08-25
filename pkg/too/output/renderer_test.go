package output

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatMultilineText(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		baseIndent  string
		columnWidth int
		expected    string
		description string
	}{
		{
			name:        "single line text unchanged",
			text:        "Single line todo",
			baseIndent:  "",
			columnWidth: 6,
			expected:    "Single line todo",
			description: "Single line text should remain unchanged",
		},
		{
			name:        "two line text with indentation",
			text:        "First line\nSecond line",
			baseIndent:  "",
			columnWidth: 6,
			expected:    "First line\n                 Second line",
			description: "Second line should be indented to align with first line text",
		},
		{
			name:        "multiple lines with base indent",
			text:        "Line one\nLine two\nLine three",
			baseIndent:  "    ",
			columnWidth: 6,
			expected:    "Line one\n                     Line two\n                     Line three",
			description: "All continuation lines should have consistent indentation",
		},
		{
			name:        "empty lines preserved",
			text:        "First\n\nThird",
			baseIndent:  "",
			columnWidth: 6,
			expected:    "First\n                 \n                 Third",
			description: "Empty lines should be preserved with proper indentation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatMultilineText(tt.text, tt.baseIndent, tt.columnWidth)
			assert.Equal(t, tt.expected, result, tt.description)

			// Verify that continuation lines have the expected indentation
			if strings.Contains(tt.text, "\n") {
				lines := strings.Split(result, "\n")
				if len(lines) > 1 {
					// Calculate expected indent: baseIndent + columnWidth + 11 (for " | X ")
					expectedIndentLen := len(tt.baseIndent) + tt.columnWidth + 11
					for i := 1; i < len(lines); i++ {
						if lines[i] != "" { // Skip empty lines for this check
							actualIndent := len(lines[i]) - len(strings.TrimLeft(lines[i], " "))
							assert.GreaterOrEqual(t, actualIndent, expectedIndentLen,
								"Line %d should have at least %d spaces of indentation", i+1, expectedIndentLen)
						}
					}
				}
			}
		})
	}
}
