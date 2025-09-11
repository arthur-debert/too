package modify

import (
	"fmt"

	"github.com/arthur-debert/too/pkg/too/store"
)

// ExecuteDirect modifies the text of an existing todo without adapters.
func ExecuteDirect(positionStr string, newText string, opts Options) (*Result, error) {
	if newText == "" {
		return nil, fmt.Errorf("new todo text cannot be empty")
	}

	s := store.NewStore(opts.CollectionPath)
	
	// Create direct workflow manager
	manager, err := store.NewDirectWorkflowManager(s, opts.CollectionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create direct workflow manager: %w", err)
	}

	// Resolve position to UID
	uid, err := manager.ResolvePositionPath(store.RootScope, positionStr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve todo position '%s': %w", positionStr, err)
	}

	// Find and modify the todo
	collection := manager.GetCollection()
	todo := collection.FindItemByID(uid)
	if todo == nil {
		return nil, fmt.Errorf("todo with ID '%s' not found", uid)
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