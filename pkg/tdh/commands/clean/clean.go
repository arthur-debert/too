package clean

import (
	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/arthur-debert/tdh/pkg/tdh/store"
)

// Options contains options for the clean command
type Options struct {
	CollectionPath string
}

// Result contains the result of the clean command
type Result struct {
	RemovedCount int
	RemovedTodos []*models.Todo
	ActiveCount  int
}

// Execute removes finished todos from the collection
func Execute(opts Options) (*Result, error) {
	s := store.NewStore(opts.CollectionPath)
	var removedTodos []*models.Todo
	var activeCount int

	err := s.Update(func(collection *models.Collection) error {
		// Capture the todos to be removed *before* modifying the slice
		removedTodos = findDoneTodos(collection.Todos)
		activeCount = removeFinishedTodos(collection)

		// Auto-reorder after cleaning
		collection.Reorder()

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &Result{
		RemovedCount: len(removedTodos),
		RemovedTodos: removedTodos,
		ActiveCount:  activeCount,
	}, nil
}

// findDoneTodos returns a list of done todos from the given slice.
// This creates new Todo pointers to avoid issues when the original slice is modified.
func findDoneTodos(todos []*models.Todo) []*models.Todo {
	var doneTodos []*models.Todo
	for _, todo := range todos {
		if todo.Status == models.StatusDone {
			// Create a copy to avoid issues when the original slice is modified
			todoCopy := *todo
			doneTodos = append(doneTodos, &todoCopy)
		}
	}
	return doneTodos
}

// removeFinishedTodos removes all done todos from a collection.
// Returns the count of remaining active todos.
func removeFinishedTodos(c *models.Collection) int {
	var activeTodos []*models.Todo
	for _, todo := range c.Todos {
		if todo.Status != models.StatusDone {
			activeTodos = append(activeTodos, todo)
		}
	}
	c.Todos = activeTodos
	return len(activeTodos)
}
