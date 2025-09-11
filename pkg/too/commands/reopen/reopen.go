package reopen

import (
	"fmt"

	"github.com/arthur-debert/too/pkg/logging"
	"github.com/arthur-debert/too/pkg/too/parser"
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/store"
)

// Options contains options for the reopen command
type Options struct {
	CollectionPath string
}

// Result contains the result of the reopen command
type Result struct {
	Todo      *models.Todo
	OldStatus string
	NewStatus string
}

// Execute marks a todo as pending by finding it via a user-provided reference,
// which can be either a position path (e.g., "1.2") or a short ID.
// Uses WorkflowManager for status management.
func Execute(ref string, opts Options) (*Result, error) {
	logger := logging.GetLogger("too.commands.reopen")
	logger.Debug().
		Str("ref", ref).
		Str("collectionPath", opts.CollectionPath).
		Msg("executing reopen command")

	var result *Result

	s := store.NewStore(opts.CollectionPath)
	err := s.Update(func(collection *models.Collection) error {
		// Create workflow manager for this collection
		wm, err := store.NewWorkflowManager(collection, opts.CollectionPath)
		if err != nil {
			return fmt.Errorf("failed to create workflow manager: %w", err)
		}

		var uid string

		if parser.IsPositionPath(ref) {
			// Resolve position path to UID using workflow manager
			uid, err = wm.ResolvePositionPathInContext(store.RootScope, ref, "all")
			if err != nil {
				return fmt.Errorf("todo not found: %w", err)
			}
		} else {
			// Find by short ID
			todo, err := collection.FindItemByShortID(ref)
			if err != nil {
				logger.Error().
					Err(err).
					Str("ref", ref).
					Msg("failed to find todo")
				return fmt.Errorf("todo not found with reference: %s", ref)
			}
			if todo == nil {
				return fmt.Errorf("todo not found with reference: %s", ref)
			}
			uid = todo.ID
		}

		// Get the todo for validation and logging
		todo := collection.FindItemByID(uid)
		if todo == nil {
			return fmt.Errorf("todo with ID '%s' not found", uid)
		}

		// Capture old status for result
		oldStatus, err := wm.GetStatus(uid, "completion")
		if err != nil {
			return fmt.Errorf("failed to get current status: %w", err)
		}

		// Set status to "pending" using workflow manager
		err = wm.SetStatus(uid, "completion", "pending")
		if err != nil {
			return fmt.Errorf("failed to set pending status: %w", err)
		}

		logger.Debug().
			Str("todoID", uid).
			Str("oldStatus", oldStatus).
			Str("newStatus", "pending").
			Msg("marked todo as pending using workflow manager")

		// Preserve legacy position reset behavior for compatibility
		if todo.ParentID != "" {
			collection.ResetSiblingPositions(todo.ParentID)
		} else {
			collection.ResetRootPositions()
		}

		// Capture result
		result = &Result{
			Todo:      todo,
			OldStatus: oldStatus,
			NewStatus: "pending",
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	logger.Info().
		Str("ref", ref).
		Str("todoText", result.Todo.Text).
		Msg("successfully reopened todo")

	return result, nil
}
