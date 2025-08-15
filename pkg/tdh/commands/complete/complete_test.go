package complete_test

import (
	"testing"

	"github.com/arthur-debert/tdh/pkg/tdh/commands/complete"
	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/arthur-debert/tdh/pkg/tdh/testutil"
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
		todo := testutil.AssertTodoByPosition(t, collection.Todos, 1)
		testutil.AssertTodoHasStatus(t, todo, models.StatusDone)

		// Verify other todo is still pending
		todo2 := testutil.AssertTodoByPosition(t, collection.Todos, 2)
		testutil.AssertTodoHasStatus(t, todo2, models.StatusPending)
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
		parent, err := collection.FindItemByPositionPath("1")
		testutil.AssertNoError(t, err)
		assert.Equal(t, models.StatusPending, parent.Status)

		// Verify only the specific child was marked done
		child, err := collection.FindItemByPositionPath("1.1")
		testutil.AssertNoError(t, err)
		assert.Equal(t, models.StatusDone, child.Status)

		// Verify sibling is still pending
		sibling, err := collection.FindItemByPositionPath("1.2")
		testutil.AssertNoError(t, err)
		assert.Equal(t, models.StatusPending, sibling.Status)
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

		// Verify only the specific item was affected
		collection, err := store.Load()
		testutil.AssertNoError(t, err)

		item, err := collection.FindItemByPositionPath("1.2.1")
		testutil.AssertNoError(t, err)
		assert.Equal(t, models.StatusDone, item.Status)

		// Verify bottom-up completion: 1.2 should be done (all children complete)
		parent12, err := collection.FindItemByPositionPath("1.2")
		testutil.AssertNoError(t, err)
		assert.Equal(t, models.StatusDone, parent12.Status, "Parent at 1.2 should be done (bottom-up completion)")

		// But top-level parent should remain pending (not all children complete)
		parent1, err := collection.FindItemByPositionPath("1")
		testutil.AssertNoError(t, err)
		assert.Equal(t, models.StatusPending, parent1.Status, "Parent at 1 should remain pending (1.1 still pending)")
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

	t.Run("complete invalid position path format", func(t *testing.T) {
		// Setup
		store := testutil.CreateNestedStore(t)

		// Execute - invalid format with non-numeric part
		opts := complete.Options{CollectionPath: store.Path()}
		result, err := complete.Execute("1.a.2", opts)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid position")
	})

	t.Run("complete already done todo", func(t *testing.T) {
		// Setup
		store := testutil.CreateStoreWithSpecs(t, []testutil.TodoSpec{
			{Text: "Already done", Status: models.StatusDone},
		})

		// Execute
		opts := complete.Options{CollectionPath: store.Path()}
		result, err := complete.Execute("1", opts)

		// Assert
		testutil.AssertNoError(t, err)
		assert.Equal(t, "Already done", result.Todo.Text)
		assert.Equal(t, "done", result.OldStatus)
		assert.Equal(t, "done", result.NewStatus)
		assert.Equal(t, models.StatusDone, result.Todo.Status)
	})
}
