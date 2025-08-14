package reorder

import (
	"fmt"

	"github.com/arthur-debert/tdh/pkg/tdh/internal/helpers"
	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/arthur-debert/tdh/pkg/tdh/store"
)

// Options contains options for the reorder command
type Options struct {
	CollectionPath string
}

// Result contains the result of the reorder command
type Result struct {
	TodoA *models.Todo
	TodoB *models.Todo
}

// Execute swaps the position of two todos
func Execute(idA, idB int, opts Options) (*Result, error) {
	s := store.NewStore(opts.CollectionPath)
	var todoA, todoB *models.Todo

	err := s.Update(func(collection *models.Collection) error {
		if err := helpers.Swap(collection, idA, idB); err != nil {
			return fmt.Errorf("failed to swap todos: %w", err)
		}
		todoA, _ = helpers.Find(collection, idA)
		todoB, _ = helpers.Find(collection, idB)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &Result{
		TodoA: todoA,
		TodoB: todoB,
	}, nil
}
