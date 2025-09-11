package list

import (
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

// Execute returns todos from the collection with optional filtering
func Execute(opts Options) (*Result, error) {
	s := store.NewStore(opts.CollectionPath)

	// Create direct workflow manager instead of loading collection directly
	manager, err := store.NewDirectWorkflowManager(s, opts.CollectionPath)
	if err != nil {
		return nil, err
	}

	// Use DirectWorkflowManager's IDM-aware filtering methods
	var filteredTodos []*models.Todo
	if opts.ShowAll {
		filteredTodos = manager.ListAll()
	} else if opts.ShowDone {
		filteredTodos = manager.ListArchived()
	} else {
		filteredTodos = manager.ListActive()
	}

	// Count totals from the original collection
	collection := manager.GetCollection()
	totalCount, doneCount := countTodos(collection.Todos)

	return &Result{
		Todos:      filteredTodos,
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
