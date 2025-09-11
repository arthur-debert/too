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
	Todo         *models.IDMTodo
	PositionPath string           // Position path of the newly created todo (e.g., "1", "1.2")
	Mode         string           // Output mode passed from options
	AllTodos     []*models.IDMTodo // All todos for long mode
	TotalCount   int              // Total count for long mode
	DoneCount    int              // Done count for long mode
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
	var parentUID = store.RootScope
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

	// Build result
	result := &Result{
		Todo:         todo,
		PositionPath: positionPath,
		Mode:         opts.Mode,
	}

	// Add long mode data if requested
	if opts.Mode == "long" {
		result.AllTodos = manager.ListActive()
		result.TotalCount, result.DoneCount = manager.CountTodos()
	}

	return result, nil
}

