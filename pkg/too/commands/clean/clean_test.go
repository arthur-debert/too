package clean_test

import (
	"testing"

	"github.com/arthur-debert/too/pkg/too"
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/testutil"
	"github.com/stretchr/testify/assert"
)

func TestCleanCommand(t *testing.T) {
	t.Run("removes done todos and keeps pending", func(t *testing.T) {
		// Create a store with mixed pending and done todos
		store := testutil.CreateStoreWithSpecs(t, []testutil.TodoSpec{
			{Text: "Pending todo", Status: models.StatusPending},
			{Text: "Done todo", Status: models.StatusDone},
		})

		// Run clean command
		cleanOpts := too.CleanOptions{CollectionPath: store.Path()}
		cleanResult, err := too.Clean(cleanOpts)

		testutil.AssertNoError(t, err)
		assert.Equal(t, 1, cleanResult.RemovedCount)
		assert.Equal(t, 1, cleanResult.ActiveCount)

		// Verify using testutil
		collection, err := store.LoadIDM()
		testutil.AssertNoError(t, err)

		testutil.AssertCollectionSize(t, collection, 1)
		testutil.AssertTodoInList(t, collection.Items, "Pending todo")
		testutil.AssertTodoNotInList(t, collection.Items, "Done todo")
	})

	t.Run("auto-reorders remaining todos after clean", func(t *testing.T) {
		// Create store with non-sequential positions
		store := testutil.CreateStoreWithSpecs(t, []testutil.TodoSpec{
			{Text: "First pending", Status: models.StatusPending},
			{Text: "Done todo", Status: models.StatusDone},
			{Text: "Second pending", Status: models.StatusPending},
			{Text: "Another done", Status: models.StatusDone},
			{Text: "Third pending", Status: models.StatusPending},
		})

		// Run clean command
		cleanOpts := too.CleanOptions{CollectionPath: store.Path()}
		cleanResult, err := too.Clean(cleanOpts)

		testutil.AssertNoError(t, err)
		assert.Equal(t, 2, cleanResult.RemovedCount)
		assert.Equal(t, 3, cleanResult.ActiveCount)

		// Verify remaining todos
		collection, err := store.LoadIDM()
		testutil.AssertNoError(t, err)
		testutil.AssertCollectionSize(t, collection, 3)

		// Check that the pending todos remain (order may vary in flat structure)
		testutil.AssertTodoInList(t, collection.Items, "First pending")
		testutil.AssertTodoInList(t, collection.Items, "Second pending")
		testutil.AssertTodoInList(t, collection.Items, "Third pending")
	})

	t.Run("removes pending children of a done parent", func(t *testing.T) {
		store := testutil.CreateStoreWithNestedSpecs(t, []testutil.TodoSpec{
			{Text: "Done Parent", Status: models.StatusDone, Children: []testutil.TodoSpec{
				{Text: "Pending Child", Status: models.StatusPending},
			}},
			{Text: "Pending Sibling", Status: models.StatusPending},
		})

		// Run clean command
		cleanOpts := too.CleanOptions{CollectionPath: store.Path()}
		cleanResult, err := too.Clean(cleanOpts)

		testutil.AssertNoError(t, err)
		// IDM clean removes done todos and their descendants (2 items: parent + child)
		assert.Equal(t, 2, cleanResult.RemovedCount, "Should remove done parent and its pending child")
		assert.Equal(t, 1, cleanResult.ActiveCount, "Only the pending sibling should remain")

		// Verify the collection state
		collection, err := store.LoadIDM()
		testutil.AssertNoError(t, err)
		testutil.AssertCollectionSize(t, collection, 1)
		testutil.AssertTodoInList(t, collection.Items, "Pending Sibling")
	})
}
