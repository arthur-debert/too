package add

import (
	"fmt"

	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/arthur-debert/tdh/pkg/tdh/store"
)

// Options contains options for the add command
type Options struct {
	CollectionPath string
}

// Result contains the result of the add command
type Result struct {
	Todo *models.Todo
}

// Execute adds a new todo to the collection
func Execute(text string, opts Options) (*Result, error) {
	if text == "" {
		return nil, fmt.Errorf("todo text cannot be empty")
	}

	s := store.NewStore(opts.CollectionPath)
	var todo *models.Todo

	err := s.Update(func(collection *models.Collection) error {
		var err error
		todo, err = collection.CreateTodo(text, "")
		return err
	})

	if err != nil {
		return nil, fmt.Errorf("failed to add todo: %w", err)
	}

	return &Result{Todo: todo}, nil
}
