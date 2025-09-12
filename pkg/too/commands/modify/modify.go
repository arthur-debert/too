package modify

import (
	"fmt"

	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/store"
)

// Options contains options for the modify command
type Options struct {
	CollectionPath string
}

// Result contains the result of the modify command
type Result struct {
	Todo    *models.IDMTodo
	OldText string
	NewText string
}

// Execute modifies the text of an existing todo using the pure IDM manager.
func Execute(positionStr string, newText string, opts Options) (*Result, error) {
	if newText == "" {
		return nil, fmt.Errorf("new todo text cannot be empty")
	}

	idmStore := store.NewIDMStore(opts.CollectionPath)
	
	// Create pure IDM workflow manager
	manager, err := store.NewPureIDMManager(idmStore, opts.CollectionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create pure IDM manager: %w", err)
	}

	// Resolve position to UID
	uid, err := manager.ResolvePositionPath(store.RootScope, positionStr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve todo position '%s': %w", positionStr, err)
	}

	// Find and modify the todo using PureIDMManager method
	todo := manager.GetTodoByUID(uid)
	if todo == nil {
		return nil, fmt.Errorf("todo with UID '%s' not found", uid)
	}

	oldText := todo.Text
	todo.Text = newText
	todo.SetModified()

	// Save the collection
	if err := manager.Save(); err != nil {
		return nil, err
	}

	return &Result{
		Todo:    todo,
		OldText: oldText,
		NewText: newText,
	}, nil
}
