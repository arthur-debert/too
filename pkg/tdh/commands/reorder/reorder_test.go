package reorder_test

import (
	"testing"

	"github.com/arthur-debert/tdh/pkg/tdh/commands/reorder"
	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/arthur-debert/tdh/pkg/tdh/testutil"
	"github.com/stretchr/testify/assert"
)

func TestReorderCommand(t *testing.T) {
	t.Run("successfully swaps two todos", func(t *testing.T) {
		// Create store with multiple todos using testutil
		store := testutil.CreatePopulatedStore(t,
			"First todo",
			"Second todo",
			"Third todo",
			"Fourth todo",
		)

		// Get the collection to verify initial state
		collection, _ := store.Load()
		firstPosition := collection.Todos[0].Position
		thirdPosition := collection.Todos[2].Position

		// Execute reorder command to swap first and third
		opts := reorder.Options{CollectionPath: store.Path()}
		result, err := reorder.Execute(int(firstPosition), int(thirdPosition), opts)

		// Verify result
		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.TodoA)
		assert.NotNil(t, result.TodoB)

		// Verify the swap in persistence
		collection, err = store.Load()
		testutil.AssertNoError(t, err)

		// After swap, the IDs are also swapped, so:
		// Position 0 should have third todo with first ID
		// Position 2 should have first todo with third ID
		assert.Equal(t, firstPosition, collection.Todos[0].Position)
		assert.Equal(t, "Third todo", collection.Todos[0].Text)
		assert.Equal(t, thirdPosition, collection.Todos[2].Position)
		assert.Equal(t, "First todo", collection.Todos[2].Text)

		// Other todos should remain unchanged
		assert.Equal(t, "Second todo", collection.Todos[1].Text)
		assert.Equal(t, "Fourth todo", collection.Todos[3].Text)
	})

	t.Run("swaps adjacent todos correctly", func(t *testing.T) {
		// Create store with todos
		store := testutil.CreatePopulatedStore(t,
			"Todo A",
			"Todo B",
			"Todo C",
		)

		// Get IDs
		collection, _ := store.Load()
		posA := collection.Todos[0].Position
		posB := collection.Todos[1].Position

		// Swap adjacent todos
		opts := reorder.Options{CollectionPath: store.Path()}
		result, err := reorder.Execute(posA, posB, opts)

		// Verify
		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)

		// Check persistence
		collection, err = store.Load()
		testutil.AssertNoError(t, err)

		// After swap: B, A, C (with swapped IDs)
		assert.Equal(t, posA, collection.Todos[0].Position)
		assert.Equal(t, "Todo B", collection.Todos[0].Text)
		assert.Equal(t, posB, collection.Todos[1].Position)
		assert.Equal(t, "Todo A", collection.Todos[1].Text)
		assert.Equal(t, "Todo C", collection.Todos[2].Text)
	})

	t.Run("preserves todo status when reordering", func(t *testing.T) {
		// Create store with mixed todo states
		store := testutil.CreateStoreWithSpecs(t, []testutil.TodoSpec{
			{Text: "Pending todo", Status: models.StatusPending},
			{Text: "Done todo", Status: models.StatusDone},
			{Text: "Another pending", Status: models.StatusPending},
		})

		// Get IDs
		collection, _ := store.Load()
		pendingPosition := collection.Todos[0].Position
		donePosition := collection.Todos[1].Position

		// Swap pending and done todos
		opts := reorder.Options{CollectionPath: store.Path()}
		_, err := reorder.Execute(pendingPosition, donePosition, opts)

		// Verify status preservation
		testutil.AssertNoError(t, err)

		collection, err = store.Load()
		testutil.AssertNoError(t, err)

		// First position now has done todo
		testutil.AssertTodoHasStatus(t, collection.Todos[0], models.StatusDone)
		assert.Equal(t, "Done todo", collection.Todos[0].Text)

		// Second position now has pending todo
		testutil.AssertTodoHasStatus(t, collection.Todos[1], models.StatusPending)
		assert.Equal(t, "Pending todo", collection.Todos[1].Text)
	})

	t.Run("returns error when first todo not found", func(t *testing.T) {
		// Create store with todos
		store := testutil.CreatePopulatedStore(t, "Todo 1", "Todo 2")

		// Get valid ID
		collection, _ := store.Load()
		validPosition := collection.Todos[0].Position

		// Try to reorder with non-existent ID
		opts := reorder.Options{CollectionPath: store.Path()}
		result, err := reorder.Execute(999, int(validPosition), opts)

		// Verify error
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "one or both todos not found")

		// Verify no changes were made
		collection, err = store.Load()
		testutil.AssertNoError(t, err)
		assert.Equal(t, "Todo 1", collection.Todos[0].Text)
		assert.Equal(t, "Todo 2", collection.Todos[1].Text)
	})

	t.Run("returns error when second todo not found", func(t *testing.T) {
		// Create store with todos
		store := testutil.CreatePopulatedStore(t, "Todo 1", "Todo 2")

		// Get valid ID
		collection, _ := store.Load()
		validPosition := collection.Todos[0].Position

		// Try to reorder with non-existent ID
		opts := reorder.Options{CollectionPath: store.Path()}
		result, err := reorder.Execute(int(validPosition), 999, opts)

		// Verify error
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "one or both todos not found")
	})

	t.Run("handles reordering same todo gracefully", func(t *testing.T) {
		// Create store with todos
		store := testutil.CreatePopulatedStore(t, "Todo 1", "Todo 2", "Todo 3")

		// Get ID
		collection, _ := store.Load()
		todoPosition := collection.Todos[1].Position

		// Try to swap todo with itself
		opts := reorder.Options{CollectionPath: store.Path()}
		result, err := reorder.Execute(int(todoPosition), int(todoPosition), opts)

		// Should succeed (no-op)
		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)

		// Verify nothing changed
		collection, err = store.Load()
		testutil.AssertNoError(t, err)
		assert.Equal(t, "Todo 1", collection.Todos[0].Text)
		assert.Equal(t, "Todo 2", collection.Todos[1].Text)
		assert.Equal(t, "Todo 3", collection.Todos[2].Text)
	})
}
