package output

import (
	"github.com/arthur-debert/too/pkg/too/models"
)

// HierarchicalTodo represents a todo with its children for display purposes
type HierarchicalTodo struct {
	*models.Todo
	Children        []*HierarchicalTodo
	EffectiveStatus string // Computed status considering children
}

// BuildHierarchy converts a flat list of Todos into a hierarchical structure
func BuildHierarchy(todos []*models.Todo) []*HierarchicalTodo {
	// Create a map for quick lookup
	todoMap := make(map[string]*HierarchicalTodo)
	var roots []*HierarchicalTodo

	// First pass: create HierarchicalTodo wrappers
	for _, todo := range todos {
		todoMap[todo.UID] = &HierarchicalTodo{
			Todo:     todo,
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

	// Third pass: compute effective status for each node
	// We need to do this recursively from leaves to roots
	var computeStatusRecursive func(*HierarchicalTodo)
	computeStatusRecursive = func(htodo *HierarchicalTodo) {
		// First compute status for all children
		for _, child := range htodo.Children {
			computeStatusRecursive(child)
		}
		// Then compute this node's effective status
		htodo.EffectiveStatus = computeEffectiveStatus(htodo)
	}
	
	// Start from roots
	for _, root := range roots {
		computeStatusRecursive(root)
	}

	return roots
}

// computeEffectiveStatus calculates the effective status for a hierarchical todo
func computeEffectiveStatus(htodo *HierarchicalTodo) string {
	// No children - return own status
	if len(htodo.Children) == 0 {
		if htodo.GetStatus() == models.StatusDone {
			return "done"
		}
		return "pending"
	}
	
	// Check children's states
	hasComplete := false
	hasPending := false
	
	for _, child := range htodo.Children {
		// Use the child's effective status which considers its own children
		switch child.EffectiveStatus {
		case "done":
			hasComplete = true
		case "pending":
			hasPending = true
		case "mixed":
			// If any child is mixed, parent is mixed
			return "mixed"
		}
		
		if hasComplete && hasPending {
			return "mixed"
		}
	}
	
	// All children have same state
	if hasComplete {
		return "done"
	}
	return "pending"
}

// FlattenHierarchy converts a hierarchical structure back to a flat list
func FlattenHierarchy(todos []*HierarchicalTodo) []*models.Todo {
	var flat []*models.Todo
	
	for _, todo := range todos {
		flat = append(flat, todo.Todo)
		if len(todo.Children) > 0 {
			flat = append(flat, FlattenHierarchy(todo.Children)...)
		}
	}
	
	return flat
}