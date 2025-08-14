package list

import (
	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/arthur-debert/tdh/pkg/tdh/store"
)

// Options contains options for listing todos
type Options struct {
	CollectionPath string
	ShowDone       bool
	ShowAll        bool
}

// Result contains the result of listing todos
type Result struct {
	Todos      []*models.Todo
	TotalCount int
	DoneCount  int
}

// Execute returns todos from the collection with optional filtering
func Execute(opts Options) (*Result, error) {
	s := store.NewStore(opts.CollectionPath)

	// For nested lists, we need to load the full collection to preserve hierarchy
	collection, err := s.Load()
	if err != nil {
		return nil, err
	}

	// Calculate counts
	totalCount, doneCount := countTodos(collection.Todos)

	// If showing all, return the full hierarchical structure
	if opts.ShowAll {
		return &Result{
			Todos:      collection.Todos,
			TotalCount: totalCount,
			DoneCount:  doneCount,
		}, nil
	}

	// For filtered views, we need to filter while preserving hierarchy
	status := models.StatusPending
	if opts.ShowDone {
		status = models.StatusDone
	}

	filteredTodos := filterTodosPreservingHierarchy(collection.Todos, status)

	return &Result{
		Todos:      filteredTodos,
		TotalCount: totalCount,
		DoneCount:  doneCount,
	}, nil
}

// filterTodosPreservingHierarchy filters todos by status while preserving the tree structure
// It includes a parent if it matches OR if any of its descendants match
func filterTodosPreservingHierarchy(todos []*models.Todo, status models.TodoStatus) []*models.Todo {
	var result []*models.Todo

	for _, todo := range todos {
		// Check if this todo or any descendant matches
		if todoOrDescendantMatches(todo, status) {
			// Clone the todo to avoid modifying the original
			filteredTodo := &models.Todo{
				ID:       todo.ID,
				ParentID: todo.ParentID,
				Position: todo.Position,
				Text:     todo.Text,
				Status:   todo.Status,
				Modified: todo.Modified,
				Items:    filterTodosPreservingHierarchy(todo.Items, status),
			}
			result = append(result, filteredTodo)
		}
	}

	return result
}

// todoOrDescendantMatches checks if a todo or any of its descendants matches the status
func todoOrDescendantMatches(todo *models.Todo, status models.TodoStatus) bool {
	// Check self
	if todo.Status == status {
		return true
	}

	// Check descendants
	for _, child := range todo.Items {
		if todoOrDescendantMatches(child, status) {
			return true
		}
	}

	return false
}

// countTodos recursively counts all todos and done todos in the tree
func countTodos(todos []*models.Todo) (total, done int) {
	for _, todo := range todos {
		total++
		if todo.Status == models.StatusDone {
			done++
		}

		// Count children
		childTotal, childDone := countTodos(todo.Items)
		total += childTotal
		done += childDone
	}
	return
}
