package output

import (
	"github.com/arthur-debert/too/pkg/too/models"
)

// HierarchicalTodo represents a todo with its children for display purposes
type HierarchicalTodo struct {
	*models.IDMTodo
	Children []*HierarchicalTodo
}

// BuildHierarchy converts a flat list of IDMTodos into a hierarchical structure
func BuildHierarchy(todos []*models.IDMTodo) []*HierarchicalTodo {
	// Create a map for quick lookup
	todoMap := make(map[string]*HierarchicalTodo)
	var roots []*HierarchicalTodo

	// First pass: create HierarchicalTodo wrappers
	for _, todo := range todos {
		todoMap[todo.UID] = &HierarchicalTodo{
			IDMTodo:  todo,
			Children: []*HierarchicalTodo{},
		}
	}

	// Second pass: build parent-child relationships
	for _, todo := range todos {
		htodo := todoMap[todo.UID]
		
		if todo.ParentID == "" {
			// This is a root todo
			roots = append(roots, htodo)
		} else if parent, exists := todoMap[todo.ParentID]; exists {
			// Add as child to parent
			parent.Children = append(parent.Children, htodo)
		} else {
			// Parent not in list, treat as root
			roots = append(roots, htodo)
		}
	}

	return roots
}

// FlattenHierarchy converts a hierarchical structure back to a flat list
func FlattenHierarchy(todos []*HierarchicalTodo) []*models.IDMTodo {
	var flat []*models.IDMTodo
	
	for _, todo := range todos {
		flat = append(flat, todo.IDMTodo)
		if len(todo.Children) > 0 {
			flat = append(flat, FlattenHierarchy(todo.Children)...)
		}
	}
	
	return flat
}