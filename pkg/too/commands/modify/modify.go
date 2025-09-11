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
	Todo    *models.Todo
	OldText string
	NewText string
}

// Execute modifies the text of an existing todo
func Execute(positionStr string, newText string, opts Options) (*Result, error) {
	if newText == "" {
		return nil, fmt.Errorf("new todo text cannot be empty")
	}

	s := store.NewStore(opts.CollectionPath)
	manager, err := store.NewManagerFromStore(s)
	if err != nil {
		return nil, fmt.Errorf("failed to create idm manager: %w", err)
	}

	uid, err := manager.Registry().ResolvePositionPath(store.RootScope, positionStr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve todo position '%s': %w", positionStr, err)
	}

	var result *Result
	err = s.Update(func(collection *models.Collection) error {
		todo := collection.FindItemByID(uid)
		if todo == nil {
			return fmt.Errorf("todo with ID '%s' not found", uid)
		}

		oldText := todo.Text
		todo.Text = newText

		result = &Result{
			Todo:    todo,
			OldText: oldText,
			NewText: newText,
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}
