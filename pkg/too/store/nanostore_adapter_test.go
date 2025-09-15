package store_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a test adapter with a temp database
func createTestAdapter(t *testing.T) (*store.NanoStoreAdapter, func()) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	
	adapter, err := store.NewNanoStoreAdapter(dbPath)
	require.NoError(t, err, "Failed to create adapter")
	
	cleanup := func() {
		err := adapter.Close()
		assert.NoError(t, err, "Failed to close adapter")
	}
	
	return adapter, cleanup
}

// Test basic creation and initialization
func TestNewNanoStoreAdapter(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		adapter, cleanup := createTestAdapter(t)
		defer cleanup()
		assert.NotNil(t, adapter)
	})
	
	t.Run("home directory expansion", func(t *testing.T) {
		// Save original home
		originalHome := os.Getenv("HOME")
		if originalHome == "" {
			originalHome = os.Getenv("USERPROFILE") // Windows
		}
		
		// Create a temp dir to act as home
		tmpHome := t.TempDir()
		os.Setenv("HOME", tmpHome)
		defer os.Setenv("HOME", originalHome)
		
		// Create adapter with ~ path
		adapter, err := store.NewNanoStoreAdapter("~/.todos.db")
		require.NoError(t, err)
		defer adapter.Close()
		
		// Verify the file was created in the temp home
		expectedPath := filepath.Join(tmpHome, ".todos.db")
		_, err = os.Stat(expectedPath)
		assert.NoError(t, err, "Database file should exist in expanded home directory")
	})
	
	t.Run("creates parent directories", func(t *testing.T) {
		tmpDir := t.TempDir()
		deepPath := filepath.Join(tmpDir, "deep", "nested", "path", "todos.db")
		
		adapter, err := store.NewNanoStoreAdapter(deepPath)
		require.NoError(t, err)
		defer adapter.Close()
		
		// Verify parent directories were created
		parentDir := filepath.Dir(deepPath)
		info, err := os.Stat(parentDir)
		assert.NoError(t, err)
		assert.True(t, info.IsDir())
	})
}

// Test basic CRUD operations
func TestBasicCRUD(t *testing.T) {
	t.Run("add and retrieve todo", func(t *testing.T) {
		adapter, cleanup := createTestAdapter(t)
		defer cleanup()
		
		// Add a todo
		todo, err := adapter.Add("Test todo", nil)
		require.NoError(t, err)
		assert.NotNil(t, todo)
		assert.Equal(t, "Test todo", todo.Text)
		assert.Equal(t, "1", todo.PositionPath)
		assert.Equal(t, string(models.StatusPending), todo.Statuses["completion"])
		assert.NotEmpty(t, todo.UID)
		assert.Empty(t, todo.ParentID)
		
		// List todos
		todos, err := adapter.List(false)
		require.NoError(t, err)
		assert.Len(t, todos, 1)
		assert.Equal(t, todo.UID, todos[0].UID)
	})
	
	t.Run("add with parent", func(t *testing.T) {
		adapter, cleanup := createTestAdapter(t)
		defer cleanup()
		
		// Add parent
		parent, err := adapter.Add("Parent todo", nil)
		require.NoError(t, err)
		
		// Add child using parent's position path
		child, err := adapter.Add("Child todo", &parent.PositionPath)
		require.NoError(t, err)
		assert.Equal(t, parent.UID, child.ParentID)
		assert.Equal(t, "1.1", child.PositionPath)
		
		// Verify hierarchy in list
		todos, err := adapter.List(false)
		require.NoError(t, err)
		assert.Len(t, todos, 2)
	})
	
	t.Run("complete and reopen", func(t *testing.T) {
		adapter, cleanup := createTestAdapter(t)
		defer cleanup()
		
		// Add todo
		todo, err := adapter.Add("Test todo", nil)
		require.NoError(t, err)
		
		// Complete it
		err = adapter.Complete(todo.PositionPath)
		require.NoError(t, err)
		
		// Verify it's completed
		todos, err := adapter.List(true) // show all
		require.NoError(t, err)
		assert.Len(t, todos, 1)
		assert.Equal(t, string(models.StatusDone), todos[0].Statuses["completion"])
		assert.True(t, strings.HasPrefix(todos[0].PositionPath, "c"))
		
		// Reopen it
		err = adapter.Reopen(todos[0].PositionPath)
		require.NoError(t, err)
		
		// Verify it's pending again
		todos, err = adapter.List(false)
		require.NoError(t, err)
		assert.Len(t, todos, 1)
		assert.Equal(t, string(models.StatusPending), todos[0].Statuses["completion"])
		assert.Equal(t, "1", todos[0].PositionPath)
	})
	
	t.Run("update todo text", func(t *testing.T) {
		adapter, cleanup := createTestAdapter(t)
		defer cleanup()
		
		// Add todo
		todo, err := adapter.Add("Original text", nil)
		require.NoError(t, err)
		
		// Update it
		err = adapter.Update(todo.PositionPath, "Updated text")
		require.NoError(t, err)
		
		// Verify update
		todos, err := adapter.List(false)
		require.NoError(t, err)
		assert.Len(t, todos, 1)
		assert.Equal(t, "Updated text", todos[0].Text)
		assert.Equal(t, todo.UID, todos[0].UID)
	})
	
	t.Run("delete todo", func(t *testing.T) {
		adapter, cleanup := createTestAdapter(t)
		defer cleanup()
		
		// Add todos
		todo1, err := adapter.Add("Todo 1", nil)
		require.NoError(t, err)
		todo2, err := adapter.Add("Todo 2", nil)
		require.NoError(t, err)
		
		// Delete first todo
		err = adapter.Delete(todo1.PositionPath, false)
		require.NoError(t, err)
		
		// Verify only second remains
		todos, err := adapter.List(false)
		require.NoError(t, err)
		assert.Len(t, todos, 1)
		assert.Equal(t, todo2.UID, todos[0].UID)
	})
}

// Test smart ID detection
func TestSmartIDDetection(t *testing.T) {
	t.Run("operations with UUID", func(t *testing.T) {
		adapter, cleanup := createTestAdapter(t)
		defer cleanup()
		
		// Add todo
		todo, err := adapter.Add("Test todo", nil)
		require.NoError(t, err)
		uuid := todo.UID
		
		// Complete using UUID
		err = adapter.CompleteByUUID(uuid)
		require.NoError(t, err)
		
		// Update using UUID  
		err = adapter.UpdateByUUID(uuid, "Updated via UUID")
		require.NoError(t, err)
		
		// Reopen using UUID
		err = adapter.ReopenByUUID(uuid)
		require.NoError(t, err)
		
		// Verify final state
		todos, err := adapter.List(false)
		require.NoError(t, err)
		assert.Len(t, todos, 1)
		assert.Equal(t, "Updated via UUID", todos[0].Text)
		assert.Equal(t, string(models.StatusPending), todos[0].Statuses["completion"])
	})
	
	t.Run("operations with user-facing ID", func(t *testing.T) {
		adapter, cleanup := createTestAdapter(t)
		defer cleanup()
		
		// Add todo
		_, err := adapter.Add("Test todo", nil)
		require.NoError(t, err)
		
		// Complete using position path
		err = adapter.Complete("1")
		require.NoError(t, err)
		
		// Update using completed position path
		err = adapter.Update("c1", "Updated via user ID")
		require.NoError(t, err)
		
		// Reopen using completed position path
		err = adapter.Reopen("c1")
		require.NoError(t, err)
		
		// Verify final state
		todos, err := adapter.List(false)
		require.NoError(t, err)
		assert.Len(t, todos, 1)
		assert.Equal(t, "Updated via user ID", todos[0].Text)
		assert.Equal(t, string(models.StatusPending), todos[0].Statuses["completion"])
	})
}

// Test move operations
func TestMoveOperations(t *testing.T) {
	t.Run("move to different parent", func(t *testing.T) {
		adapter, cleanup := createTestAdapter(t)
		defer cleanup()
		
		// Create structure: parent1, parent2, child (under parent1)
		parent1, err := adapter.Add("Parent 1", nil)
		require.NoError(t, err)
		parent2, err := adapter.Add("Parent 2", nil) 
		require.NoError(t, err)
		child, err := adapter.Add("Child", &parent1.PositionPath)
		require.NoError(t, err)
		
		// Move child to parent2
		err = adapter.Move(child.PositionPath, &parent2.PositionPath)
		require.NoError(t, err)
		
		// Verify new structure
		todos, err := adapter.List(false)
		require.NoError(t, err)
		
		// Find the moved child
		var movedChild *models.Todo
		for _, todo := range todos {
			if todo.Text == "Child" {
				movedChild = todo
				break
			}
		}
		require.NotNil(t, movedChild)
		assert.Equal(t, parent2.UID, movedChild.ParentID)
		// Position path might be "3" if move changes the order
		assert.True(t, strings.HasPrefix(movedChild.PositionPath, "2.") || movedChild.PositionPath == "3",
			"Child should be under parent2 or reordered")
	})
	
	t.Run("move to root", func(t *testing.T) {
		adapter, cleanup := createTestAdapter(t)
		defer cleanup()
		
		// Create parent and child
		parent, err := adapter.Add("Parent", nil)
		require.NoError(t, err)
		child, err := adapter.Add("Child", &parent.PositionPath)
		require.NoError(t, err)
		
		// Move child to root
		err = adapter.Move(child.PositionPath, nil)
		require.NoError(t, err)
		
		// Verify it's at root
		todos, err := adapter.List(false)
		require.NoError(t, err)
		
		var movedChild *models.Todo
		for _, todo := range todos {
			if todo.Text == "Child" {
				movedChild = todo
				break
			}
		}
		require.NotNil(t, movedChild)
		assert.Empty(t, movedChild.ParentID)
		assert.Equal(t, "2", movedChild.PositionPath) // Should be second root item
	})
}

// Test filtering
func TestFiltering(t *testing.T) {
	t.Run("filter completed todos", func(t *testing.T) {
		adapter, cleanup := createTestAdapter(t)
		defer cleanup()
		
		// Add todos
		todo1, err := adapter.Add("Todo 1", nil)
		require.NoError(t, err)
		todo2, err := adapter.Add("Todo 2", nil)
		require.NoError(t, err)
		_, err = adapter.Add("Todo 3", nil)
		require.NoError(t, err)
		
		// Complete first two
		err = adapter.Complete(todo1.PositionPath)
		require.NoError(t, err)
		err = adapter.Complete(todo2.PositionPath)
		require.NoError(t, err)
		
		// List pending only (default)
		pending, err := adapter.List(false)
		require.NoError(t, err)
		assert.Len(t, pending, 1)
		// The only pending todo should have "pending" status
		assert.Equal(t, string(models.StatusPending), pending[0].Statuses["completion"])
		
		// List all
		all, err := adapter.List(true)
		require.NoError(t, err)
		assert.Len(t, all, 3)
	})
	
	t.Run("search functionality", func(t *testing.T) {
		adapter, cleanup := createTestAdapter(t)
		defer cleanup()
		
		// Add todos with different content
		_, err := adapter.Add("Buy groceries", nil)
		require.NoError(t, err)
		_, err = adapter.Add("Review PR", nil)
		require.NoError(t, err)
		_, err = adapter.Add("Buy coffee", nil) 
		require.NoError(t, err)
		
		// Search for "buy"
		results, err := adapter.Search("buy", false)
		require.NoError(t, err)
		assert.Len(t, results, 2)
		
		// Search for "PR"
		results, err = adapter.Search("PR", false)
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "Review PR", results[0].Text)
	})
}

// Test batch operations
func TestBatchOperations(t *testing.T) {
	t.Run("delete completed todos", func(t *testing.T) {
		adapter, cleanup := createTestAdapter(t)
		defer cleanup()
		
		// Add and complete some todos
		for i := 1; i <= 5; i++ {
			todo, err := adapter.Add(fmt.Sprintf("Todo %d", i), nil)
			require.NoError(t, err)
			if i <= 3 {
				err = adapter.Complete(todo.PositionPath)
				require.NoError(t, err)
			}
		}
		
		// Delete completed
		count, err := adapter.DeleteCompleted()
		require.NoError(t, err)
		assert.Equal(t, 3, count)
		
		// Verify only pending remain
		todos, err := adapter.List(true)
		require.NoError(t, err)
		assert.Len(t, todos, 2)
		for _, todo := range todos {
			assert.Equal(t, string(models.StatusPending), todo.Statuses["completion"])
		}
	})
	
	t.Run("cascade delete", func(t *testing.T) {
		adapter, cleanup := createTestAdapter(t)
		defer cleanup()
		
		// Create hierarchy
		parent, err := adapter.Add("Parent", nil)
		require.NoError(t, err)
		child1, err := adapter.Add("Child 1", &parent.PositionPath)
		require.NoError(t, err)
		_, err = adapter.Add("Grandchild", &child1.PositionPath)
		require.NoError(t, err)
		_, err = adapter.Add("Child 2", &parent.PositionPath)
		require.NoError(t, err)
		
		// Delete parent with cascade
		err = adapter.Delete(parent.PositionPath, true)
		require.NoError(t, err)
		
		// Verify all are gone
		todos, err := adapter.List(false)
		require.NoError(t, err)
		assert.Len(t, todos, 0)
	})
}

// Test error scenarios
func TestErrorScenarios(t *testing.T) {
	t.Run("invalid parent ID", func(t *testing.T) {
		adapter, cleanup := createTestAdapter(t)
		defer cleanup()
		
		invalidParent := "nonexistent"
		_, err := adapter.Add("Child", &invalidParent)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to resolve parent ID")
	})
	
	t.Run("delete with children without cascade", func(t *testing.T) {
		adapter, cleanup := createTestAdapter(t)
		defer cleanup()
		
		// Create parent and child
		parent, err := adapter.Add("Parent", nil)
		require.NoError(t, err)
		_, err = adapter.Add("Child", &parent.PositionPath)
		require.NoError(t, err)
		
		// Try to delete parent without cascade
		err = adapter.Delete(parent.PositionPath, false)
		assert.Error(t, err)
	})
	
	t.Run("operations on non-existent todo", func(t *testing.T) {
		adapter, cleanup := createTestAdapter(t)
		defer cleanup()
		
		// Try operations on non-existent ID
		err := adapter.Complete("999")
		assert.Error(t, err)
		
		err = adapter.Update("999", "New text")
		assert.Error(t, err)
		
		err = adapter.Delete("999", false)
		assert.Error(t, err)
	})
}

// Test helper methods
func TestHelperMethods(t *testing.T) {
	t.Run("resolve position path", func(t *testing.T) {
		adapter, cleanup := createTestAdapter(t)
		defer cleanup()
		
		// Add todo
		todo, err := adapter.Add("Test", nil)
		require.NoError(t, err)
		
		// Resolve position path to UUID
		uuid, err := adapter.ResolvePositionPath(todo.PositionPath)
		require.NoError(t, err)
		assert.Equal(t, todo.UID, uuid)
		
		// Complete and try with completed ID
		err = adapter.Complete(todo.PositionPath)
		require.NoError(t, err)
		
		uuid2, err := adapter.ResolvePositionPath("c1")
		require.NoError(t, err)
		assert.Equal(t, todo.UID, uuid2)
	})
	
	t.Run("get by UUID", func(t *testing.T) {
		adapter, cleanup := createTestAdapter(t)
		defer cleanup()
		
		// Add todo
		original, err := adapter.Add("Test todo", nil)
		require.NoError(t, err)
		
		// Get by UUID
		retrieved, err := adapter.GetByUUID(original.UID)
		require.NoError(t, err)
		assert.Equal(t, original.UID, retrieved.UID)
		assert.Equal(t, original.Text, retrieved.Text)
		assert.Equal(t, original.PositionPath, retrieved.PositionPath)
	})
}