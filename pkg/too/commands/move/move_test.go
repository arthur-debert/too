package move_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/arthur-debert/too/pkg/too"
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/testutil"
)

func TestMoveCommand(t *testing.T) {
	t.Run("moves a todo to be a child of another todo", func(t *testing.T) {
		// Create simple store with flat todos
		store := testutil.CreateStoreWithSpecs(t, []testutil.TodoSpec{
			{Text: "Parent todo", Status: models.StatusPending},
			{Text: "Todo to move", Status: models.StatusPending},
		})

		// Move second todo to be child of first
		opts := too.MoveOptions{CollectionPath: store.Path()}
		result, err := too.Move("2", "1", opts) // Move item at pos 2 to be child of item at pos 1

		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "2", result.OldPath)
		assert.NotEmpty(t, result.NewPath)

		// Verify the move worked
		collection, err := store.LoadIDM()
		testutil.AssertNoError(t, err)
		
		// Should still have 2 todos
		assert.Equal(t, 2, len(collection.Items))
		
		// Find the todos
		var parent, child *models.IDMTodo
		for _, todo := range collection.Items {
			if todo.Text == "Parent todo" && todo.ParentID == "" {
				parent = todo
			} else if todo.Text == "Todo to move" && todo.ParentID != "" {
				child = todo
			}
		}
		
		assert.NotNil(t, parent, "Should have parent todo")
		assert.NotNil(t, child, "Should have child todo")
		if parent != nil && child != nil {
			assert.Equal(t, parent.UID, child.ParentID, "Child should have parent's UID as ParentID")
		}
	})

	t.Run("moves a child todo to become top-level", func(t *testing.T) {
		// Create nested structure
		store := testutil.CreateStoreWithNestedSpecs(t, []testutil.TodoSpec{
			{Text: "Parent todo", Status: models.StatusPending, Children: []testutil.TodoSpec{
				{Text: "Child to move", Status: models.StatusPending},
			}},
		})

		// Move child to root level
		opts := too.MoveOptions{CollectionPath: store.Path()}
		result, err := too.Move("1.1", "", opts) // Move child to root

		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)

		// Verify the move worked
		collection, err := store.LoadIDM()
		testutil.AssertNoError(t, err)
		
		// Find the moved todo
		var movedTodo *models.IDMTodo
		for _, todo := range collection.Items {
			if todo.Text == "Child to move" {
				movedTodo = todo
				break
			}
		}
		
		assert.NotNil(t, movedTodo, "Should find the moved todo")
		if movedTodo != nil {
			assert.Equal(t, "", movedTodo.ParentID, "Moved todo should have empty ParentID (top-level)")
		}
	})

	t.Run("handles invalid source position", func(t *testing.T) {
		store := testutil.CreateStoreWithSpecs(t, []testutil.TodoSpec{
			{Text: "Test todo", Status: models.StatusPending},
		})

		opts := too.MoveOptions{CollectionPath: store.Path()}
		result, err := too.Move("99", "1", opts) // Try to move non-existent todo

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("handles invalid target position", func(t *testing.T) {
		store := testutil.CreateStoreWithSpecs(t, []testutil.TodoSpec{
			{Text: "Todo 1", Status: models.StatusPending},
			{Text: "Todo 2", Status: models.StatusPending},
		})

		opts := too.MoveOptions{CollectionPath: store.Path()}
		result, err := too.Move("1", "99", opts) // Try to move to non-existent target

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}