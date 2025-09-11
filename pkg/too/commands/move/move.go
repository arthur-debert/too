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

		// Get new path after reordering
		// Calculate the new position path directly from the collection
		newPath := calculatePositionPath(collection, sourceTodo)
		if newPath == "" {
			logger.Error().
				Str("todoID", sourceTodo.ID).
				Str("todoText", sourceTodo.Text).
				Str("parentID", sourceTodo.ParentID).
				Int("position", sourceTodo.Position).
				Msg("failed to get new position path")
			return fmt.Errorf("failed to determine new position path")
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

// calculatePositionPath calculates the position path for a todo directly from the collection
func calculatePositionPath(collection *models.Collection, todo *models.Todo) string {
	if todo == nil {
		return ""
	}

	// Build the path from the todo up to the root
	path := []string{}
	current := todo
	
	for current != nil {
		// Find position among pending siblings
		var siblings []*models.Todo
		if current.ParentID == "" {
			// Root level
			siblings = collection.Todos
		} else {
			parent := collection.FindItemByID(current.ParentID)
			if parent == nil {
				return ""
			}
			siblings = parent.Items
		}
		
		// Count position among pending siblings
		position := 0
		for _, sibling := range siblings {
			if sibling.Status == models.StatusPending {
				position++
				if sibling.ID == current.ID {
					break
				}
			}
		}
		
		if position == 0 {
			return "" // Todo not found or not pending
		}
		
		// Prepend position to path
		path = append([]string{fmt.Sprintf("%d", position)}, path...)
		
		// Move up to parent
		if current.ParentID == "" {
			break // Reached root
		}
		current = collection.FindItemByID(current.ParentID)
	}
	
	// Join path components
	if len(path) == 0 {
		return ""
	}
	result := path[0]
	for i := 1; i < len(path); i++ {
		result += "." + path[i]
	}
	return result
}
