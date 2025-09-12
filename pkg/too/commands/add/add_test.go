package add_test

import (
	"testing"

	"github.com/arthur-debert/too/pkg/too/commands/add"
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/testutil"
	"github.com/stretchr/testify/assert"
)

func TestAddCommand(t *testing.T) {
	t.Run("adds todo to empty collection", func(t *testing.T) {
		// Create empty IDM store
		idmStore := testutil.CreateIDMStore(t) // Empty store

		opts := add.Options{CollectionPath: idmStore.Path()}
		result, err := add.Execute("My first todo", opts)

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Todo)
		assert.Equal(t, "My first todo", result.Todo.Text)
		assert.Equal(t, models.StatusPending, result.Todo.GetStatus())

		// Verify it was saved
		collection, err := idmStore.LoadIDM()
		assert.NoError(t, err)
		assert.Equal(t, 1, collection.Count())
		assert.Equal(t, "My first todo", collection.Items[0].Text)
	})

	t.Run("adds todo to existing collection", func(t *testing.T) {
		// Create store with existing todos
		idmStore := testutil.CreateIDMStore(t, "Existing 1", "Existing 2")

		opts := add.Options{CollectionPath: idmStore.Path()}
		result, err := add.Execute("New todo", opts)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "New todo", result.Todo.Text)

		// Verify all todos are present
		collection, err := idmStore.LoadIDM()
		assert.NoError(t, err)

		assert.Equal(t, 3, collection.Count())
		// Verify each todo exists
		texts := []string{}
		for _, item := range collection.Items {
			texts = append(texts, item.Text)
		}
		assert.Contains(t, texts, "Existing 1")
		assert.Contains(t, texts, "Existing 2")
		assert.Contains(t, texts, "New todo")
	})

	t.Run("returns error for empty text", func(t *testing.T) {
		idmStore := testutil.CreateIDMStore(t)

		opts := add.Options{CollectionPath: idmStore.Path()}
		result, err := add.Execute("", opts)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "todo text cannot be empty")
	})

	t.Run("adds sub-todo to parent", func(t *testing.T) {
		// Create store with parent todo
		idmStore := testutil.CreateIDMStore(t, "Parent task")

		// Add sub-task
		opts := add.Options{
			CollectionPath: idmStore.Path(),
			ParentPath:     "1",
		}
		result, err := add.Execute("Sub-task", opts)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Sub-task", result.Todo.Text)

		// Verify structure
		collection, err := idmStore.LoadIDM()
		assert.NoError(t, err)
		assert.Equal(t, 2, collection.Count())

		// Find the child todo
		var childTodo *models.IDMTodo
		for _, item := range collection.Items {
			if item.Text == "Sub-task" {
				childTodo = item
				break
			}
		}
		assert.NotNil(t, childTodo)
		assert.NotEmpty(t, childTodo.ParentID) // Should have parent
	})

	t.Run("returns error for non-existent parent", func(t *testing.T) {
		idmStore := testutil.CreateIDMStore(t, "Only todo")

		opts := add.Options{
			CollectionPath: idmStore.Path(),
			ParentPath:     "99",
		}
		result, err := add.Execute("Orphan task", opts)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "parent todo not found")
	})
}