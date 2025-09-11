package commands_test

import (
	"testing"

	"github.com/arthur-debert/too/pkg/too/commands"
	"github.com/arthur-debert/too/pkg/too/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModifyDirect(t *testing.T) {
	t.Run("modify todo text", func(t *testing.T) {
		store := testutil.CreatePopulatedStore(t, "Original text")
		
		// Get the todo position
		todo, oldText, err := commands.ModifyDirect(store, "file://test", "1", "Modified text")
		require.NoError(t, err)
		require.NotNil(t, todo)
		
		assert.Equal(t, "Modified text", todo.Text)
		assert.Equal(t, "Original text", oldText)
		
		// Verify persistence
		collection, _ := store.Load()
		updatedTodo := collection.Todos[0]
		assert.Equal(t, "Modified text", updatedTodo.Text)
	})

	t.Run("modify nested todo", func(t *testing.T) {
		store := testutil.CreateNestedStore(t)
		
		// Modify "Sub-task 1.2" which is at position 1.2
		todo, oldText, err := commands.ModifyDirect(store, "file://test", "1.2", "Updated sub-task")
		require.NoError(t, err)
		
		assert.Equal(t, "Updated sub-task", todo.Text)
		assert.Equal(t, "Sub-task 1.2", oldText)
	})

	t.Run("error on empty text", func(t *testing.T) {
		store := testutil.CreatePopulatedStore(t, "Test todo")
		
		_, _, err := commands.ModifyDirect(store, "file://test", "1", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})

	t.Run("error on invalid position", func(t *testing.T) {
		store := testutil.CreatePopulatedStore(t, "Test todo")
		
		_, _, err := commands.ModifyDirect(store, "file://test", "99", "New text")
		assert.Error(t, err)
	})
}