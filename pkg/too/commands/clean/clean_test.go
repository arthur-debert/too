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
		collection, err := store.Load()
		testutil.AssertNoError(t, err)

		testutil.AssertCollectionSize(t, collection, 1)
		testutil.AssertTodoInList(t, collection.Todos, "Pending todo")
		testutil.AssertTodoNotInList(t, collection.Todos, "Done todo")
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

		// Save the collection
		collection, _ := store.Load()
		err := store.Save(collection)
		testutil.AssertNoError(t, err)

		// Run clean command
		cleanOpts := too.CleanOptions{CollectionPath: store.Path()}
		cleanResult, err := too.Clean(cleanOpts)

		testutil.AssertNoError(t, err)
		assert.Equal(t, 2, cleanResult.RemovedCount)
		assert.Equal(t, 3, cleanResult.ActiveCount)

		// Verify remaining todos have sequential positions
		collection, err = store.Load()
		testutil.AssertNoError(t, err)
		testutil.AssertCollectionSize(t, collection, 3)

		// Check that positions are now 1, 2, 3
		assert.Equal(t, "First pending", collection.Todos[0].Text)
		assert.Equal(t, "Second pending", collection.Todos[1].Text)
		assert.Equal(t, "Third pending", collection.Todos[2].Text)
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
		assert.Equal(t, 1, cleanResult.RemovedCount, "Should report the Done Parent as removed")
		assert.Equal(t, 1, cleanResult.ActiveCount, "Only the pending sibling should remain")

		// Verify the collection state
		collection, err := store.Load()
		testutil.AssertNoError(t, err)
		testutil.AssertCollectionSize(t, collection, 1)
		assert.Equal(t, "Pending Sibling", collection.Todos[0].Text)
	})
}
