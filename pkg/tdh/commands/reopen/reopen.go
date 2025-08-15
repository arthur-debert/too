package reopen

import (
	"fmt"

	"github.com/arthur-debert/tdh/pkg/logging"
	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/arthur-debert/tdh/pkg/tdh/store"
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

// Execute marks a todo as pending
func Execute(positionPath string, opts Options) (*Result, error) {
	logger := logging.GetLogger("tdh.commands.reopen")
	logger.Debug().
		Str("positionPath", positionPath).
		Str("collectionPath", opts.CollectionPath).
		Msg("executing reopen command")

	var result *Result

	s := store.NewStore(opts.CollectionPath)
	err := s.Update(func(collection *models.Collection) error {
		// Find the todo by position path
		todo, err := collection.FindItemByPositionPath(positionPath)
		if err != nil {
			logger.Error().
				Err(err).
				Str("positionPath", positionPath).
				Msg("failed to find todo")
			return fmt.Errorf("todo not found: %w", err)
		}

		// Capture old status
		oldStatus := string(todo.Status)

		// According to the spec, reopen only affects the specified item
		// No propagation in any direction
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
		Str("positionPath", positionPath).
		Str("todoText", result.Todo.Text).
		Msg("successfully reopened todo")

	return result, nil
}
