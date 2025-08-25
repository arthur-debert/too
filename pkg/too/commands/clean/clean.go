package clean

import (
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/store"
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
		// Find all done items (not their pending descendants)
		removedTodos = findDoneItems(collection.Todos)

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

// findDoneItems finds all done todos (not including their pending descendants)
func findDoneItems(todos []*models.Todo) []*models.Todo {
	var doneItems []*models.Todo
	for _, todo := range todos {
		if todo.Status == models.StatusDone {
			doneItems = append(doneItems, todo.Clone())
		}
		// Always recurse, as a pending parent can have done children
		doneItems = append(doneItems, findDoneItems(todo.Items)...)
	}
	return doneItems
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
