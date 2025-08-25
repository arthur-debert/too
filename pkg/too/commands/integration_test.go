package commands_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/arthur-debert/too/pkg/too"
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/testutil"
)

func TestIntegration_CompleteReopenList(t *testing.T) {
	t.Run("reopening an item correctly places it back in the active list", func(t *testing.T) {
		// Setup
		store := testutil.CreatePopulatedStore(t, "Item 1", "Item 2", "Item 3")
		path := store.Path()

		// 1. Complete the middle item
		_, err := too.Complete("2", too.CompleteOptions{CollectionPath: path})
		testutil.AssertNoError(t, err)

		// 2. Verify the active list is correct
		listResult1, err := too.List(too.ListOptions{CollectionPath: path})
		testutil.AssertNoError(t, err)
		assert.Len(t, listResult1.Todos, 2)
		assert.Equal(t, "Item 1", listResult1.Todos[0].Text)
		assert.Equal(t, "Item 3", listResult1.Todos[1].Text)

		// 3. Find the archived todo's short ID
		fullCollection, err := store.Load()
		testutil.AssertNoError(t, err)
		var archivedTodo *models.Todo
		fullCollection.Walk(func(t *models.Todo) {
			if t.Text == "Item 2" {
				archivedTodo = t
			}
		})
		assert.NotNil(t, archivedTodo, "Could not find archived item in collection")
		archivedRef := archivedTodo.GetShortID()

		// 4. Reopen "Item 2" using its short ID
		_, err = too.Reopen(archivedRef, too.ReopenOptions{CollectionPath: path})
		testutil.AssertNoError(t, err)

		// 5. List the active items again
		listResult2, err := too.List(too.ListOptions{CollectionPath: path})
		testutil.AssertNoError(t, err)

		// 6. Assert that "Item 2" is back, at the end of the list
		assert.Len(t, listResult2.Todos, 3)
		assert.Equal(t, "Item 1", listResult2.Todos[0].Text)
		assert.Equal(t, "Item 3", listResult2.Todos[1].Text)
		assert.Equal(t, "Item 2", listResult2.Todos[2].Text, "Reopened item should be last")
		assert.Equal(t, 3, listResult2.Todos[2].Position, "Reopened item should have the last position")
	})
}
