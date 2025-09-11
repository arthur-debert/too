package add

import (
	"fmt"

	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/store"
)

// RunDirect executes the add command using the direct workflow manager without adapters.
func RunDirect(s store.Store, text string, opts Options) (*Result, error) {
	// Create direct workflow manager
	manager, err := store.NewDirectWorkflowManager(s, s.Path())
	if err != nil {
		return nil, fmt.Errorf("failed to create direct workflow manager: %w", err)
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

	// Add the todo directly
	newUID, err := manager.Add(parentUID, text)
	if err != nil {
		return nil, fmt.Errorf("failed to add todo: %w", err)
	}

	// Save the changes
	if err := manager.Save(); err != nil {
		return nil, fmt.Errorf("failed to save collection: %w", err)
	}

	// Get the created todo using DirectWorkflowManager method
	todoInterface := manager.GetTodoByID(newUID)
	if todoInterface == nil {
		return nil, fmt.Errorf("todo with ID %s not found after creation", newUID)
	}
	todo := todoInterface.(*models.Todo)

	// Get the position path of the newly created todo
	positionPath, err := manager.GetPositionPath(store.RootScope, newUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get position path: %w", err)
	}

	result := &Result{
		Todo:         todo,
		PositionPath: positionPath,
		Mode:         opts.Mode,
	}

	// If in long mode, get additional data using manager's IDM-aware methods
	if opts.Mode == "long" {
		result.AllTodos = manager.ListActive().([]*models.Todo)
		result.TotalCount, result.DoneCount = manager.CountTodos()
	}

	return result, nil
}