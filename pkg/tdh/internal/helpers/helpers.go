package helpers

import (
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
