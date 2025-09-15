package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/arthur-debert/too/pkg/too"
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/stretchr/testify/assert"
)

func TestRenderChange(t *testing.T) {
	tests := []struct {
		name           string
		command        string
		affectedTodos  []*models.Todo
		expectedInMsg  []string
		notExpectedMsg []string
	}{
		{
			name:    "single add",
			command: "add",
			affectedTodos: []*models.Todo{
				{UID: "123", Text: "New todo", PositionPath: "1"},
			},
			expectedInMsg: []string{"Added todo: 1"},
		},
		{
			name:    "multiple complete",
			command: "completed",
			affectedTodos: []*models.Todo{
				{UID: "123", Text: "Todo 1", PositionPath: "1"},
				{UID: "456", Text: "Todo 2", PositionPath: "2"},
			},
			expectedInMsg: []string{"Completed todos: 1, 2"},
		},
		{
			name:    "single modified",
			command: "modified",
			affectedTodos: []*models.Todo{
				{UID: "123", Text: "Modified todo", PositionPath: "3.1"},
			},
			expectedInMsg: []string{"Modified todo: 3.1"},
		},
		{
			name:    "clean with no todos",
			command: "cleaned",
			affectedTodos: []*models.Todo{},
			expectedInMsg: []string{"Cleaned: no todos affected"},
		},
		{
			name:    "reopened multiple",
			command: "reopened",
			affectedTodos: []*models.Todo{
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
			renderer := NewRenderer(&buf)

			// Create change result
			allTodos := []*models.Todo{
				{UID: "1", Text: "Todo 1"},
				{UID: "2", Text: "Todo 2"},
			}
			// Generate expected message based on command
			var message string
			switch tt.command {
			case "add":
				if len(tt.affectedTodos) > 0 {
					message = "Added todo: " + tt.affectedTodos[0].PositionPath
				}
			case "completed":
				if len(tt.affectedTodos) > 0 {
					positions := make([]string, len(tt.affectedTodos))
					for i, todo := range tt.affectedTodos {
						positions[i] = todo.PositionPath
					}
					message = "Completed todos: " + strings.Join(positions, ", ")
				}
			case "modified":
				if len(tt.affectedTodos) > 0 {
					message = "Modified todo: " + tt.affectedTodos[0].PositionPath
				}
			case "cleaned":
				if len(tt.affectedTodos) == 0 {
					message = "Cleaned: no todos affected"
				}
			case "reopened":
				if len(tt.affectedTodos) > 0 {
					positions := make([]string, len(tt.affectedTodos))
					for i, todo := range tt.affectedTodos {
						positions[i] = todo.PositionPath
					}
					message = "Reopened todos: " + strings.Join(positions, ", ")
				}
			}

			result := too.NewChangeResult(
				tt.command,
				message,
				tt.affectedTodos,
				allTodos,
				2,
				0,
			)

			// Render
			err := renderer.RenderChange(result)
			assert.NoError(t, err)

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
			renderer := NewRenderer(&buf)

			// Generate expected message based on command
			var message string
			switch tt.command {
			case "add":
				message = "Added todo: 1"
			case "completed":
				message = "Completed todo: 1"
			case "modified":
				message = "Modified todo: 1"
			case "reopened":
				message = "Reopened todo: 1"
			case "cleaned":
				message = "Cleaned todo: 1"
			case "moved":
				message = "Moved todo: 1"
			}

			result := too.NewChangeResult(
				tt.command,
				message,
				[]*models.Todo{{UID: "1", Text: "Test", PositionPath: "1"}},
				[]*models.Todo{{UID: "1", Text: "Test"}},
				1,
				0,
			)

			err := renderer.RenderChange(result)
			assert.NoError(t, err)

			output := buf.String()
			// The message should be present regardless of style tags when useColor is false
			expectedMessage := strings.Title(tt.command)
			if !strings.HasSuffix(tt.command, "ed") {
				expectedMessage = expectedMessage + "ed"
			}
			assert.Contains(t, output, expectedMessage)
		})
	}
}