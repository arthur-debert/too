package commands_test

import (
	"testing"

	"github.com/arthur-debert/too/pkg/too/commands"
	"github.com/arthur-debert/too/pkg/too/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMoveDirect(t *testing.T) {
	t.Run("move todo to root", func(t *testing.T) {
		store := testutil.CreateNestedStore(t)
		
		// Move "Sub-task 1.1" (position 1.1) to root
		todo, oldPath, newPath, err := commands.MoveDirect(store, "file://test", "1.1", "")
		require.NoError(t, err)
		require.NotNil(t, todo)
		
		assert.Equal(t, "Sub-task 1.1", todo.Text)
		assert.Equal(t, "1.1", oldPath)
		// After moving, positions are reordered
		assert.NotEqual(t, oldPath, newPath)
		
		// Verify it's now at root level
		collection, _ := store.Load()
		assert.Equal(t, "", todo.ParentID)
		assert.Len(t, collection.Todos, 3)
	})

	t.Run("move todo to different parent", func(t *testing.T) {
		store := testutil.CreateNestedStore(t)
		
		// Move "Grandchild 1.2.1" to be under "Sub-task 1.1"
		todo, oldPath, newPath, err := commands.MoveDirect(store, "file://test", "1.2.1", "1.1")
		require.NoError(t, err)
		
		assert.Equal(t, "Grandchild 1.2.1", todo.Text)
		assert.Equal(t, "1.2.1", oldPath)
		assert.Equal(t, "1.1.1", newPath) // Now first child of 1.1
	})

	t.Run("error on circular reference", func(t *testing.T) {
		store := testutil.CreateNestedStore(t)
		
		// Try to move parent into its own child
		_, _, _, err := commands.MoveDirect(store, "file://test", "1", "1.2")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot move a parent into its own descendant")
	})

	t.Run("error on invalid source", func(t *testing.T) {
		store := testutil.CreateNestedStore(t)
		
		_, _, _, err := commands.MoveDirect(store, "file://test", "99", "1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("error on invalid destination", func(t *testing.T) {
		store := testutil.CreateNestedStore(t)
		
		_, _, _, err := commands.MoveDirect(store, "file://test", "1.1", "99")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "destination parent not found")
	})
}