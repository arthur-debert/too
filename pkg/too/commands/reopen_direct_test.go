package commands_test

import (
	"testing"

	"github.com/arthur-debert/too/pkg/too/commands"
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReopenDirect(t *testing.T) {
	t.Run("reopen single todo", func(t *testing.T) {
		store := testutil.CreateStoreWithSpecs(t, []testutil.TodoSpec{
			{Text: "Completed todo", Status: models.StatusDone},
		})
		
		// Get the todo's short ID
		collection, _ := store.Load()
		shortID := collection.Todos[0].ID[:8] // Use first 8 chars as short ID
		
		reopened, err := commands.ReopenDirect(store, "file://test", shortID, false)
		require.NoError(t, err)
		assert.Len(t, reopened, 1)
		
		// Verify status changed
		collection, _ = store.Load()
		todo := collection.Todos[0]
		assert.Equal(t, models.StatusPending, todo.GetStatus())
	})

	t.Run("reopen todo with children recursively", func(t *testing.T) {
		// Create a nested structure with all done
		store := testutil.CreateStoreWithNestedSpecs(t, []testutil.TodoSpec{
			{
				Text:   "Parent done",
				Status: models.StatusDone,
				Children: []testutil.TodoSpec{
					{Text: "Child 1 done", Status: models.StatusDone},
					{
						Text:   "Child 2 done",
						Status: models.StatusDone,
						Children: []testutil.TodoSpec{
							{Text: "Grandchild done", Status: models.StatusDone},
						},
					},
				},
			},
		})
		
		// Get the parent's short ID
		collection, _ := store.Load()
		parentShortID := collection.Todos[0].ID[:8]
		
		// Reopen parent recursively
		reopened, err := commands.ReopenDirect(store, "file://test", parentShortID, true)
		require.NoError(t, err)
		assert.Len(t, reopened, 4) // Parent + 2 children + 1 grandchild
		
		// Verify all are pending
		collection, _ = store.Load()
		collection.Walk(func(todo *models.Todo) {
			assert.Equal(t, models.StatusPending, todo.GetStatus())
		})
	})

	t.Run("reopen todo non-recursively", func(t *testing.T) {
		// Create nested structure
		store := testutil.CreateStoreWithNestedSpecs(t, []testutil.TodoSpec{
			{
				Text:   "Parent done",
				Status: models.StatusDone,
				Children: []testutil.TodoSpec{
					{Text: "Child done", Status: models.StatusDone},
				},
			},
		})
		
		// Get the parent's short ID
		collection, _ := store.Load()
		parentShortID := collection.Todos[0].ID[:8]
		
		// Reopen parent only
		reopened, err := commands.ReopenDirect(store, "file://test", parentShortID, false)
		require.NoError(t, err)
		assert.Len(t, reopened, 1) // Only parent
		
		// Verify parent is pending but child is still done
		collection, _ = store.Load()
		parent := collection.Todos[0]
		assert.Equal(t, models.StatusPending, parent.GetStatus())
		assert.Equal(t, models.StatusDone, parent.Items[0].GetStatus())
	})

	t.Run("error on invalid position", func(t *testing.T) {
		store := testutil.CreatePopulatedStore(t, "Test todo")
		
		_, err := commands.ReopenDirect(store, "file://test", "99", false)
		assert.Error(t, err)
	})
}