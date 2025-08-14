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

		// Determine new status
		newStatus := models.StatusDone
		if todo.Status == models.StatusDone {
			newStatus = models.StatusPending
		}

		// Apply the new status to the todo and all its descendants.
		// Business Rule: A parent's status must match all its children.
		// A "done" item cannot have "pending" children, and vice versa.
		// This ensures logical consistency in the task hierarchy.
		applyStatusRecursive(todo, newStatus, time.Now())

		newStatusStr := string(newStatus)

		// Auto-reorder after toggle
		collection.Reorder()

		// Capture result
		result = &Result{
			Todo:      todo,
			OldStatus: oldStatus,
			NewStatus: newStatusStr,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// applyStatusRecursive applies a status to a todo and all its descendants.
// This ensures the invariant that a parent and all its children must have
// the same status - a completed task cannot have incomplete subtasks.
func applyStatusRecursive(todo *models.Todo, status models.TodoStatus, timestamp time.Time) {
	todo.Status = status
	todo.Modified = timestamp

	for _, child := range todo.Items {
		applyStatusRecursive(child, status, timestamp)
	}
}
