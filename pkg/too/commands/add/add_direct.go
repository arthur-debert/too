package add

import (
	"fmt"

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

	// Get the created todo
	collection := manager.GetCollection()
	todo := collection.FindItemByID(newUID)
	if todo == nil {
		return nil, fmt.Errorf("todo with ID %s not found after creation", newUID)
	}

	result := &Result{
		Todo: todo,
		Mode: opts.Mode,
	}

	// If in long mode, get additional data
	if opts.Mode == "long" {
		result.AllTodos = collection.ListActive()
		result.TotalCount, result.DoneCount = countTodos(collection.Todos)
	}

	return result, nil
}