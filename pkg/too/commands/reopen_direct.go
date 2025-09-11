package commands

import (
	"fmt"

	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/store"
)

// ReopenDirect marks todos as pending without adapters.
func ReopenDirect(s store.Store, collectionPath string, ref string, recursive bool) ([]*models.Todo, error) {
	// Create direct workflow manager
	manager, err := store.NewDirectWorkflowManager(s, collectionPath)
	if err != nil {
		return nil, err
	}

	collection := manager.GetCollection()
	var uid string

	// Try to find by short ID first (works for any status)
	todo, err := collection.FindItemByShortID(ref)
	if err == nil && todo != nil {
		uid = todo.ID
	} else {
		// If not found by short ID, try position resolution (only works for active items)
		uid, err = manager.ResolvePositionPath(store.RootScope, ref)
		if err != nil {
			return nil, fmt.Errorf("todo not found with reference: %s", ref)
		}
	}

	todo = collection.FindItemByID(uid)
	if todo == nil {
		return nil, fmt.Errorf("todo with ID %s not found", uid)
	}

	var reopened []*models.Todo

	// Reopen the todo
	if err := manager.SetStatus(uid, "completion", string(models.StatusPending)); err != nil {
		return nil, err
	}
	reopened = append(reopened, todo)

	// Reopen children recursively if requested
	if recursive {
		err := reopenChildren(manager, todo, &reopened)
		if err != nil {
			return nil, err
		}
	}

	// Save the collection
	if err := manager.Save(); err != nil {
		return nil, err
	}

	return reopened, nil
}

func reopenChildren(manager *store.DirectWorkflowManager, parent *models.Todo, reopened *[]*models.Todo) error {
	for _, child := range parent.Items {
		if err := manager.SetStatus(child.ID, "completion", string(models.StatusPending)); err != nil {
			return err
		}
		*reopened = append(*reopened, child)

		// Recursively reopen grandchildren
		if err := reopenChildren(manager, child, reopened); err != nil {
			return err
		}
	}
	return nil
}