package complete

import (
	"fmt"

	"github.com/arthur-debert/too/pkg/logging"
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/store"
)

// ExecuteDirect marks a todo as complete using the direct workflow manager without adapters.
func ExecuteDirect(positionPath string, opts Options) (*Result, error) {
	logger := logging.GetLogger("too.commands.complete")
	logger.Debug().
		Str("positionPath", positionPath).
		Str("collectionPath", opts.CollectionPath).
		Msg("executing complete command with direct manager")

	s := store.NewStore(opts.CollectionPath)
	
	// Create direct workflow manager
	manager, err := store.NewDirectWorkflowManager(s, opts.CollectionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create direct workflow manager: %w", err)
	}

	// Resolve position path to UID
	uid, err := manager.ResolvePositionPath(store.RootScope, positionPath)
	if err != nil {
		return nil, fmt.Errorf("todo not found: %w", err)
	}

	// Get the todo using DirectWorkflowManager method
	todo := manager.GetTodoByID(uid)
	if todo == nil {
		return nil, fmt.Errorf("todo with ID '%s' not found", uid)
	}

	// Capture old status
	oldStatus, err := manager.GetStatus(uid, "completion")
	if err != nil {
		oldStatus = string(models.StatusPending)
	}

	// Set status to done
	if err := manager.SetStatus(uid, "completion", "done"); err != nil {
		return nil, fmt.Errorf("failed to set completion status: %w", err)
	}

	// Position management is handled by IDM, no need to reset

	// Save changes
	if err := manager.Save(); err != nil {
		return nil, fmt.Errorf("failed to save collection: %w", err)
	}

	logger.Debug().
		Str("todoID", uid).
		Str("oldStatus", oldStatus).
		Str("newStatus", "done").
		Msg("marked todo as complete using direct workflow manager")

	// Build result
	result := &Result{
		Todo:      todo,
		OldStatus: oldStatus,
		NewStatus: "done",
		Mode:      opts.Mode,
	}

	// Add long mode data if requested using manager's IDM-aware methods
	if opts.Mode == "long" {
		result.AllTodos = manager.ListActive()
		if result.AllTodos == nil {
			result.AllTodos = []*models.Todo{}
		}
		result.TotalCount, result.DoneCount = manager.CountTodos()
	}

	return result, nil
}

