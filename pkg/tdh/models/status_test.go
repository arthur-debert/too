package models_test

import (
	"testing"
	"time"

	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/stretchr/testify/assert"
)

func TestSetStatus(t *testing.T) {
	t.Run("marking todo as done sets position to 0", func(t *testing.T) {
		collection := &models.Collection{
			Todos: []*models.Todo{
				{ID: "1", Position: 1, Text: "First", Status: models.StatusPending},
				{ID: "2", Position: 2, Text: "Second", Status: models.StatusPending},
			},
		}

		todo := collection.Todos[0]
		originalModified := todo.Modified

		// Sleep to ensure modified time changes
		time.Sleep(time.Millisecond)

		// Mark as done with skip reorder to test just the position change
		todo.SetStatus(models.StatusDone, collection, true)

		assert.Equal(t, models.StatusDone, todo.Status)
		assert.Equal(t, 0, todo.Position)
		assert.True(t, todo.Modified.After(originalModified))
	})

	t.Run("marking todo as done triggers sibling reorder by default", func(t *testing.T) {
		collection := &models.Collection{
			Todos: []*models.Todo{
				{ID: "1", Position: 1, Text: "First", Status: models.StatusPending},
				{ID: "2", Position: 2, Text: "Second", Status: models.StatusPending},
				{ID: "3", Position: 3, Text: "Third", Status: models.StatusPending},
			},
		}

		// Mark first todo as done
		collection.Todos[0].SetStatus(models.StatusDone, collection)

		// First todo should have position 0
		assert.Equal(t, 0, collection.Todos[0].Position)
		// Other todos should be renumbered
		assert.Equal(t, 1, collection.Todos[1].Position)
		assert.Equal(t, 2, collection.Todos[2].Position)
	})

	t.Run("marking nested todo as done triggers sibling reorder", func(t *testing.T) {
		parent := &models.Todo{
			ID:       "parent",
			Position: 1,
			Text:     "Parent",
			Status:   models.StatusPending,
			Items: []*models.Todo{
				{ID: "1", ParentID: "parent", Position: 1, Text: "Child 1", Status: models.StatusPending},
				{ID: "2", ParentID: "parent", Position: 2, Text: "Child 2", Status: models.StatusPending},
				{ID: "3", ParentID: "parent", Position: 3, Text: "Child 3", Status: models.StatusPending},
			},
		}
		collection := &models.Collection{Todos: []*models.Todo{parent}}

		// Mark first child as done
		parent.Items[0].SetStatus(models.StatusDone, collection)

		// First child should have position 0
		assert.Equal(t, 0, parent.Items[0].Position)
		// Other children should be renumbered
		assert.Equal(t, 1, parent.Items[1].Position)
		assert.Equal(t, 2, parent.Items[2].Position)
		// Parent position unchanged
		assert.Equal(t, 1, parent.Position)
	})

	t.Run("skipReorder parameter prevents automatic reordering", func(t *testing.T) {
		collection := &models.Collection{
			Todos: []*models.Todo{
				{ID: "1", Position: 1, Text: "First", Status: models.StatusPending},
				{ID: "2", Position: 2, Text: "Second", Status: models.StatusPending},
				{ID: "3", Position: 3, Text: "Third", Status: models.StatusPending},
			},
		}

		// Mark first todo as done with skipReorder
		collection.Todos[0].SetStatus(models.StatusDone, collection, true)

		// First todo should have position 0
		assert.Equal(t, 0, collection.Todos[0].Position)
		// Other todos should NOT be renumbered
		assert.Equal(t, 2, collection.Todos[1].Position)
		assert.Equal(t, 3, collection.Todos[2].Position)
	})

	t.Run("setting same status does not trigger reorder", func(t *testing.T) {
		collection := &models.Collection{
			Todos: []*models.Todo{
				{ID: "1", Position: 1, Text: "First", Status: models.StatusPending},
				{ID: "2", Position: 3, Text: "Second", Status: models.StatusPending}, // Gap in position
			},
		}

		// Set to same status
		collection.Todos[0].SetStatus(models.StatusPending, collection)

		// Positions should remain unchanged (no reorder triggered)
		assert.Equal(t, 1, collection.Todos[0].Position)
		assert.Equal(t, 3, collection.Todos[1].Position)
	})
}

func TestMarkComplete(t *testing.T) {
	t.Run("MarkComplete is convenience for SetStatus", func(t *testing.T) {
		collection := &models.Collection{
			Todos: []*models.Todo{
				{ID: "1", Position: 1, Text: "First", Status: models.StatusPending},
				{ID: "2", Position: 2, Text: "Second", Status: models.StatusPending},
			},
		}

		collection.Todos[0].MarkComplete(collection)

		assert.Equal(t, models.StatusDone, collection.Todos[0].Status)
		assert.Equal(t, 0, collection.Todos[0].Position)
		// Reorder should have happened
		assert.Equal(t, 1, collection.Todos[1].Position)
	})

	t.Run("MarkComplete with skipReorder", func(t *testing.T) {
		collection := &models.Collection{
			Todos: []*models.Todo{
				{ID: "1", Position: 1, Text: "First", Status: models.StatusPending},
				{ID: "2", Position: 2, Text: "Second", Status: models.StatusPending},
			},
		}

		collection.Todos[0].MarkComplete(collection, true)

		assert.Equal(t, models.StatusDone, collection.Todos[0].Status)
		assert.Equal(t, 0, collection.Todos[0].Position)
		// No reorder
		assert.Equal(t, 2, collection.Todos[1].Position)
	})
}

func TestMarkPending(t *testing.T) {
	t.Run("MarkPending assigns next available position", func(t *testing.T) {
		collection := &models.Collection{
			Todos: []*models.Todo{
				{ID: "1", Position: 0, Text: "First", Status: models.StatusDone},
				{ID: "2", Position: 1, Text: "Second", Status: models.StatusPending},
				{ID: "3", Position: 2, Text: "Third", Status: models.StatusPending},
			},
		}

		collection.Todos[0].MarkPending(collection)

		assert.Equal(t, models.StatusPending, collection.Todos[0].Status)
		// Should get position 3 (after existing 1 and 2)
		assert.Equal(t, 3, collection.Todos[0].Position)
	})

	t.Run("MarkPending in nested context", func(t *testing.T) {
		parent := &models.Todo{
			ID:       "parent",
			Position: 1,
			Text:     "Parent",
			Status:   models.StatusPending,
			Items: []*models.Todo{
				{ID: "1", ParentID: "parent", Position: 0, Text: "Child 1", Status: models.StatusDone},
				{ID: "2", ParentID: "parent", Position: 1, Text: "Child 2", Status: models.StatusPending},
			},
		}
		collection := &models.Collection{Todos: []*models.Todo{parent}}

		parent.Items[0].MarkPending(collection)

		assert.Equal(t, models.StatusPending, parent.Items[0].Status)
		// Should get position 2 (after existing position 1)
		assert.Equal(t, 2, parent.Items[0].Position)
	})
}

func TestResetActivePositions(t *testing.T) {
	t.Run("resets positions for pending items only", func(t *testing.T) {
		todos := []*models.Todo{
			{ID: "1", Position: 1, Text: "First", Status: models.StatusDone},
			{ID: "2", Position: 2, Text: "Second", Status: models.StatusPending},
			{ID: "3", Position: 3, Text: "Third", Status: models.StatusDone},
			{ID: "4", Position: 4, Text: "Fourth", Status: models.StatusPending},
		}

		models.ResetActivePositions(todos)

		// Done items should have position 0
		assert.Equal(t, 0, todos[0].Position)
		assert.Equal(t, 0, todos[2].Position)
		// Pending items should be renumbered sequentially
		assert.Equal(t, 1, todos[1].Position)
		assert.Equal(t, 2, todos[3].Position)
	})

	t.Run("handles newly reopened items with position 0", func(t *testing.T) {
		todos := []*models.Todo{
			{ID: "1", Position: 1, Text: "First", Status: models.StatusPending},
			{ID: "2", Position: 0, Text: "Reopened", Status: models.StatusPending}, // Position 0
			{ID: "3", Position: 2, Text: "Third", Status: models.StatusPending},
		}

		models.ResetActivePositions(todos)

		// Items should be ordered with position 0 items at the end
		assert.Equal(t, 1, todos[0].Position) // First stays first
		assert.Equal(t, 3, todos[1].Position) // Reopened goes to end
		assert.Equal(t, 2, todos[2].Position) // Third becomes second
	})

	t.Run("empty list", func(t *testing.T) {
		todos := []*models.Todo{}
		// Should not panic
		models.ResetActivePositions(todos)
	})

	t.Run("all done items", func(t *testing.T) {
		todos := []*models.Todo{
			{ID: "1", Position: 1, Text: "First", Status: models.StatusDone},
			{ID: "2", Position: 2, Text: "Second", Status: models.StatusDone},
		}

		models.ResetActivePositions(todos)

		// All should have position 0
		assert.Equal(t, 0, todos[0].Position)
		assert.Equal(t, 0, todos[1].Position)
	})
}

func TestCollectionResetMethods(t *testing.T) {
	t.Run("ResetRootPositions affects only root level", func(t *testing.T) {
		collection := &models.Collection{
			Todos: []*models.Todo{
				{ID: "1", Position: 1, Text: "First", Status: models.StatusPending},
				{ID: "2", Position: 2, Text: "Second", Status: models.StatusDone},
				{ID: "3", Position: 3, Text: "Third", Status: models.StatusPending,
					Items: []*models.Todo{
						{ID: "3.1", ParentID: "3", Position: 99, Text: "Child", Status: models.StatusPending},
					},
				},
			},
		}

		collection.ResetRootPositions()

		// Root level should be reset
		assert.Equal(t, 1, collection.Todos[0].Position)
		assert.Equal(t, 0, collection.Todos[1].Position) // Done
		assert.Equal(t, 2, collection.Todos[2].Position)
		// Child should be unchanged
		assert.Equal(t, 99, collection.Todos[2].Items[0].Position)
	})

	t.Run("ResetSiblingPositions affects only specified parent's children", func(t *testing.T) {
		parent1 := &models.Todo{
			ID: "p1", Position: 1, Text: "Parent 1", Status: models.StatusPending,
			Items: []*models.Todo{
				{ID: "1.1", ParentID: "p1", Position: 1, Text: "Child 1.1", Status: models.StatusDone},
				{ID: "1.2", ParentID: "p1", Position: 2, Text: "Child 1.2", Status: models.StatusPending},
			},
		}
		parent2 := &models.Todo{
			ID: "p2", Position: 2, Text: "Parent 2", Status: models.StatusPending,
			Items: []*models.Todo{
				{ID: "2.1", ParentID: "p2", Position: 99, Text: "Child 2.1", Status: models.StatusPending},
			},
		}
		collection := &models.Collection{Todos: []*models.Todo{parent1, parent2}}

		collection.ResetSiblingPositions("p1")

		// Parent 1's children should be reset
		assert.Equal(t, 0, parent1.Items[0].Position) // Done
		assert.Equal(t, 1, parent1.Items[1].Position)
		// Parent 2's children should be unchanged
		assert.Equal(t, 99, parent2.Items[0].Position)
	})

	t.Run("ResetSiblingPositions with non-existent parent", func(t *testing.T) {
		collection := &models.Collection{
			Todos: []*models.Todo{
				{ID: "1", Position: 1, Text: "First", Status: models.StatusPending},
			},
		}

		// Should not panic
		collection.ResetSiblingPositions("non-existent")

		// Nothing should change
		assert.Equal(t, 1, collection.Todos[0].Position)
	})
}
