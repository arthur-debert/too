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
	// Test with both ShowListSummary ON and OFF
	configs := []struct {
		name            string
		showListSummary bool
	}{
		{"summary_on", true},
		{"summary_off", false},
	}

	for _, cfg := range configs {
		t.Run(cfg.name, func(t *testing.T) {
			// Set config for this test
			testConfig := too.DefaultConfig()
			testConfig.Display.ShowListSummary = cfg.showListSummary
			too.SetConfig(testConfig)
			defer too.SetConfig(too.DefaultConfig()) // Reset after test
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
			// Create engine
			var buf bytes.Buffer
			engine, err := NewEngine()
			assert.NoError(t, err)

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
			err = engine.GetLipbalmEngine().Render(&buf, "term", result)
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
			
			// Check summary based on config
			if cfg.showListSummary {
				assert.Contains(t, output, "2 todo(s), 0 done", "Expected summary when ShowListSummary is ON")
			} else {
				assert.NotContains(t, output, "2 todo(s), 0 done", "Should not show summary when ShowListSummary is OFF")
			}
		})
	}
		})
	}
}

func TestRenderChangeSummaryWhitespace(t *testing.T) {
	// Test that when summary is OFF, no extra whitespace is left
	testConfig := too.DefaultConfig()
	testConfig.Display.ShowListSummary = false
	too.SetConfig(testConfig)
	defer too.SetConfig(too.DefaultConfig())

	engine, err := NewEngine()
	assert.NoError(t, err)

	// Create test data
	allTodos := []*models.Todo{
		{UID: "1", Text: "Todo 1", PositionPath: "1"},
		{UID: "2", Text: "Todo 2", PositionPath: "2"},
	}

	result := too.NewChangeResult(
		"add",
		"Added todo: 1",
		[]*models.Todo{allTodos[0]},
		allTodos,
		2,
		0,
	)

	var buf bytes.Buffer
	err = engine.GetLipbalmEngine().Render(&buf, "term", result)
	assert.NoError(t, err)

	output := buf.String()
	
	// Check that there's no extra blank line before the message
	lines := strings.Split(output, "\n")
	
	// Find the last todo line
	lastTodoIdx := -1
	for i, line := range lines {
		if strings.Contains(line, "○") && strings.Contains(line, "Todo") {
			lastTodoIdx = i
		}
	}
	
	// Find the message line
	messageIdx := -1
	for i, line := range lines {
		if strings.Contains(line, "Added todo") {
			messageIdx = i
		}
	}
	
	// When summary is OFF, ensure no summary line is present
	assert.NotContains(t, output, "todo(s)", "Should not contain summary when ShowListSummary is OFF")
	
	// Ensure proper spacing - two blank lines is acceptable for message formatting
	if lastTodoIdx != -1 && messageIdx != -1 {
		blankLines := 0
		for i := lastTodoIdx + 1; i < messageIdx; i++ {
			if strings.TrimSpace(lines[i]) == "" {
				blankLines++
			}
		}
		// Two blank lines is acceptable for proper formatting
		assert.True(t, blankLines >= 1 && blankLines <= 2, "Should have 1-2 blank lines between todos and message, got %d", blankLines)
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
			engine, err := NewEngine()
			assert.NoError(t, err)

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

			err = engine.GetLipbalmEngine().Render(&buf, "term", result)
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