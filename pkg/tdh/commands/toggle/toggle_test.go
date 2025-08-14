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

		// Get the todo ID (it will be 1 since it's the first todo)
		collection, _ := store.Load()
		todoPosition := collection.Todos[0].Position

		// Toggle the todo
		opts := toggle.Options{CollectionPath: store.Path()}
		result, err := toggle.Execute(todoPosition, opts)

		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "pending", result.OldStatus)
		assert.Equal(t, "done", result.NewStatus)
		assert.Equal(t, models.StatusDone, result.Todo.Status)

		// Verify it was saved using testutil
		collection, err = store.Load()
		testutil.AssertNoError(t, err)

		// Use testutil to find and verify the todo
		todo := testutil.AssertTodoByPosition(t, collection.Todos, todoPosition)
		testutil.AssertTodoHasStatus(t, todo, models.StatusDone)
	})

	t.Run("toggles done todo to pending", func(t *testing.T) {
		// Create a store with a done todo
		store := testutil.CreateStoreWithSpecs(t, []testutil.TodoSpec{
			{Text: "Completed task", Status: models.StatusDone},
		})

		// Get the todo ID
		collection, _ := store.Load()
		todoPosition := collection.Todos[0].Position

		// Toggle the todo
		opts := toggle.Options{CollectionPath: store.Path()}
		result, err := toggle.Execute(todoPosition, opts)

		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "done", result.OldStatus)
		assert.Equal(t, "pending", result.NewStatus)
		assert.Equal(t, models.StatusPending, result.Todo.Status)

		// Verify persistence
		collection, err = store.Load()
		testutil.AssertNoError(t, err)
		todo := testutil.AssertTodoByPosition(t, collection.Todos, todoPosition)
		testutil.AssertTodoHasStatus(t, todo, models.StatusPending)
	})

	t.Run("returns error for non-existent todo", func(t *testing.T) {
		// Create store with one todo
		store := testutil.CreatePopulatedStore(t, "Existing todo")

		// Try to toggle non-existent todo
		opts := toggle.Options{CollectionPath: store.Path()}
		result, err := toggle.Execute(999, opts)

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

		// Get the middle todo's position
		collection, _ := store.Load()
		middleTodoPosition := collection.Todos[1].Position

		// Toggle the middle todo
		opts := toggle.Options{CollectionPath: store.Path()}
		result, err := toggle.Execute(middleTodoPosition, opts)

		testutil.AssertNoError(t, err)
		assert.Equal(t, "done", result.OldStatus)
		assert.Equal(t, "pending", result.NewStatus)

		// Verify only the middle todo was changed
		collection, err = store.Load()
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
		result, err := toggle.Execute(1, opts)

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("handles negative todo ID", func(t *testing.T) {
		store := testutil.CreatePopulatedStore(t, "Test todo")

		opts := toggle.Options{CollectionPath: store.Path()}
		result, err := toggle.Execute(-1, opts)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "todo not found")
		assert.Contains(t, err.Error(), "-1")
	})

	t.Run("handles zero todo ID", func(t *testing.T) {
		store := testutil.CreatePopulatedStore(t, "Test todo")

		opts := toggle.Options{CollectionPath: store.Path()}
		result, err := toggle.Execute(0, opts)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "todo not found")
		assert.Contains(t, err.Error(), "0")
	})

	t.Run("toggle updates modified timestamp", func(t *testing.T) {
		store := testutil.CreatePopulatedStore(t, "Test todo")

		// Get initial state
		collection, _ := store.Load()
		originalModified := collection.Todos[0].Modified
		todoPosition := collection.Todos[0].Position

		// Toggle the todo
		opts := toggle.Options{CollectionPath: store.Path()}
		result, err := toggle.Execute(todoPosition, opts)

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
		result, err := toggle.Execute(1, opts)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "todo not found")
	})

	t.Run("result contains correct todo data", func(t *testing.T) {
		store := testutil.CreatePopulatedStore(t, "My todo")

		collection, _ := store.Load()
		todoPosition := collection.Todos[0].Position

		opts := toggle.Options{CollectionPath: store.Path()}
		result, err := toggle.Execute(todoPosition, opts)

		testutil.AssertNoError(t, err)
		assert.NotNil(t, result.Todo)
		assert.Equal(t, todoPosition, result.Todo.Position)
		assert.Equal(t, "My todo", result.Todo.Text)
		assert.Equal(t, models.StatusDone, result.Todo.Status)
		assert.Equal(t, "pending", result.OldStatus)
		assert.Equal(t, "done", result.NewStatus)
	})
}
