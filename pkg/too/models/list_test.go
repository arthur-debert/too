package models_test

import (
	"testing"

	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/stretchr/testify/assert"
)

func TestListActive(t *testing.T) {
	t.Run("returns only pending todos", func(t *testing.T) {
		collection := &models.Collection{
			Todos: []*models.Todo{
				{ID: "1",  Text: "Active 1", Statuses: map[string]string{"completion": string(models.StatusPending)}, Items: []*models.Todo{}},
				{ID: "2",  Text: "Done 1", Statuses: map[string]string{"completion": string(models.StatusDone)}, Items: []*models.Todo{}},
				{ID: "3",  Text: "Active 2", Statuses: map[string]string{"completion": string(models.StatusPending)}, Items: []*models.Todo{}},
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
			
			Text:     "Done Parent",
			Statuses: map[string]string{"completion": string(models.StatusDone)},
			Items: []*models.Todo{
				{ID: "child1", ParentID: "parent",  Text: "Pending Child", Statuses: map[string]string{"completion": string(models.StatusPending)}, Items: []*models.Todo{}},
				{ID: "child2", ParentID: "parent",  Text: "Done Child", Statuses: map[string]string{"completion": string(models.StatusDone)}, Items: []*models.Todo{}},
			},
		}

		collection := &models.Collection{
			Todos: []*models.Todo{
				parent,
				{ID: "2",  Text: "Active Sibling", Statuses: map[string]string{"completion": string(models.StatusPending)}, Items: []*models.Todo{}},
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
			
			Text:     "Active Parent",
			Statuses: map[string]string{"completion": string(models.StatusPending)},
			Items: []*models.Todo{
				{ID: "child1", ParentID: "parent",  Text: "Pending Child", Statuses: map[string]string{"completion": string(models.StatusPending)}, Items: []*models.Todo{}},
				{ID: "child2", ParentID: "parent",  Text: "Done Child", Statuses: map[string]string{"completion": string(models.StatusDone)}, Items: []*models.Todo{}},
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
				{ID: "1",  Text: "Done 1", Statuses: map[string]string{"completion": string(models.StatusDone)}, Items: []*models.Todo{}},
				{ID: "2",  Text: "Done 2", Statuses: map[string]string{"completion": string(models.StatusDone)}, Items: []*models.Todo{}},
			},
		}

		active := collection.ListActive()
		assert.Empty(t, active)
	})

	t.Run("returns clones not originals", func(t *testing.T) {
		collection := &models.Collection{
			Todos: []*models.Todo{
				{ID: "1",  Text: "Active", Statuses: map[string]string{"completion": string(models.StatusPending)}, Items: []*models.Todo{}},
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
				{ID: "1",  Text: "Active", Statuses: map[string]string{"completion": string(models.StatusPending)}, Items: []*models.Todo{}},
				{ID: "2",  Text: "Done 1", Statuses: map[string]string{"completion": string(models.StatusDone)}, Items: []*models.Todo{}},
				{ID: "3",  Text: "Done 2", Statuses: map[string]string{"completion": string(models.StatusDone)}, Items: []*models.Todo{}},
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
			
			Text:     "Done Parent",
			Statuses: map[string]string{"completion": string(models.StatusDone)},
			Items: []*models.Todo{
				{ID: "child1", ParentID: "parent",  Text: "Pending Child", Statuses: map[string]string{"completion": string(models.StatusPending)}, Items: []*models.Todo{}},
				{ID: "child2", ParentID: "parent",  Text: "Done Child", Statuses: map[string]string{"completion": string(models.StatusDone)}, Items: []*models.Todo{}},
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
			
			Text:     "Active Parent",
			Statuses: map[string]string{"completion": string(models.StatusPending)},
			Items: []*models.Todo{
				{ID: "child1", ParentID: "parent",  Text: "Pending Child", Statuses: map[string]string{"completion": string(models.StatusPending)}, Items: []*models.Todo{}},
				{ID: "child2", ParentID: "parent",  Text: "Done Child", Statuses: map[string]string{"completion": string(models.StatusDone)}, Items: []*models.Todo{}},
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
				{ID: "1",  Text: "Active 1", Statuses: map[string]string{"completion": string(models.StatusPending)}, Items: []*models.Todo{}},
				{ID: "2",  Text: "Active 2", Statuses: map[string]string{"completion": string(models.StatusPending)}, Items: []*models.Todo{}},
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
				{ID: "1",  Text: "Active", Statuses: map[string]string{"completion": string(models.StatusPending)}, Items: []*models.Todo{}},
				{ID: "2",  Text: "Done", Statuses: map[string]string{"completion": string(models.StatusDone)}, Items: []*models.Todo{}},
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
			
			Text:     "Done Parent",
			Statuses: map[string]string{"completion": string(models.StatusDone)},
			Items: []*models.Todo{
				{ID: "child1", ParentID: "parent",  Text: "Pending Child", Statuses: map[string]string{"completion": string(models.StatusPending)}, Items: []*models.Todo{}},
				{ID: "child2", ParentID: "parent",  Text: "Done Child", Statuses: map[string]string{"completion": string(models.StatusDone)}, Items: []*models.Todo{}},
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
				{ID: "1",  Text: "Original", Statuses: map[string]string{"completion": string(models.StatusPending)}, Items: []*models.Todo{}},
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
			
			Text:     "Grandchild",
			Statuses: map[string]string{"completion": string(models.StatusPending)},
			Items:    []*models.Todo{},
		}

		child := &models.Todo{
			ID:       "child",
			ParentID: "parent",
			
			Text:     "Child",
			Statuses: map[string]string{"completion": string(models.StatusDone)},
			Items:    []*models.Todo{grandchild},
		}

		parent := &models.Todo{
			ID:       "parent",
			
			Text:     "Parent",
			Statuses: map[string]string{"completion": string(models.StatusPending)},
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