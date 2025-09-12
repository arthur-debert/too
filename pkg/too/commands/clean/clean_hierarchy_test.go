package clean_test

import (
	"testing"

	"github.com/arthur-debert/too/pkg/too/commands/clean"
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/testutil"
	"github.com/stretchr/testify/assert"
)

func TestClean_IDMBased(t *testing.T) {
	t.Run("should remove done todos only", func(t *testing.T) {
		// Create a simple mix of done and pending todos
		store := testutil.CreateStoreWithSpecs(t, []testutil.TodoSpec{
			{Text: "Done task 1", Status: models.StatusDone},
			{Text: "Pending task 1", Status: models.StatusPending},
			{Text: "Done task 2", Status: models.StatusDone},
			{Text: "Pending task 2", Status: models.StatusPending},
		})

		// Execute clean
		result, err := clean.Execute(clean.Options{
			CollectionPath: store.Path(),
		})
		testutil.AssertNoError(t, err)

		// Should have removed 2 done tasks
		assert.Equal(t, 2, result.RemovedCount)
		assert.Equal(t, 2, result.ActiveCount)

		// Verify the removed todos
		assert.Equal(t, 2, len(result.RemovedTodos))
		testutil.AssertTodoInList(t, result.RemovedTodos, "Done task 1")
		testutil.AssertTodoInList(t, result.RemovedTodos, "Done task 2")

		// Verify only pending tasks remain
		collection, err := store.LoadIDM()
		testutil.AssertNoError(t, err)
		assert.Equal(t, 2, len(collection.Items))
		testutil.AssertTodoInList(t, collection.Items, "Pending task 1")
		testutil.AssertTodoInList(t, collection.Items, "Pending task 2")
	})

	t.Run("should handle empty collection", func(t *testing.T) {
		// Create empty store
		store := testutil.CreatePopulatedStore(t)

		// Execute clean on empty collection
		result, err := clean.Execute(clean.Options{
			CollectionPath: store.Path(),
		})
		testutil.AssertNoError(t, err)

		assert.Equal(t, 0, result.RemovedCount)
		assert.Equal(t, 0, result.ActiveCount)
		assert.Equal(t, 0, len(result.RemovedTodos))
	})

	t.Run("should handle no done items", func(t *testing.T) {
		// Create store with only pending todos
		store := testutil.CreateStoreWithSpecs(t, []testutil.TodoSpec{
			{Text: "Pending task 1", Status: models.StatusPending},
			{Text: "Pending task 2", Status: models.StatusPending},
		})

		// Execute clean
		result, err := clean.Execute(clean.Options{
			CollectionPath: store.Path(),
		})
		testutil.AssertNoError(t, err)

		assert.Equal(t, 0, result.RemovedCount)
		assert.Equal(t, 2, result.ActiveCount)

		// Verify nothing changed
		collection, err := store.LoadIDM()
		testutil.AssertNoError(t, err)
		assert.Equal(t, 2, len(collection.Items))
	})

	t.Run("should handle nested done parent with pending children", func(t *testing.T) {
		// Create nested structure with done parent
		store := testutil.CreateStoreWithNestedSpecs(t, []testutil.TodoSpec{
			{Text: "Done parent", Status: models.StatusDone, Children: []testutil.TodoSpec{
				{Text: "Pending child", Status: models.StatusPending},
			}},
			{Text: "Pending sibling", Status: models.StatusPending},
		})

		// Execute clean
		result, err := clean.Execute(clean.Options{
			CollectionPath: store.Path(),
		})
		testutil.AssertNoError(t, err)

		// IDM clean removes done todos and their descendants
		// So we should have removed 2 items (done parent + pending child)
		// But only report the done parent in removed count
		assert.Equal(t, 2, result.RemovedCount)
		assert.Equal(t, 1, result.ActiveCount)

		// Verify only the sibling remains
		collection, err := store.LoadIDM()
		testutil.AssertNoError(t, err)
		assert.Equal(t, 1, len(collection.Items))
		testutil.AssertTodoInList(t, collection.Items, "Pending sibling")
	})
}