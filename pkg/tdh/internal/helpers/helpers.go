package helpers

import (
	"fmt"

	"github.com/arthur-debert/tdh/pkg/tdh/models"
)

// Find finds a todo by its position in a collection.
func Find(c *models.Collection, position int) (*models.Todo, error) {
	for _, todo := range c.Todos {
		if position == todo.Position {
			return todo, nil
		}
	}
	return nil, fmt.Errorf("todo with position %d was not found", position)
}
