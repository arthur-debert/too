package reorder

import (
	"errors"
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
		if err := swap(collection, idA, idB); err != nil {
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

// swap swaps the position of two todos in a collection by their positions.
// Note: This also swaps the positions, which maintains the visual order.
func swap(c *models.Collection, posA, posB int) error {
	var indexA, indexB = -1, -1

	for i, todo := range c.Todos {
		if todo.Position == posA {
			indexA = i
		}
		if todo.Position == posB {
			indexB = i
		}
	}

	if indexA == -1 || indexB == -1 {
		return errors.New("one or both todos not found")
	}

	c.Todos[indexA], c.Todos[indexB] = c.Todos[indexB], c.Todos[indexA]
	c.Todos[indexA].Position, c.Todos[indexB].Position = c.Todos[indexB].Position, c.Todos[indexA].Position
	return nil
}
