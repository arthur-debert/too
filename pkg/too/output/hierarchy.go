package output

import (
	"github.com/arthur-debert/too/pkg/too/models"
)

// HierarchicalTodo represents a todo with its children for display purposes
type HierarchicalTodo struct {
	*models.IDMTodo
	Children        []*HierarchicalTodo
	EffectiveStatus string // Computed status considering children
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

	// Third pass: compute effective status for each node
	for _, htodo := range todoMap {
		htodo.EffectiveStatus = computeEffectiveStatus(htodo)
	}

	return roots
}

// computeEffectiveStatus calculates the effective status for a hierarchical todo
func computeEffectiveStatus(htodo *HierarchicalTodo) string {
	// Check if deleted
	if status, exists := htodo.GetWorkflowStatus("status"); exists && status == "deleted" {
		return "deleted"
	}
	
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