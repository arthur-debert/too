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
		opts := complete.Options{CollectionPath: store.Path()}
		result, err := complete.Execute("1", opts)

		// Assert
		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Test todo 1", result.Todo.Text)
		assert.Equal(t, "pending", result.OldStatus)
		assert.Equal(t, "done", result.NewStatus)
		assert.Equal(t, models.StatusDone, result.Todo.Status)

		// Verify it was saved
		collection, err := store.Load()
		testutil.AssertNoError(t, err)

		// With new behavior: slice is reordered with active items first
		// Active todo is now first in slice with position 1
		activeTodo := collection.Todos[0]
		assert.Equal(t, "Test todo 2", activeTodo.Text)
		assert.Equal(t, 1, activeTodo.Position)
		testutil.AssertTodoHasStatus(t, activeTodo, models.StatusPending)

		// Completed todo is now second in slice with position 0
		completedTodo := collection.Todos[1]
		assert.Equal(t, "Test todo 1", completedTodo.Text)
		assert.Equal(t, 0, completedTodo.Position)
		testutil.AssertTodoHasStatus(t, completedTodo, models.StatusDone)
	})

	t.Run("complete nested todo", func(t *testing.T) {
		// Setup - create nested structure
		store := testutil.CreateNestedStore(t)

		// Execute - complete child todo
		opts := complete.Options{CollectionPath: store.Path()}
		result, err := complete.Execute("1.1", opts)

		// Assert
		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Sub-task 1.1", result.Todo.Text)
		assert.Equal(t, "pending", result.OldStatus)
		assert.Equal(t, "done", result.NewStatus)
		assert.Equal(t, models.StatusDone, result.Todo.Status)

		// Verify parent is still pending
		collection, err := store.Load()
		testutil.AssertNoError(t, err)

		// Find parent by text
		var parent *models.Todo
		for _, todo := range collection.Todos {
			if todo.Text == "Parent todo" {
				parent = todo
				break
			}
		}
		assert.NotNil(t, parent)
		if parent == nil {
			t.FailNow()
		}
		assert.Equal(t, models.StatusPending, parent.Status)
		assert.Equal(t, 2, len(parent.Items))

		// Active sibling is now first in slice with position 1
		sibling := parent.Items[0]
		assert.Equal(t, "Sub-task 1.2", sibling.Text)
		assert.Equal(t, 1, sibling.Position)
		assert.Equal(t, models.StatusPending, sibling.Status)

		// Completed child is now second in slice with position 0
		completedChild := parent.Items[1]
		assert.Equal(t, "Sub-task 1.1", completedChild.Text)
		assert.Equal(t, 0, completedChild.Position)
		assert.Equal(t, models.StatusDone, completedChild.Status)
	})

	t.Run("complete grandchild todo", func(t *testing.T) {
		// Setup - create nested structure
		store := testutil.CreateNestedStore(t)

		// Execute - complete grandchild
		opts := complete.Options{CollectionPath: store.Path()}
		result, err := complete.Execute("1.2.1", opts)

		// Assert
		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Grandchild 1.2.1", result.Todo.Text)
		assert.Equal(t, "pending", result.OldStatus)
		assert.Equal(t, "done", result.NewStatus)

		// Verify the changes
		collection, err := store.Load()
		testutil.AssertNoError(t, err)

		// Verify bottom-up completion: 1.2 should be done (all children complete)
		// After reordering, 1.2 becomes 1.1 (since 1.1 was pending and 1.2 had higher position)
		var parent *models.Todo
		for _, todo := range collection.Todos {
			if todo.Text == "Parent todo" {
				parent = todo
				break
			}
		}
		assert.NotNil(t, parent)
		assert.NotNil(t, parent)
		if parent == nil {
			t.FailNow()
		}
		assert.Equal(t, 2, len(parent.Items))

		// Find the completed subtask (with grandchild) - it should have position 0
		var completedSubtask *models.Todo
		for _, item := range parent.Items {
			if item.Text == "Sub-task 1.2" {
				completedSubtask = item
				break
			}
		}
		assert.NotNil(t, completedSubtask)
		assert.Equal(t, 0, completedSubtask.Position)
		assert.Equal(t, models.StatusDone, completedSubtask.Status, "Parent at 1.2 should be done (bottom-up completion)")

		// The grandchild should also have position 0
		assert.Equal(t, 1, len(completedSubtask.Items))
		assert.Equal(t, 0, completedSubtask.Items[0].Position)
		assert.Equal(t, models.StatusDone, completedSubtask.Items[0].Status)

		// But top-level parent should remain pending (not all children complete)
		assert.Equal(t, models.StatusPending, parent.Status, "Parent at 1 should remain pending (1.1 still pending)")
	})

	t.Run("complete invalid position", func(t *testing.T) {
		// Setup
		store := testutil.CreatePopulatedStore(t, "Test todo")

		// Execute
		opts := complete.Options{CollectionPath: store.Path()}
		result, err := complete.Execute("99", opts)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "todo not found")
	})

	t.Run("complete already done todo", func(t *testing.T) {
		// Setup - create a pending todo and a done todo
		store := testutil.CreateStoreWithSpecs(t, []testutil.TodoSpec{
			{Text: "Pending todo", Status: models.StatusPending},
			{Text: "Already done", Status: models.StatusDone},
		})

		// Try to complete the done todo using its UID (since done todos don't have HIDs)
		// This test now expects an error because we can't reference done todos by position
		opts := complete.Options{CollectionPath: store.Path()}
		result, err := complete.Execute("2", opts)

		// Assert - should fail because position 2 doesn't exist (only 1 pending todo)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no item found at position")
		assert.Nil(t, result)
	})
}
