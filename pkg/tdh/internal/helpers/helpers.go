package helpers

import (
	"fmt"

	"github.com/arthur-debert/tdh/pkg/tdh/models"
)

// FindByPosition finds a todo by its position in a collection.
func FindByPosition(c *models.Collection, position int) (*models.Todo, error) {
	for _, todo := range c.Todos {
		if position == todo.Position {
			return todo, nil
		}
	}
	return nil, fmt.Errorf("todo with position %d was not found", position)
}

// Find is deprecated - use FindByPosition instead
// Kept for backward compatibility
func Find(c *models.Collection, position int) (*models.Todo, error) {
	return FindByPosition(c, position)
}
