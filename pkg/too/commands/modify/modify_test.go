package modify_test

import (
	"testing"

	"github.com/arthur-debert/too/pkg/too/commands/modify"
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/testutil"
	"github.com/stretchr/testify/assert"
)

func TestModifyCommand(t *testing.T) {
	t.Run("successfully modifies todo text", func(t *testing.T) {
		// Create store with test todos using testutil
		store := testutil.CreatePopulatedStore(t, "Original todo text", "Another todo")

		// Execute modify command
		opts := modify.Options{CollectionPath: store.Path()}
		result, err := modify.Execute("1", "Modified todo text", opts)

		// Verify result
		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Original todo text", result.OldText)
		assert.Equal(t, "Modified todo text", result.NewText)
		assert.Equal(t, "Modified todo text", result.Todo.Text)

		// Verify persistence using testutil
		collection, err := store.LoadIDM()
		testutil.AssertNoError(t, err)
		testutil.AssertTodoInList(t, collection.Items, "Modified todo text")
		testutil.AssertTodoInList(t, collection.Items, "Another todo")
	})

	t.Run("preserves todo status when modifying", func(t *testing.T) {
		// Create store with a done todo
		store := testutil.CreateStoreWithSpecs(t, []testutil.TodoSpec{
			{Text: "Completed task", Status: models.StatusDone},
			{Text: "Active task", Status: models.StatusPending},
		})

		// Modify the active task
		opts := modify.Options{CollectionPath: store.Path()}
		result, err := modify.Execute("1", "Active task (updated)", opts)

		// Verify status is preserved
		testutil.AssertNoError(t, err)
		testutil.AssertTodoHasStatus(t, result.Todo, models.StatusPending)
		assert.Equal(t, "Active task (updated)", result.Todo.Text)

		// Verify in persistence
		collection, err := store.LoadIDM()
		testutil.AssertNoError(t, err)
		
		// Find the modified todo by text
		var modifiedTodo *models.IDMTodo
		for _, todo := range collection.Items {
			if todo.Text == "Active task (updated)" {
				modifiedTodo = todo
				break
			}
		}
		assert.NotNil(t, modifiedTodo)
		testutil.AssertTodoHasStatus(t, modifiedTodo, models.StatusPending)
	})

	t.Run("returns error for non-existent todo", func(t *testing.T) {
		// Create store with one todo
		store := testutil.CreatePopulatedStore(t, "Existing todo")

		// Try to modify non-existent todo
		opts := modify.Options{CollectionPath: store.Path()}
		result, err := modify.Execute("999", "New text", opts)

		// Verify error
		assert.Error(t, err)
		assert.Nil(t, result)

		// Verify no changes were made
		collection, err := store.LoadIDM()
		testutil.AssertNoError(t, err)
		testutil.AssertCollectionSize(t, collection, 1)
		testutil.AssertTodoInList(t, collection.Items, "Existing todo")
	})

	t.Run("returns error for empty new text", func(t *testing.T) {
		// Create store with test todo
		store := testutil.CreatePopulatedStore(t, "Original text")

		// Try to modify with empty text
		opts := modify.Options{CollectionPath: store.Path()}
		result, err := modify.Execute("1", "", opts)

		// Verify error
		assert.Error(t, err)
		assert.Nil(t, result)

		// Verify no changes were made
		collection, err := store.LoadIDM()
		testutil.AssertNoError(t, err)
		testutil.AssertTodoInList(t, collection.Items, "Original text")
	})

	t.Run("handles modification of multiple todos correctly", func(t *testing.T) {
		// Create store with multiple todos
		store := testutil.CreatePopulatedStore(t,
			"First todo",
			"Second todo", 
			"Third todo",
		)

		// Modify only the second todo
		opts := modify.Options{CollectionPath: store.Path()}
		result, err := modify.Execute("2", "Second todo (modified)", opts)

		// Verify correct todo was modified
		testutil.AssertNoError(t, err)
		assert.Equal(t, "Second todo", result.OldText)
		assert.Equal(t, "Second todo (modified)", result.NewText)

		// Verify persistence and other todos unchanged
		collection, err := store.LoadIDM()
		testutil.AssertNoError(t, err)
		testutil.AssertCollectionSize(t, collection, 3)
		testutil.AssertTodoInList(t, collection.Items, "First todo")
		testutil.AssertTodoInList(t, collection.Items, "Second todo (modified)")
		testutil.AssertTodoInList(t, collection.Items, "Third todo")
	})

	t.Run("handles nested todo modification", func(t *testing.T) {
		// Create store with nested todos
		store := testutil.CreateStoreWithNestedSpecs(t, []testutil.TodoSpec{
			{Text: "Parent todo", Status: models.StatusPending, Children: []testutil.TodoSpec{
				{Text: "Child todo", Status: models.StatusPending},
			}},
		})

		// Modify the child todo
		opts := modify.Options{CollectionPath: store.Path()}
		result, err := modify.Execute("1.1", "Child todo (modified)", opts)

		// Verify correct todo was modified
		testutil.AssertNoError(t, err)
		assert.Equal(t, "Child todo", result.OldText)
		assert.Equal(t, "Child todo (modified)", result.NewText)

		// Verify persistence
		collection, err := store.LoadIDM()
		testutil.AssertNoError(t, err)
		testutil.AssertTodoInList(t, collection.Items, "Child todo (modified)")
		testutil.AssertTodoInList(t, collection.Items, "Parent todo")
	})
}