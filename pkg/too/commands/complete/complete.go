package complete

import (
	"fmt"

	"github.com/arthur-debert/too/pkg/logging"
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/store"
	"github.com/rs/zerolog"
)

// Options contains options for the complete command
type Options struct {
	CollectionPath string
	Mode           string // Output mode: "short" or "long"
}

// Result contains the result of the complete command
type Result struct {
	Todo       *models.Todo
	OldStatus  string
	NewStatus  string
	Mode       string         // Output mode passed from options
	AllTodos   []*models.Todo // All todos for long mode
	TotalCount int            // Total count for long mode
	DoneCount  int            // Done count for long mode
}

// Execute marks a todo as complete
func Execute(positionPath string, opts Options) (*Result, error) {
	logger := logging.GetLogger("too.commands.complete")
	logger.Debug().
		Str("positionPath", positionPath).
		Str("collectionPath", opts.CollectionPath).
		Msg("executing complete command")

	var result *Result

	s := store.NewStore(opts.CollectionPath)
	err := s.Update(func(collection *models.Collection) error {
		// Resolve the position path to a UID
		manager, err := store.NewManagerFromStore(s)
		if err != nil {
			return fmt.Errorf("failed to create idm manager: %w", err)
		}

		uid, err := manager.Registry().ResolvePositionPath(store.RootScope, positionPath)
		if err != nil {
			return fmt.Errorf("todo not found: %w", err)
		}

		todo := collection.FindItemByID(uid)
		if todo == nil {
			return fmt.Errorf("todo with ID '%s' not found", uid)
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
			Mode:      opts.Mode,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// If in long mode, get all active todos
	if opts.Mode == "long" {
		collection, err := s.Load()
		if err != nil {
			logger.Error().Err(err).Msg("failed to load collection for long mode")
			return nil, fmt.Errorf("failed to load collection for long mode: %w", err)
		}

		result.AllTodos = collection.ListActive()
		result.TotalCount, result.DoneCount = countTodos(collection.Todos)
	}

	logger.Info().
		Str("positionPath", positionPath).
		Str("todoText", result.Todo.Text).
		Msg("successfully completed todo")

	return result, nil
}

// countTodos recursively counts total and done todos
func countTodos(todos []*models.Todo) (total int, done int) {
	for _, todo := range todos {
		total++
		if todo.Status == models.StatusDone {
			done++
		}
		// Recursively count children
		childTotal, childDone := countTodos(todo.Items)
		total += childTotal
		done += childDone
	}
	return total, done
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
