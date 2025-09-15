package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildHierarchy(t *testing.T) {
	t.Run("Simple parent-child relationship", func(t *testing.T) {
		parent := &Todo{
			UID:          "1",
			Text:         "Parent",
			PositionPath: "1",
		}
		child := &Todo{
			UID:          "2",
			Text:         "Child", 
			ParentID:     "1",
			PositionPath: "1.1",
		}
		
		todos := []*Todo{parent, child}
		hierarchical := BuildHierarchy(todos)
		
		require.Len(t, hierarchical, 1)
		assert.Equal(t, "Parent", hierarchical[0].Text)
		assert.Len(t, hierarchical[0].Children, 1)
		assert.Equal(t, "Child", hierarchical[0].Children[0].Text)
	})

	t.Run("Multiple roots", func(t *testing.T) {
		todo1 := &Todo{
			UID:          "1",
			Text:         "First Root",
			PositionPath: "1",
		}
		todo2 := &Todo{
			UID:          "2", 
			Text:         "Second Root",
			PositionPath: "2",
		}
		
		todos := []*Todo{todo1, todo2}
		hierarchical := BuildHierarchy(todos)
		
		require.Len(t, hierarchical, 2)
		assert.Equal(t, "First Root", hierarchical[0].Text)
		assert.Equal(t, "Second Root", hierarchical[1].Text)
		assert.Len(t, hierarchical[0].Children, 0)
		assert.Len(t, hierarchical[1].Children, 0)
	})

	t.Run("Orphaned child becomes root", func(t *testing.T) {
		child := &Todo{
			UID:          "2",
			Text:         "Orphaned Child",
			ParentID:     "nonexistent",
			PositionPath: "1.1",
		}
		
		todos := []*Todo{child}
		hierarchical := BuildHierarchy(todos)
		
		require.Len(t, hierarchical, 1)
		assert.Equal(t, "Orphaned Child", hierarchical[0].Text)
		assert.Len(t, hierarchical[0].Children, 0)
	})
}

func TestComputeEffectiveStatus(t *testing.T) {
	t.Run("Leaf node - pending", func(t *testing.T) {
		todo := &Todo{
			UID:      "1", 
			Text:     "Test",
			Statuses: map[string]string{"completion": string(StatusPending)},
		}
		htodo := &HierarchicalTodo{Todo: todo, Children: []*HierarchicalTodo{}}
		
		status := ComputeEffectiveStatus(htodo)
		assert.Equal(t, "pending", status)
	})

	t.Run("Leaf node - done", func(t *testing.T) {
		todo := &Todo{
			UID:      "1",
			Text:     "Test", 
			Statuses: map[string]string{"completion": string(StatusDone)},
		}
		htodo := &HierarchicalTodo{Todo: todo, Children: []*HierarchicalTodo{}}
		
		status := ComputeEffectiveStatus(htodo)
		assert.Equal(t, "done", status)
	})

	t.Run("Parent with all done children", func(t *testing.T) {
		parent := &Todo{
			UID:      "1",
			Text:     "Parent",
			Statuses: map[string]string{"completion": string(StatusPending)},
		}
		child1 := &HierarchicalTodo{
			Todo:            &Todo{UID: "2", Text: "Child1"},
			EffectiveStatus: "done",
		}
		child2 := &HierarchicalTodo{
			Todo:            &Todo{UID: "3", Text: "Child2"},
			EffectiveStatus: "done",
		}
		htodo := &HierarchicalTodo{
			Todo:     parent,
			Children: []*HierarchicalTodo{child1, child2},
		}
		
		status := ComputeEffectiveStatus(htodo)
		assert.Equal(t, "done", status)
	})

	t.Run("Parent with all pending children", func(t *testing.T) {
		parent := &Todo{
			UID:      "1",
			Text:     "Parent",
			Statuses: map[string]string{"completion": string(StatusDone)},
		}
		child1 := &HierarchicalTodo{
			Todo:            &Todo{UID: "2", Text: "Child1"},
			EffectiveStatus: "pending",
		}
		child2 := &HierarchicalTodo{
			Todo:            &Todo{UID: "3", Text: "Child2"},
			EffectiveStatus: "pending",
		}
		htodo := &HierarchicalTodo{
			Todo:     parent,
			Children: []*HierarchicalTodo{child1, child2},
		}
		
		status := ComputeEffectiveStatus(htodo)
		assert.Equal(t, "pending", status)
	})

	t.Run("Parent with mixed children", func(t *testing.T) {
		parent := &Todo{
			UID:      "1",
			Text:     "Parent",
			Statuses: map[string]string{"completion": string(StatusPending)},
		}
		child1 := &HierarchicalTodo{
			Todo:            &Todo{UID: "2", Text: "Child1"},
			EffectiveStatus: "done",
		}
		child2 := &HierarchicalTodo{
			Todo:            &Todo{UID: "3", Text: "Child2"}, 
			EffectiveStatus: "pending",
		}
		htodo := &HierarchicalTodo{
			Todo:     parent,
			Children: []*HierarchicalTodo{child1, child2},
		}
		
		status := ComputeEffectiveStatus(htodo)
		assert.Equal(t, "mixed", status)
	})

	t.Run("Parent with mixed child propagates mixed", func(t *testing.T) {
		parent := &Todo{
			UID:      "1",
			Text:     "Parent",
			Statuses: map[string]string{"completion": string(StatusPending)},
		}
		child1 := &HierarchicalTodo{
			Todo:            &Todo{UID: "2", Text: "Child1"},
			EffectiveStatus: "mixed",
		}
		child2 := &HierarchicalTodo{
			Todo:            &Todo{UID: "3", Text: "Child2"},
			EffectiveStatus: "done",
		}
		htodo := &HierarchicalTodo{
			Todo:     parent,
			Children: []*HierarchicalTodo{child1, child2},
		}
		
		status := ComputeEffectiveStatus(htodo)
		assert.Equal(t, "mixed", status)
	})
}

func TestFlattenHierarchy(t *testing.T) {
	t.Run("Single level", func(t *testing.T) {
		todo1 := &Todo{UID: "1", Text: "First"}
		todo2 := &Todo{UID: "2", Text: "Second"}
		
		hierarchical := []*HierarchicalTodo{
			{Todo: todo1, Children: []*HierarchicalTodo{}},
			{Todo: todo2, Children: []*HierarchicalTodo{}},
		}
		
		flat := FlattenHierarchy(hierarchical)
		require.Len(t, flat, 2)
		assert.Equal(t, "First", flat[0].Text)
		assert.Equal(t, "Second", flat[1].Text)
	})

	t.Run("Nested hierarchy", func(t *testing.T) {
		parent := &Todo{UID: "1", Text: "Parent"}
		child := &Todo{UID: "2", Text: "Child"}
		
		hierarchical := []*HierarchicalTodo{
			{
				Todo: parent,
				Children: []*HierarchicalTodo{
					{Todo: child, Children: []*HierarchicalTodo{}},
				},
			},
		}
		
		flat := FlattenHierarchy(hierarchical)
		require.Len(t, flat, 2)
		assert.Equal(t, "Parent", flat[0].Text)
		assert.Equal(t, "Child", flat[1].Text)
	})
}