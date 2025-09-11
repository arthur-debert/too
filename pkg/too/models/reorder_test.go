package models_test

import (
	"testing"

	"github.com/arthur-debert/too/pkg/too/models"
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

		models.ReorderTodos(todos)

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

		models.ReorderTodos(todos)

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

		models.ReorderTodos(todos)

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

		models.ReorderTodos(todos)

		assert.Len(t, todos, 0)
	})

	t.Run("handles single todo", func(t *testing.T) {
		todos := []*models.Todo{
			{ID: "1", Position: 5, Text: "Single"},
		}

		models.ReorderTodos(todos)

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

		models.ReorderTodos(todos)

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
			Statuses: map[string]string{"completion": string(models.StatusPending)},
			Items:    []*models.Todo{},
		}
		todos := []*models.Todo{todo}

		models.ReorderTodos(todos)

		assert.Equal(t, 1, todo.Position) // Only position changed
		assert.Equal(t, "test-id", todo.ID)
		assert.Equal(t, "Test todo", todo.Text)
		assert.Equal(t, models.StatusPending, todo.GetStatus())
	})
	t.Run("recursively reorders nested todos", func(t *testing.T) {
		// Create a nested structure with position gaps at each level
		todos := []*models.Todo{
			{ID: "1", Position: 5, Text: "Parent 1", Items: []*models.Todo{
				{ID: "1.1", Position: 3, Text: "Child 1.1", Items: []*models.Todo{}},
				{ID: "1.2", Position: 1, Text: "Child 1.2", Items: []*models.Todo{}},
			}},
			{ID: "2", Position: 2, Text: "Parent 2", Items: []*models.Todo{
				{ID: "2.1", Position: 10, Text: "Child 2.1", Items: []*models.Todo{}},
			}},
		}

		models.ReorderTodos(todos)

		// Verify parent reordering
		assert.Equal(t, "Parent 2", todos[0].Text)
		assert.Equal(t, 1, todos[0].Position)
		assert.Equal(t, "Parent 1", todos[1].Text)
		assert.Equal(t, 2, todos[1].Position)

		// Verify children of Parent 1 were reordered
		parent1Children := todos[1].Items
		assert.Equal(t, "Child 1.2", parent1Children[0].Text)
		assert.Equal(t, 1, parent1Children[0].Position)
		assert.Equal(t, "Child 1.1", parent1Children[1].Text)
		assert.Equal(t, 2, parent1Children[1].Position)

		// Verify children of Parent 2 were reordered
		parent2Children := todos[0].Items
		assert.Equal(t, "Child 2.1", parent2Children[0].Text)
		assert.Equal(t, 1, parent2Children[0].Position)
	})
}

func TestCollectionReorder(t *testing.T) {
	t.Run("reorders only active todos", func(t *testing.T) {
		collection := &models.Collection{
			Todos: []*models.Todo{
				{ID: "1", Position: 5, Text: "First", Statuses: map[string]string{"completion": string(models.StatusDone)}, Items: []*models.Todo{}},
				{ID: "2", Position: 2, Text: "Second", Statuses: map[string]string{"completion": string(models.StatusPending)}, Items: []*models.Todo{}},
				{ID: "3", Position: 8, Text: "Third", Statuses: map[string]string{"completion": string(models.StatusPending)}, Items: []*models.Todo{}},
				{ID: "4", Position: 4, Text: "Fourth", Statuses: map[string]string{"completion": string(models.StatusDone)}, Items: []*models.Todo{}},
			},
		}

		collection.Reorder()

		// After reordering, slice should have active items first, then done items
		// Active: Second (pos 1), Third (pos 2)
		// Done: First (pos 0), Fourth (pos 0)
		assert.Equal(t, "Second", collection.Todos[0].Text)
		assert.Equal(t, 1, collection.Todos[0].Position)
		assert.Equal(t, "Third", collection.Todos[1].Text)
		assert.Equal(t, 2, collection.Todos[1].Position)
		assert.Equal(t, "First", collection.Todos[2].Text)
		assert.Equal(t, 0, collection.Todos[2].Position)
		assert.Equal(t, "Fourth", collection.Todos[3].Text)
		assert.Equal(t, 0, collection.Todos[3].Position)
	})

	t.Run("reorders all pending todos", func(t *testing.T) {
		collection := &models.Collection{
			Todos: []*models.Todo{
				{ID: "1", Position: 5, Text: "First", Statuses: map[string]string{"completion": string(models.StatusPending)}, Items: []*models.Todo{}},
				{ID: "2", Position: 2, Text: "Second", Statuses: map[string]string{"completion": string(models.StatusPending)}, Items: []*models.Todo{}},
				{ID: "3", Position: 8, Text: "Third", Statuses: map[string]string{"completion": string(models.StatusPending)}, Items: []*models.Todo{}},
			},
		}

		collection.Reorder()

		// The todos should be reordered by their original position values
		// Second (was pos 2) -> pos 1, First (was pos 5) -> pos 2, Third (was pos 8) -> pos 3
		// And slice order should match: Second, First, Third
		assert.Equal(t, "Second", collection.Todos[0].Text)
		assert.Equal(t, 1, collection.Todos[0].Position)
		assert.Equal(t, "First", collection.Todos[1].Text)
		assert.Equal(t, 2, collection.Todos[1].Position)
		assert.Equal(t, "Third", collection.Todos[2].Text)
		assert.Equal(t, 3, collection.Todos[2].Position)
	})

	t.Run("handles empty collection", func(t *testing.T) {
		collection := models.NewCollection()

		collection.Reorder()

		assert.Len(t, collection.Todos, 0)
	})

	t.Run("handles all done todos", func(t *testing.T) {
		collection := &models.Collection{
			Todos: []*models.Todo{
				{ID: "1", Position: 1, Text: "First", Statuses: map[string]string{"completion": string(models.StatusDone)}, Items: []*models.Todo{}},
				{ID: "2", Position: 2, Text: "Second", Statuses: map[string]string{"completion": string(models.StatusDone)}, Items: []*models.Todo{}},
			},
		}

		collection.Reorder()

		// All should have position 0
		assert.Equal(t, 0, collection.Todos[0].Position)
		assert.Equal(t, 0, collection.Todos[1].Position)
	})
}
