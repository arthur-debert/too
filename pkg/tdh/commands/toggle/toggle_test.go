package toggle_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/arthur-debert/tdh/pkg/tdh/commands/toggle"
	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/arthur-debert/tdh/pkg/tdh/testutil"
	"github.com/stretchr/testify/assert"
)

func TestToggleCommand(t *testing.T) {
	t.Run("toggles pending todo to done", func(t *testing.T) {
		// Create a store with a pending todo
		store := testutil.CreatePopulatedStore(t, "Todo to toggle")

		// Toggle the todo using position path
		opts := toggle.Options{CollectionPath: store.Path()}
		result, err := toggle.Execute("1", opts)

		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "pending", result.OldStatus)
		assert.Equal(t, "done", result.NewStatus)
		assert.Equal(t, models.StatusDone, result.Todo.Status)

		// Verify it was saved using testutil
		collection, err := store.Load()
		testutil.AssertNoError(t, err)

		// Use testutil to find and verify the todo
		todo := testutil.AssertTodoByPosition(t, collection.Todos, 1)
		testutil.AssertTodoHasStatus(t, todo, models.StatusDone)
	})

	t.Run("toggles done todo to pending", func(t *testing.T) {
		// Create a store with a done todo
		store := testutil.CreateStoreWithSpecs(t, []testutil.TodoSpec{
			{Text: "Completed task", Status: models.StatusDone},
		})

		// Toggle the todo
		opts := toggle.Options{CollectionPath: store.Path()}
		result, err := toggle.Execute("1", opts)

		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "done", result.OldStatus)
		assert.Equal(t, "pending", result.NewStatus)
		assert.Equal(t, models.StatusPending, result.Todo.Status)

		// Verify persistence
		collection, err := store.Load()
		testutil.AssertNoError(t, err)
		todo := testutil.AssertTodoByPosition(t, collection.Todos, 1)
		testutil.AssertTodoHasStatus(t, todo, models.StatusPending)
	})

	t.Run("returns error for non-existent todo", func(t *testing.T) {
		// Create store with one todo
		store := testutil.CreatePopulatedStore(t, "Existing todo")

		// Try to toggle non-existent todo
		opts := toggle.Options{CollectionPath: store.Path()}
		result, err := toggle.Execute("999", opts)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "todo not found")

		// Verify no changes were made
		collection, err := store.Load()
		testutil.AssertNoError(t, err)
		testutil.AssertCollectionSize(t, collection, 1)
		testutil.AssertTodoHasStatus(t, collection.Todos[0], models.StatusPending)
	})

	t.Run("toggles correct todo when multiple exist", func(t *testing.T) {
		// Create store with multiple todos of different statuses
		store := testutil.CreateStoreWithSpecs(t, []testutil.TodoSpec{
			{Text: "First pending", Status: models.StatusPending},
			{Text: "Middle done", Status: models.StatusDone},
			{Text: "Last pending", Status: models.StatusPending},
		})

		// Toggle the middle todo
		opts := toggle.Options{CollectionPath: store.Path()}
		result, err := toggle.Execute("2", opts)

		testutil.AssertNoError(t, err)
		assert.Equal(t, "done", result.OldStatus)
		assert.Equal(t, "pending", result.NewStatus)

		// Verify only the middle todo was changed
		collection, err := store.Load()
		testutil.AssertNoError(t, err)
		testutil.AssertCollectionSize(t, collection, 3)
		testutil.AssertTodoHasStatus(t, collection.Todos[0], models.StatusPending)
		testutil.AssertTodoHasStatus(t, collection.Todos[1], models.StatusPending) // Changed
		testutil.AssertTodoHasStatus(t, collection.Todos[2], models.StatusPending)
	})

	t.Run("returns error when store operation fails", func(t *testing.T) {
		// Create a read-only directory to force a store error
		dir := t.TempDir()
		dbPath := filepath.Join(dir, "todos.json")

		// Create a store with a todo
		_ = testutil.CreatePopulatedStore(t, "Test todo")

		// Write to the read-only path
		err := os.WriteFile(dbPath, []byte(`[{"id":1,"text":"Test todo","status":"pending"}]`), 0644)
		assert.NoError(t, err)

		// Make the directory read-only
		err = os.Chmod(dir, 0555)
		assert.NoError(t, err)
		defer func() { _ = os.Chmod(dir, 0755) }()

		// Try to toggle
		opts := toggle.Options{CollectionPath: dbPath}
		result, err := toggle.Execute("1", opts)

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("handles invalid position path", func(t *testing.T) {
		store := testutil.CreatePopulatedStore(t, "Test todo")

		opts := toggle.Options{CollectionPath: store.Path()}
		result, err := toggle.Execute("invalid", opts)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid position")
	})

	t.Run("handles zero position", func(t *testing.T) {
		store := testutil.CreatePopulatedStore(t, "Test todo")

		opts := toggle.Options{CollectionPath: store.Path()}
		result, err := toggle.Execute("0", opts)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "position must be >= 1")
	})

	t.Run("toggle updates modified timestamp", func(t *testing.T) {
		store := testutil.CreatePopulatedStore(t, "Test todo")

		// Get initial state
		collection, _ := store.Load()
		originalModified := collection.Todos[0].Modified

		// Toggle the todo
		opts := toggle.Options{CollectionPath: store.Path()}
		result, err := toggle.Execute("1", opts)

		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)

		// Verify timestamp was updated
		assert.True(t, result.Todo.Modified.After(originalModified))

		// Verify in persistence
		collection, err = store.Load()
		testutil.AssertNoError(t, err)
		assert.True(t, collection.Todos[0].Modified.After(originalModified))
	})

	t.Run("handles empty collection", func(t *testing.T) {
		// Create empty store
		store := testutil.CreatePopulatedStore(t)

		opts := toggle.Options{CollectionPath: store.Path()}
		result, err := toggle.Execute("1", opts)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "todo not found")
	})

	t.Run("result contains correct todo data", func(t *testing.T) {
		store := testutil.CreatePopulatedStore(t, "My todo")

		opts := toggle.Options{CollectionPath: store.Path()}
		result, err := toggle.Execute("1", opts)

		testutil.AssertNoError(t, err)
		assert.NotNil(t, result.Todo)
		assert.Equal(t, 1, result.Todo.Position)
		assert.Equal(t, "My todo", result.Todo.Text)
		assert.Equal(t, models.StatusDone, result.Todo.Status)
		assert.Equal(t, "pending", result.OldStatus)
		assert.Equal(t, "done", result.NewStatus)
	})

	t.Run("auto-reorders todos after toggle", func(t *testing.T) {
		// Create store with todos having non-sequential positions
		store := testutil.CreateStoreWithSpecs(t, []testutil.TodoSpec{
			{Text: "First", Status: models.StatusPending},
			{Text: "Second", Status: models.StatusPending},
			{Text: "Third", Status: models.StatusPending},
		})

		// Manually set non-sequential positions to simulate gaps
		collection, _ := store.Load()
		collection.Todos[0].Position = 1
		collection.Todos[1].Position = 5
		collection.Todos[2].Position = 8
		err := store.Save(collection)
		testutil.AssertNoError(t, err)

		// Toggle the middle todo (position 5)
		opts := toggle.Options{CollectionPath: store.Path()}
		_, err = toggle.Execute("5", opts)
		testutil.AssertNoError(t, err)

		// Verify todos were reordered to sequential positions
		collection, err = store.Load()
		testutil.AssertNoError(t, err)
		assert.Equal(t, 1, collection.Todos[0].Position)
		assert.Equal(t, 2, collection.Todos[1].Position)
		assert.Equal(t, 3, collection.Todos[2].Position)
	})
}

func TestToggleCommandWithNestedTodos(t *testing.T) {
	t.Run("toggles nested todo using position path", func(t *testing.T) {
		// Create store with nested todos
		store := testutil.CreateNestedStore(t)

		// Toggle todo at position 1.1
		opts := toggle.Options{CollectionPath: store.Path()}
		result, err := toggle.Execute("1.1", opts)

		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "pending", result.OldStatus)
		assert.Equal(t, "done", result.NewStatus)
		assert.Equal(t, "Sub-task 1.1", result.Todo.Text)

		// Verify it was saved
		collection, err := store.Load()
		testutil.AssertNoError(t, err)

		// Find the nested todo and verify its status
		parent := collection.Todos[0]
		child := parent.Items[0]
		assert.Equal(t, models.StatusDone, child.Status)
	})

	t.Run("recursively toggles parent and all children", func(t *testing.T) {
		// Create store with nested todos
		store := testutil.CreateNestedStore(t)

		// Toggle parent todo at position 1
		opts := toggle.Options{CollectionPath: store.Path()}
		result, err := toggle.Execute("1", opts)

		testutil.AssertNoError(t, err)
		assert.Equal(t, "done", result.NewStatus)

		// Verify parent and all children are done
		collection, err := store.Load()
		testutil.AssertNoError(t, err)

		parent := collection.Todos[0]
		assert.Equal(t, models.StatusDone, parent.Status)

		// Check all children
		for _, child := range parent.Items {
			assert.Equal(t, models.StatusDone, child.Status)
			// Check grandchildren
			for _, grandchild := range child.Items {
				assert.Equal(t, models.StatusDone, grandchild.Status)
			}
		}
	})

	t.Run("recursively toggles done parent to pending with all children", func(t *testing.T) {
		// Create store with done nested todos
		store := testutil.CreateNestedStore(t)

		// First toggle everything to done
		collection, _ := store.Load()
		setStatusRecursive(collection.Todos[0], models.StatusDone)
		err := store.Save(collection)
		testutil.AssertNoError(t, err)

		// Toggle parent back to pending
		opts := toggle.Options{CollectionPath: store.Path()}
		result, err := toggle.Execute("1", opts)

		testutil.AssertNoError(t, err)
		assert.Equal(t, "pending", result.NewStatus)

		// Verify parent and all children are pending
		collection, err = store.Load()
		testutil.AssertNoError(t, err)

		parent := collection.Todos[0]
		assert.Equal(t, models.StatusPending, parent.Status)

		// Check all children
		for _, child := range parent.Items {
			assert.Equal(t, models.StatusPending, child.Status)
			// Check grandchildren
			for _, grandchild := range child.Items {
				assert.Equal(t, models.StatusPending, grandchild.Status)
			}
		}
	})

	t.Run("toggles deeply nested todo", func(t *testing.T) {
		// Create store with deeply nested todos
		store := testutil.CreateNestedStore(t)

		// Toggle todo at position 1.2.1
		opts := toggle.Options{CollectionPath: store.Path()}
		result, err := toggle.Execute("1.2.1", opts)

		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Grandchild 1.2.1", result.Todo.Text)
		assert.Equal(t, "done", result.NewStatus)

		// Verify it was saved
		collection, err := store.Load()
		testutil.AssertNoError(t, err)

		// Find the deeply nested todo
		parent := collection.Todos[0]
		subTask := parent.Items[1] // position 2 in Items array
		grandchild := subTask.Items[0]
		assert.Equal(t, models.StatusDone, grandchild.Status)

		// Verify its children are also done
		for _, greatGrandchild := range grandchild.Items {
			assert.Equal(t, models.StatusDone, greatGrandchild.Status)
		}
	})

	t.Run("returns error for invalid nested path", func(t *testing.T) {
		store := testutil.CreateNestedStore(t)

		opts := toggle.Options{CollectionPath: store.Path()}
		result, err := toggle.Execute("1.99.1", opts)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "no item found at position 99")
	})

	t.Run("all children get same modified timestamp as parent", func(t *testing.T) {
		store := testutil.CreateNestedStore(t)

		// Toggle parent
		opts := toggle.Options{CollectionPath: store.Path()}
		_, err := toggle.Execute("1", opts)

		testutil.AssertNoError(t, err)

		// Verify all items have the same modified timestamp
		collection, err := store.Load()
		testutil.AssertNoError(t, err)

		parent := collection.Todos[0]
		parentTime := parent.Modified

		// Check all children have same timestamp
		for _, child := range parent.Items {
			assert.Equal(t, parentTime, child.Modified)
			for _, grandchild := range child.Items {
				assert.Equal(t, parentTime, grandchild.Modified)
			}
		}
	})

	t.Run("handles empty position path", func(t *testing.T) {
		store := testutil.CreateNestedStore(t)

		opts := toggle.Options{CollectionPath: store.Path()}
		result, err := toggle.Execute("", opts)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "empty path")
	})

	t.Run("handles position path with trailing dot", func(t *testing.T) {
		store := testutil.CreateNestedStore(t)

		opts := toggle.Options{CollectionPath: store.Path()}
		result, err := toggle.Execute("1.2.", opts)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid position")
	})

	t.Run("handles position path with leading dot", func(t *testing.T) {
		store := testutil.CreateNestedStore(t)

		opts := toggle.Options{CollectionPath: store.Path()}
		result, err := toggle.Execute(".1.2", opts)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid position")
	})

	t.Run("toggles only specified child, not siblings", func(t *testing.T) {
		store := testutil.CreateNestedStore(t)

		// Toggle only the second child
		opts := toggle.Options{CollectionPath: store.Path()}
		_, err := toggle.Execute("1.2", opts)
		testutil.AssertNoError(t, err)

		// Verify siblings are unaffected
		collection, err := store.Load()
		testutil.AssertNoError(t, err)

		parent := collection.Todos[0]
		assert.Equal(t, models.StatusPending, parent.Status)          // Parent unchanged
		assert.Equal(t, models.StatusPending, parent.Items[0].Status) // First sibling unchanged
		assert.Equal(t, models.StatusDone, parent.Items[1].Status)    // Target changed

		// But its children should all be done
		for _, grandchild := range parent.Items[1].Items {
			assert.Equal(t, models.StatusDone, grandchild.Status)
		}
	})
}

// Helper function to set status recursively
func setStatusRecursive(todo *models.Todo, status models.TodoStatus) {
	todo.Status = status
	for _, child := range todo.Items {
		setStatusRecursive(child, status)
	}
}
