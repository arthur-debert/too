package clean

import (
	"fmt"

	"github.com/arthur-debert/too/pkg/logging"
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/store"
)

// IDMResult contains the result of the clean command using pure IDM data structures.
type IDMResult struct {
	RemovedTodos []*models.IDMTodo
	ActiveCount  int
}

// ExecuteIDM cleans finished todos using the pure IDM manager.
func ExecuteIDM(opts Options) (*IDMResult, error) {
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

	// Use the manager's integrated clean operation
	removedTodos, activeCount, err := manager.CleanFinishedTodos()
	if err != nil {
		return nil, fmt.Errorf("failed to clean finished todos: %w", err)
	}

	// Save the updated collection
	if err := manager.Save(); err != nil {
		return nil, fmt.Errorf("failed to save collection after clean: %w", err)
	}

	result := &IDMResult{
		RemovedTodos: removedTodos,
		ActiveCount:  activeCount,
	}

	logger.Info().
		Int("removedCount", len(removedTodos)).
		Int("activeCount", activeCount).
		Msg("clean command completed with pure IDM manager")

	return result, nil
}

// ConvertIDMResultToResult converts an IDMResult to the traditional Result format.
// This enables backward compatibility while migrating to pure IDM structures.
func ConvertIDMResultToResult(idmResult *IDMResult) *Result {
	result := &Result{
		RemovedCount: len(idmResult.RemovedTodos),
		ActiveCount:  idmResult.ActiveCount,
	}

	// Convert removed todos
	if idmResult.RemovedTodos != nil {
		result.RemovedTodos = make([]*models.Todo, len(idmResult.RemovedTodos))
		for i, idmTodo := range idmResult.RemovedTodos {
			result.RemovedTodos[i] = convertIDMTodoToTodo(idmTodo)
		}
	}

	return result
}

// convertIDMTodoToTodo converts a single IDMTodo to Todo for backward compatibility.
func convertIDMTodoToTodo(idmTodo *models.IDMTodo) *models.Todo {
	todo := &models.Todo{
		ID:       idmTodo.UID,
		ParentID: idmTodo.ParentID,
		Text:     idmTodo.Text,
		Statuses: make(map[string]string),
		Modified: idmTodo.Modified,
		Items:    []*models.Todo{}, // Empty - hierarchy managed by IDM
	}

	// Copy statuses
	for k, v := range idmTodo.Statuses {
		todo.Statuses[k] = v
	}

	return todo
}