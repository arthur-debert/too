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

// Execute marks a todo as complete using the WorkflowManager
func Execute(positionPath string, opts Options) (*Result, error) {
	logger := logging.GetLogger("too.commands.complete")
	logger.Debug().
		Str("positionPath", positionPath).
		Str("collectionPath", opts.CollectionPath).
		Msg("executing complete command")

	var result *Result

	s := store.NewStore(opts.CollectionPath)
	err := s.Update(func(collection *models.Collection) error {
		// Create workflow manager for this collection
		wm, err := store.NewWorkflowManager(collection, opts.CollectionPath)
		if err != nil {
			return fmt.Errorf("failed to create workflow manager: %w", err)
		}

		// Resolve position path to UID using workflow manager
		uid, err := wm.ResolvePositionPathInContext(store.RootScope, positionPath, "active")
		if err != nil {
			return fmt.Errorf("todo not found: %w", err)
		}

		// Get the todo for logging and validation
		todo := collection.FindItemByID(uid)
		if todo == nil {
			return fmt.Errorf("todo with ID '%s' not found", uid)
		}

		// Capture old status for result
		oldStatus, err := wm.GetStatus(uid, "completion")
		if err != nil {
			return fmt.Errorf("failed to get current status: %w", err)
		}

		// Set status to "done" - this will trigger auto-transitions including bottom-up completion
		err = wm.SetStatus(uid, "completion", "done")
		if err != nil {
			return fmt.Errorf("failed to set completion status: %w", err)
		}

		logger.Debug().
			Str("todoID", uid).
			Str("oldStatus", oldStatus).
			Str("newStatus", "done").
			Msg("marked todo as complete using workflow manager")

		// Preserve legacy bottom-up completion behavior while auto-transitions are disabled
		if todo.ParentID != "" {
			checkAndCompleteParent(collection, todo.ParentID, logger)
		}

		// Preserve legacy position reset behavior for compatibility
		if todo.ParentID != "" {
			collection.ResetSiblingPositions(todo.ParentID)
		} else {
			collection.ResetRootPositions()
		}

		// Build result using workflow manager
		wfResult, err := wm.BuildResult(uid, opts.Mode, oldStatus)
		if err != nil {
			return fmt.Errorf("failed to build result: %w", err)
		}

		// Convert to command-specific result structure
		result = &Result{
			Todo:       wfResult.Todo,
			OldStatus:  wfResult.OldStatus,
			NewStatus:  wfResult.NewStatus,
			Mode:       opts.Mode,
			AllTodos:   wfResult.AllTodos,
			TotalCount: wfResult.TotalCount,
			DoneCount:  wfResult.DoneCount,
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
// and if so, marks the parent as complete and continues up the hierarchy.
// This is a temporary implementation while auto-transitions are being fixed.
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

