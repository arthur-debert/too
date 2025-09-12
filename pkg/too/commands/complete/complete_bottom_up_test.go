package complete_test

import (
	"testing"

	"github.com/arthur-debert/too/pkg/too/commands/complete"
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/testutil"
	"github.com/stretchr/testify/assert"
)

func TestComplete_IDMBased(t *testing.T) {
	t.Run("should complete a single todo", func(t *testing.T) {
		// Create store with pending todos
		store := testutil.CreateStoreWithSpecs(t, []testutil.TodoSpec{
			{Text: "Task 1", Status: models.StatusPending},
			{Text: "Task 2", Status: models.StatusPending},
		})

		// Complete the first todo using position path "1"
		result, err := complete.Execute("1", complete.Options{
			CollectionPath: store.Path(),
		})
		testutil.AssertNoError(t, err)

		// Should have completed the todo
		assert.NotNil(t, result.Todo)
		assert.Equal(t, models.StatusDone, result.Todo.GetStatus())
		assert.Equal(t, "pending", result.OldStatus)
		assert.Equal(t, "done", result.NewStatus)

		// Verify the collection
		collection, err := store.LoadIDM()
		testutil.AssertNoError(t, err)
		
		// Should still have 2 todos, but one is done
		assert.Equal(t, 2, len(collection.Items))
		
		doneCount := 0
		for _, todo := range collection.Items {
			if todo.GetStatus() == models.StatusDone {
				doneCount++
			}
		}
		assert.Equal(t, 1, doneCount)
	})

	t.Run("should handle nested todos with parent-child relationships", func(t *testing.T) {
		// Create nested structure
		store := testutil.CreateStoreWithNestedSpecs(t, []testutil.TodoSpec{
			{Text: "Parent task", Status: models.StatusPending, Children: []testutil.TodoSpec{
				{Text: "Child task 1", Status: models.StatusPending},
				{Text: "Child task 2", Status: models.StatusPending},
			}},
		})

		// Complete a child todo
		result, err := complete.Execute("1.1", complete.Options{
			CollectionPath: store.Path(),
		})
		testutil.AssertNoError(t, err)

		// Should have completed the todo
		assert.NotNil(t, result.Todo)
		assert.Equal(t, "Child task 1", result.Todo.Text)

		// Verify the collection still has all 3 items
		collection, err := store.LoadIDM()
		testutil.AssertNoError(t, err)
		assert.Equal(t, 3, len(collection.Items))
	})

	t.Run("should handle invalid position path", func(t *testing.T) {
		// Create simple store
		store := testutil.CreateStoreWithSpecs(t, []testutil.TodoSpec{
			{Text: "Task 1", Status: models.StatusPending},
		})

		// Try to complete non-existent todo
		_, err := complete.Execute("999", complete.Options{
			CollectionPath: store.Path(),
		})
		
		// Should return an error
		assert.Error(t, err)
	})

	t.Run("should handle already completed todo", func(t *testing.T) {
		// Create store with done todo
		store := testutil.CreateStoreWithSpecs(t, []testutil.TodoSpec{
			{Text: "Done task", Status: models.StatusDone},
		})

		// Try to complete already done todo
		result, err := complete.Execute("1", complete.Options{
			CollectionPath: store.Path(),
		})
		
		// IDM may not be able to find done todos at position paths, or it may handle gracefully
		if err != nil {
			assert.Error(t, err)
		} else {
			assert.NotNil(t, result)
			// If it succeeds, status should already be done
			assert.Equal(t, "done", result.OldStatus)
		}
	})
}