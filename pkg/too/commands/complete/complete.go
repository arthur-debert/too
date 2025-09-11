package complete

import (
	"github.com/arthur-debert/too/pkg/too/models"
)

// Options contains options for the complete command
type Options struct {
	CollectionPath string
	Mode           string // Output mode: "short" or "long"
}

// Result contains the result of the complete command
type Result struct {
	Todo       *models.Todo
	OldStatus  string
	NewStatus  string
	Mode       string         // Output mode passed from options
	AllTodos   []*models.Todo // All todos for long mode
	TotalCount int            // Total count for long mode
	DoneCount  int            // Done count for long mode
}

// Execute marks a todo as complete using the appropriate manager.
// This function automatically detects the storage format and uses the correct
// implementation while maintaining backward compatibility.
func Execute(positionPath string, opts Options) (*Result, error) {
	// Use unified implementation that auto-detects storage format
	unifiedResult, err := ExecuteUnified(positionPath, opts)
	if err != nil {
		return nil, err
	}

	return ConvertUnifiedToResult(unifiedResult), nil
}


