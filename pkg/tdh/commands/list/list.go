package list

import (
	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/arthur-debert/tdh/pkg/tdh/store"
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

	// Load the full collection to apply behavioral propagation
	collection, err := s.Load()
	if err != nil {
		return nil, err
	}

	// Apply behavioral propagation: filter out done branches
	filteredTodos := filterWithBehavioralPropagation(collection.Todos, opts)

	// Count totals from the original collection
	totalCount, doneCount := countTodos(collection.Todos)

	return &Result{
		Todos:      filteredTodos,
		TotalCount: totalCount,
		DoneCount:  doneCount,
	}, nil
}

// filterWithBehavioralPropagation filters todos respecting behavioral propagation rules
// When a parent is done, all its descendants are hidden regardless of their status
func filterWithBehavioralPropagation(todos []*models.Todo, opts Options) []*models.Todo {
	var filtered []*models.Todo

	for _, todo := range todos {
		// Skip done items when not showing all (unless explicitly showing done)
		if !opts.ShowAll && !opts.ShowDone && todo.Status == models.StatusDone {
			continue
		}

		// Skip pending items when only showing done
		if !opts.ShowAll && opts.ShowDone && todo.Status != models.StatusDone {
			continue
		}

		// Clone the todo to avoid modifying the original
		filteredTodo := &models.Todo{
			ID:       todo.ID,
			ParentID: todo.ParentID,
			Position: todo.Position,
			Text:     todo.Text,
			Status:   todo.Status,
			Modified: todo.Modified,
			Items:    []*models.Todo{},
		}

		// If this todo is done, behavioral propagation stops here - don't process children
		// Exception: when ShowAll is true, we show everything
		if todo.Status == models.StatusDone && !opts.ShowAll {
			// Add the done item but with no children
			filtered = append(filtered, filteredTodo)
		} else {
			// Recursively filter children
			filteredTodo.Items = filterWithBehavioralPropagation(todo.Items, opts)
			filtered = append(filtered, filteredTodo)
		}
	}

	return filtered
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
