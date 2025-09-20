package too

import (
	"strings"
	"testing"

	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFreeTextAddressing(t *testing.T) {
	tests := []struct {
		name        string
		setupTodos  []testutil.TodoSpec
		reference   string
		expectError bool
		errorMsg    string
	}{
		{
			name: "exact match single result",
			setupTodos: []testutil.TodoSpec{
				{Text: "Buy milk"},
				{Text: "Walk the dog"},
				{Text: "Write tests"},
			},
			reference:   "Walk the dog",
			expectError: false,
		},
		{
			name: "partial match single result",
			setupTodos: []testutil.TodoSpec{
				{Text: "Buy milk"},
				{Text: "Walk the dog"},
				{Text: "Write tests"},
			},
			reference:   "dog",
			expectError: false,
		},
		{
			name: "no match",
			setupTodos: []testutil.TodoSpec{
				{Text: "Buy milk"},
				{Text: "Walk the dog"},
				{Text: "Write tests"},
			},
			reference:   "cat",
			expectError: true,
			errorMsg:    "no todo found matching 'cat'",
		},
		{
			name: "multiple matches",
			setupTodos: []testutil.TodoSpec{
				{Text: "Buy milk"},
				{Text: "Buy bread"},
				{Text: "Walk the dog"},
			},
			reference:   "Buy",
			expectError: true,
			errorMsg:    "Multiple todos found matching 'Buy'",
		},
		{
			name: "exact match resolves ambiguity",
			setupTodos: []testutil.TodoSpec{
				{Text: "Test"},
				{Text: "Test the feature"},
				{Text: "Testing framework"},
			},
			reference:   "Test",
			expectError: false, // Should match the exact "Test" todo
		},
		{
			name: "numeric position path still works",
			setupTodos: []testutil.TodoSpec{
				{Text: "First todo"},
				{Text: "Second todo"},
			},
			reference:   "1",
			expectError: false,
		},
		{
			name: "completed todo position path",
			setupTodos: []testutil.TodoSpec{
				{Text: "First todo", Complete: true},
				{Text: "Second todo"},
			},
			reference:   "c1",
			expectError: false,
		},
		{
			name: "invalid position path falls back to search",
			setupTodos: []testutil.TodoSpec{
				{Text: "Todo number 99"},
			},
			reference:   "99",
			expectError: true, // Position 99 doesn't exist
			errorMsg:    "could not resolve '99'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter, dbPath := testutil.CreateStoreWithSpecs(t, tt.setupTodos...)
			defer adapter.Close()

			engine, err := NewNanoEngine(dbPath)
			require.NoError(t, err)
			defer engine.Close()

			uuid, err := engine.ResolveReference(tt.reference)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, uuid)
				
				// Verify we got the right todo
				todo, err := engine.GetTodoByUID(uuid)
				assert.NoError(t, err)
				
				// For numeric references, just check it exists
				if looksLikePositionPath(tt.reference) {
					assert.NotNil(t, todo)
				} else {
					// For text references, verify the text matches
					assert.True(t, 
						strings.Contains(strings.ToLower(todo.Text), strings.ToLower(tt.reference)) ||
						strings.ToLower(todo.Text) == strings.ToLower(tt.reference),
						"Expected todo text '%s' to contain '%s'", todo.Text, tt.reference)
				}
			}
		})
	}
}

func TestFreeTextInCommands(t *testing.T) {
	// Test that free text works with actual commands
	adapter, dbPath := testutil.CreateStoreWithSpecs(t,
		testutil.TodoSpec{Text: "Buy groceries"},
		testutil.TodoSpec{Text: "Walk the dog"},
		testutil.TodoSpec{Text: "Write documentation"},
	)
	defer adapter.Close()

	t.Run("complete by text", func(t *testing.T) {
		opts := map[string]interface{}{
			"collectionPath": dbPath,
		}
		
		result, err := ExecuteUnifiedCommand("complete", []string{"Walk the dog"}, opts)
		require.NoError(t, err)
		
		// Find the completed todo
		var found bool
		for _, todo := range result.AllTodos {
			if todo.Text == "Walk the dog" {
				assert.Equal(t, models.StatusDone, todo.GetStatus())
				found = true
				break
			}
		}
		assert.True(t, found, "Todo 'Walk the dog' should be marked as done")
	})

	t.Run("edit by partial text", func(t *testing.T) {
		opts := map[string]interface{}{
			"collectionPath": dbPath,
		}
		
		result, err := ExecuteUnifiedCommand("edit", []string{"groceries", "Buy milk and bread"}, opts)
		require.NoError(t, err)
		
		// Find the edited todo
		var found bool
		for _, todo := range result.AllTodos {
			if todo.Text == "Buy milk and bread" {
				found = true
				break
			}
		}
		assert.True(t, found, "Todo should be edited to 'Buy milk and bread'")
	})

	t.Run("ambiguous text shows suggestions", func(t *testing.T) {
		// Add another "Write" todo
		engine, _ := NewNanoEngine(dbPath)
		engine.Add("Write tests", nil)
		engine.Close()

		opts := map[string]interface{}{
			"collectionPath": dbPath,
		}
		
		_, err := ExecuteUnifiedCommand("complete", []string{"Write"}, opts)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Multiple todos found")
		assert.Contains(t, err.Error(), "Write documentation")
		assert.Contains(t, err.Error(), "Write tests")
	})
}

func TestLooksLikePositionPath(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		// Valid position paths
		{"1", true},
		{"123", true},
		{"1.2", true},
		{"1.2.3", true},
		{"c1", true},
		{"p2", true},
		{"c1.2", true},
		{"1.c2", true},
		{"c1.p2.3", true},
		
		// Invalid - not position paths
		{"", false},
		{"abc", false},
		{"1a", false},
		{"a1", false},
		{"1.a", false},
		{"1.", false},
		{".1", false},
		{"1..2", false},
		{"buy milk", false},
		{"task-1", false},
		{"#1", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := looksLikePositionPath(tt.input)
			assert.Equal(t, tt.expected, result, 
				"looksLikePositionPath(%q) = %v, want %v", tt.input, result, tt.expected)
		})
	}
}