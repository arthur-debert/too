package toggle_test

import (
	"testing"

	"github.com/arthur-debert/tdh/pkg/tdh"
	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/arthur-debert/tdh/pkg/tdh/testutil"
	"github.com/stretchr/testify/assert"
)

func TestToggleCommand(t *testing.T) {
	// Create a store with a pending todo
	store := testutil.CreatePopulatedStore(t, "Todo to toggle")

	// Get the todo ID (it will be 1 since it's the first todo)
	collection, _ := store.Load()
	todoID := collection.Todos[0].ID

	// Toggle the todo
	toggleOpts := tdh.ToggleOptions{CollectionPath: store.Path()}
	toggleResult, err := tdh.Toggle(int(todoID), toggleOpts)

	testutil.AssertNoError(t, err)
	assert.Equal(t, string(models.StatusDone), toggleResult.NewStatus)

	// Verify it was saved using testutil
	collection, err = store.Load()
	testutil.AssertNoError(t, err)

	// Use testutil to find and verify the todo
	todo := testutil.AssertTodoByID(t, collection.Todos, todoID)
	testutil.AssertTodoHasStatus(t, todo, models.StatusDone)
}
