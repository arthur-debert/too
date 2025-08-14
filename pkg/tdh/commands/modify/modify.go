package modify

import (
	"fmt"

	"github.com/arthur-debert/tdh/pkg/tdh/internal/helpers"
	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/arthur-debert/tdh/pkg/tdh/store"
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
func Execute(id int, newText string, opts Options) (*Result, error) {
	if newText == "" {
		return nil, fmt.Errorf("new todo text cannot be empty")
	}

	s := store.NewStore(opts.CollectionPath)
	var todo *models.Todo
	var oldText string

	err := s.Update(func(collection *models.Collection) error {
		var err error
		todo, err = helpers.Find(collection, id)
		if err != nil {
			return fmt.Errorf("todo not found: %w", err)
		}
		oldText = todo.Text
		todo.Text = newText
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &Result{
		Todo:    todo,
		OldText: oldText,
		NewText: newText,
	}, nil
}
