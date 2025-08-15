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
		// Capture all todos that will be removed (including descendants)
		removedTodos = findAllDoneTodosRecursive(collection.Todos)

		// Remove done todos and their descendants
		collection.Todos = removeFinishedTodosRecursive(collection.Todos)

		// Count remaining active todos
		activeCount = countActiveTodos(collection.Todos)

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

// findAllDoneTodosRecursive finds all done todos and their descendants
func findAllDoneTodosRecursive(todos []*models.Todo) []*models.Todo {
	var removedTodos []*models.Todo

	for _, todo := range todos {
		if todo.Status == models.StatusDone {
			// Add this todo and all its descendants
			removedTodos = append(removedTodos, collectTodoAndDescendants(todo)...)
		} else {
			// If not done, still check children for done items
			removedTodos = append(removedTodos, findAllDoneTodosRecursive(todo.Items)...)
		}
	}

	return removedTodos
}

// collectTodoAndDescendants collects a todo and all its descendants
func collectTodoAndDescendants(todo *models.Todo) []*models.Todo {
	// Create a copy to avoid issues when the original is modified
	todoCopy := *todo
	todoCopy.Items = make([]*models.Todo, len(todo.Items))

	// Recursively copy all descendants
	for i, child := range todo.Items {
		childCopy := *child
		todoCopy.Items[i] = &childCopy
	}

	result := []*models.Todo{&todoCopy}

	// Add all descendants
	for _, child := range todo.Items {
		result = append(result, collectTodoAndDescendants(child)...)
	}

	return result
}

// removeFinishedTodosRecursive removes done todos and their descendants
func removeFinishedTodosRecursive(todos []*models.Todo) []*models.Todo {
	var activeTodos []*models.Todo

	for _, todo := range todos {
		if todo.Status != models.StatusDone {
			// Keep this todo but recursively clean its children
			todoCopy := *todo
			todoCopy.Items = removeFinishedTodosRecursive(todo.Items)
			activeTodos = append(activeTodos, &todoCopy)
		}
		// If done, skip this todo and all its descendants
	}

	return activeTodos
}

// countActiveTodos recursively counts all active (non-done) todos
func countActiveTodos(todos []*models.Todo) int {
	count := 0
	for _, todo := range todos {
		if todo.Status != models.StatusDone {
			count++
			count += countActiveTodos(todo.Items)
		}
	}
	return count
}
