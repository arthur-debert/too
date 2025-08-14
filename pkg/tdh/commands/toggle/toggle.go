package toggle

import (
	"fmt"

	"github.com/arthur-debert/tdh/pkg/tdh/internal/helpers"
	"github.com/arthur-debert/tdh/pkg/tdh/models"
)

// Options contains options for the toggle command
type Options struct {
	CollectionPath string
}

// Result contains the result of the toggle command
type Result struct {
	Todo      *models.Todo
	OldStatus string
	NewStatus string
}

// Execute toggles the status of a todo
func Execute(position int, opts Options) (*Result, error) {
	var result *Result

	err := helpers.TransactOnTodo(opts.CollectionPath, position, func(todo *models.Todo, collection *models.Collection) error {
		oldStatus := string(todo.Status)
		todo.Toggle()
		newStatus := string(todo.Status)

		// Auto-reorder after toggle
		collection.Reorder()

		// Capture result
		result = &Result{
			Todo:      todo,
			OldStatus: oldStatus,
			NewStatus: newStatus,
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("todo not found: %w", err)
	}

	return result, nil
}
