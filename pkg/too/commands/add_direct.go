package commands

import (
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/store"
)

// AddDirect creates a new todo item without adapters.
func AddDirect(s store.Store, collectionPath string, text string, parentUID string) (*models.Todo, error) {
	// Create direct workflow manager
	manager, err := store.NewDirectWorkflowManager(s, collectionPath)
	if err != nil {
		return nil, err
	}

	// Determine parent scope
	parentScope := store.RootScope
	if parentUID != "" {
		parentScope = parentUID
	}

	// Add the todo
	uid, err := manager.Add(parentScope, text)
	if err != nil {
		return nil, err
	}

	// Save the collection
	if err := manager.Save(); err != nil {
		return nil, err
	}

	// Return the created todo
	return manager.GetCollection().FindItemByID(uid), nil
}