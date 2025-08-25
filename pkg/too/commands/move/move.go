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
		// Find the source todo
		sourceTodo, err := collection.FindItemByPositionPath(sourcePath)
		if err != nil {
			logger.Error().
				Str("sourcePath", sourcePath).
				Err(err).
				Msg("failed to find source todo")
			return fmt.Errorf("todo not found at position: %s", sourcePath)
		}

		// Find the destination parent (empty string means root)
		var destParent *models.Todo
		if destParentPath != "" {
			destParent, err = collection.FindItemByPositionPath(destParentPath)
			if err != nil {
				logger.Error().
					Str("destParentPath", destParentPath).
					Err(err).
					Msg("failed to find destination parent")
				return fmt.Errorf("destination parent not found at position: %s", destParentPath)
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
		// Source location (where item was removed from)
		if oldParentID != "" {
			collection.ResetSiblingPositions(oldParentID)
		} else {
			collection.ResetRootPositions()
		}

		// Destination location (where item was added to)
		if destParent != nil {
			collection.ResetSiblingPositions(destParent.ID)
		} else if oldParentID != "" {
			// Only reset root if we actually moved to root from elsewhere
			collection.ResetRootPositions()
		}

		// Get new path after reordering
		newPath := getPositionPath(collection, sourceTodo)
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
			Msg("successfully moved todo")

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// isDescendantOf checks if child is a descendant of parent
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

// getPositionPath builds the dot-notation position path for a todo
func getPositionPath(collection *models.Collection, todo *models.Todo) string {
	path := buildPath(collection.Todos, todo, "")
	return path
}

// buildPath recursively builds the position path
func buildPath(todos []*models.Todo, target *models.Todo, currentPath string) string {
	for _, t := range todos {
		// Skip done items (position 0) when building paths
		if t.Position == 0 {
			continue
		}

		newPath := currentPath
		if newPath == "" {
			newPath = fmt.Sprintf("%d", t.Position)
		} else {
			newPath = fmt.Sprintf("%s.%d", currentPath, t.Position)
		}

		if t.ID == target.ID {
			return newPath
		}

		// Recursively check children
		if foundPath := buildPath(t.Items, target, newPath); foundPath != "" {
			return foundPath
		}
	}
	return ""
}
