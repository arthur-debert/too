package list

import (
	"github.com/arthur-debert/too/pkg/too/models"
)

// Options contains options for listing todos
type Options struct {
	CollectionPath string
	ShowDone       bool
	ShowAll        bool
}

// Result contains the result of listing todos
type Result struct {
	Todos      []*models.Todo
	TotalCount int
	DoneCount  int
}

// Execute returns todos from the collection using the appropriate manager.
// This function automatically detects the storage format and uses the correct
// implementation while maintaining backward compatibility.
func Execute(opts Options) (*Result, error) {
	// Use unified implementation that auto-detects storage format
	unifiedResult, err := ExecuteUnified(opts)
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
