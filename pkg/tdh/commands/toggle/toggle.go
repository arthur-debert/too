package toggle

import (
	"fmt"

	"github.com/arthur-debert/tdh/pkg/tdh/internal/helpers"
	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/arthur-debert/tdh/pkg/tdh/store"
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
func Execute(id int, opts Options) (*Result, error) {
	s := store.NewStore(opts.CollectionPath)
	var todo *models.Todo
	var oldStatus string
	var newStatus string

	err := s.Update(func(collection *models.Collection) error {
		var err error
		todo, err = helpers.Find(collection, id)
		if err != nil {
			return fmt.Errorf("todo not found: %w", err)
		}
		oldStatus = string(todo.Status)
		todo.Toggle()
		newStatus = string(todo.Status)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &Result{
		Todo:      todo,
		OldStatus: oldStatus,
		NewStatus: newStatus,
	}, nil
}
