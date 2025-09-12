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
	Todo       *models.IDMTodo
	OldPath    string
	NewPath    string
	OldParent  *models.IDMTodo
	NewParent  *models.IDMTodo
	AllTodos   []*models.IDMTodo // All todos for display
	TotalCount int               // Total count for display
	DoneCount  int               // Done count for display
}

// Execute moves a todo from one parent to another using the pure IDM manager.
func Execute(sourcePath string, destParentPath string, opts Options) (*Result, error) {
	logger := logging.GetLogger("too.commands.move")
	logger.Debug().
		Str("sourcePath", sourcePath).
		Str("destParentPath", destParentPath).
		Str("collectionPath", opts.CollectionPath).
		Msg("executing move command with pure IDM manager")

	idmStore := store.NewIDMStore(opts.CollectionPath)
	
	// Create pure IDM workflow manager
	manager, err := store.NewPureIDMManager(idmStore, opts.CollectionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create pure IDM manager: %w", err)
	}

	// Resolve source todo
	sourceUID, err := manager.ResolvePositionPath(store.RootScope, sourcePath)
	if err != nil {
		return nil, fmt.Errorf("todo not found at position: %s", sourcePath)
	}

	sourceTodo := manager.GetTodoByUID(sourceUID)
	if sourceTodo == nil {
		return nil, fmt.Errorf("todo with UID '%s' not found", sourceUID)
	}

	// Resolve destination parent
	var destParentUID = store.RootScope
	if destParentPath != "" {
		uid, err := manager.ResolvePositionPath(store.RootScope, destParentPath)
		if err != nil {
			return nil, fmt.Errorf("destination parent not found at position: %s", destParentPath)
		}
		destParentUID = uid
	}

	// Check for circular reference
	if destParentUID != store.RootScope {
		destParent := manager.GetTodoByUID(destParentUID)
		if destParent != nil && isIDMDescendantOf(sourceTodo, destParent, manager) {
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
	var oldParent *models.IDMTodo
	if oldParentUID != store.RootScope {
		oldParent = manager.GetTodoByUID(oldParentUID)
	}
	
	var newParent *models.IDMTodo
	if destParentUID != store.RootScope {
		newParent = manager.GetTodoByUID(destParentUID)
	}

	// Perform the move
	if err := manager.Move(sourceUID, oldParentUID, destParentUID); err != nil {
		return nil, fmt.Errorf("failed to move todo: %w", err)
	}

	// Get new path
	newPath, err := manager.GetPositionPath(store.RootScope, sourceUID)
	if err != nil {
		logger.Error().
			Str("todoUID", sourceTodo.UID).
			Str("todoText", sourceTodo.Text).
			Str("parentID", sourceTodo.ParentID).
			Err(err).
			Msg("failed to get new position path")
		return nil, fmt.Errorf("failed to determine new position path: %w", err)
	}

	// Save the collection
	if err := manager.Save(); err != nil {
		return nil, err
	}

	// Get all todos for display
	allTodos := manager.ListActive()
	manager.AttachActiveOnlyPositionPaths(allTodos)
	totalCount, doneCount := manager.CountTodos()

	result := &Result{
		Todo:       sourceTodo,
		OldPath:    oldPath,
		NewPath:    newPath,
		OldParent:  oldParent,
		NewParent:  newParent,
		AllTodos:   allTodos,
		TotalCount: totalCount,
		DoneCount:  doneCount,
	}

	logger.Debug().
		Str("todoUID", sourceTodo.UID).
		Str("oldPath", oldPath).
		Str("newPath", newPath).
		Msg("successfully moved todo")

	return result, nil
}

// isIDMDescendantOf checks if child is a descendant of parent using IDM collection.
func isIDMDescendantOf(child, parent *models.IDMTodo, manager *store.PureIDMManager) bool {
	// Get all descendants of the parent
	descendants := manager.GetCollection().GetDescendants(parent.UID)
	for _, descendant := range descendants {
		if descendant.UID == child.UID {
			return true
		}
	}
	return false
}