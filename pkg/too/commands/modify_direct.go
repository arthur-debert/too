package commands

import (
	"fmt"

	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/store"
)

// ModifyDirect modifies a todo's text without adapters.
func ModifyDirect(s store.Store, collectionPath string, positionStr string, newText string) (*models.Todo, string, error) {
	if newText == "" {
		return nil, "", fmt.Errorf("new todo text cannot be empty")
	}

	// Create direct workflow manager
	manager, err := store.NewDirectWorkflowManager(s, collectionPath)
	if err != nil {
		return nil, "", err
	}

	// Resolve position to UID
	uid, err := manager.ResolvePositionPath(store.RootScope, positionStr)
	if err != nil {
		return nil, "", fmt.Errorf("failed to resolve todo position '%s': %w", positionStr, err)
	}

	// Find and modify the todo
	collection := manager.GetCollection()
	todo := collection.FindItemByID(uid)
	if todo == nil {
		return nil, "", fmt.Errorf("todo with ID '%s' not found", uid)
	}

	oldText := todo.Text
	todo.Text = newText
	todo.SetModified()

	// Save the collection
	if err := manager.Save(); err != nil {
		return nil, "", err
	}

	return todo, oldText, nil
}