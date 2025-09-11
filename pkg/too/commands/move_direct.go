package commands

import (
	"fmt"

	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/store"
)

// MoveDirect moves a todo from one parent to another without adapters.
func MoveDirect(s store.Store, collectionPath string, sourcePath, destParentPath string) (*models.Todo, string, string, error) {
	// Create direct workflow manager
	manager, err := store.NewDirectWorkflowManager(s, collectionPath)
	if err != nil {
		return nil, "", "", err
	}

	// Resolve source todo
	sourceUID, err := manager.ResolvePositionPath(store.RootScope, sourcePath)
	if err != nil {
		return nil, "", "", fmt.Errorf("todo not found at position: %s", sourcePath)
	}

	collection := manager.GetCollection()
	sourceTodo := collection.FindItemByID(sourceUID)
	if sourceTodo == nil {
		return nil, "", "", fmt.Errorf("todo with ID '%s' not found", sourceUID)
	}

	// Resolve destination parent
	var destParentUID string = store.RootScope
	if destParentPath != "" {
		uid, err := manager.ResolvePositionPath(store.RootScope, destParentPath)
		if err != nil {
			return nil, "", "", fmt.Errorf("destination parent not found at position: %s", destParentPath)
		}
		destParentUID = uid
	}

	// Check for circular reference
	if destParentUID != store.RootScope {
		destParent := collection.FindItemByID(destParentUID)
		if destParent != nil && isDescendantOf(destParent, sourceTodo) {
			return nil, "", "", fmt.Errorf("cannot move a parent into its own descendant")
		}
	}

	// Get old parent UID
	oldParentUID := store.RootScope
	if sourceTodo.ParentID != "" {
		oldParentUID = sourceTodo.ParentID
	}

	// Store old path
	oldPath := sourcePath

	// Perform the move
	if err := manager.Move(sourceUID, oldParentUID, destParentUID); err != nil {
		return nil, "", "", fmt.Errorf("failed to move todo: %w", err)
	}

	// Get new path
	newPath, err := manager.GetPositionPath(store.RootScope, sourceUID)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to determine new position path: %w", err)
	}

	// Save the collection
	if err := manager.Save(); err != nil {
		return nil, "", "", err
	}

	return sourceTodo, oldPath, newPath, nil
}

func isDescendantOf(child, parent *models.Todo) bool {
	for _, item := range parent.Items {
		if item.ID == child.ID {
			return true
		}
		if isDescendantOf(child, item) {
			return true
		}
	}
	return false
}