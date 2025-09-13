package complete

import (
	"fmt"

	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/store"
)

// Options contains options for the complete command
type Options struct {
	CollectionPath string
}

// Result contains the result of the complete command
type Result struct {
	Todo       *models.IDMTodo
	OldStatus  string
	NewStatus  string
	AllTodos   []*models.IDMTodo // All todos after the operation
	TotalCount int               // Total count after the operation
	DoneCount  int               // Done count after the operation
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
	
	// Set the position path that was used to find it
	idmTodo.PositionPath = positionPath

	// Get all todos and counts
	allTodos := manager.ListActive()
	// CRITICAL: Use active-only position paths for consecutive IDs in command output
	manager.AttachActiveOnlyPositionPaths(allTodos)
	totalCount, doneCount := manager.CountTodos()

	// Build result
	result := &Result{
		Todo:       idmTodo,
		OldStatus:  oldStatus,
		NewStatus:  newStatus,
		AllTodos:   allTodos,
		TotalCount: totalCount,
		DoneCount:  doneCount,
	}

	return result, nil
}


