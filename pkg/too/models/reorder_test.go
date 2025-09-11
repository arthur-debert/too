package models_test

import (
	"testing"

	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/stretchr/testify/assert"
)

func TestReorderTodos(t *testing.T) {
	t.Run("is now a no-op since IDM handles positioning", func(t *testing.T) {
		// Create todos
		todos := []*models.Todo{
			{ID: "1", Text: "First"},
			{ID: "2", Text: "Second"},
			{ID: "3", Text: "Third"},
		}

		// Store original order
		originalOrder := make([]string, len(todos))
		for i, todo := range todos {
			originalOrder[i] = todo.ID
		}

		models.ReorderTodos(todos)

		// Verify order is unchanged
		for i, todo := range todos {
			assert.Equal(t, originalOrder[i], todo.ID)
		}
	})

	t.Run("handles empty slice", func(t *testing.T) {
		var todos []*models.Todo

		models.ReorderTodos(todos)

		assert.Len(t, todos, 0)
	})

	t.Run("preserves todo fields", func(t *testing.T) {
		todo := &models.Todo{
			ID:       "test-id",
			Text:     "Test todo",
			Statuses: map[string]string{"completion": string(models.StatusPending)},
			Items:    []*models.Todo{},
		}
		todos := []*models.Todo{todo}

		models.ReorderTodos(todos)

		assert.Equal(t, "test-id", todo.ID)
		assert.Equal(t, "Test todo", todo.Text)
		assert.Equal(t, models.StatusPending, todo.GetStatus())
	})
}

func TestResetActivePositions(t *testing.T) {
	t.Run("is now a no-op since IDM handles positioning", func(t *testing.T) {
		todos := []*models.Todo{
			{ID: "1", Text: "First", Statuses: map[string]string{"completion": string(models.StatusDone)}, Items: []*models.Todo{}},
			{ID: "2", Text: "Second", Statuses: map[string]string{"completion": string(models.StatusPending)}, Items: []*models.Todo{}},
			{ID: "3", Text: "Third", Statuses: map[string]string{"completion": string(models.StatusPending)}, Items: []*models.Todo{}},
		}

		// Store original order
		originalOrder := make([]string, len(todos))
		for i, todo := range todos {
			originalOrder[i] = todo.ID
		}

		models.ResetActivePositions(&todos)

		// Verify order is unchanged
		for i, todo := range todos {
			assert.Equal(t, originalOrder[i], todo.ID)
		}
	})

	t.Run("handles nil slice", func(t *testing.T) {
		var todos *[]*models.Todo = nil
		
		// Should not panic
		models.ResetActivePositions(todos)
	})

	t.Run("handles empty slice", func(t *testing.T) {
		todos := []*models.Todo{}
		
		models.ResetActivePositions(&todos)
		
		assert.Len(t, todos, 0)
	})
}

func TestCollectionReorder(t *testing.T) {
	t.Run("is now a no-op since IDM handles positioning", func(t *testing.T) {
		collection := &models.Collection{
			Todos: []*models.Todo{
				{ID: "1", Text: "First", Statuses: map[string]string{"completion": string(models.StatusDone)}, Items: []*models.Todo{}},
				{ID: "2", Text: "Second", Statuses: map[string]string{"completion": string(models.StatusPending)}, Items: []*models.Todo{}},
				{ID: "3", Text: "Third", Statuses: map[string]string{"completion": string(models.StatusPending)}, Items: []*models.Todo{}},
			},
		}

		// Store original order
		originalOrder := make([]string, len(collection.Todos))
		for i, todo := range collection.Todos {
			originalOrder[i] = todo.ID
		}

		collection.Reorder()

		// Verify order is unchanged
		for i, todo := range collection.Todos {
			assert.Equal(t, originalOrder[i], todo.ID)
		}
	})

	t.Run("handles empty collection", func(t *testing.T) {
		collection := models.NewCollection()

		collection.Reorder()

		assert.Len(t, collection.Todos, 0)
	})
}