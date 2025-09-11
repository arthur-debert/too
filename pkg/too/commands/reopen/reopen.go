package reopen

import (
	"fmt"

	"github.com/arthur-debert/too/pkg/logging"
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/parser"
	"github.com/arthur-debert/too/pkg/too/store"
)

// Options contains options for the reopen command
type Options struct {
	CollectionPath string
}

// Result contains the result of the reopen command
type Result struct {
	Todo      *models.IDMTodo
	OldStatus string
	NewStatus string
}

// Execute marks a todo as pending using the pure IDM manager.
func Execute(ref string, opts Options) (*Result, error) {
	logger := logging.GetLogger("too.commands.reopen")
	logger.Debug().
		Str("ref", ref).
		Str("collectionPath", opts.CollectionPath).
		Msg("executing reopen command with pure IDM manager")

	idmStore := store.NewIDMStore(opts.CollectionPath)
	
	// Create pure IDM workflow manager
	manager, err := store.NewPureIDMManager(idmStore, opts.CollectionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create pure IDM manager: %w", err)
	}

	var uid string

	// Try to resolve as position path or find by short ID
	if parser.IsPositionPath(ref) {
		// Position paths only work for active items
		uid, err = manager.ResolvePositionPath(store.RootScope, ref)
		if err != nil {
			// Try as short ID instead
			todo, shortErr := manager.GetTodoByShortID(ref)
			if shortErr != nil || todo == nil {
				return nil, fmt.Errorf("todo not found with reference: %s", ref)
			}
			uid = todo.UID
		}
	} else {
		// Find by short ID using PureIDMManager method
		todo, err := manager.GetTodoByShortID(ref)
		if err != nil || todo == nil {
			return nil, fmt.Errorf("todo not found with reference: %s", ref)
		}
		uid = todo.UID
	}

	// Get the todo for validation using PureIDMManager method
	todo := manager.GetTodoByUID(uid)
	if todo == nil {
		return nil, fmt.Errorf("todo with UID '%s' not found", uid)
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
		Str("todoUID", uid).
		Str("oldStatus", oldStatus).
		Str("newStatus", "pending").
		Msg("marked todo as pending using pure IDM manager")

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