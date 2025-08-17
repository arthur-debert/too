package editor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessTodoText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
		{
			name:     "single line",
			input:    "Hello todo",
			expected: "Hello todo",
		},
		{
			name:     "single line with whitespace",
			input:    "  Hello todo  ",
			expected: "Hello todo",
		},
		{
			name:     "multiple lines with blank lines before and after",
			input:    "\n\nHello\nWorld\n\n\n",
			expected: "Hello\nWorld",
		},
		{
			name:     "preserve blank lines between text",
			input:    "\n\nHello\n\nWorld\n\n",
			expected: "Hello\n\nWorld",
		},
		{
			name:     "complex example from issue",
			input:    "\n\n  hi\n...............\n  the line above has many spaces , this one has trailing spaces.....................  \n\n\n",
			expected: "hi\n...............\nthe line above has many spaces , this one has trailing spaces.....................",
		},
		{
			name:     "all blank lines",
			input:    "\n\n\n\n",
			expected: "",
		},
		{
			name:     "lines with only spaces",
			input:    "   \n   \n   ",
			expected: "",
		},
		{
			name:     "preserve multiple blank lines between text",
			input:    "First\n\n\n\nLast",
			expected: "First\n\n\n\nLast",
		},
		{
			name:     "trim each line but preserve structure",
			input:    "  Line 1  \n  \n  Line 2  ",
			expected: "Line 1\n\nLine 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ProcessTodoText(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
