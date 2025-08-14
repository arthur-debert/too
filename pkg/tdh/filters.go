package tdh

import (
	"github.com/arthur-debert/tdh/pkg/tdh/models"
)

// ListPendingTodos filters the collection to only show pending todos.
// WARNING: This modifies the collection in place, removing all non-pending todos.
func ListPendingTodos(c *models.Collection) {
	for i := len(c.Todos) - 1; i >= 0; i-- {
		if c.Todos[i].Status != models.StatusPending {
			RemoveAtIndex(c, i)
		}
	}
}

// ListDoneTodos filters the collection to only show done todos.
// WARNING: This modifies the collection in place, removing all non-done todos.
func ListDoneTodos(c *models.Collection) {
	for i := len(c.Todos) - 1; i >= 0; i-- {
		if c.Todos[i].Status != models.StatusDone {
			RemoveAtIndex(c, i)
		}
	}
}

// FilterPending returns a new slice containing only pending todos without modifying the collection.
func FilterPending(c *models.Collection) []*models.Todo {
	var pending []*models.Todo
	for _, todo := range c.Todos {
		if todo.Status == models.StatusPending {
			pending = append(pending, todo)
		}
	}
	return pending
}

// FilterDone returns a new slice containing only done todos without modifying the collection.
func FilterDone(c *models.Collection) []*models.Todo {
	var done []*models.Todo
	for _, todo := range c.Todos {
		if todo.Status == models.StatusDone {
			done = append(done, todo)
		}
	}
	return done
}
