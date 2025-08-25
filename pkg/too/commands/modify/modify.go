package modify

import (
	"fmt"

	"github.com/arthur-debert/too/pkg/too/internal/helpers"
	"github.com/arthur-debert/too/pkg/too/models"
)

// Options contains options for the modify command
type Options struct {
	CollectionPath string
}

// Result contains the result of the modify command
type Result struct {
	Todo    *models.Todo
	OldText string
	NewText string
}

// Execute modifies the text of an existing todo
func Execute(position int, newText string, opts Options) (*Result, error) {
	if newText == "" {
		return nil, fmt.Errorf("new todo text cannot be empty")
	}

	var result *Result

	err := helpers.TransactOnTodo(opts.CollectionPath, position, func(todo *models.Todo, collection *models.Collection) error {
		oldText := todo.Text
		todo.Text = newText

		// Capture result
		result = &Result{
			Todo:    todo,
			OldText: oldText,
			NewText: newText,
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("todo not found: %w", err)
	}

	return result, nil
}
