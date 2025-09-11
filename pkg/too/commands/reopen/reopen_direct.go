package reopen

import (
	"fmt"

	"github.com/arthur-debert/too/pkg/logging"
	"github.com/arthur-debert/too/pkg/too/parser"
	"github.com/arthur-debert/too/pkg/too/store"
)

// ExecuteDirect marks a todo as pending without adapters.
func ExecuteDirect(ref string, opts Options) (*Result, error) {
	logger := logging.GetLogger("too.commands.reopen")
	logger.Debug().
		Str("ref", ref).
		Str("collectionPath", opts.CollectionPath).
		Msg("executing reopen command with direct manager")

	s := store.NewStore(opts.CollectionPath)
	
	// Create direct workflow manager
	manager, err := store.NewDirectWorkflowManager(s, opts.CollectionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create direct workflow manager: %w", err)
	}

	collection := manager.GetCollection()
	var uid string

	// Try to resolve as position path or find by short ID
	if parser.IsPositionPath(ref) {
		// Position paths only work for active items
		uid, err = manager.ResolvePositionPath(store.RootScope, ref)
		if err != nil {
			// Try as short ID instead
			todo, shortErr := collection.FindItemByShortID(ref)
			if shortErr != nil || todo == nil {
				return nil, fmt.Errorf("todo not found with reference: %s", ref)
			}
			uid = todo.ID
		}
	} else {
		// Find by short ID
		todo, err := collection.FindItemByShortID(ref)
		if err != nil || todo == nil {
			return nil, fmt.Errorf("todo not found with reference: %s", ref)
		}
		uid = todo.ID
	}

	// Get the todo for validation
	todo := collection.FindItemByID(uid)
	if todo == nil {
		return nil, fmt.Errorf("todo with ID '%s' not found", uid)
	}

	// Capture old status for result
	oldStatus, err := manager.GetStatus(uid, "completion")
	if err != nil {
		oldStatus = "done" // Assume it was done if we can't get status
	}

	// Set status to "pending"
	err = manager.SetStatus(uid, "completion", "pending")
	if err != nil {
		return nil, fmt.Errorf("failed to set pending status: %w", err)
	}

	logger.Debug().
		Str("todoID", uid).
		Str("oldStatus", oldStatus).
		Str("newStatus", "pending").
		Msg("marked todo as pending using direct workflow manager")

	// Reset positions to maintain sort order
	if todo.ParentID != "" {
		collection.ResetSiblingPositions(todo.ParentID)
	} else {
		collection.ResetRootPositions()
	}

	// Save changes
	if err := manager.Save(); err != nil {
		return nil, err
	}

	// Capture result
	result := &Result{
		Todo:      todo,
		OldStatus: oldStatus,
		NewStatus: "pending",
	}

	logger.Info().
		Str("ref", ref).
		Str("todoText", result.Todo.Text).
		Msg("successfully reopened todo")

	return result, nil
}