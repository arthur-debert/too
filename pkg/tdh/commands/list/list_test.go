package list_test

import (
	"testing"

	"github.com/arthur-debert/tdh/pkg/tdh/commands/list"
	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/arthur-debert/tdh/pkg/tdh/testutil"
	"github.com/stretchr/testify/assert"
)

func TestListCommand(t *testing.T) {
	t.Run("lists only pending todos by default", func(t *testing.T) {
		// Create store with mixed todo states using testutil
		store := testutil.CreateStoreWithSpecs(t, []testutil.TodoSpec{
			{Text: "Pending task 1", Status: models.StatusPending},
			{Text: "Done task 1", Status: models.StatusDone},
			{Text: "Pending task 2", Status: models.StatusPending},
			{Text: "Done task 2", Status: models.StatusDone},
		})

		// Execute list command with default options
		opts := list.Options{
			CollectionPath: store.Path(),
			ShowDone:       false,
			ShowAll:        false,
		}
		result, err := list.Execute(opts)

		// Verify using testutil assertions
		testutil.AssertNoError(t, err)
		assert.Equal(t, 2, len(result.Todos))
		assert.Equal(t, 4, result.TotalCount)
		assert.Equal(t, 2, result.DoneCount)

		// Verify only pending todos are returned
		for _, todo := range result.Todos {
			testutil.AssertTodoHasStatus(t, todo, models.StatusPending)
		}
		testutil.AssertTodoInList(t, result.Todos, "Pending task 1")
		testutil.AssertTodoInList(t, result.Todos, "Pending task 2")
	})

	t.Run("lists only done todos with ShowDone flag", func(t *testing.T) {
		// Create store with mixed todo states
		store := testutil.CreateStoreWithSpecs(t, []testutil.TodoSpec{
			{Text: "Pending task", Status: models.StatusPending},
			{Text: "Done task 1", Status: models.StatusDone},
			{Text: "Done task 2", Status: models.StatusDone},
		})

		// Execute list command with ShowDone flag
		opts := list.Options{
			CollectionPath: store.Path(),
			ShowDone:       true,
			ShowAll:        false,
		}
		result, err := list.Execute(opts)

		// Verify results
		testutil.AssertNoError(t, err)
		assert.Equal(t, 2, len(result.Todos))
		assert.Equal(t, 3, result.TotalCount)
		assert.Equal(t, 2, result.DoneCount)

		// Verify only done todos are returned
		for _, todo := range result.Todos {
			testutil.AssertTodoHasStatus(t, todo, models.StatusDone)
		}
		testutil.AssertTodoInList(t, result.Todos, "Done task 1")
		testutil.AssertTodoInList(t, result.Todos, "Done task 2")
	})

	t.Run("lists all todos with ShowAll flag", func(t *testing.T) {
		// Create store with mixed todo states
		store := testutil.CreateStoreWithSpecs(t, []testutil.TodoSpec{
			{Text: "Pending task 1", Status: models.StatusPending},
			{Text: "Done task 1", Status: models.StatusDone},
			{Text: "Pending task 2", Status: models.StatusPending},
			{Text: "Done task 2", Status: models.StatusDone},
		})

		// Execute list command with ShowAll flag
		opts := list.Options{
			CollectionPath: store.Path(),
			ShowDone:       false, // This should be ignored when ShowAll is true
			ShowAll:        true,
		}
		result, err := list.Execute(opts)

		// Verify all todos are returned
		testutil.AssertNoError(t, err)
		assert.Equal(t, 4, len(result.Todos))
		assert.Equal(t, 4, result.TotalCount)
		assert.Equal(t, 2, result.DoneCount)

		// Verify all todos are present
		testutil.AssertTodoInList(t, result.Todos, "Pending task 1")
		testutil.AssertTodoInList(t, result.Todos, "Done task 1")
		testutil.AssertTodoInList(t, result.Todos, "Pending task 2")
		testutil.AssertTodoInList(t, result.Todos, "Done task 2")
	})

	t.Run("handles empty collection gracefully", func(t *testing.T) {
		// Create empty store
		store := testutil.CreatePopulatedStore(t) // Empty store

		// Execute list command
		opts := list.Options{
			CollectionPath: store.Path(),
			ShowDone:       false,
			ShowAll:        false,
		}
		result, err := list.Execute(opts)

		// Verify empty results
		testutil.AssertNoError(t, err)
		assert.Empty(t, result.Todos)
		assert.Equal(t, 0, result.TotalCount)
		assert.Equal(t, 0, result.DoneCount)
	})

	t.Run("returns counts even when filtering", func(t *testing.T) {
		// Create store with specific todo distribution
		store := testutil.CreateStoreWithSpecs(t, []testutil.TodoSpec{
			{Text: "Pending 1", Status: models.StatusPending},
			{Text: "Pending 2", Status: models.StatusPending},
			{Text: "Pending 3", Status: models.StatusPending},
			{Text: "Done 1", Status: models.StatusDone},
			{Text: "Done 2", Status: models.StatusDone},
		})

		// Test with pending filter
		opts := list.Options{
			CollectionPath: store.Path(),
			ShowDone:       false,
			ShowAll:        false,
		}
		result, err := list.Execute(opts)

		testutil.AssertNoError(t, err)
		assert.Equal(t, 3, len(result.Todos)) // Only pending todos returned
		assert.Equal(t, 5, result.TotalCount) // But total count includes all
		assert.Equal(t, 2, result.DoneCount)  // And done count is accurate
	})

	t.Run("hides pending children under a done parent", func(t *testing.T) {
		store := testutil.CreateStoreWithNestedSpecs(t, []testutil.TodoSpec{
			{Text: "Done Parent", Status: models.StatusDone, Children: []testutil.TodoSpec{
				{Text: "Pending Child", Status: models.StatusPending},
			}},
			{Text: "Pending Sibling", Status: models.StatusPending},
		})

		// Default list should hide the done parent and its pending child
		opts := list.Options{CollectionPath: store.Path()}
		result, err := list.Execute(opts)

		testutil.AssertNoError(t, err)
		assert.Equal(t, 1, len(result.Todos), "Only the pending sibling should be visible")
		assert.Equal(t, "Pending Sibling", result.Todos[0].Text)
		assert.Equal(t, 3, result.TotalCount, "Total count should include all items")
		assert.Equal(t, 1, result.DoneCount, "Done count should include the parent")
	})

	t.Run("ShowAll reveals pending children under a done parent", func(t *testing.T) {
		store := testutil.CreateStoreWithNestedSpecs(t, []testutil.TodoSpec{
			{Text: "Done Parent", Status: models.StatusDone, Children: []testutil.TodoSpec{
				{Text: "Pending Child", Status: models.StatusPending},
			}},
		})

		// --all should show the inconsistent state
		opts := list.Options{CollectionPath: store.Path(), ShowAll: true}
		result, err := list.Execute(opts)

		testutil.AssertNoError(t, err)
		assert.Equal(t, 1, len(result.Todos), "Should show the top-level parent")
		assert.Equal(t, "Done Parent", result.Todos[0].Text)
		assert.Equal(t, 1, len(result.Todos[0].Items), "Should reveal the child")
		assert.Equal(t, "Pending Child", result.Todos[0].Items[0].Text)
	})
}
