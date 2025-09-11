package list

import (
	"fmt"

	"github.com/arthur-debert/too/pkg/logging"
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/store"
)

// IDMResult contains the result of the list command using pure IDM data structures.
type IDMResult struct {
	Todos      []*models.IDMTodo
	TotalCount int
	DoneCount  int
}

// ExecuteIDM lists todos using the pure IDM manager.
func ExecuteIDM(opts Options) (*IDMResult, error) {
	logger := logging.GetLogger("too.commands.list")
	logger.Debug().
		Str("collectionPath", opts.CollectionPath).
		Bool("showAll", opts.ShowAll).
		Bool("showDone", opts.ShowDone).
		Msg("executing list command with pure IDM manager")

	idmStore := store.NewIDMStore(opts.CollectionPath)

	// Create pure IDM workflow manager
	manager, err := store.NewPureIDMManager(idmStore, opts.CollectionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create pure IDM manager: %w", err)
	}

	// Use PureIDMManager's IDM-aware filtering methods
	var filteredTodos []*models.IDMTodo
	if opts.ShowAll {
		filteredTodos = manager.ListAll()
	} else if opts.ShowDone {
		filteredTodos = manager.ListArchived()
	} else {
		filteredTodos = manager.ListActive()
	}

	// Count all todos for statistics
	totalCount, doneCount := manager.CountTodos()

	result := &IDMResult{
		Todos:      filteredTodos,
		TotalCount: totalCount,
		DoneCount:  doneCount,
	}

	logger.Debug().
		Int("filteredCount", len(filteredTodos)).
		Int("totalCount", totalCount).
		Int("doneCount", doneCount).
		Msg("list command completed with pure IDM manager")

	return result, nil
}

// ConvertIDMResultToResult converts an IDMResult to the traditional Result format.
// This enables backward compatibility while migrating to pure IDM structures.
func ConvertIDMResultToResult(idmResult *IDMResult) *Result {
	result := &Result{
		TotalCount: idmResult.TotalCount,
		DoneCount:  idmResult.DoneCount,
	}

	// Convert all todos
	if idmResult.Todos != nil {
		result.Todos = make([]*models.Todo, len(idmResult.Todos))
		for i, idmTodo := range idmResult.Todos {
			result.Todos[i] = convertIDMTodoToTodo(idmTodo)
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