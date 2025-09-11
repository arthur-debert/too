package move

import (
	"github.com/arthur-debert/too/pkg/too/models"
)

// Options holds the options for the move command
type Options struct {
	CollectionPath string
}

// Result represents the result of a move operation
type Result struct {
	Todo      *models.Todo
	OldPath   string
	NewPath   string
	OldParent *models.Todo
	NewParent *models.Todo
}

// Execute moves a todo from one parent to another
func Execute(sourcePath string, destParentPath string, opts Options) (*Result, error) {
	return ExecuteDirect(sourcePath, destParentPath, opts)
}

func isDescendantOf(child, parent *models.Todo) bool {
	// Check all children recursively
	for _, item := range parent.Items {
		if item.ID == child.ID {
			return true
		}
		if isDescendantOf(child, item) {
			return true
		}
	}
	return false
}

