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
	Todo         *models.Todo
	PositionPath string         // Position path of the newly created todo (e.g., "1", "1.2")
	Mode         string         // Output mode passed from options
	AllTodos     []*models.Todo // All todos for long mode
	TotalCount   int            // Total count for long mode
	DoneCount    int            // Done count for long mode
}

// Execute adds a new todo to the collection using pure IDM.
func Execute(text string, opts Options) (*Result, error) {
	if text == "" {
		return nil, fmt.Errorf("todo text cannot be empty")
	}

	// Create IDM store and manager
	idmStore := store.NewIDMStore(opts.CollectionPath)
	manager, err := store.NewPureIDMManager(idmStore, opts.CollectionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create manager: %w", err)
	}

	// Resolve parent if specified
	var parentUID string = store.RootScope
	if opts.ParentPath != "" {
		uid, err := manager.ResolvePositionPath(store.RootScope, opts.ParentPath)
		if err != nil {
			return nil, fmt.Errorf("parent todo not found: %w", err)
		}
		parentUID = uid
	}

	// Add the todo
	newUID, err := manager.Add(parentUID, text)
	if err != nil {
		return nil, fmt.Errorf("failed to add todo: %w", err)
	}

	// Save changes
	if err := manager.Save(); err != nil {
		return nil, fmt.Errorf("failed to save collection: %w", err)
	}

	// Get the created todo
	todo := manager.GetTodoByUID(newUID)
	if todo == nil {
		return nil, fmt.Errorf("todo not found after creation")
	}

	// Get position path
	positionPath, err := manager.GetPositionPath(store.RootScope, newUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get position path: %w", err)
	}

	// Build result (converting IDMTodo to Todo for API compatibility)
	result := &Result{
		Todo: &models.Todo{
			ID:       todo.UID,
			ParentID: todo.ParentID,
			Text:     todo.Text,
			Modified: todo.Modified,
			Items:    []*models.Todo{},
		},
		PositionPath: positionPath,
		Mode:         opts.Mode,
	}

	// Copy statuses
	if todo.Statuses != nil {
		result.Todo.Statuses = make(map[string]string)
		for k, v := range todo.Statuses {
			result.Todo.Statuses[k] = v
		}
	}

	// Add long mode data if requested
	if opts.Mode == "long" {
		allTodos := manager.ListActive()
		result.AllTodos = make([]*models.Todo, len(allTodos))
		for i, idmTodo := range allTodos {
			result.AllTodos[i] = &models.Todo{
				ID:       idmTodo.UID,
				ParentID: idmTodo.ParentID,
				Text:     idmTodo.Text,
				Modified: idmTodo.Modified,
				Items:    []*models.Todo{},
			}
			if idmTodo.Statuses != nil {
				result.AllTodos[i].Statuses = make(map[string]string)
				for k, v := range idmTodo.Statuses {
					result.AllTodos[i].Statuses[k] = v
				}
			}
		}
		result.TotalCount, result.DoneCount = manager.CountTodos()
	}

	return result, nil
}

// countTodos recursively counts total and done todos
func countTodos(todos []*models.Todo) (total int, done int) {
	for _, todo := range todos {
		total++
		if todo.GetStatus() == models.StatusDone {
			done++
		}
		// Recursively count children
		childTotal, childDone := countTodos(todo.Items)
		total += childTotal
		done += childDone
	}
	return total, done
}
