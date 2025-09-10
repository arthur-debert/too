package swap

import (
	"fmt"

	"github.com/arthur-debert/too/pkg/idm"
	"github.com/arthur-debert/too/pkg/logging"
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/store"
)

// Options holds the options for the swap command
type Options struct {
	CollectionPath string
}

// Result represents the result of a swap operation
type Result struct {
	Todo      *models.Todo
	OldPath   string
	NewPath   string
	OldParent *models.Todo
	NewParent *models.Todo
}

// Execute swaps a todo from one parent to another
func Execute(sourcePath string, destParentPath string, opts Options) (*Result, error) {
	logger := logging.GetLogger("too.commands.swap")
	logger.Debug().
		Str("sourcePath", sourcePath).
		Str("destParentPath", destParentPath).
		Str("collectionPath", opts.CollectionPath).
		Msg("executing swap command")

	s := store.NewStore(opts.CollectionPath)
	var result *Result

	err := s.Update(func(collection *models.Collection) error {
		// Set up the IDM registry
		adapter, err := store.NewIDMStoreAdapter(s)
		if err != nil {
			return fmt.Errorf("failed to create idm adapter: %w", err)
		}
		reg := idm.NewRegistry()
		scopes, err := adapter.GetScopes()
		if err != nil {
			return fmt.Errorf("failed to get scopes: %w", err)
		}
		for _, scope := range scopes {
			if err := reg.RebuildScope(adapter, scope); err != nil {
				return fmt.Errorf("failed to build idm scope '%s': %w", scope, err)
			}
		}

		// Find the source todo
		sourceUID, err := reg.ResolvePositionPath(store.RootScope, sourcePath)
		if err != nil {
			return fmt.Errorf("todo not found at position: %s", sourcePath)
		}
		sourceTodo := collection.FindItemByID(sourceUID)
		if sourceTodo == nil {
			return fmt.Errorf("todo with ID '%s' not found", sourceUID)
		}

		// Find the destination parent (empty string means root)
		var destParent *models.Todo
		if destParentPath != "" {
			destParentUID, err := reg.ResolvePositionPath(store.RootScope, destParentPath)
			if err != nil {
				return fmt.Errorf("destination parent not found at position: %s", destParentPath)
			}
			destParent = collection.FindItemByID(destParentUID)
			if destParent == nil {
				return fmt.Errorf("destination parent with ID '%s' not found", destParentUID)
			}
		}

		// Check for circular reference (can't move a parent into its own child)
		if destParent != nil && isDescendantOf(destParent, sourceTodo) {
			logger.Error().
				Str("sourcePath", sourcePath).
				Str("destParentPath", destParentPath).
				Msg("attempted to move parent into its own descendant")
			return fmt.Errorf("cannot move a parent into its own descendant")
		}

		// Find the old parent
		var oldParent *models.Todo
		oldParentID := sourceTodo.ParentID
		if oldParentID != "" {
			oldParent = collection.FindItemByID(oldParentID)
		}

		// Store old path for result
		oldPath := sourcePath

		// Remove from old location
		if oldParent != nil {
			// Remove from parent's Items slice
			for i, item := range oldParent.Items {
				if item.ID == sourceTodo.ID {
					oldParent.Items = append(oldParent.Items[:i], oldParent.Items[i+1:]...)
					break
				}
			}
		} else {
			// Remove from root todos
			for i, item := range collection.Todos {
				if item.ID == sourceTodo.ID {
					collection.Todos = append(collection.Todos[:i], collection.Todos[i+1:]...)
					break
				}
			}
		}

		// Update parent ID
		if destParent != nil {
			sourceTodo.ParentID = destParent.ID
		} else {
			sourceTodo.ParentID = ""
		}

		// Add to new location
		if destParent != nil {
			// Set a high position to ensure it's placed at the end before reordering
			sourceTodo.Position = len(destParent.Items) + 1
			destParent.Items = append(destParent.Items, sourceTodo)
		} else {
			sourceTodo.Position = len(collection.Todos) + 1
			collection.Todos = append(collection.Todos, sourceTodo)
		}

		// Reset positions at both source and destination
		if oldParentID != "" {
			collection.ResetSiblingPositions(oldParentID)
		} else {
			collection.ResetRootPositions()
		}

		if destParent != nil {
			collection.ResetSiblingPositions(destParent.ID)
		} else if oldParentID != "" {
			collection.ResetRootPositions()
		}

		// Get new path after reordering
		newPath := "" // TODO: This needs to be recalculated
		if newPath == "" {
			logger.Error().
				Str("todoID", sourceTodo.ID).
				Msg("failed to get new position path")
			return fmt.Errorf("failed to determine new position path")
		}

		result = &Result{
			Todo:      sourceTodo,
			OldPath:   oldPath,
			NewPath:   newPath,
			OldParent: oldParent,
			NewParent: destParent,
		}

		logger.Debug().
			Str("todoID", sourceTodo.ID).
			Str("oldPath", oldPath).
			Str("newPath", newPath).
			Msg("successfully swapped todo")

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
