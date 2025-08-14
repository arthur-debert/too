package clean_test

import (
	"testing"

	"github.com/arthur-debert/tdh/pkg/tdh"
	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/arthur-debert/tdh/pkg/tdh/testutil"
	"github.com/stretchr/testify/assert"
)

func TestCleanCommand(t *testing.T) {
	// Create a store with mixed pending and done todos
	store := testutil.CreateStoreWithSpecs(t, []testutil.TodoSpec{
		{Text: "Pending todo", Status: models.StatusPending},
		{Text: "Done todo", Status: models.StatusDone},
	})

	// Run clean command
	cleanOpts := tdh.CleanOptions{CollectionPath: store.Path()}
	cleanResult, err := tdh.Clean(cleanOpts)

	testutil.AssertNoError(t, err)
	assert.Equal(t, 1, cleanResult.RemovedCount)
	assert.Equal(t, 1, cleanResult.ActiveCount)

	// Verify using testutil
	collection, err := store.Load()
	testutil.AssertNoError(t, err)

	testutil.AssertCollectionSize(t, collection, 1)
	testutil.AssertTodoInList(t, collection.Todos, "Pending todo")
	testutil.AssertTodoNotInList(t, collection.Todos, "Done todo")
}
