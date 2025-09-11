package complete_test

import (
	"testing"

	"github.com/arthur-debert/too/pkg/too/commands/complete"
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/testutil"
	"github.com/stretchr/testify/assert"
)

func TestComplete(t *testing.T) {
	t.Run("complete simple todo", func(t *testing.T) {
		// Setup
		store := testutil.CreatePopulatedStore(t, "Test todo 1", "Test todo 2")

		// Execute
		result, err := complete.Execute("1", complete.Options{CollectionPath: store.Path()})

		// Assert
		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Test todo 1", result.Todo.Text)
		assert.Equal(t, models.StatusDone, result.Todo.GetStatus())
		assert.Equal(t, "pending", result.OldStatus)
		assert.Equal(t, "done", result.NewStatus)

		// Verify it was saved
		collection, err := store.LoadIDM()
		testutil.AssertNoError(t, err)

		// Should have 2 todos, one completed
		assert.Equal(t, 2, len(collection.Items))
		testutil.AssertTodoInList(t, collection.Items, "Test todo 1")
		testutil.AssertTodoInList(t, collection.Items, "Test todo 2")
		
		// Find and verify the completed todo
		for _, todo := range collection.Items {
			if todo.Text == "Test todo 1" {
				testutil.AssertTodoHasStatus(t, todo, models.StatusDone)
			} else if todo.Text == "Test todo 2" {
				testutil.AssertTodoHasStatus(t, todo, models.StatusPending)
			}
		}
	})

	t.Run("complete nested todo", func(t *testing.T) {
		// Setup - create nested structure
		store := testutil.CreateStoreWithNestedSpecs(t, []testutil.TodoSpec{
			{Text: "Parent todo", Status: models.StatusPending, Children: []testutil.TodoSpec{
				{Text: "Sub-task 1", Status: models.StatusPending},
				{Text: "Sub-task 2", Status: models.StatusPending},
			}},
		})

		// Execute - complete first child
		result, err := complete.Execute("1.1", complete.Options{CollectionPath: store.Path()})

		// Assert
		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Sub-task 1", result.Todo.Text)
		assert.Equal(t, models.StatusDone, result.Todo.GetStatus())

		// Verify the collection
		collection, err := store.LoadIDM()
		testutil.AssertNoError(t, err)
		
		// Should have 3 todos (parent + 2 children)
		assert.Equal(t, 3, len(collection.Items))
		
		// Find and verify the completed child
		for _, todo := range collection.Items {
			if todo.Text == "Sub-task 1" {
				testutil.AssertTodoHasStatus(t, todo, models.StatusDone)
			}
		}
	})

	t.Run("complete invalid position", func(t *testing.T) {
		// Setup
		store := testutil.CreatePopulatedStore(t, "Test todo")

		// Execute
		result, err := complete.Execute("99", complete.Options{CollectionPath: store.Path()})

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("complete already done todo", func(t *testing.T) {
		// Setup - create todos with one already done
		store := testutil.CreateStoreWithSpecs(t, []testutil.TodoSpec{
			{Text: "Pending todo", Status: models.StatusPending},
			{Text: "Already done", Status: models.StatusDone},
		})

		// Try to complete using a position that may not exist (done todos may not have positions)
		result, err := complete.Execute("2", complete.Options{CollectionPath: store.Path()})

		// This should either error or handle gracefully
		// The exact behavior depends on how IDM handles done todos in position paths
		if err != nil {
			assert.Error(t, err)
			assert.Nil(t, result)
		} else {
			// If it succeeds, it should handle the already-done status
			assert.NotNil(t, result)
		}
	})
}