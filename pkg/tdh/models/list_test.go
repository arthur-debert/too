package models_test

import (
	"testing"

	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/stretchr/testify/assert"
)

func TestListActive(t *testing.T) {
	t.Run("returns only pending todos", func(t *testing.T) {
		collection := &models.Collection{
			Todos: []*models.Todo{
				{ID: "1", Position: 1, Text: "Active 1", Status: models.StatusPending},
				{ID: "2", Position: 2, Text: "Done 1", Status: models.StatusDone},
				{ID: "3", Position: 3, Text: "Active 2", Status: models.StatusPending},
			},
		}

		active := collection.ListActive()

		assert.Len(t, active, 2)
		assert.Equal(t, "Active 1", active[0].Text)
		assert.Equal(t, "Active 2", active[1].Text)
	})

	t.Run("hides children of done parents (behavioral propagation)", func(t *testing.T) {
		parent := &models.Todo{
			ID:       "parent",
			Position: 1,
			Text:     "Done Parent",
			Status:   models.StatusDone,
			Items: []*models.Todo{
				{ID: "child1", ParentID: "parent", Position: 1, Text: "Pending Child", Status: models.StatusPending},
				{ID: "child2", ParentID: "parent", Position: 2, Text: "Done Child", Status: models.StatusDone},
			},
		}

		collection := &models.Collection{
			Todos: []*models.Todo{
				parent,
				{ID: "2", Position: 2, Text: "Active Sibling", Status: models.StatusPending},
			},
		}

		active := collection.ListActive()

		// Should only see the active sibling, not the done parent or its children
		assert.Len(t, active, 1)
		assert.Equal(t, "Active Sibling", active[0].Text)
	})

	t.Run("shows pending children of pending parents", func(t *testing.T) {
		parent := &models.Todo{
			ID:       "parent",
			Position: 1,
			Text:     "Active Parent",
			Status:   models.StatusPending,
			Items: []*models.Todo{
				{ID: "child1", ParentID: "parent", Position: 1, Text: "Pending Child", Status: models.StatusPending},
				{ID: "child2", ParentID: "parent", Position: 2, Text: "Done Child", Status: models.StatusDone},
			},
		}

		collection := &models.Collection{Todos: []*models.Todo{parent}}

		active := collection.ListActive()

		assert.Len(t, active, 1)
		assert.Equal(t, "Active Parent", active[0].Text)
		// Should have only the pending child
		assert.Len(t, active[0].Items, 1)
		assert.Equal(t, "Pending Child", active[0].Items[0].Text)
	})

	t.Run("empty collection returns empty list", func(t *testing.T) {
		collection := models.NewCollection()
		active := collection.ListActive()
		assert.Empty(t, active)
	})

	t.Run("all done returns empty list", func(t *testing.T) {
		collection := &models.Collection{
			Todos: []*models.Todo{
				{ID: "1", Position: 1, Text: "Done 1", Status: models.StatusDone},
				{ID: "2", Position: 2, Text: "Done 2", Status: models.StatusDone},
			},
		}

		active := collection.ListActive()
		assert.Empty(t, active)
	})

	t.Run("returns clones not originals", func(t *testing.T) {
		collection := &models.Collection{
			Todos: []*models.Todo{
				{ID: "1", Position: 1, Text: "Active", Status: models.StatusPending},
			},
		}

		active := collection.ListActive()
		active[0].Text = "Modified"

		// Original should be unchanged
		assert.Equal(t, "Active", collection.Todos[0].Text)
	})
}

func TestListArchived(t *testing.T) {
	t.Run("returns only done todos", func(t *testing.T) {
		collection := &models.Collection{
			Todos: []*models.Todo{
				{ID: "1", Position: 1, Text: "Active", Status: models.StatusPending},
				{ID: "2", Position: 2, Text: "Done 1", Status: models.StatusDone},
				{ID: "3", Position: 3, Text: "Done 2", Status: models.StatusDone},
			},
		}

		archived := collection.ListArchived()

		assert.Len(t, archived, 2)
		assert.Equal(t, "Done 1", archived[0].Text)
		assert.Equal(t, "Done 2", archived[1].Text)
	})

	t.Run("shows done parents without their children", func(t *testing.T) {
		parent := &models.Todo{
			ID:       "parent",
			Position: 1,
			Text:     "Done Parent",
			Status:   models.StatusDone,
			Items: []*models.Todo{
				{ID: "child1", ParentID: "parent", Position: 1, Text: "Pending Child", Status: models.StatusPending},
				{ID: "child2", ParentID: "parent", Position: 2, Text: "Done Child", Status: models.StatusDone},
			},
		}

		collection := &models.Collection{Todos: []*models.Todo{parent}}

		archived := collection.ListArchived()

		assert.Len(t, archived, 1)
		assert.Equal(t, "Done Parent", archived[0].Text)
		// Children should be hidden (behavioral propagation)
		assert.Empty(t, archived[0].Items)
	})

	t.Run("shows done children of pending parents", func(t *testing.T) {
		parent := &models.Todo{
			ID:       "parent",
			Position: 1,
			Text:     "Active Parent",
			Status:   models.StatusPending,
			Items: []*models.Todo{
				{ID: "child1", ParentID: "parent", Position: 1, Text: "Pending Child", Status: models.StatusPending},
				{ID: "child2", ParentID: "parent", Position: 2, Text: "Done Child", Status: models.StatusDone},
			},
		}

		collection := &models.Collection{Todos: []*models.Todo{parent}}

		archived := collection.ListArchived()

		// Should not see the pending parent or pending child
		assert.Empty(t, archived)
	})

	t.Run("empty collection returns empty list", func(t *testing.T) {
		collection := models.NewCollection()
		archived := collection.ListArchived()
		assert.Empty(t, archived)
	})

	t.Run("all pending returns empty list", func(t *testing.T) {
		collection := &models.Collection{
			Todos: []*models.Todo{
				{ID: "1", Position: 1, Text: "Active 1", Status: models.StatusPending},
				{ID: "2", Position: 2, Text: "Active 2", Status: models.StatusPending},
			},
		}

		archived := collection.ListArchived()
		assert.Empty(t, archived)
	})
}

func TestListAll(t *testing.T) {
	t.Run("returns all todos regardless of status", func(t *testing.T) {
		collection := &models.Collection{
			Todos: []*models.Todo{
				{ID: "1", Position: 1, Text: "Active", Status: models.StatusPending},
				{ID: "2", Position: 2, Text: "Done", Status: models.StatusDone},
			},
		}

		all := collection.ListAll()

		assert.Len(t, all, 2)
		assert.Equal(t, "Active", all[0].Text)
		assert.Equal(t, "Done", all[1].Text)
	})

	t.Run("shows all children including inconsistent states", func(t *testing.T) {
		parent := &models.Todo{
			ID:       "parent",
			Position: 1,
			Text:     "Done Parent",
			Status:   models.StatusDone,
			Items: []*models.Todo{
				{ID: "child1", ParentID: "parent", Position: 1, Text: "Pending Child", Status: models.StatusPending},
				{ID: "child2", ParentID: "parent", Position: 2, Text: "Done Child", Status: models.StatusDone},
			},
		}

		collection := &models.Collection{Todos: []*models.Todo{parent}}

		all := collection.ListAll()

		assert.Len(t, all, 1)
		assert.Equal(t, "Done Parent", all[0].Text)
		// Should show all children (revealing inconsistent state)
		assert.Len(t, all[0].Items, 2)
		assert.Equal(t, "Pending Child", all[0].Items[0].Text)
		assert.Equal(t, "Done Child", all[0].Items[1].Text)
	})

	t.Run("empty collection returns empty list", func(t *testing.T) {
		collection := models.NewCollection()
		all := collection.ListAll()
		assert.Empty(t, all)
	})

	t.Run("returns clones not originals", func(t *testing.T) {
		collection := &models.Collection{
			Todos: []*models.Todo{
				{ID: "1", Position: 1, Text: "Original", Status: models.StatusPending},
			},
		}

		all := collection.ListAll()
		all[0].Text = "Modified"

		// Original should be unchanged
		assert.Equal(t, "Original", collection.Todos[0].Text)
	})

	t.Run("preserves deep structure", func(t *testing.T) {
		grandchild := &models.Todo{
			ID:       "grandchild",
			ParentID: "child",
			Position: 1,
			Text:     "Grandchild",
			Status:   models.StatusPending,
			Items:    []*models.Todo{},
		}

		child := &models.Todo{
			ID:       "child",
			ParentID: "parent",
			Position: 1,
			Text:     "Child",
			Status:   models.StatusDone,
			Items:    []*models.Todo{grandchild},
		}

		parent := &models.Todo{
			ID:       "parent",
			Position: 1,
			Text:     "Parent",
			Status:   models.StatusPending,
			Items:    []*models.Todo{child},
		}

		collection := &models.Collection{Todos: []*models.Todo{parent}}

		all := collection.ListAll()

		// Verify structure is preserved
		assert.Len(t, all, 1)
		assert.Len(t, all[0].Items, 1)
		assert.Len(t, all[0].Items[0].Items, 1)
		assert.Equal(t, "Grandchild", all[0].Items[0].Items[0].Text)
	})
}
