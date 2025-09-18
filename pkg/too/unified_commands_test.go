package too_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/arthur-debert/too/pkg/too"
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to create test database path
func createTestDB(t *testing.T) string {
	tmpDir := t.TempDir()
	return filepath.Join(tmpDir, "test.db")
}

// Helper to execute command and get result
func executeCommand(t *testing.T, cmdName string, args []string, opts map[string]interface{}) *too.ChangeResult {
	result, err := too.ExecuteUnifiedCommand(cmdName, args, opts)
	require.NoError(t, err, "Command %s should not error", cmdName)
	require.NotNil(t, result, "Result should not be nil")
	return result
}

// Test command registration and aliases
func TestCommandRegistration(t *testing.T) {
	t.Run("find command by name", func(t *testing.T) {
		dbPath := createTestDB(t)
		opts := map[string]interface{}{"collectionPath": dbPath}
		
		// Test known command
		result, err := too.ExecuteUnifiedCommand("add", []string{"Test todo"}, opts)
		require.NoError(t, err)
		assert.NotNil(t, result)
	})
	
	t.Run("find command by alias", func(t *testing.T) {
		dbPath := createTestDB(t)
		opts := map[string]interface{}{"collectionPath": dbPath}
		
		// Add a todo first
		result, err := too.ExecuteUnifiedCommand("add", []string{"Test todo"}, opts)
		require.NoError(t, err)
		
		// Test alias 'c' for complete
		result, err = too.ExecuteUnifiedCommand("c", []string{"1"}, opts)
		require.NoError(t, err)
		assert.Empty(t, result.Message)
	})
	
	t.Run("unknown command", func(t *testing.T) {
		dbPath := createTestDB(t)
		opts := map[string]interface{}{"collectionPath": dbPath}
		
		_, err := too.ExecuteUnifiedCommand("unknown", []string{}, opts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown command")
	})
}

// Test basic command execution with options
func TestCommandExecutionWithOptions(t *testing.T) {
	t.Run("add with parent option", func(t *testing.T) {
		dbPath := createTestDB(t)
		opts := map[string]interface{}{"collectionPath": dbPath}
		
		// Add parent
		parent := executeCommand(t, "add", []string{"Parent todo"}, opts)
		require.GreaterOrEqual(t, len(parent.AffectedTodos), 1)
		// Find the parent todo we just added
		var parentTodo *models.Todo
		for _, todo := range parent.AffectedTodos {
			if todo.Text == "Parent todo" {
				parentTodo = todo
				break
			}
		}
		require.NotNil(t, parentTodo)
		parentPath := parentTodo.PositionPath
		
		// Add child with parent option
		opts["parent"] = parentPath
		child := executeCommand(t, "add", []string{"Child todo"}, opts)
		
		// The result might include both parent and child in AllTodos
		require.GreaterOrEqual(t, len(child.AffectedTodos), 1)
		// Find the child todo (the one we just added)
		var childTodo *models.Todo
		for _, todo := range child.AffectedTodos {
			if todo.Text == "Child todo" {
				childTodo = todo
				break
			}
		}
		require.NotNil(t, childTodo, "Child todo should be in affected todos")
		assert.Equal(t, parentTodo.UID, childTodo.ParentID)
		assert.Equal(t, "1.1", childTodo.PositionPath)
	})
	
	t.Run("list with filter options", func(t *testing.T) {
		dbPath := createTestDB(t)
		opts := map[string]interface{}{"collectionPath": dbPath}
		
		// Add and complete some todos
		executeCommand(t, "add", []string{"Todo 1"}, opts)
		executeCommand(t, "add", []string{"Todo 2"}, opts)
		executeCommand(t, "add", []string{"Todo 3"}, opts)
		executeCommand(t, "complete", []string{"1", "2"}, opts)
		
		// List without filters (pending only)
		result := executeCommand(t, "list", []string{}, opts)
		assert.Len(t, result.AllTodos, 1)
		assert.Equal(t, "Todo 3", result.AllTodos[0].Text)
		
		// List with done filter
		opts["done"] = true
		result = executeCommand(t, "list", []string{}, opts)
		assert.Len(t, result.AllTodos, 2)
		
		// List with all filter
		delete(opts, "done")
		opts["all"] = true
		result = executeCommand(t, "list", []string{}, opts)
		assert.Len(t, result.AllTodos, 3)
	})
	
	t.Run("search command", func(t *testing.T) {
		dbPath := createTestDB(t)
		baseOpts := map[string]interface{}{"collectionPath": dbPath}
		
		// Add todos
		executeCommand(t, "add", []string{"Buy groceries"}, baseOpts)
		executeCommand(t, "add", []string{"Review PR"}, baseOpts)
		executeCommand(t, "add", []string{"Buy coffee"}, baseOpts)
		
		// Search for "buy" - create new opts with query
		searchOpts := map[string]interface{}{
			"collectionPath": dbPath,
			"query": "buy",
		}
		result := executeCommand(t, "search", []string{"buy"}, searchOpts)
		// The result should contain the search message
		assert.Contains(t, result.Message, "Found")
		
		// Search for "PR"
		searchOpts["query"] = "PR"
		result = executeCommand(t, "search", []string{"PR"}, searchOpts)
		assert.Contains(t, result.Message, "Found")
	})
}

// Test multi-todo operations
func TestMultiTodoOperations(t *testing.T) {
	t.Run("complete multiple todos", func(t *testing.T) {
		dbPath := createTestDB(t)
		opts := map[string]interface{}{"collectionPath": dbPath}
		
		// Add todos
		executeCommand(t, "add", []string{"Todo 1"}, opts)
		executeCommand(t, "add", []string{"Todo 2"}, opts)
		executeCommand(t, "add", []string{"Todo 3"}, opts)
		
		// Complete multiple
		result := executeCommand(t, "complete", []string{"1", "2", "3"}, opts)
		assert.Len(t, result.AffectedTodos, 3)
		assert.Empty(t, result.Message)
		
		// Verify all are completed
		for _, todo := range result.AffectedTodos {
			assert.Equal(t, string(models.StatusDone), todo.Statuses["completion"])
		}
	})
	
	t.Run("reopen multiple todos", func(t *testing.T) {
		dbPath := createTestDB(t)
		opts := map[string]interface{}{"collectionPath": dbPath}
		
		// Add and complete todos
		executeCommand(t, "add", []string{"Todo 1"}, opts)
		executeCommand(t, "add", []string{"Todo 2"}, opts)
		executeCommand(t, "complete", []string{"1", "2"}, opts)
		
		// Reopen both
		result := executeCommand(t, "reopen", []string{"c1", "c2"}, opts)
		assert.Len(t, result.AffectedTodos, 2)
		assert.Empty(t, result.Message) // No message, visual highlight is sufficient
		
		// Verify all are pending
		for _, todo := range result.AffectedTodos {
			assert.Equal(t, string(models.StatusPending), todo.Statuses["completion"])
		}
	})
	
	t.Run("mixed references", func(t *testing.T) {
		dbPath := createTestDB(t)
		opts := map[string]interface{}{"collectionPath": dbPath}
		
		// Create hierarchy
		parent := executeCommand(t, "add", []string{"Parent"}, opts)
		opts["parent"] = parent.AffectedTodos[0].PositionPath
		executeCommand(t, "add", []string{"Child 1"}, opts)
		executeCommand(t, "add", []string{"Child 2"}, opts)
		delete(opts, "parent")
		
		// Complete using mixed references
		result := executeCommand(t, "complete", []string{"1", "1.1", "1.2"}, opts)
		assert.Len(t, result.AffectedTodos, 3)
	})
}

// Test error handling
func TestErrorHandling(t *testing.T) {
	t.Run("invalid reference", func(t *testing.T) {
		dbPath := createTestDB(t)
		opts := map[string]interface{}{"collectionPath": dbPath}
		
		_, err := too.ExecuteUnifiedCommand("complete", []string{"999"}, opts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to resolve reference")
	})
	
	t.Run("missing required arguments", func(t *testing.T) {
		dbPath := createTestDB(t)
		opts := map[string]interface{}{"collectionPath": dbPath}
		
		// Add without text
		_, err := too.ExecuteUnifiedCommand("add", []string{}, opts)
		assert.Error(t, err)
		
		// Edit without text
		executeCommand(t, "add", []string{"Test"}, opts)
		_, err = too.ExecuteUnifiedCommand("edit", []string{"1"}, opts)
		assert.Error(t, err)
	})
	
	t.Run("invalid parent reference", func(t *testing.T) {
		dbPath := createTestDB(t)
		opts := map[string]interface{}{
			"collectionPath": dbPath,
			"parent": "nonexistent",
		}
		
		_, err := too.ExecuteUnifiedCommand("add", []string{"Child"}, opts)
		assert.Error(t, err)
	})
}

// Test command-specific behaviors
func TestCommandSpecificBehaviors(t *testing.T) {
	t.Run("edit command", func(t *testing.T) {
		dbPath := createTestDB(t)
		opts := map[string]interface{}{"collectionPath": dbPath}
		
		// Add todo
		executeCommand(t, "add", []string{"Original text"}, opts)
		
		// Edit it
		result := executeCommand(t, "edit", []string{"1", "Updated text"}, opts)
		assert.Len(t, result.AffectedTodos, 1)
		assert.Equal(t, "Updated text", result.AffectedTodos[0].Text)
		assert.Empty(t, result.Message)
	})
	
	t.Run("move command", func(t *testing.T) {
		dbPath := createTestDB(t)
		opts := map[string]interface{}{"collectionPath": dbPath}
		
		// Create structure
		parent1 := executeCommand(t, "add", []string{"Parent 1"}, opts)
		parent2 := executeCommand(t, "add", []string{"Parent 2"}, opts)
		opts["parent"] = parent1.AffectedTodos[0].PositionPath
		child := executeCommand(t, "add", []string{"Child"}, opts)
		delete(opts, "parent")
		
		// Move child to parent2
		result := executeCommand(t, "move", []string{child.AffectedTodos[0].PositionPath, parent2.AffectedTodos[0].PositionPath}, opts)
		assert.Len(t, result.AffectedTodos, 1)
		assert.Equal(t, parent2.AffectedTodos[0].UID, result.AffectedTodos[0].ParentID)
		assert.Empty(t, result.Message)
	})
	
	t.Run("swap command uses move", func(t *testing.T) {
		dbPath := createTestDB(t)
		opts := map[string]interface{}{"collectionPath": dbPath}
		
		// Add todos
		todo1 := executeCommand(t, "add", []string{"Todo 1"}, opts)
		executeCommand(t, "add", []string{"Todo 2"}, opts)
		
		// Swap is a CLI command that calls move internally
		// In unified commands, we should use move directly
		// Moving todo 2 under todo 1
		result := executeCommand(t, "move", []string{"2", "1"}, opts)
		assert.Empty(t, result.Message)
		assert.Equal(t, "1.1", result.AffectedTodos[0].PositionPath)
		assert.Equal(t, todo1.AffectedTodos[0].UID, result.AffectedTodos[0].ParentID)
	})
	
	t.Run("clean command", func(t *testing.T) {
		dbPath := createTestDB(t)
		opts := map[string]interface{}{"collectionPath": dbPath}
		
		// Add and complete todos
		executeCommand(t, "add", []string{"Todo 1"}, opts)
		executeCommand(t, "add", []string{"Todo 2"}, opts)
		executeCommand(t, "add", []string{"Todo 3"}, opts)
		executeCommand(t, "complete", []string{"1", "2"}, opts)
		
		// Clean completed
		result := executeCommand(t, "clean", []string{}, opts)
		assert.Len(t, result.AffectedTodos, 2)
		assert.Contains(t, result.Message, "Cleaned")
		
		// Verify only pending remain
		listResult := executeCommand(t, "list", []string{}, opts)
		assert.Len(t, listResult.AllTodos, 1)
		assert.Equal(t, "Todo 3", listResult.AllTodos[0].Text)
	})
}

// Test edge cases
func TestEdgeCases(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		dbPath := createTestDB(t)
		opts := map[string]interface{}{"collectionPath": dbPath}
		
		result := executeCommand(t, "list", []string{}, opts)
		assert.Empty(t, result.AllTodos)
		// The message for empty list might vary - just check it executed
		assert.NotNil(t, result)
	})
	
	t.Run("single todo operations", func(t *testing.T) {
		dbPath := createTestDB(t)
		opts := map[string]interface{}{"collectionPath": dbPath}
		
		// Add single todo
		add := executeCommand(t, "add", []string{"Single todo"}, opts)
		assert.GreaterOrEqual(t, len(add.AffectedTodos), 1)
		
		// Complete it
		complete := executeCommand(t, "complete", []string{"1"}, opts)
		assert.Len(t, complete.AffectedTodos, 1)
		
		// Reopen it
		reopen := executeCommand(t, "reopen", []string{"c1"}, opts)
		assert.Len(t, reopen.AffectedTodos, 1)
		
		// Note: too does not support deletion, only completion
		// This is by design to maintain a complete history
	})
	
	t.Run("deep hierarchy", func(t *testing.T) {
		dbPath := createTestDB(t)
		opts := map[string]interface{}{"collectionPath": dbPath}
		
		// Create deep hierarchy
		root := executeCommand(t, "add", []string{"Root"}, opts)
		
		currentParent := root.AffectedTodos[0].PositionPath
		for i := 1; i <= 5; i++ {
			opts["parent"] = currentParent
			child := executeCommand(t, "add", []string{fmt.Sprintf("Level %d", i)}, opts)
			currentParent = child.AffectedTodos[0].PositionPath
		}
		delete(opts, "parent")
		
		// List all
		opts["all"] = true
		result := executeCommand(t, "list", []string{}, opts)
		assert.Len(t, result.AllTodos, 6) // root + 5 levels
		
		// Complete entire hierarchy
		result = executeCommand(t, "complete", []string{"1"}, opts)
		assert.Empty(t, result.Message)
	})
	
	t.Run("special characters in text", func(t *testing.T) {
		dbPath := createTestDB(t)
		opts := map[string]interface{}{"collectionPath": dbPath}
		
		specialTexts := []string{
			"Todo with 'quotes'",
			"Todo with \"double quotes\"",
			"Todo with\nnewline",
			"Todo with\ttab",
			"Todo with unicode: 🎉",
		}
		
		for _, text := range specialTexts {
			result := executeCommand(t, "add", []string{text}, opts)
			assert.GreaterOrEqual(t, len(result.AffectedTodos), 1)
			// Find the todo we just added
			found := false
			for _, todo := range result.AffectedTodos {
				if todo.Text == text {
					found = true
					break
				}
			}
			assert.True(t, found, "Should find todo with text: %s", text)
		}
	})
}

// Test validation functions
func TestValidation(t *testing.T) {
	t.Run("move validation", func(t *testing.T) {
		dbPath := createTestDB(t)
		opts := map[string]interface{}{"collectionPath": dbPath}
		
		// Try move with no arguments
		_, err := too.ExecuteUnifiedCommand("move", []string{}, opts)
		assert.Error(t, err)
		
		// Try move with only one argument
		executeCommand(t, "add", []string{"Todo"}, opts)
		_, err = too.ExecuteUnifiedCommand("move", []string{"1"}, opts)
		assert.Error(t, err)
	})
}

// Test custom message functions
func TestCustomMessages(t *testing.T) {
	t.Run("complete message variations", func(t *testing.T) {
		dbPath := createTestDB(t)
		opts := map[string]interface{}{"collectionPath": dbPath}
		
		// Single todo
		executeCommand(t, "add", []string{"Single"}, opts)
		result := executeCommand(t, "complete", []string{"1"}, opts)
		assert.Empty(t, result.Message) // No message, visual highlight is sufficient
		
		// Multiple todos in new test env
		dbPath2 := createTestDB(t)
		opts2 := map[string]interface{}{"collectionPath": dbPath2}
		executeCommand(t, "add", []string{"Multi 1"}, opts2)
		executeCommand(t, "add", []string{"Multi 2"}, opts2)
		result = executeCommand(t, "complete", []string{"1", "2"}, opts2)
		assert.Empty(t, result.Message) // No message, visual highlight is sufficient
		// Verify both todos were completed
		assert.Len(t, result.AffectedTodos, 2)
	})
	
	t.Run("clean message variations", func(t *testing.T) {
		dbPath := createTestDB(t)
		opts := map[string]interface{}{"collectionPath": dbPath}
		
		// No completed todos
		result := executeCommand(t, "clean", []string{}, opts)
		assert.Contains(t, result.Message, "No finished todos")
		
		// With completed todos
		executeCommand(t, "add", []string{"Todo 1"}, opts)
		executeCommand(t, "add", []string{"Todo 2"}, opts)
		executeCommand(t, "complete", []string{"1", "2"}, opts)
		result = executeCommand(t, "clean", []string{}, opts)
		assert.Contains(t, result.Message, "Cleaned 2")
	})
}

// Test database path handling
func TestDatabasePathHandling(t *testing.T) {
	t.Run("current directory default", func(t *testing.T) {
		// Create a temp directory and change to it
		tmpDir := t.TempDir()
		originalWd, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(originalWd)
		
		// Execute with empty collection path
		// When empty, datapath.ResolveCollectionPath would default to .todos.json in current dir
		opts := map[string]interface{}{"collectionPath": ".todos.json"}
		result, err := too.ExecuteUnifiedCommand("add", []string{"Test"}, opts)
		require.NoError(t, err)
		assert.NotNil(t, result)
		
		// Verify file was created in current directory
		expectedPath := filepath.Join(tmpDir, ".todos.json")
		_, err = os.Stat(expectedPath)
		assert.NoError(t, err)
	})
	
	t.Run("custom path", func(t *testing.T) {
		customPath := filepath.Join(t.TempDir(), "custom", "todos.db")
		opts := map[string]interface{}{"collectionPath": customPath}
		
		result, err := too.ExecuteUnifiedCommand("add", []string{"Test"}, opts)
		require.NoError(t, err)
		assert.NotNil(t, result)
		
		// Verify file was created
		_, err = os.Stat(customPath)
		assert.NoError(t, err)
	})
}