package add

import (
	"fmt"

	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/store"
)

// Options contains options for the add command
type Options struct {
	CollectionPath string
	ParentPath     string // Position path of parent todo (e.g., "1.2")
	Mode           string // Output mode: "short" or "long"
}

// Result contains the result of the add command
type Result struct {
	Todo       *models.Todo
	Mode       string         // Output mode passed from options
	AllTodos   []*models.Todo // All todos for long mode
	TotalCount int            // Total count for long mode
	DoneCount  int            // Done count for long mode
}

// Execute adds a new todo to the collection
func Execute(text string, opts Options) (*Result, error) {
	if text == "" {
		return nil, fmt.Errorf("todo text cannot be empty")
	}

	s := store.NewStore(opts.CollectionPath)
	var todo *models.Todo

	err := s.Update(func(collection *models.Collection) error {
		var parentID string

		// If parent path is specified, resolve it to a UID
		if opts.ParentPath != "" {
			manager, err := store.NewManagerFromStore(s)
			if err != nil {
				return fmt.Errorf("failed to create idm manager: %w", err)
			}

			uid, err := manager.Registry().ResolvePositionPath(store.RootScope, opts.ParentPath)
			if err != nil {
				return fmt.Errorf("parent todo not found: %w", err)
			}
			parentID = uid
		}

		var err error
		todo, err = collection.CreateTodo(text, parentID)
		return err
	})

	if err != nil {
		return nil, fmt.Errorf("failed to add todo: %w", err)
	}

	result := &Result{
		Todo: todo,
		Mode: opts.Mode,
	}

	// If in long mode, get all active todos
	if opts.Mode == "long" {
		// Reload to get the fresh state including the newly added todo
		collection, err := s.Load()
		if err != nil {
			return nil, fmt.Errorf("failed to load collection for long mode: %w", err)
		}

		result.AllTodos = collection.ListActive()
		result.TotalCount, result.DoneCount = countTodos(collection.Todos)
	}

	return result, nil
}

// countTodos recursively counts total and done todos
func countTodos(todos []*models.Todo) (total int, done int) {
	for _, todo := range todos {
		total++
		if todo.Status == models.StatusDone {
			done++
		}
		// Recursively count children
		childTotal, childDone := countTodos(todo.Items)
		total += childTotal
		done += childDone
	}
	return total, done
}
