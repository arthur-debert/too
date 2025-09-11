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
		collection, _ := store.Load()
		collection.Reorder()
		testutil.AssertNoError(t, store.Save(collection))

		// Execute
		// Since done todos don't have positions in the new system, we need to use short ID
		// Find the done todo
		var doneTodo *models.Todo
		for _, todo := range collection.Todos {
			if todo.GetStatus() == models.StatusDone {
				doneTodo = todo
				break
			}
		}
		assert.NotNil(t, doneTodo, "Should have a done todo")
		shortID := doneTodo.ID[:8]  // Use first 8 chars as short ID
		
		opts := reopen.Options{CollectionPath: store.Path()}
		result, err := reopen.Execute(shortID, opts)

		// Assert
		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Test todo 1", result.Todo.Text)
		assert.Equal(t, "done", result.OldStatus)
		assert.Equal(t, "pending", result.NewStatus)
		assert.Equal(t, models.StatusPending, result.Todo.GetStatus())

		// Verify it was saved
		collection, err = store.Load()
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

		var child *models.Todo
		for _, todo := range collection.Todos {
			if todo.Text == "Parent todo" {
				child = todo.Items[0]
				break
			}
		}
		assert.NotNil(t, child)
		if child == nil {
			t.FailNow()
		}
		child.Statuses = map[string]string{"completion": string(models.StatusDone)}
		collection.Reorder()
		err = store.Save(collection)
		testutil.AssertNoError(t, err)

		// Execute - reopen child todo using short ID since it's done
		// and done todos don't have positions in the current implementation
		opts := reopen.Options{CollectionPath: store.Path()}
		childShortID := child.ID[:8]
		result, err := reopen.Execute(childShortID, opts)

		// Assert
		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Sub-task 1.1", result.Todo.Text)
		assert.Equal(t, "done", result.OldStatus)
		assert.Equal(t, "pending", result.NewStatus)
		assert.Equal(t, models.StatusPending, result.Todo.GetStatus())

		// Verify parent remains unchanged
		collection, err = store.Load()
		testutil.AssertNoError(t, err)
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
		assert.Equal(t, models.StatusPending, parent.GetStatus())

		// Verify only the specific child was reopened
		child = parent.Items[0]
		assert.Equal(t, models.StatusPending, child.GetStatus())
	})

	t.Run("reopen grandchild todo", func(t *testing.T) {
		// Setup - create nested structure
		store := testutil.CreateNestedStore(t)

		// Mark grandchild as done
		collection, err := store.Load()
		testutil.AssertNoError(t, err)
		var item *models.Todo
		for _, todo := range collection.Todos {
			if todo.Text == "Parent todo" {
				item = todo.Items[1].Items[0]
				break
			}
		}
		assert.NotNil(t, item)
		item.Statuses = map[string]string{"completion": string(models.StatusDone)}
		err = store.Save(collection)
		testutil.AssertNoError(t, err)

		// Execute - reopen grandchild using short ID since it's done
		opts := reopen.Options{CollectionPath: store.Path()}
		grandchildShortID := item.ID[:8]
		result, err := reopen.Execute(grandchildShortID, opts)

		// Assert
		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Grandchild 1.2.1", result.Todo.Text)
		assert.Equal(t, "done", result.OldStatus)
		assert.Equal(t, "pending", result.NewStatus)

		// Verify only the specific item was affected
		collection, err = store.Load()
		testutil.AssertNoError(t, err)

		for _, todo := range collection.Todos {
			if todo.Text == "Parent todo" {
				item = todo.Items[1].Items[0]
				break
			}
		}
		assert.NotNil(t, item)
		assert.Equal(t, models.StatusPending, item.GetStatus())

		// Verify no propagation happened
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
		assert.Equal(t, models.StatusPending, parent.GetStatus())
		assert.Equal(t, models.StatusPending, parent.Items[1].GetStatus())
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
		assert.Equal(t, models.StatusPending, result.Todo.GetStatus())
	})

	t.Run("reopen with parent done", func(t *testing.T) {
		// Setup - create nested structure
		store := testutil.CreateNestedStore(t)

		// Mark parent and child as done
		collection, err := store.Load()
		testutil.AssertNoError(t, err)
		var parent *models.Todo
		for _, todo := range collection.Todos {
			if todo.Text == "Parent todo" {
				parent = todo
				break
			}
		}
		assert.NotNil(t, parent)
		parent.Statuses = map[string]string{"completion": string(models.StatusDone)}
		child := parent.Items[0]
		child.Statuses = map[string]string{"completion": string(models.StatusDone)}
		err = store.Save(collection)
		testutil.AssertNoError(t, err)

		// Execute - reopen child when parent is done using short ID
		opts := reopen.Options{CollectionPath: store.Path()}
		childShortID := child.ID[:8]
		result, err := reopen.Execute(childShortID, opts)

		// Assert - should still work per spec (no propagation)
		testutil.AssertNoError(t, err)
		assert.Equal(t, "Sub-task 1.1", result.Todo.Text)
		assert.Equal(t, "done", result.OldStatus)
		assert.Equal(t, "pending", result.NewStatus)

		// Verify parent remains done
		collection, err = store.Load()
		testutil.AssertNoError(t, err)
		for _, todo := range collection.Todos {
			if todo.Text == "Parent todo" {
				parent = todo
				break
			}
		}
		assert.NotNil(t, parent)
		assert.Equal(t, models.StatusDone, parent.GetStatus())

		// Verify child is now pending
		child = parent.Items[0]
		assert.Equal(t, models.StatusPending, child.GetStatus())
	})
}
