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

// Execute returns todos from the collection using the pure IDM data model.
// This function now uses IDM internally but maintains backward compatibility 
// by returning the traditional Result format.
func Execute(opts Options) (*Result, error) {
	// Use IDM implementation and convert result for backward compatibility
	idmResult, err := ExecuteIDM(opts)
	if err != nil {
		return nil, err
	}

	return ConvertIDMResultToResult(idmResult), nil
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
