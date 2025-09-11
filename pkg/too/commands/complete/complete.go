package complete

import (
	"fmt"

	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/store"
)

// Options contains options for the complete command
type Options struct {
	CollectionPath string
	Mode           string // Output mode: "short" or "long"
}

// Result contains the result of the complete command
type Result struct {
	Todo       *models.Todo
	OldStatus  string
	NewStatus  string
	Mode       string         // Output mode passed from options
	AllTodos   []*models.Todo // All todos for long mode
	TotalCount int            // Total count for long mode
	DoneCount  int            // Done count for long mode
}

// Execute marks a todo as complete using pure IDM.
func Execute(positionPath string, opts Options) (*Result, error) {
	// Create IDM store and manager
	idmStore := store.NewIDMStore(opts.CollectionPath)
	manager, err := store.NewPureIDMManager(idmStore, opts.CollectionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create manager: %w", err)
	}

	// Resolve position path
	uid, err := manager.ResolvePositionPath(store.RootScope, positionPath)
	if err != nil {
		return nil, fmt.Errorf("todo not found at position %s: %w", positionPath, err)
	}

	// Get current status
	oldStatus, err := manager.GetStatus(uid, "completion")
	if err != nil {
		oldStatus = string(models.StatusPending)
	}

	// Set to done
	newStatus := string(models.StatusDone)
	if err := manager.SetStatus(uid, "completion", newStatus); err != nil {
		return nil, fmt.Errorf("failed to mark todo as complete: %w", err)
	}

	// Save changes
	if err := manager.Save(); err != nil {
		return nil, fmt.Errorf("failed to save collection: %w", err)
	}

	// Get the todo
	idmTodo := manager.GetTodoByUID(uid)
	if idmTodo == nil {
		return nil, fmt.Errorf("todo not found after update")
	}

	// Build result
	result := &Result{
		Todo: &models.Todo{
			ID:       idmTodo.UID,
			ParentID: idmTodo.ParentID,
			Text:     idmTodo.Text,
			Modified: idmTodo.Modified,
			Items:    []*models.Todo{},
		},
		OldStatus: oldStatus,
		NewStatus: newStatus,
		Mode:      opts.Mode,
	}

	// Copy statuses
	if idmTodo.Statuses != nil {
		result.Todo.Statuses = make(map[string]string)
		for k, v := range idmTodo.Statuses {
			result.Todo.Statuses[k] = v
		}
	}

	// Add long mode data if requested
	if opts.Mode == "long" {
		allTodos := manager.ListActive()
		result.AllTodos = make([]*models.Todo, len(allTodos))
		for i, todo := range allTodos {
			result.AllTodos[i] = &models.Todo{
				ID:       todo.UID,
				ParentID: todo.ParentID,
				Text:     todo.Text,
				Modified: todo.Modified,
				Items:    []*models.Todo{},
			}
			if todo.Statuses != nil {
				result.AllTodos[i].Statuses = make(map[string]string)
				for k, v := range todo.Statuses {
					result.AllTodos[i].Statuses[k] = v
				}
			}
		}
		result.TotalCount, result.DoneCount = manager.CountTodos()
	}

	return result, nil
}


