package commands_test

import (
	"testing"

	"github.com/arthur-debert/too/pkg/too/commands"
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompleteDirect(t *testing.T) {
	t.Run("complete single todo", func(t *testing.T) {
		store := testutil.CreatePopulatedStore(t)
		
		// Create a todo
		todo, err := commands.AddDirect(store, "file://test", "Test todo", "")
		require.NoError(t, err)
		
		// Complete it
		completed, err := commands.CompleteDirect(store, "file://test", []string{todo.ID})
		require.NoError(t, err)
		assert.Len(t, completed, 1)
		
		// Verify status
		collection, _ := store.Load()
		updatedTodo := collection.FindItemByID(todo.ID)
		assert.Equal(t, models.StatusDone, updatedTodo.GetStatus())
	})

	t.Run("complete multiple todos", func(t *testing.T) {
		store := testutil.CreatePopulatedStore(t)
		
		// Create todos
		todo1, _ := commands.AddDirect(store, "file://test", "Todo 1", "")
		todo2, _ := commands.AddDirect(store, "file://test", "Todo 2", "")
		todo3, _ := commands.AddDirect(store, "file://test", "Todo 3", "")
		
		// Complete them
		completed, err := commands.CompleteDirect(store, "file://test", []string{todo1.ID, todo3.ID})
		require.NoError(t, err)
		assert.Len(t, completed, 2)
		
		// Verify statuses
		collection, _ := store.Load()
		assert.Equal(t, models.StatusDone, collection.FindItemByID(todo1.ID).GetStatus())
		assert.Equal(t, models.StatusPending, collection.FindItemByID(todo2.ID).GetStatus())
		assert.Equal(t, models.StatusDone, collection.FindItemByID(todo3.ID).GetStatus())
	})

	t.Run("error on non-existent todo", func(t *testing.T) {
		store := testutil.CreatePopulatedStore(t)
		
		_, err := commands.CompleteDirect(store, "file://test", []string{"non-existent"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}