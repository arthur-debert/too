package complete

import (
	"fmt"

	"github.com/arthur-debert/tdh/pkg/logging"
	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/arthur-debert/tdh/pkg/tdh/store"
	"github.com/rs/zerolog"
)

// Options contains options for the complete command
type Options struct {
	CollectionPath string
}

// Result contains the result of the complete command
type Result struct {
	Todo      *models.Todo
	OldStatus string
	NewStatus string
}

// Execute marks a todo as complete
func Execute(positionPath string, opts Options) (*Result, error) {
	logger := logging.GetLogger("tdh.commands.complete")
	logger.Debug().
		Str("positionPath", positionPath).
		Str("collectionPath", opts.CollectionPath).
		Msg("executing complete command")

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

		// Mark the todo as complete using the new method
		// Skip reorder for now since we may need to handle bottom-up completion
		todo.MarkComplete(collection, true)

		logger.Debug().
			Str("todoID", todo.ID).
			Str("oldStatus", oldStatus).
			Str("newStatus", string(todo.Status)).
			Msg("marked todo as complete")

		// Bottom-Up Completion: Check if all siblings are complete and propagate up
		if todo.ParentID != "" {
			logger.Debug().
				Str("parentID", todo.ParentID).
				Msg("checking bottom-up completion for parent")

			checkAndCompleteParent(collection, todo.ParentID, logger)
		}

		// Now trigger position reset at the appropriate level
		if todo.ParentID != "" {
			collection.ResetSiblingPositions(todo.ParentID)
		} else {
			collection.ResetRootPositions()
		}

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
		Msg("successfully completed todo")

	return result, nil
}

// checkAndCompleteParent recursively checks if all children of a parent are complete,
// and if so, marks the parent as complete and continues up the hierarchy
func checkAndCompleteParent(collection *models.Collection, parentID string, logger zerolog.Logger) {
	// Find the parent todo
	parent := collection.FindItemByID(parentID)
	if parent == nil {
		logger.Error().
			Str("parentID", parentID).
			Msg("parent not found during bottom-up completion")
		return
	}

	// Check if all children are complete
	allChildrenComplete := true
	for _, child := range parent.Items {
		if child.Status != models.StatusDone {
			allChildrenComplete = false
			break
		}
	}

	// If all children are complete, mark parent as complete
	// Only mark parent as complete if it actually has children to check
	// This prevents childless parents from being auto-completed
	if allChildrenComplete && len(parent.Items) > 0 {
		logger.Debug().
			Str("parentID", parentID).
			Int("childCount", len(parent.Items)).
			Msg("all children complete, marking parent as complete")

		// Use the new method which handles status, position, and timestamp
		parent.MarkComplete(collection, true) // Skip reorder during recursion

		// Continue up the hierarchy
		if parent.ParentID != "" {
			logger.Debug().
				Str("grandparentID", parent.ParentID).
				Msg("checking grandparent for bottom-up completion")

			checkAndCompleteParent(collection, parent.ParentID, logger)
		}
	} else {
		logger.Debug().
			Str("parentID", parentID).
			Bool("allChildrenComplete", allChildrenComplete).
			Int("childCount", len(parent.Items)).
			Msg("parent not marked complete")
	}
}
