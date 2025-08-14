package toggle

import (
	"fmt"
	"time"

	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/arthur-debert/tdh/pkg/tdh/store"
)

// Options contains options for the toggle command
type Options struct {
	CollectionPath string
}

// Result contains the result of the toggle command
type Result struct {
	Todo      *models.Todo
	OldStatus string
	NewStatus string
}

// Execute toggles the status of a todo and all its children
func Execute(positionPath string, opts Options) (*Result, error) {
	var result *Result

	s := store.NewStore(opts.CollectionPath)
	err := s.Update(func(collection *models.Collection) error {
		// Find the todo by position path
		todo, err := collection.FindItemByPositionPath(positionPath)
		if err != nil {
			return fmt.Errorf("todo not found: %w", err)
		}

		// Capture old status
		oldStatus := string(todo.Status)

		// Toggle the todo and all its children recursively
		toggleRecursive(todo)

		newStatus := string(todo.Status)

		// Auto-reorder after toggle
		collection.Reorder()

		// Capture result
		result = &Result{
			Todo:      todo,
			OldStatus: oldStatus,
			NewStatus: newStatus,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// toggleRecursive toggles a todo and all its children to match the parent's new status
func toggleRecursive(todo *models.Todo) {
	// Toggle the parent
	todo.Toggle()

	// Apply the parent's new status to all children recursively
	timestamp := time.Now()
	for _, child := range todo.Items {
		child.Status = todo.Status
		child.Modified = timestamp
		// Recursively apply to grandchildren
		applyStatusRecursive(child, todo.Status, timestamp)
	}
}

// applyStatusRecursive applies a status to a todo and all its descendants
func applyStatusRecursive(todo *models.Todo, status models.TodoStatus, timestamp time.Time) {
	todo.Status = status
	todo.Modified = timestamp

	for _, child := range todo.Items {
		child.Status = status
		child.Modified = timestamp
		applyStatusRecursive(child, status, timestamp)
	}
}
