package helpers

import (
	"errors"
	"fmt"

	"github.com/arthur-debert/tdh/pkg/tdh/models"
)

// Find finds a todo by its ID in a collection.
func Find(c *models.Collection, id int) (*models.Todo, error) {
	id64 := int64(id)
	for _, todo := range c.Todos {
		if id64 == todo.ID {
			return todo, nil
		}
	}
	return nil, fmt.Errorf("todo with id %d was not found", id)
}

// RemoveAtIndex removes a todo from a collection at a specific index.
func RemoveAtIndex(c *models.Collection, item int) {
	c.Todos = append(c.Todos[:item], c.Todos[item+1:]...)
}

// RemoveFinishedTodos removes all done todos from a collection.
// Returns the count of remaining active todos.
func RemoveFinishedTodos(c *models.Collection) int {
	var activeTodos []*models.Todo
	for _, todo := range c.Todos {
		if todo.Status != models.StatusDone {
			activeTodos = append(activeTodos, todo)
		}
	}
	c.Todos = activeTodos
	return len(activeTodos)
}

// Swap swaps the position of two todos in a collection by their IDs.
// Note: This also swaps the IDs, which maintains the visual order.
func Swap(c *models.Collection, idA, idB int) error {
	var positionA, positionB = -1, -1
	idA64 := int64(idA)
	idB64 := int64(idB)

	for i, todo := range c.Todos {
		if todo.ID == idA64 {
			positionA = i
		}
		if todo.ID == idB64 {
			positionB = i
		}
	}

	if positionA == -1 || positionB == -1 {
		return errors.New("one or both todos not found")
	}

	c.Todos[positionA], c.Todos[positionB] = c.Todos[positionB], c.Todos[positionA]
	c.Todos[positionA].ID, c.Todos[positionB].ID = c.Todos[positionB].ID, c.Todos[positionA].ID
	return nil
}
