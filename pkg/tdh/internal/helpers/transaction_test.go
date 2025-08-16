package helpers_test

import (
	"fmt"
	"testing"

	"github.com/arthur-debert/tdh/pkg/tdh/internal/helpers"
	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/arthur-debert/tdh/pkg/tdh/testutil"
	"github.com/stretchr/testify/assert"
)

func TestTransactOnTodo(t *testing.T) {
	t.Run("successfully executes action on todo", func(t *testing.T) {
		// Create a store with test todos
		store := testutil.CreateStoreWithSpecs(t, []testutil.TodoSpec{
			{Text: "First todo", Status: models.StatusPending},
			{Text: "Second todo", Status: models.StatusPending},
		})

		// Define an action that modifies the todo
		actionExecuted := false
		err := helpers.TransactOnTodo(store.Path(), 1, func(todo *models.Todo, collection *models.Collection) error {
			actionExecuted = true
			todo.Text = "Modified first todo"
			return nil
		})

		// Verify no error and action was executed
		testutil.AssertNoError(t, err)
		assert.True(t, actionExecuted)

		// Verify the change was persisted
		collection, err := store.Load()
		testutil.AssertNoError(t, err)
		assert.Equal(t, "Modified first todo", collection.Todos[0].Text)
	})

	t.Run("returns error when todo not found", func(t *testing.T) {
		store := testutil.CreatePopulatedStore(t, "Single todo")

		err := helpers.TransactOnTodo(store.Path(), 999, func(todo *models.Todo, collection *models.Collection) error {
			t.Fatal("Action should not be called for non-existent todo")
			return nil
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "todo with position 999 was not found")
	})

	t.Run("rolls back changes when action returns error", func(t *testing.T) {
		store := testutil.CreatePopulatedStore(t, "Original text")

		// Define an action that modifies todo but returns error
		err := helpers.TransactOnTodo(store.Path(), 1, func(todo *models.Todo, collection *models.Collection) error {
			todo.Text = "This should be rolled back"
			return fmt.Errorf("simulated error")
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "simulated error")

		// Verify the change was NOT persisted
		collection, err := store.Load()
		testutil.AssertNoError(t, err)
		assert.Equal(t, "Original text", collection.Todos[0].Text)
	})

	t.Run("provides access to collection for operations like reorder", func(t *testing.T) {
		store := testutil.CreateStoreWithSpecs(t, []testutil.TodoSpec{
			{Text: "First", Status: models.StatusPending},
			{Text: "Second", Status: models.StatusPending},
			{Text: "Third", Status: models.StatusPending},
		})

		// Manually set non-sequential positions
		collection, _ := store.Load()
		collection.Todos[0].Position = 1
		collection.Todos[1].Position = 5
		collection.Todos[2].Position = 8
		err := store.Save(collection)
		testutil.AssertNoError(t, err)

		// Action that modifies todo and reorders collection
		err = helpers.TransactOnTodo(store.Path(), 5, func(todo *models.Todo, collection *models.Collection) error {
			todo.Status = models.StatusDone
			collection.Reorder()
			return nil
		})

		testutil.AssertNoError(t, err)

		// Verify both the status change and reorder happened
		collection, err = store.Load()
		testutil.AssertNoError(t, err)

		// After reordering, slice should have active items first (pos 1, 2), then done item (pos 0)
		assert.Equal(t, "First", collection.Todos[0].Text)
		assert.Equal(t, models.StatusPending, collection.Todos[0].Status)
		assert.Equal(t, 1, collection.Todos[0].Position)

		assert.Equal(t, "Third", collection.Todos[1].Text)
		assert.Equal(t, models.StatusPending, collection.Todos[1].Status)
		assert.Equal(t, 2, collection.Todos[1].Position)

		assert.Equal(t, "Second", collection.Todos[2].Text)
		assert.Equal(t, models.StatusDone, collection.Todos[2].Status)
		assert.Equal(t, 0, collection.Todos[2].Position)
	})
}
