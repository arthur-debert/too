package reorder

import (
	"sort"

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
		count = reorder(collection)
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

// reorder sorts todos by their current position and reassigns sequential positions starting from 1
// Returns the number of todos that had their position changed
func reorder(c *models.Collection) int {
	if len(c.Todos) == 0 {
		return 0
	}

	// Sort todos by their current position
	// Using a stable sort to maintain relative order of todos with same position
	sort.SliceStable(c.Todos, func(i, j int) bool {
		return c.Todos[i].Position < c.Todos[j].Position
	})

	// Reassign positions sequentially starting from 1
	changed := 0
	for i := range c.Todos {
		newPosition := i + 1
		if c.Todos[i].Position != newPosition {
			c.Todos[i].Position = newPosition
			changed++
		}
	}

	return changed
}
