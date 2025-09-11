package commands

import (
	"fmt"

	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/store"
)

// CompleteDirect marks todos as complete without adapters.
func CompleteDirect(s store.Store, collectionPath string, uids []string) ([]*models.Todo, error) {
	// Create direct workflow manager
	manager, err := store.NewDirectWorkflowManager(s, collectionPath)
	if err != nil {
		return nil, err
	}

	var completed []*models.Todo
	collection := manager.GetCollection()

	// Complete each todo
	for _, uid := range uids {
		todo := collection.FindItemByID(uid)
		if todo == nil {
			return nil, fmt.Errorf("todo with ID %s not found", uid)
		}

		// Set status to done
		if err := manager.SetStatus(uid, "completion", string(models.StatusDone)); err != nil {
			return nil, err
		}

		completed = append(completed, todo)
	}

	// Save the collection
	if err := manager.Save(); err != nil {
		return nil, err
	}

	return completed, nil
}