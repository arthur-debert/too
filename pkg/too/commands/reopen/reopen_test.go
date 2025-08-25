package reopen_test

import (
	"testing"

	"github.com/arthur-debert/too/pkg/too/commands/reopen"
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/testutil"
	"github.com/stretchr/testify/assert"
)

func TestReopen(t *testing.T) {
	t.Run("reopen simple todo", func(t *testing.T) {
		// Setup
		store := testutil.CreateStoreWithSpecs(t, []testutil.TodoSpec{
			{Text: "Test todo 1", Status: models.StatusDone},
			{Text: "Test todo 2", Status: models.StatusPending},
		})

		// Execute
		opts := reopen.Options{CollectionPath: store.Path()}
		result, err := reopen.Execute("1", opts)

		// Assert
		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Test todo 1", result.Todo.Text)
		assert.Equal(t, "done", result.OldStatus)
		assert.Equal(t, "pending", result.NewStatus)
		assert.Equal(t, models.StatusPending, result.Todo.Status)

		// Verify it was saved
		collection, err := store.Load()
		testutil.AssertNoError(t, err)
		todo := testutil.AssertTodoByPosition(t, collection.Todos, 1)
		testutil.AssertTodoHasStatus(t, todo, models.StatusPending)
		todo2 := testutil.AssertTodoByPosition(t, collection.Todos, 2)
		testutil.AssertTodoHasStatus(t, todo2, models.StatusPending)
	})

	t.Run("reopen nested todo", func(t *testing.T) {
		// Setup - create nested structure
		store := testutil.CreateNestedStore(t)

		// Mark a child as done first
		collection, err := store.Load()
		testutil.AssertNoError(t, err)
		child, err := collection.FindItemByPositionPath("1.1")
		testutil.AssertNoError(t, err)
		child.Status = models.StatusDone
		err = store.Save(collection)
		testutil.AssertNoError(t, err)

		// Execute - reopen child todo
		opts := reopen.Options{CollectionPath: store.Path()}
		result, err := reopen.Execute("1.1", opts)

		// Assert
		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Sub-task 1.1", result.Todo.Text)
		assert.Equal(t, "done", result.OldStatus)
		assert.Equal(t, "pending", result.NewStatus)
		assert.Equal(t, models.StatusPending, result.Todo.Status)

		// Verify parent remains unchanged
		collection, err = store.Load()
		testutil.AssertNoError(t, err)
		parent, err := collection.FindItemByPositionPath("1")
		testutil.AssertNoError(t, err)
		assert.Equal(t, models.StatusPending, parent.Status)

		// Verify only the specific child was reopened
		child, err = collection.FindItemByPositionPath("1.1")
		testutil.AssertNoError(t, err)
		assert.Equal(t, models.StatusPending, child.Status)
	})

	t.Run("reopen grandchild todo", func(t *testing.T) {
		// Setup - create nested structure
		store := testutil.CreateNestedStore(t)

		// Mark grandchild as done
		collection, err := store.Load()
		testutil.AssertNoError(t, err)
		item, err := collection.FindItemByPositionPath("1.2.1")
		testutil.AssertNoError(t, err)
		item.Status = models.StatusDone
		err = store.Save(collection)
		testutil.AssertNoError(t, err)

		// Execute - reopen grandchild
		opts := reopen.Options{CollectionPath: store.Path()}
		result, err := reopen.Execute("1.2.1", opts)

		// Assert
		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Grandchild 1.2.1", result.Todo.Text)
		assert.Equal(t, "done", result.OldStatus)
		assert.Equal(t, "pending", result.NewStatus)

		// Verify only the specific item was affected
		collection, err = store.Load()
		testutil.AssertNoError(t, err)

		item, err = collection.FindItemByPositionPath("1.2.1")
		testutil.AssertNoError(t, err)
		assert.Equal(t, models.StatusPending, item.Status)

		// Verify no propagation happened
		paths := []string{"1", "1.2"}
		for _, path := range paths {
			parent, err := collection.FindItemByPositionPath(path)
			testutil.AssertNoError(t, err)
			assert.Equal(t, models.StatusPending, parent.Status, "Parent at %s should remain unchanged", path)
		}
	})

	t.Run("reopen invalid position", func(t *testing.T) {
		// Setup
		store := testutil.CreatePopulatedStore(t, "Test todo")

		// Execute
		opts := reopen.Options{CollectionPath: store.Path()}
		result, err := reopen.Execute("99", opts)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "todo not found")
	})

	t.Run("reopen invalid position path format", func(t *testing.T) {
		// Setup
		store := testutil.CreateNestedStore(t)

		// Execute - invalid format with non-numeric part
		opts := reopen.Options{CollectionPath: store.Path()}
		result, err := reopen.Execute("1.a.2", opts)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "todo not found with reference")
	})

	t.Run("reopen already pending todo", func(t *testing.T) {
		// Setup
		store := testutil.CreatePopulatedStore(t, "Already pending")

		// Execute
		opts := reopen.Options{CollectionPath: store.Path()}
		result, err := reopen.Execute("1", opts)

		// Assert
		testutil.AssertNoError(t, err)
		assert.Equal(t, "Already pending", result.Todo.Text)
		assert.Equal(t, "pending", result.OldStatus)
		assert.Equal(t, "pending", result.NewStatus)
		assert.Equal(t, models.StatusPending, result.Todo.Status)
	})

	t.Run("reopen with parent done", func(t *testing.T) {
		// Setup - create nested structure
		store := testutil.CreateNestedStore(t)

		// Mark parent and child as done
		collection, err := store.Load()
		testutil.AssertNoError(t, err)
		parent, err := collection.FindItemByPositionPath("1")
		testutil.AssertNoError(t, err)
		parent.Status = models.StatusDone
		child, err := collection.FindItemByPositionPath("1.1")
		testutil.AssertNoError(t, err)
		child.Status = models.StatusDone
		err = store.Save(collection)
		testutil.AssertNoError(t, err)

		// Execute - reopen child when parent is done
		opts := reopen.Options{CollectionPath: store.Path()}
		result, err := reopen.Execute("1.1", opts)

		// Assert - should still work per spec (no propagation)
		testutil.AssertNoError(t, err)
		assert.Equal(t, "Sub-task 1.1", result.Todo.Text)
		assert.Equal(t, "done", result.OldStatus)
		assert.Equal(t, "pending", result.NewStatus)

		// Verify parent remains done
		collection, err = store.Load()
		testutil.AssertNoError(t, err)
		parent, err = collection.FindItemByPositionPath("1")
		testutil.AssertNoError(t, err)
		assert.Equal(t, models.StatusDone, parent.Status)

		// Verify child is now pending
		child, err = collection.FindItemByPositionPath("1.1")
		testutil.AssertNoError(t, err)
		assert.Equal(t, models.StatusPending, child.Status)
	})
}
