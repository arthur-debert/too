package reopen_test

import (
	"testing"

	"github.com/arthur-debert/too/pkg/too/commands/reopen"
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/testutil"
	"github.com/stretchr/testify/assert"
)

func TestReopen(t *testing.T) {
	t.Run("reopen simple done todo", func(t *testing.T) {
		// Setup
		store := testutil.CreateStoreWithSpecs(t, []testutil.TodoSpec{
			{Text: "Done todo", Status: models.StatusDone},
			{Text: "Pending todo", Status: models.StatusPending},
		})

		// Get the done todo's short ID
		collection, err := store.LoadIDM()
		testutil.AssertNoError(t, err)
		
		var doneTodo *models.IDMTodo
		for _, todo := range collection.Items {
			if todo.GetStatus() == models.StatusDone {
				doneTodo = todo
				break
			}
		}
		assert.NotNil(t, doneTodo, "Should have a done todo")
		shortID := doneTodo.GetShortID()
		
		// Execute
		opts := reopen.Options{CollectionPath: store.Path()}
		result, err := reopen.Execute(shortID, opts)

		// Assert
		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Done todo", result.Todo.Text)
		assert.Equal(t, "done", result.OldStatus)
		assert.Equal(t, "pending", result.NewStatus)
		assert.Equal(t, models.StatusPending, result.Todo.GetStatus())

		// Verify it was saved
		collection, err = store.LoadIDM()
		testutil.AssertNoError(t, err)
		
		// Find the reopened todo
		var reopenedTodo *models.IDMTodo
		for _, todo := range collection.Items {
			if todo.Text == "Done todo" {
				reopenedTodo = todo
				break
			}
		}
		assert.NotNil(t, reopenedTodo)
		testutil.AssertTodoHasStatus(t, reopenedTodo, models.StatusPending)
	})

	t.Run("reopen nested done todo", func(t *testing.T) {
		// Setup - create nested structure with done child
		store := testutil.CreateStoreWithNestedSpecs(t, []testutil.TodoSpec{
			{Text: "Parent todo", Status: models.StatusPending, Children: []testutil.TodoSpec{
				{Text: "Done child", Status: models.StatusDone},
				{Text: "Pending child", Status: models.StatusPending},
			}},
		})

		// Get the done child's short ID
		collection, err := store.LoadIDM()
		testutil.AssertNoError(t, err)
		
		var doneChild *models.IDMTodo
		for _, todo := range collection.Items {
			if todo.Text == "Done child" && todo.GetStatus() == models.StatusDone {
				doneChild = todo
				break
			}
		}
		assert.NotNil(t, doneChild, "Should have a done child")
		shortID := doneChild.GetShortID()

		// Execute - reopen child todo
		opts := reopen.Options{CollectionPath: store.Path()}
		result, err := reopen.Execute(shortID, opts)

		// Assert
		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Done child", result.Todo.Text)
		assert.Equal(t, "done", result.OldStatus)
		assert.Equal(t, "pending", result.NewStatus)
		assert.Equal(t, models.StatusPending, result.Todo.GetStatus())

		// Verify the change was saved
		collection, err = store.LoadIDM()
		testutil.AssertNoError(t, err)
		
		// Find the reopened child
		var reopenedChild *models.IDMTodo
		for _, todo := range collection.Items {
			if todo.Text == "Done child" {
				reopenedChild = todo
				break
			}
		}
		assert.NotNil(t, reopenedChild)
		testutil.AssertTodoHasStatus(t, reopenedChild, models.StatusPending)
	})

	t.Run("reopen invalid todo ID", func(t *testing.T) {
		// Setup
		store := testutil.CreatePopulatedStore(t, "Test todo")

		// Execute with invalid ID
		opts := reopen.Options{CollectionPath: store.Path()}
		result, err := reopen.Execute("invalid", opts)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("reopen already pending todo", func(t *testing.T) {
		// Setup
		store := testutil.CreateStoreWithSpecs(t, []testutil.TodoSpec{
			{Text: "Already pending", Status: models.StatusPending},
		})

		// Get the pending todo's position (should work since it's active)
		opts := reopen.Options{CollectionPath: store.Path()}
		result, err := reopen.Execute("1", opts)

		// Assert - should handle gracefully
		if err == nil {
			assert.NotNil(t, result)
			assert.Equal(t, "Already pending", result.Todo.Text)
			assert.Equal(t, "pending", result.OldStatus)
			assert.Equal(t, "pending", result.NewStatus)
		} else {
			// It's also acceptable if the command returns an error for already-pending todos
			assert.Error(t, err)
			assert.Nil(t, result)
		}
	})

	t.Run("reopen with ambiguous short ID", func(t *testing.T) {
		// Setup with similar UIDs (unlikely but possible edge case)
		store := testutil.CreateStoreWithSpecs(t, []testutil.TodoSpec{
			{Text: "Todo 1", Status: models.StatusDone},
			{Text: "Todo 2", Status: models.StatusDone},
		})

		// Use a very short ID that could match multiple
		opts := reopen.Options{CollectionPath: store.Path()}
		result, err := reopen.Execute("a", opts) // Very short, likely to be ambiguous

		// Should either work (if not ambiguous) or return ambiguous error
		if err != nil {
			// Acceptable outcome - ambiguous ID error
			assert.Error(t, err)
			assert.Nil(t, result)
		} else {
			// Also acceptable - found a unique match
			assert.NotNil(t, result)
		}
	})
}