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

// Execute moves a todo from one parent to another using the pure IDM data model.
// This function now uses IDM internally but maintains backward compatibility 
// by returning the traditional Result format.
func Execute(sourcePath string, destParentPath string, opts Options) (*Result, error) {
	// Use IDM implementation and convert result for backward compatibility
	idmResult, err := ExecuteIDM(sourcePath, destParentPath, opts)
	if err != nil {
		return nil, err
	}

	return ConvertIDMResultToResult(idmResult), nil
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

