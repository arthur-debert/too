package reorder_test

import (
	"testing"

	"github.com/arthur-debert/tdh/pkg/tdh/commands/reorder"
	"github.com/arthur-debert/tdh/pkg/tdh/testutil"
	"github.com/stretchr/testify/assert"
)

func TestReorderCommand(t *testing.T) {
	t.Run("executes reorder and returns result", func(t *testing.T) {
		// Create store with todos that have gaps
		store := testutil.CreatePopulatedStore(t,
			"First todo",
			"Second todo",
			"Third todo",
		)

		// Manually set positions with gaps
		collection, _ := store.Load()
		collection.Todos[0].Position = 1
		collection.Todos[1].Position = 5
		collection.Todos[2].Position = 10
		if err := store.Save(collection); err != nil {
			t.Fatalf("failed to save collection: %v", err)
		}

		// Execute reorder command
		opts := reorder.Options{CollectionPath: store.Path()}
		result, err := reorder.Execute(opts)

		// Verify result
		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 2, result.ReorderedCount)
		assert.Len(t, result.Todos, 3)

		// Verify persistence
		collection, err = store.Load()
		testutil.AssertNoError(t, err)
		assert.Equal(t, 1, collection.Todos[0].Position)
		assert.Equal(t, 2, collection.Todos[1].Position)
		assert.Equal(t, 3, collection.Todos[2].Position)
	})

	t.Run("handles empty collection", func(t *testing.T) {
		// Create empty store
		store := testutil.CreatePopulatedStore(t)

		// Execute reorder command
		opts := reorder.Options{CollectionPath: store.Path()}
		result, err := reorder.Execute(opts)

		// Verify result
		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 0, result.ReorderedCount)
		assert.Len(t, result.Todos, 0)
	})

	t.Run("returns copy of reordered todos", func(t *testing.T) {
		// Create store with todos
		store := testutil.CreatePopulatedStore(t, "Todo 1", "Todo 2")

		// Execute reorder command
		opts := reorder.Options{CollectionPath: store.Path()}
		result, err := reorder.Execute(opts)

		// Verify we got a copy in the result
		testutil.AssertNoError(t, err)
		assert.Len(t, result.Todos, 2)
		assert.Equal(t, "Todo 1", result.Todos[0].Text)
		assert.Equal(t, "Todo 2", result.Todos[1].Text)
	})

	t.Run("only reorders active todos", func(t *testing.T) {
		// Create store with mixed status todos
		store := testutil.CreateStoreWithSpecs(t, []testutil.TodoSpec{
			{Text: "Done todo", Status: "done"},
			{Text: "Active 1", Status: "pending"},
			{Text: "Active 2", Status: "pending"},
			{Text: "Another done", Status: "done"},
		})

		// Manually set positions with gaps
		collection, _ := store.Load()
		collection.Todos[0].Position = 1  // Done
		collection.Todos[1].Position = 5  // Active
		collection.Todos[2].Position = 10 // Active
		collection.Todos[3].Position = 3  // Done
		if err := store.Save(collection); err != nil {
			t.Fatalf("failed to save collection: %v", err)
		}

		// Execute reorder command
		opts := reorder.Options{CollectionPath: store.Path()}
		result, err := reorder.Execute(opts)

		// Verify result
		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)

		// Verify persistence
		collection, err = store.Load()
		testutil.AssertNoError(t, err)

		// Done items should have position 0
		assert.Equal(t, 0, collection.Todos[0].Position) // Done todo
		assert.Equal(t, 0, collection.Todos[3].Position) // Another done

		// Active items should be reordered 1, 2
		assert.Equal(t, 1, collection.Todos[1].Position) // Active 1
		assert.Equal(t, 2, collection.Todos[2].Position) // Active 2
	})
}
