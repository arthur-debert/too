package reorder

import (
	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/arthur-debert/tdh/pkg/tdh/store"
)

// Options contains options for the reorder command
type Options struct {
	CollectionPath string
}

// Result contains the result of the reorder command
type Result struct {
	ReorderedCount int
	Todos          []*models.Todo
}

// Execute reorders todos by sorting them by their current position and reassigning sequential positions
func Execute(opts Options) (*Result, error) {
	s := store.NewStore(opts.CollectionPath)
	var reorderedTodos []*models.Todo
	var count int

	err := s.Update(func(collection *models.Collection) error {
		// Use the collection's Reorder method
		count = collection.Reorder()
		// Make a copy of the todos for the result
		reorderedTodos = make([]*models.Todo, len(collection.Todos))
		copy(reorderedTodos, collection.Todos)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &Result{
		ReorderedCount: count,
		Todos:          reorderedTodos,
	}, nil
}
