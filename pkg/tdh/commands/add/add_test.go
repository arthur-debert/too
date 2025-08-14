package add_test

import (
	"testing"

	"github.com/arthur-debert/tdh/pkg/tdh"
	"github.com/arthur-debert/tdh/pkg/tdh/testutil"
	"github.com/stretchr/testify/assert"
)

func TestAddCommand(t *testing.T) {
	// Use testutil to create a clean store
	store := testutil.CreatePopulatedStore(t) // Empty store

	opts := tdh.AddOptions{CollectionPath: store.Path()}
	result, err := tdh.Add("My first todo", opts)

	// Use testutil assertions
	testutil.AssertNoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "My first todo", result.Todo.Text)
	assert.Equal(t, int64(1), result.Todo.ID)

	// Verify it was saved using testutil
	collection, err := store.Load()
	testutil.AssertNoError(t, err)

	testutil.AssertCollectionSize(t, collection, 1)
	testutil.AssertTodoInList(t, collection.Todos, "My first todo")
}
