package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSingleTodo(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []*TodoItem
	}{
		{
			name:  "single todo with dash",
			input: "- Call aunt May",
			expected: []*TodoItem{
				{Text: "Call aunt May", Level: 0, Children: []*TodoItem{}},
			},
		},
		{
			name:  "single todo with asterisk",
			input: "* Pick up groceries",
			expected: []*TodoItem{
				{Text: "Pick up groceries", Level: 0, Children: []*TodoItem{}},
			},
		},
		{
			name:  "todo with leading spaces",
			input: "  - Trim the hedges",
			expected: []*TodoItem{
				{Text: "Trim the hedges", Level: 0, Children: []*TodoItem{}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := DefaultParseOptions()
			result := ParseMultipleTodos(tt.input, opts)
			assertTodosEqual(t, tt.expected, result)
		})
	}
}

func TestParseMultipleTodos(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []*TodoItem
	}{
		{
			name: "three simple todos",
			input: `- Call aunt May
- Pick up groceries
- Walk the dog`,
			expected: []*TodoItem{
				{Text: "Call aunt May", Level: 0, Children: []*TodoItem{}},
				{Text: "Pick up groceries", Level: 0, Children: []*TodoItem{}},
				{Text: "Walk the dog", Level: 0, Children: []*TodoItem{}},
			},
		},
		{
			name: "todos with blank lines between",
			input: `- First task

- Second task

- Third task`,
			expected: []*TodoItem{
				{Text: "First task", Level: 0, Children: []*TodoItem{}},
				{Text: "Second task", Level: 0, Children: []*TodoItem{}},
				{Text: "Third task", Level: 0, Children: []*TodoItem{}},
			},
		},
		{
			name: "mixed markers",
			input: `- First with dash
* Second with asterisk
- Third with dash`,
			expected: []*TodoItem{
				{Text: "First with dash", Level: 0, Children: []*TodoItem{}},
				{Text: "Second with asterisk", Level: 0, Children: []*TodoItem{}},
				{Text: "Third with dash", Level: 0, Children: []*TodoItem{}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := DefaultParseOptions()
			result := ParseMultipleTodos(tt.input, opts)
			assertTodosEqual(t, tt.expected, result)
		})
	}
}

func TestParseMultiLineTodos(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []*TodoItem
	}{
		{
			name: "single multi-line todo",
			input: `Call aunt May
Remember that she doesn't have a cellphone.
If needed, ask Robert for the number`,
			expected: []*TodoItem{},
		},
		{
			name: "todo with continuation lines",
			input: `- Call aunt May
  Remember she doesn't have a cellphone
  Ask Robert for the number if needed`,
			expected: []*TodoItem{
				{
					Text:     "Call aunt May\nRemember she doesn't have a cellphone\nAsk Robert for the number if needed",
					Level:    0,
					Children: []*TodoItem{},
				},
			},
		},
		{
			name: "multiple todos with continuations",
			input: `- Call aunt May
  Remember the birthday
- Pick up groceries
  bring milk, bread and butter
- Walk the dog`,
			expected: []*TodoItem{
				{
					Text:     "Call aunt May\nRemember the birthday",
					Level:    0,
					Children: []*TodoItem{},
				},
				{
					Text:     "Pick up groceries\nbring milk, bread and butter",
					Level:    0,
					Children: []*TodoItem{},
				},
				{
					Text:     "Walk the dog",
					Level:    0,
					Children: []*TodoItem{},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := DefaultParseOptions()
			result := ParseMultipleTodos(tt.input, opts)
			assertTodosEqual(t, tt.expected, result)
		})
	}
}

func TestParseNestedTodos(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []*TodoItem
	}{
		{
			name: "simple nested todos",
			input: `- Parent task
  - Child task 1
  - Child task 2`,
			expected: []*TodoItem{
				{
					Text:  "Parent task",
					Level: 0,
					Children: []*TodoItem{
						{Text: "Child task 1", Level: 1, Children: []*TodoItem{}},
						{Text: "Child task 2", Level: 1, Children: []*TodoItem{}},
					},
				},
			},
		},
		{
			name: "complex nested structure",
			input: `- Call aunt May
  - ask about her health
  - ask about her job
- Pick up groceries
  - bring milk, bread and butter
    Do not go to the expensive store
- Walk the dog`,
			expected: []*TodoItem{
				{
					Text:  "Call aunt May",
					Level: 0,
					Children: []*TodoItem{
						{Text: "ask about her health", Level: 1, Children: []*TodoItem{}},
						{Text: "ask about her job", Level: 1, Children: []*TodoItem{}},
					},
				},
				{
					Text:  "Pick up groceries",
					Level: 0,
					Children: []*TodoItem{
						{
							Text:     "bring milk, bread and butter\nDo not go to the expensive store",
							Level:    1,
							Children: []*TodoItem{},
						},
					},
				},
				{
					Text:     "Walk the dog",
					Level:    0,
					Children: []*TodoItem{},
				},
			},
		},
		{
			name: "deeply nested todos",
			input: `- Level 1
  - Level 2
    - Level 3
      - Level 4`,
			expected: []*TodoItem{
				{
					Text:  "Level 1",
					Level: 0,
					Children: []*TodoItem{
						{
							Text:  "Level 2",
							Level: 1,
							Children: []*TodoItem{
								{
									Text:  "Level 3",
									Level: 2,
									Children: []*TodoItem{
										{Text: "Level 4", Level: 3, Children: []*TodoItem{}},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "4-space indentation",
			input: `- Parent
    - Child with 4 spaces
    - Another child`,
			expected: []*TodoItem{
				{
					Text:  "Parent",
					Level: 0,
					Children: []*TodoItem{
						{Text: "Child with 4 spaces", Level: 2, Children: []*TodoItem{}},
						{Text: "Another child", Level: 2, Children: []*TodoItem{}},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := DefaultParseOptions()
			result := ParseMultipleTodos(tt.input, opts)
			assertTodosEqual(t, tt.expected, result)
		})
	}
}

func TestTabHandling(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []*TodoItem
	}{
		{
			name:  "tabs converted to spaces",
			input: "- Parent\n\t- Child with tab",
			expected: []*TodoItem{
				{
					Text:  "Parent",
					Level: 0,
					Children: []*TodoItem{
						{Text: "Child with tab", Level: 2, Children: []*TodoItem{}},
					},
				},
			},
		},
		{
			name: "mixed tabs and spaces",
			input: `- Level 1
	- Tab child
  - Space child`,
			expected: []*TodoItem{
				{
					Text:  "Level 1",
					Level: 0,
					Children: []*TodoItem{
						{Text: "Tab child", Level: 2, Children: []*TodoItem{}},
						{Text: "Space child", Level: 1, Children: []*TodoItem{}},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := DefaultParseOptions()
			result := ParseMultipleTodos(tt.input, opts)
			assertTodosEqual(t, tt.expected, result)
		})
	}
}

func TestEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []*TodoItem
	}{
		{
			name:     "empty input",
			input:    "",
			expected: []*TodoItem{},
		},
		{
			name:     "only whitespace",
			input:    "   \n\t\n   ",
			expected: []*TodoItem{},
		},
		{
			name:     "no todo markers",
			input:    "This is just text\nWithout any markers",
			expected: []*TodoItem{},
		},
		{
			name: "invalid markers",
			input: `+ Not a valid marker
# Also not valid
- This one is valid`,
			expected: []*TodoItem{
				{Text: "This one is valid", Level: 0, Children: []*TodoItem{}},
			},
		},
		{
			name: "marker without space",
			input: `-No space after marker
- Proper todo`,
			expected: []*TodoItem{
				{Text: "Proper todo", Level: 0, Children: []*TodoItem{}},
			},
		},
		{
			name: "inconsistent indentation",
			input: `- Parent
   - Three space child
  - Two space child
    - Four space grandchild`,
			expected: []*TodoItem{
				{
					Text:  "Parent",
					Level: 0,
					Children: []*TodoItem{
						{Text: "Three space child", Level: 1, Children: []*TodoItem{}},
						{
							Text:  "Two space child",
							Level: 1,
							Children: []*TodoItem{
								{Text: "Four space grandchild", Level: 2, Children: []*TodoItem{}},
							},
						},
					},
				},
			},
		},
		{
			name: "orphaned nested todo",
			input: `    - This has no parent
- Now a parent
  - With a child`,
			expected: []*TodoItem{
				{Text: "This has no parent", Level: 0, Children: []*TodoItem{}},
				{
					Text:  "Now a parent",
					Level: 0,
					Children: []*TodoItem{
						{Text: "With a child", Level: 1, Children: []*TodoItem{}},
					},
				},
			},
		},
		{
			name: "multiline with nested",
			input: `- Parent todo
  This continues on next line
  - Child todo
    Also multiline
  And this continues the child`,
			expected: []*TodoItem{
				{
					Text:  "Parent todo\nThis continues on next line",
					Level: 0,
					Children: []*TodoItem{
						{
							Text:     "Child todo\nAlso multiline\nAnd this continues the child",
							Level:    1,
							Children: []*TodoItem{},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := DefaultParseOptions()
			result := ParseMultipleTodos(tt.input, opts)
			assertTodosEqual(t, tt.expected, result)
		})
	}
}

func TestFlatten(t *testing.T) {
	tests := []struct {
		name     string
		todos    []*TodoItem
		expected []string
	}{
		{
			name: "simple list",
			todos: []*TodoItem{
				{Text: "First"},
				{Text: "Second"},
				{Text: "Third"},
			},
			expected: []string{"First", "Second", "Third"},
		},
		{
			name: "nested list",
			todos: []*TodoItem{
				{
					Text: "Parent 1",
					Children: []*TodoItem{
						{Text: "Child 1.1"},
						{Text: "Child 1.2"},
					},
				},
				{
					Text: "Parent 2",
					Children: []*TodoItem{
						{Text: "Child 2.1"},
					},
				},
			},
			expected: []string{
				"Parent 1",
				"1.Child 1.1",
				"1.Child 1.2",
				"Parent 2",
				"2.Child 2.1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Flatten(tt.todos)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function to compare TodoItems
func assertTodosEqual(t *testing.T, expected, actual []*TodoItem) {
	t.Helper()

	assert.Equal(t, len(expected), len(actual), "Number of todos mismatch")

	for i := range expected {
		assertTodoEqual(t, expected[i], actual[i], "")
	}
}

func assertTodoEqual(t *testing.T, expected, actual *TodoItem, path string) {
	t.Helper()

	if path == "" {
		path = "root"
	}

	assert.Equal(t, expected.Text, actual.Text, "Text mismatch at %s", path)
	assert.Equal(t, expected.Level, actual.Level, "Level mismatch at %s", path)
	assert.Equal(t, len(expected.Children), len(actual.Children), "Number of children mismatch at %s", path)

	for i := range expected.Children {
		childPath := path + "[" + string(rune('0'+i)) + "]"
		assertTodoEqual(t, expected.Children[i], actual.Children[i], childPath)
	}
}
