package output

import (
	"github.com/arthur-debert/too/pkg/too/models"
)

// ContextualNode represents a node in the contextual view
type ContextualNode struct {
	Todo               *models.HierarchicalTodo
	ShowEllipsisBefore bool
	ShowEllipsisAfter  bool
	SiblingsBefore     []*models.HierarchicalTodo
	SiblingsAfter      []*models.HierarchicalTodo
	Children           []*ContextualNode
}

// buildContextualView creates a contextual view of the hierarchy focused on the highlighted item
func buildContextualView(hierarchy []*models.HierarchicalTodo, highlightID string) *ContextualNode {
	// Find the path to the highlighted item
	path := findPathToHighlight(hierarchy, highlightID, []string{})
	if path == nil {
		return nil
	}

	// Build the contextual view following the path
	return buildContextFromPath(hierarchy, path, 0)
}

// findPathToHighlight finds the path of UIDs from root to the highlighted item
func findPathToHighlight(todos []*models.HierarchicalTodo, highlightID string, currentPath []string) []string {
	for _, todo := range todos {
		newPath := append(append([]string{}, currentPath...), todo.UID)
		
		if todo.UID == highlightID {
			return newPath
		}
		
		if todo.Children != nil {
			if found := findPathToHighlight(todo.Children, highlightID, newPath); found != nil {
				return found
			}
		}
	}
	return nil
}

// buildContextFromPath builds the contextual view following the path
func buildContextFromPath(todos []*models.HierarchicalTodo, path []string, pathIndex int) *ContextualNode {
	if pathIndex >= len(path) {
		return nil
	}

	targetUID := path[pathIndex]
	isLastInPath := pathIndex == len(path)-1

	// Find the target todo and its index
	var targetTodo *models.HierarchicalTodo
	var targetIndex int
	for i, todo := range todos {
		if todo.UID == targetUID {
			targetTodo = todo
			targetIndex = i
			break
		}
	}

	if targetTodo == nil {
		return nil
	}

	node := &ContextualNode{
		Todo: targetTodo,
	}

	if isLastInPath {
		// This is the highlighted item - add context siblings
		const contextSize = 2

		// Calculate siblings before
		startIdx := targetIndex - contextSize
		if startIdx < 0 {
			startIdx = 0
		} else if startIdx > 0 {
			node.ShowEllipsisBefore = true
		}

		// Add siblings before
		for i := startIdx; i < targetIndex; i++ {
			node.SiblingsBefore = append(node.SiblingsBefore, todos[i])
		}

		// Calculate siblings after
		endIdx := targetIndex + contextSize + 1
		if endIdx > len(todos) {
			endIdx = len(todos)
		} else if endIdx < len(todos) {
			node.ShowEllipsisAfter = true
		}

		// Add siblings after
		for i := targetIndex + 1; i < endIdx; i++ {
			node.SiblingsAfter = append(node.SiblingsAfter, todos[i])
		}
	} else {
		// This is a parent node - just add the child in the path
		if targetTodo.Children != nil {
			childNode := buildContextFromPath(targetTodo.Children, path, pathIndex+1)
			if childNode != nil {
				node.Children = []*ContextualNode{childNode}
			}
		}
	}

	return node
}