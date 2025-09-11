package move

import (
	"fmt"

	"github.com/arthur-debert/too/pkg/logging"
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/store"
)

// ExecuteDirect moves a todo from one parent to another without adapters.
func ExecuteDirect(sourcePath string, destParentPath string, opts Options) (*Result, error) {
	logger := logging.GetLogger("too.commands.move")
	logger.Debug().
		Str("sourcePath", sourcePath).
		Str("destParentPath", destParentPath).
		Str("collectionPath", opts.CollectionPath).
		Msg("executing move command with direct manager")

	s := store.NewStore(opts.CollectionPath)
	
	// Create direct workflow manager
	manager, err := store.NewDirectWorkflowManager(s, opts.CollectionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create direct workflow manager: %w", err)
	}

	// Resolve source todo
	sourceUID, err := manager.ResolvePositionPath(store.RootScope, sourcePath)
	if err != nil {
		return nil, fmt.Errorf("todo not found at position: %s", sourcePath)
	}

	collection := manager.GetCollection()
	sourceTodo := collection.FindItemByID(sourceUID)
	if sourceTodo == nil {
		return nil, fmt.Errorf("todo with ID '%s' not found", sourceUID)
	}

	// Resolve destination parent
	var destParentUID string = store.RootScope
	if destParentPath != "" {
		uid, err := manager.ResolvePositionPath(store.RootScope, destParentPath)
		if err != nil {
			return nil, fmt.Errorf("destination parent not found at position: %s", destParentPath)
		}
		destParentUID = uid
	}

	// Check for circular reference
	if destParentUID != store.RootScope {
		destParent := collection.FindItemByID(destParentUID)
		if destParent != nil && isDescendantOf(destParent, sourceTodo) {
			logger.Error().
				Str("sourcePath", sourcePath).
				Str("destParentPath", destParentPath).
				Msg("attempted to move parent into its own descendant")
			return nil, fmt.Errorf("cannot move a parent into its own descendant")
		}
	}

	// Store old path for result
	oldPath := sourcePath

	// Get old parent UID
	oldParentUID := store.RootScope
	if sourceTodo.ParentID != "" {
		oldParentUID = sourceTodo.ParentID
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

	// Perform the move
	if err := manager.Move(sourceUID, oldParentUID, destParentUID); err != nil {
		return nil, fmt.Errorf("failed to move todo: %w", err)
	}

	// Get new path
	newPath, err := manager.GetPositionPath(store.RootScope, sourceUID)
	if err != nil {
		logger.Error().
			Str("todoID", sourceTodo.ID).
			Str("todoText", sourceTodo.Text).
			Str("parentID", sourceTodo.ParentID).
			Int("position", sourceTodo.Position).
			Err(err).
			Msg("failed to get new position path")
		return nil, fmt.Errorf("failed to determine new position path: %w", err)
	}

	// Save the collection
	if err := manager.Save(); err != nil {
		return nil, err
	}

	result := &Result{
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

	return result, nil
}