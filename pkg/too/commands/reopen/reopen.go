package reopen

import (
	"fmt"

	"github.com/arthur-debert/too/pkg/logging"
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
func Execute(ref string, opts Options) (*Result, error) {
	logger := logging.GetLogger("too.commands.reopen")
	logger.Debug().
		Str("ref", ref).
		Str("collectionPath", opts.CollectionPath).
		Msg("executing reopen command")

	var result *Result

	s := store.NewStore(opts.CollectionPath)
	err := s.Update(func(collection *models.Collection) error {
		// Find the todo by ref (position or short ID)
		todo, err := collection.FindItemByRef(ref)
		if err != nil {
			logger.Error().
				Err(err).
				Str("ref", ref).
				Msg("failed to find todo")
			return fmt.Errorf("todo not found with reference: %s", ref)
		}

		// Capture old status
		oldStatus := string(todo.Status)

		// Use the new method which handles status change and position reset
		todo.MarkPending(collection)

		logger.Debug().
			Str("todoID", todo.ID).
			Str("oldStatus", oldStatus).
			Str("newStatus", string(todo.Status)).
			Msg("marked todo as pending")

		// Capture result
		result = &Result{
			Todo:      todo,
			OldStatus: oldStatus,
			NewStatus: string(todo.Status),
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
