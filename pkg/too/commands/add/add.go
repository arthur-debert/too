package add

import (
	"fmt"

	"github.com/arthur-debert/too/pkg/too/models"
)

// Options contains options for the add command
type Options struct {
	CollectionPath string
	ParentPath     string // Position path of parent todo (e.g., "1.2")
	Mode           string // Output mode: "short" or "long"
}

// Result contains the result of the add command
type Result struct {
	Todo         *models.Todo
	PositionPath string         // Position path of the newly created todo (e.g., "1", "1.2")
	Mode         string         // Output mode passed from options
	AllTodos     []*models.Todo // All todos for long mode
	TotalCount   int            // Total count for long mode
	DoneCount    int            // Done count for long mode
}

// Execute adds a new todo to the collection using the appropriate manager.
// This function automatically detects the storage format and uses the correct
// implementation while maintaining backward compatibility.
func Execute(text string, opts Options) (*Result, error) {
	if text == "" {
		return nil, fmt.Errorf("todo text cannot be empty")
	}

	// Use unified implementation that auto-detects storage format
	unifiedResult, err := ExecuteUnified(text, opts)
	if err != nil {
		return nil, err
	}

	return ConvertUnifiedToResult(unifiedResult), nil
}

// countTodos recursively counts total and done todos
func countTodos(todos []*models.Todo) (total int, done int) {
	for _, todo := range todos {
		total++
		if todo.GetStatus() == models.StatusDone {
			done++
		}
		// Recursively count children
		childTotal, childDone := countTodos(todo.Items)
		total += childTotal
		done += childDone
	}
	return total, done
}
