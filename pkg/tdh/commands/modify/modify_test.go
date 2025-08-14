package modify_test

import (
	"testing"

	"github.com/arthur-debert/tdh/pkg/tdh/commands/modify"
	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/arthur-debert/tdh/pkg/tdh/testutil"
	"github.com/stretchr/testify/assert"
)

func TestModifyCommand(t *testing.T) {
	t.Run("successfully modifies todo text", func(t *testing.T) {
		// Create store with test todos using testutil
		store := testutil.CreatePopulatedStore(t, "Original todo text", "Another todo")

		// Get the first todo's ID
		collection, _ := store.Load()
		todoID := collection.Todos[0].ID

		// Execute modify command
		opts := modify.Options{CollectionPath: store.Path()}
		result, err := modify.Execute(int(todoID), "Modified todo text", opts)

		// Verify result
		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Original todo text", result.OldText)
		assert.Equal(t, "Modified todo text", result.NewText)
		assert.Equal(t, "Modified todo text", result.Todo.Text)

		// Verify persistence using testutil
		collection, err = store.Load()
		testutil.AssertNoError(t, err)
		todo := testutil.AssertTodoByID(t, collection.Todos, todoID)
		assert.Equal(t, "Modified todo text", todo.Text)

		// Ensure other todos are unchanged
		testutil.AssertTodoInList(t, collection.Todos, "Another todo")
	})

	t.Run("preserves todo status when modifying", func(t *testing.T) {
		// Create store with a done todo
		store := testutil.CreateStoreWithSpecs(t, []testutil.TodoSpec{
			{Text: "Completed task", Status: models.StatusDone},
		})

		// Get the todo's ID
		collection, _ := store.Load()
		todoID := collection.Todos[0].ID

		// Modify the todo
		opts := modify.Options{CollectionPath: store.Path()}
		result, err := modify.Execute(int(todoID), "Completed task (updated)", opts)

		// Verify status is preserved
		testutil.AssertNoError(t, err)
		testutil.AssertTodoHasStatus(t, result.Todo, models.StatusDone)

		// Verify in persistence
		collection, err = store.Load()
		testutil.AssertNoError(t, err)
		todo := testutil.AssertTodoByID(t, collection.Todos, todoID)
		testutil.AssertTodoHasStatus(t, todo, models.StatusDone)
	})

	t.Run("returns error for non-existent todo", func(t *testing.T) {
		// Create store with one todo
		store := testutil.CreatePopulatedStore(t, "Existing todo")

		// Try to modify non-existent todo
		opts := modify.Options{CollectionPath: store.Path()}
		result, err := modify.Execute(999, "New text", opts)

		// Verify error
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "todo not found")

		// Verify no changes were made
		collection, err := store.Load()
		testutil.AssertNoError(t, err)
		testutil.AssertCollectionSize(t, collection, 1)
		testutil.AssertTodoInList(t, collection.Todos, "Existing todo")
	})

	t.Run("returns error for empty new text", func(t *testing.T) {
		// Create store with test todo
		store := testutil.CreatePopulatedStore(t, "Original text")

		// Get the todo's ID
		collection, _ := store.Load()
		todoID := collection.Todos[0].ID

		// Try to modify with empty text
		opts := modify.Options{CollectionPath: store.Path()}
		result, err := modify.Execute(int(todoID), "", opts)

		// Verify error
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "new todo text cannot be empty")

		// Verify no changes were made
		collection, err = store.Load()
		testutil.AssertNoError(t, err)
		todo := testutil.AssertTodoByID(t, collection.Todos, todoID)
		assert.Equal(t, "Original text", todo.Text)
	})

	t.Run("handles modification of multiple todos correctly", func(t *testing.T) {
		// Create store with multiple todos
		store := testutil.CreatePopulatedStore(t,
			"First todo",
			"Second todo",
			"Third todo",
		)

		// Get the second todo's ID
		collection, _ := store.Load()
		secondTodoID := collection.Todos[1].ID

		// Modify only the second todo
		opts := modify.Options{CollectionPath: store.Path()}
		result, err := modify.Execute(int(secondTodoID), "Second todo (modified)", opts)

		// Verify correct todo was modified
		testutil.AssertNoError(t, err)
		assert.Equal(t, "Second todo", result.OldText)
		assert.Equal(t, "Second todo (modified)", result.NewText)

		// Verify persistence and other todos unchanged
		collection, err = store.Load()
		testutil.AssertNoError(t, err)
		testutil.AssertCollectionSize(t, collection, 3)
		testutil.AssertTodoInList(t, collection.Todos, "First todo")
		testutil.AssertTodoInList(t, collection.Todos, "Second todo (modified)")
		testutil.AssertTodoInList(t, collection.Todos, "Third todo")
	})
}
