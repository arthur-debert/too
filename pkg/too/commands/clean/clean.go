package clean

import (
	"fmt"

	"github.com/arthur-debert/too/pkg/logging"
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/store"
)

// Options contains options for the clean command
type Options struct {
	CollectionPath string
}

// Result contains the result of the clean command
type Result struct {
	RemovedCount int
	RemovedTodos []*models.IDMTodo
	ActiveCount  int
	ActiveTodos  []*models.IDMTodo // Remaining active todos for display
	TotalCount   int                // Total todos before clean
	DoneCount    int                // Done todos before clean
}

// Execute removes finished todos from the collection using the pure IDM manager.
func Execute(opts Options) (*Result, error) {
	logger := logging.GetLogger("too.commands.clean")
	logger.Debug().
		Str("collectionPath", opts.CollectionPath).
		Msg("executing clean command with pure IDM manager")

	idmStore := store.NewIDMStore(opts.CollectionPath)

	// Create pure IDM workflow manager
	manager, err := store.NewPureIDMManager(idmStore, opts.CollectionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create pure IDM manager: %w", err)
	}

	// Get counts before clean
	totalCount, doneCount := manager.CountTodos()

	// Use the manager's integrated clean operation
	removedTodos, activeCount, err := manager.CleanFinishedTodos()
	if err != nil {
		return nil, fmt.Errorf("failed to clean finished todos: %w", err)
	}

	// Save the updated collection
	if err := manager.Save(); err != nil {
		return nil, fmt.Errorf("failed to save collection after clean: %w", err)
	}

	// Get remaining active todos for display
	activeTodos := manager.ListActive()
	// Attach active-only position paths for consecutive numbering
	manager.AttachActiveOnlyPositionPaths(activeTodos)

	result := &Result{
		RemovedCount: len(removedTodos),
		RemovedTodos: removedTodos,
		ActiveCount:  activeCount,
		ActiveTodos:  activeTodos,
		TotalCount:   totalCount,
		DoneCount:    doneCount,
	}

	logger.Info().
		Int("removedCount", len(removedTodos)).
		Int("activeCount", activeCount).
		Msg("clean command completed with pure IDM manager")

	return result, nil
}