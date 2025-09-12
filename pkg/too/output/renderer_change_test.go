package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/arthur-debert/too/pkg/too"
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderChange(t *testing.T) {
	tests := []struct {
		name           string
		command        string
		affectedTodos  []*models.IDMTodo
		expectedInMsg  []string
		notExpectedMsg []string
	}{
		{
			name:    "single add",
			command: "add",
			affectedTodos: []*models.IDMTodo{
				{UID: "123", Text: "New todo", PositionPath: "1"},
			},
			expectedInMsg: []string{"Added todo: 1"},
		},
		{
			name:    "multiple complete",
			command: "completed",
			affectedTodos: []*models.IDMTodo{
				{UID: "123", Text: "Todo 1", PositionPath: "1"},
				{UID: "456", Text: "Todo 2", PositionPath: "2"},
			},
			expectedInMsg: []string{"Completed todos: 1, 2"},
		},
		{
			name:    "single modified",
			command: "modified",
			affectedTodos: []*models.IDMTodo{
				{UID: "123", Text: "Modified todo", PositionPath: "3.1"},
			},
			expectedInMsg: []string{"Modified todo: 3.1"},
		},
		{
			name:    "clean with no todos",
			command: "cleaned",
			affectedTodos: []*models.IDMTodo{},
			expectedInMsg: []string{"Cleaned: no todos affected"},
		},
		{
			name:    "reopened multiple",
			command: "reopened",
			affectedTodos: []*models.IDMTodo{
				{UID: "123", Text: "Todo 1", PositionPath: "1"},
				{UID: "456", Text: "Todo 2", PositionPath: "2.1"},
				{UID: "789", Text: "Todo 3", PositionPath: "2.2"},
			},
			expectedInMsg: []string{"Reopened todos: 1, 2.1, 2.2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create renderer
			var buf bytes.Buffer
			renderer, err := NewLipbamlRenderer(&buf, false)
			require.NoError(t, err)

			// Create change result
			allTodos := []*models.IDMTodo{
				{UID: "1", Text: "Todo 1"},
				{UID: "2", Text: "Todo 2"},
			}
			result := too.NewChangeResult(
				tt.command,
				tt.affectedTodos,
				allTodos,
				2,
				0,
			)

			// Render
			err = renderer.RenderChange(result)
			require.NoError(t, err)

			// Check output
			output := buf.String()
			
			// Message should be at the bottom (after todo list)
			lines := strings.Split(strings.TrimSpace(output), "\n")
			lastLine := lines[len(lines)-1]

			// Check expected content
			for _, expected := range tt.expectedInMsg {
				assert.Contains(t, lastLine, expected, "Expected '%s' in message line: %s", expected, lastLine)
			}

			// Check not expected content
			for _, notExpected := range tt.notExpectedMsg {
				assert.NotContains(t, output, notExpected, "Did not expect '%s' in output", notExpected)
			}
			
			// Should show count
			assert.Contains(t, output, "2 todo(s), 0 done")
		})
	}
}

func TestRenderChangeMessageTypes(t *testing.T) {
	tests := []struct {
		command      string
		expectStyle  string
	}{
		{"add", "success"},
		{"modified", "info"}, 
		{"reopened", "warning"},
		{"completed", "success"},
		{"cleaned", "success"},  // or warning if no todos
		{"moved", "success"},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			var buf bytes.Buffer
			renderer, err := NewLipbamlRenderer(&buf, false)
			require.NoError(t, err)

			result := too.NewChangeResult(
				tt.command,
				[]*models.IDMTodo{{UID: "1", Text: "Test", PositionPath: "1"}},
				[]*models.IDMTodo{{UID: "1", Text: "Test"}},
				1,
				0,
			)

			err = renderer.RenderChange(result)
			require.NoError(t, err)

			output := buf.String()
			// The message should be wrapped in the appropriate style tag
			assert.Contains(t, output, "<"+tt.expectStyle+">")
		})
	}
}