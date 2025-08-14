package helpers_test

import (
	"testing"

	"github.com/arthur-debert/tdh/pkg/tdh/internal/helpers"
	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/stretchr/testify/assert"
)

func TestFind(t *testing.T) {
	t.Run("finds existing todo by ID", func(t *testing.T) {
		collection := models.NewCollection()
		_, _ = collection.CreateTodo("First", "")
		todo2, _ := collection.CreateTodo("Second", "")
		_, _ = collection.CreateTodo("Third", "")

		// Find middle todo
		found, err := helpers.Find(collection, todo2.Position)

		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, todo2.ID, found.ID)
		assert.Equal(t, todo2.Position, found.Position)
		assert.Equal(t, "Second", found.Text)
	})

	t.Run("finds first todo", func(t *testing.T) {
		collection := models.NewCollection()
		todo1, _ := collection.CreateTodo("First", "")
		_, _ = collection.CreateTodo("Second", "")
		_, _ = collection.CreateTodo("Third", "")

		found, err := helpers.Find(collection, 1)

		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, todo1.ID, found.ID)
		assert.Equal(t, "First", found.Text)
	})

	t.Run("finds last todo", func(t *testing.T) {
		collection := models.NewCollection()
		_, _ = collection.CreateTodo("First", "")
		_, _ = collection.CreateTodo("Second", "")
		todo3, _ := collection.CreateTodo("Third", "")

		found, err := helpers.Find(collection, todo3.Position)

		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, todo3.ID, found.ID)
		assert.Equal(t, "Third", found.Text)
	})

	t.Run("returns error for non-existent ID", func(t *testing.T) {
		collection := models.NewCollection()
		_, _ = collection.CreateTodo("First", "")
		_, _ = collection.CreateTodo("Second", "")

		found, err := helpers.Find(collection, 999)

		assert.Error(t, err)
		assert.Nil(t, found)
		assert.Contains(t, err.Error(), "todo with position 999 was not found")
	})

	t.Run("returns error for negative ID", func(t *testing.T) {
		collection := models.NewCollection()
		_, _ = collection.CreateTodo("First", "")

		found, err := helpers.Find(collection, -1)

		assert.Error(t, err)
		assert.Nil(t, found)
		assert.Contains(t, err.Error(), "todo with position -1 was not found")
	})

	t.Run("returns error for zero ID", func(t *testing.T) {
		collection := models.NewCollection()
		_, _ = collection.CreateTodo("First", "")

		found, err := helpers.Find(collection, 0)

		assert.Error(t, err)
		assert.Nil(t, found)
		assert.Contains(t, err.Error(), "todo with position 0 was not found")
	})

	t.Run("works with empty collection", func(t *testing.T) {
		collection := models.NewCollection()

		found, err := helpers.Find(collection, 1)

		assert.Error(t, err)
		assert.Nil(t, found)
		assert.Contains(t, err.Error(), "todo with position 1 was not found")
	})

	t.Run("handles non-sequential IDs", func(t *testing.T) {
		collection := models.NewCollection()
		// Manually create todos with non-sequential positions
		collection.Todos = []*models.Todo{
			{ID: "id-1", Position: 5, Text: "Fifth", Status: models.StatusPending},
			{ID: "id-2", Position: 10, Text: "Tenth", Status: models.StatusPending},
			{ID: "id-3", Position: 3, Text: "Third", Status: models.StatusPending},
		}

		// Should find position 10
		found, err := helpers.Find(collection, 10)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, "id-2", found.ID)
		assert.Equal(t, 10, found.Position)
		assert.Equal(t, "Tenth", found.Text)

		// Should not find ID 7
		found, err = helpers.Find(collection, 7)
		assert.Error(t, err)
		assert.Nil(t, found)
	})

	t.Run("handles very large position", func(t *testing.T) {
		collection := models.NewCollection()
		// Create todo with large position
		collection.Todos = []*models.Todo{
			{ID: "id-large", Position: 2147483647, Text: "Max int32", Status: models.StatusPending}, // Max int32
		}

		found, err := helpers.Find(collection, 2147483647)

		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, "id-large", found.ID)
		assert.Equal(t, 2147483647, found.Position)
		assert.Equal(t, "Max int32", found.Text)
	})

	t.Run("returns actual todo reference not copy", func(t *testing.T) {
		collection := models.NewCollection()
		todo, _ := collection.CreateTodo("Original", "")

		found, err := helpers.Find(collection, todo.Position)

		assert.NoError(t, err)
		assert.NotNil(t, found)

		// Modify the found todo
		found.Text = "Modified"

		// Verify it modified the original in the collection
		assert.Equal(t, "Modified", collection.Todos[0].Text)
	})
}
