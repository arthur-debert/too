package models_test

import (
	"testing"

	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/stretchr/testify/assert"
)

func TestReorderTodos(t *testing.T) {
	t.Run("reorders todos with gaps in positions", func(t *testing.T) {
		// Create todos with gaps
		todos := []*models.Todo{
			{ID: "1", Position: 1, Text: "First"},
			{ID: "2", Position: 3, Text: "Second"},
			{ID: "3", Position: 7, Text: "Third"},
			{ID: "4", Position: 10, Text: "Fourth"},
		}

		count := models.ReorderTodos(todos)

		assert.Equal(t, 3, count) // All except the first one should be reordered
		assert.Equal(t, 1, todos[0].Position)
		assert.Equal(t, 2, todos[1].Position)
		assert.Equal(t, 3, todos[2].Position)
		assert.Equal(t, 4, todos[3].Position)
	})

	t.Run("no changes when already sequential", func(t *testing.T) {
		todos := []*models.Todo{
			{ID: "1", Position: 1, Text: "First"},
			{ID: "2", Position: 2, Text: "Second"},
			{ID: "3", Position: 3, Text: "Third"},
		}

		count := models.ReorderTodos(todos)

		assert.Equal(t, 0, count)
		assert.Equal(t, 1, todos[0].Position)
		assert.Equal(t, 2, todos[1].Position)
		assert.Equal(t, 3, todos[2].Position)
	})

	t.Run("sorts by position before reassigning", func(t *testing.T) {
		// Create todos out of order
		todos := []*models.Todo{
			{ID: "3", Position: 5, Text: "Third"},
			{ID: "1", Position: 1, Text: "First"},
			{ID: "4", Position: 8, Text: "Fourth"},
			{ID: "2", Position: 3, Text: "Second"},
		}

		count := models.ReorderTodos(todos)

		assert.Equal(t, 3, count) // Positions 3->2, 5->3, 8->4
		// Verify order and positions
		assert.Equal(t, "First", todos[0].Text)
		assert.Equal(t, 1, todos[0].Position)
		assert.Equal(t, "Second", todos[1].Text)
		assert.Equal(t, 2, todos[1].Position)
		assert.Equal(t, "Third", todos[2].Text)
		assert.Equal(t, 3, todos[2].Position)
		assert.Equal(t, "Fourth", todos[3].Text)
		assert.Equal(t, 4, todos[3].Position)
	})

	t.Run("handles empty slice", func(t *testing.T) {
		var todos []*models.Todo

		count := models.ReorderTodos(todos)

		assert.Equal(t, 0, count)
		assert.Len(t, todos, 0)
	})

	t.Run("handles single todo", func(t *testing.T) {
		todos := []*models.Todo{
			{ID: "1", Position: 5, Text: "Single"},
		}

		count := models.ReorderTodos(todos)

		assert.Equal(t, 1, count) // Position changed from 5 to 1
		assert.Equal(t, 1, todos[0].Position)
	})

	t.Run("handles duplicate positions with stable sort", func(t *testing.T) {
		// Todos with duplicate positions
		todos := []*models.Todo{
			{ID: "A", Position: 3, Text: "Todo A"},
			{ID: "B", Position: 1, Text: "Todo B"},
			{ID: "C", Position: 3, Text: "Todo C"}, // Same position as A
			{ID: "D", Position: 2, Text: "Todo D"},
		}

		count := models.ReorderTodos(todos)

		assert.Equal(t, 1, count) // Only C needs new position (from 3 to 4)
		// B(1), D(2), A(3), C(3) -> B(1), D(2), A(3), C(4)
		assert.Equal(t, "Todo B", todos[0].Text)
		assert.Equal(t, 1, todos[0].Position)
		assert.Equal(t, "Todo D", todos[1].Text)
		assert.Equal(t, 2, todos[1].Position)
		assert.Equal(t, "Todo A", todos[2].Text)
		assert.Equal(t, 3, todos[2].Position)
		assert.Equal(t, "Todo C", todos[3].Text)
		assert.Equal(t, 4, todos[3].Position)
		// Stable sort should preserve A before C
		assert.Equal(t, "A", todos[2].ID)
		assert.Equal(t, "C", todos[3].ID)
	})

	t.Run("preserves todo fields except position", func(t *testing.T) {
		todo := &models.Todo{
			ID:       "test-id",
			Position: 10,
			Text:     "Test todo",
			Status:   models.StatusPending,
		}
		todos := []*models.Todo{todo}

		count := models.ReorderTodos(todos)

		assert.Equal(t, 1, count)
		assert.Equal(t, 1, todo.Position) // Only position changed
		assert.Equal(t, "test-id", todo.ID)
		assert.Equal(t, "Test todo", todo.Text)
		assert.Equal(t, models.StatusPending, todo.Status)
	})
}

func TestCollectionReorder(t *testing.T) {
	t.Run("reorders collection todos", func(t *testing.T) {
		collection := &models.Collection{
			Todos: []*models.Todo{
				{ID: "1", Position: 5, Text: "First"},
				{ID: "2", Position: 2, Text: "Second"},
				{ID: "3", Position: 8, Text: "Third"},
			},
		}

		count := collection.Reorder()

		assert.Equal(t, 3, count) // All positions change: 2->1, 5->2, 8->3
		assert.Equal(t, 1, collection.Todos[0].Position)
		assert.Equal(t, 2, collection.Todos[1].Position)
		assert.Equal(t, 3, collection.Todos[2].Position)
	})

	t.Run("handles empty collection", func(t *testing.T) {
		collection := models.NewCollection()

		count := collection.Reorder()

		assert.Equal(t, 0, count)
		assert.Len(t, collection.Todos, 0)
	})
}
