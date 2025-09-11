package commands_test

import (
	"testing"

	"github.com/arthur-debert/too/pkg/too/commands"
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddDirect(t *testing.T) {
	t.Run("add todo to root", func(t *testing.T) {
		store := testutil.CreatePopulatedStore(t)
		
		todo, err := commands.AddDirect(store, "file://test", "Test todo", "")
		require.NoError(t, err)
		require.NotNil(t, todo)
		
		assert.Equal(t, "Test todo", todo.Text)
		assert.Equal(t, "", todo.ParentID)
		assert.Equal(t, models.StatusPending, todo.GetStatus())
	})

	t.Run("add todo to parent", func(t *testing.T) {
		store := testutil.CreatePopulatedStore(t)
		
		// Create parent
		parent, err := commands.AddDirect(store, "file://test", "Parent todo", "")
		require.NoError(t, err)
		
		// Add child
		child, err := commands.AddDirect(store, "file://test", "Child todo", parent.ID)
		require.NoError(t, err)
		require.NotNil(t, child)
		
		assert.Equal(t, "Child todo", child.Text)
		assert.Equal(t, parent.ID, child.ParentID)
		assert.Equal(t, models.StatusPending, child.GetStatus())
		
		// Verify parent has child
		collection, _ := store.Load()
		parentTodo := collection.FindItemByID(parent.ID)
		assert.Len(t, parentTodo.Items, 1)
		assert.Equal(t, child.ID, parentTodo.Items[0].ID)
	})
}