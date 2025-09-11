package move

import (
	"fmt"

	"github.com/arthur-debert/too/pkg/logging"
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/store"
)

// Options holds the options for the move command
type Options struct {
	CollectionPath string
}

// Result represents the result of a move operation
type Result struct {
	Todo      *models.Todo
	OldPath   string
	NewPath   string
	OldParent *models.Todo
	NewParent *models.Todo
}

// Execute moves a todo from one parent to another
func Execute(sourcePath string, destParentPath string, opts Options) (*Result, error) {
	logger := logging.GetLogger("too.commands.move")
	logger.Debug().
		Str("sourcePath", sourcePath).
		Str("destParentPath", destParentPath).
		Str("collectionPath", opts.CollectionPath).
		Msg("executing move command")

	s := store.NewStore(opts.CollectionPath)
	var result *Result

	err := s.Update(func(collection *models.Collection) error {
		manager, err := store.NewManagerFromCollection(collection)
		if err != nil {
			return fmt.Errorf("failed to create idm manager: %w", err)
		}

		// Find the source todo
		sourceUID, err := manager.Registry().ResolvePositionPath(store.RootScope, sourcePath)
		if err != nil {
			return fmt.Errorf("todo not found at position: %s", sourcePath)
		}
		sourceTodo := collection.FindItemByID(sourceUID)
		if sourceTodo == nil {
			return fmt.Errorf("todo with ID '%s' not found", sourceUID)
		}

		// Determine the destination parent UID
		var destParentUID string = store.RootScope
		if destParentPath != "" {
			uid, err := manager.Registry().ResolvePositionPath(store.RootScope, destParentPath)
			if err != nil {
				return fmt.Errorf("destination parent not found at position: %s", destParentPath)
			}
			destParentUID = uid
		}

		// Check for circular reference (can't move a parent into its own child)
		if destParentUID != store.RootScope {
			destParent := collection.FindItemByID(destParentUID)
			if destParent != nil && isDescendantOf(destParent, sourceTodo) {
				logger.Error().
					Str("sourcePath", sourcePath).
					Str("destParentPath", destParentPath).
					Msg("attempted to move parent into its own descendant")
				return fmt.Errorf("cannot move a parent into its own descendant")
			}
		}

		// Store old path for result
		oldPath := sourcePath

		// Get the old parent UID for the Manager.Move() call
		oldParentUID := store.RootScope
		if sourceTodo.ParentID != "" {
			oldParentUID = sourceTodo.ParentID
		}

		// Use Manager to handle the move operation
		err = manager.Move(sourceUID, oldParentUID, destParentUID)
		if err != nil {
			return fmt.Errorf("failed to move todo: %w", err)
		}

		// Get new path using IDM registry
		adapter, err := store.NewIDMStoreAdapter(s)
		if err != nil {
			return fmt.Errorf("failed to create IDM adapter: %w", err)
		}
		newPath, err := manager.Registry().GetPositionPath(store.RootScope, sourceUID, adapter)
		if err != nil {
			logger.Error().
				Str("todoID", sourceTodo.ID).
				Str("todoText", sourceTodo.Text).
				Str("parentID", sourceTodo.ParentID).
				Int("position", sourceTodo.Position).
				Err(err).
				Msg("failed to get new position path")
			return fmt.Errorf("failed to determine new position path: %w", err)
		}

		// Get parent references for the result
		var oldParent *models.Todo
		if oldParentUID != store.RootScope {
			oldParent = collection.FindItemByID(oldParentUID)
		}
		
		var newParent *models.Todo
		if destParentUID != store.RootScope {
			newParent = collection.FindItemByID(destParentUID)
		}

		result = &Result{
			Todo:      sourceTodo,
			OldPath:   oldPath,
			NewPath:   newPath,
			OldParent: oldParent,
			NewParent: newParent,
		}

		logger.Debug().
			Str("todoID", sourceTodo.ID).
			Str("oldPath", oldPath).
			Str("newPath", newPath).
			Msg("successfully moved todo")

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func isDescendantOf(child, parent *models.Todo) bool {
	// Check all children recursively
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

