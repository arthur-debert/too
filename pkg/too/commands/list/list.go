package list

import (
	"fmt"

	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/store"
)

// Options contains options for listing todos
type Options struct {
	CollectionPath string
	ShowDone       bool
	ShowAll        bool
}

// Result contains the result of listing todos
type Result struct {
	Todos      []*models.Todo
	TotalCount int
	DoneCount  int
}

// Execute returns todos from the collection using pure IDM.
func Execute(opts Options) (*Result, error) {
	// Create IDM store and manager
	idmStore := store.NewIDMStore(opts.CollectionPath)
	manager, err := store.NewPureIDMManager(idmStore, opts.CollectionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create manager: %w", err)
	}

	// Get todos based on options
	var idmTodos []*models.IDMTodo
	if opts.ShowAll {
		idmTodos = manager.ListAll()
	} else if opts.ShowDone {
		idmTodos = manager.ListArchived()
	} else {
		idmTodos = manager.ListActive()
	}

	// Convert IDMTodos to Todos for API compatibility
	todos := make([]*models.Todo, len(idmTodos))
	for i, idmTodo := range idmTodos {
		todos[i] = &models.Todo{
			ID:       idmTodo.UID,
			ParentID: idmTodo.ParentID,
			Text:     idmTodo.Text,
			Modified: idmTodo.Modified,
			Items:    []*models.Todo{},
		}
		if idmTodo.Statuses != nil {
			todos[i].Statuses = make(map[string]string)
			for k, v := range idmTodo.Statuses {
				todos[i].Statuses[k] = v
			}
		}
	}

	// Get counts
	totalCount, doneCount := manager.CountTodos()

	return &Result{
		Todos:      todos,
		TotalCount: totalCount,
		DoneCount:  doneCount,
	}, nil
}

// countTodos recursively counts total and done todos
func countTodos(todos []*models.Todo) (total int, done int) {
	for _, todo := range todos {
		total++
		if todo.GetStatus() == models.StatusDone {
			done++
		}
		// Recursively count children
		childTotal, childDone := countTodos(todo.Items)
		total += childTotal
		done += childDone
	}
	return total, done
}
