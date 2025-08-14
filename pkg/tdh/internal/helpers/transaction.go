package helpers

import (
	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/arthur-debert/tdh/pkg/tdh/store"
)

// TransactOnTodo provides a generic way to perform operations on a single todo within a transaction.
// It handles the common pattern of loading a collection, finding a todo by position,
// performing an action on it, and saving the changes.
func TransactOnTodo(collectionPath string, position int, action func(*models.Todo, *models.Collection) error) error {
	s := store.NewStore(collectionPath)
	return s.Update(func(collection *models.Collection) error {
		todo, err := FindByPosition(collection, position)
		if err != nil {
			return err
		}
		return action(todo, collection)
	})
}
