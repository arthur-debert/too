package list

import (
	"fmt"

	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/store"
)

// UnifiedResult contains the result of the list command that works with both manager types.
type UnifiedResult struct {
	Todos      interface{} // []*models.Todo or []*models.IDMTodo
	TotalCount int
	DoneCount  int
	IsPureIDM  bool
}

// ExecuteUnified lists todos using the factory pattern.
// It automatically selects the appropriate manager based on storage format.
func ExecuteUnified(opts Options) (*UnifiedResult, error) {
	// Create workflow manager using factory
	manager, err := store.CreateWorkflowManager(opts.CollectionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create workflow manager: %w", err)
	}

	// Get todos based on options
	var todos interface{}
	if opts.ShowAll {
		todos = manager.ListAll()
	} else if opts.ShowDone {
		todos = manager.ListArchived()
	} else {
		todos = manager.ListActive()
	}

	// Get counts
	totalCount, doneCount := manager.CountTodos()

	return &UnifiedResult{
		Todos:      todos,
		TotalCount: totalCount,
		DoneCount:  doneCount,
		IsPureIDM:  manager.IsPureIDM(),
	}, nil
}

// ConvertUnifiedToResult converts a UnifiedResult to the traditional Result format.
func ConvertUnifiedToResult(unified *UnifiedResult) *Result {
	result := &Result{
		TotalCount: unified.TotalCount,
		DoneCount:  unified.DoneCount,
	}

	// Convert todos based on type
	if unified.IsPureIDM {
		// Convert from IDMTodo slice
		idmTodos := unified.Todos.([]*models.IDMTodo)
		result.Todos = make([]*models.Todo, len(idmTodos))
		for i, idmTodo := range idmTodos {
			result.Todos[i] = convertIDMTodoToTodo(idmTodo)
		}
	} else {
		// Already in Todo format
		result.Todos = unified.Todos.([]*models.Todo)
	}

	return result
}

