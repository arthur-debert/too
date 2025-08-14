package tdh

import (
	"github.com/arthur-debert/tdh/pkg/tdh/internal/helpers"
	"github.com/arthur-debert/tdh/pkg/tdh/models"
)

// Find finds a todo by its ID in a collection.
func Find(c *models.Collection, id int) (*models.Todo, error) {
	return helpers.Find(c, id)
}

// RemoveAtIndex removes a todo from a collection at a specific index.
func RemoveAtIndex(c *models.Collection, item int) {
	helpers.RemoveAtIndex(c, item)
}

// RemoveFinishedTodos removes all done todos from a collection.
// Returns the count of remaining active todos.
func RemoveFinishedTodos(c *models.Collection) int {
	return helpers.RemoveFinishedTodos(c)
}

// Swap swaps the position of two todos in a collection by their IDs.
// Note: This also swaps the IDs, which maintains the visual order.
func Swap(c *models.Collection, idA, idB int) error {
	return helpers.Swap(c, idA, idB)
}
