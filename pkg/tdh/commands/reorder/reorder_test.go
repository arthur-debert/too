package reorder_test

import (
	"testing"

	"github.com/arthur-debert/tdh/pkg/tdh/commands/reorder"
	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/arthur-debert/tdh/pkg/tdh/testutil"
	"github.com/stretchr/testify/assert"
)

func TestReorderCommand(t *testing.T) {
	t.Run("reorders todos with gaps in positions", func(t *testing.T) {
		// Create store with todos
		store := testutil.CreatePopulatedStore(t,
			"First todo",
			"Second todo",
			"Third todo",
			"Fourth todo",
		)

		// Manually set positions with gaps
		collection, _ := store.Load()
		collection.Todos[0].Position = 1
		collection.Todos[1].Position = 3
		collection.Todos[2].Position = 7
		collection.Todos[3].Position = 10
		if err := store.Save(collection); err != nil {
			t.Fatalf("failed to save collection: %v", err)
		}

		// Execute reorder command
		opts := reorder.Options{CollectionPath: store.Path()}
		result, err := reorder.Execute(opts)

		// Verify result
		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 3, result.ReorderedCount) // All except the first one should be reordered
		assert.Len(t, result.Todos, 4)

		// Verify the new positions in persistence
		collection, err = store.Load()
		testutil.AssertNoError(t, err)

		// Todos should now have sequential positions 1, 2, 3, 4
		assert.Equal(t, 1, collection.Todos[0].Position)
		assert.Equal(t, "First todo", collection.Todos[0].Text)
		assert.Equal(t, 2, collection.Todos[1].Position)
		assert.Equal(t, "Second todo", collection.Todos[1].Text)
		assert.Equal(t, 3, collection.Todos[2].Position)
		assert.Equal(t, "Third todo", collection.Todos[2].Text)
		assert.Equal(t, 4, collection.Todos[3].Position)
		assert.Equal(t, "Fourth todo", collection.Todos[3].Text)
	})

	t.Run("no changes when todos are already sequential", func(t *testing.T) {
		// Create store with todos already in sequential order
		store := testutil.CreatePopulatedStore(t,
			"Todo A",
			"Todo B",
			"Todo C",
			"Todo D",
		)

		// Execute reorder command
		opts := reorder.Options{CollectionPath: store.Path()}
		result, err := reorder.Execute(opts)

		// Verify result
		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 0, result.ReorderedCount) // No changes needed
		assert.Len(t, result.Todos, 4)

		// Verify positions remain unchanged
		collection, err := store.Load()
		testutil.AssertNoError(t, err)

		for i, todo := range collection.Todos {
			assert.Equal(t, i+1, todo.Position)
		}
	})

	t.Run("sorts todos by position before reassigning", func(t *testing.T) {
		// Create store with todos
		store := testutil.CreatePopulatedStore(t,
			"Third todo",
			"First todo",
			"Fourth todo",
			"Second todo",
		)

		// Set positions that require sorting
		collection, _ := store.Load()
		collection.Todos[0].Position = 5 // Third
		collection.Todos[1].Position = 1 // First
		collection.Todos[2].Position = 8 // Fourth
		collection.Todos[3].Position = 3 // Second
		if err := store.Save(collection); err != nil {
			t.Fatalf("failed to save collection: %v", err)
		}

		// Execute reorder command
		opts := reorder.Options{CollectionPath: store.Path()}
		result, err := reorder.Execute(opts)

		// Verify result
		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 3, result.ReorderedCount) // Position 3->2, 5->3 and 8->4

		// Verify the todos are now sorted and have sequential positions
		collection, err = store.Load()
		testutil.AssertNoError(t, err)

		assert.Len(t, collection.Todos, 4)
		assert.Equal(t, "First todo", collection.Todos[0].Text)
		assert.Equal(t, 1, collection.Todos[0].Position)
		assert.Equal(t, "Second todo", collection.Todos[1].Text)
		assert.Equal(t, 2, collection.Todos[1].Position)
		assert.Equal(t, "Third todo", collection.Todos[2].Text)
		assert.Equal(t, 3, collection.Todos[2].Position)
		assert.Equal(t, "Fourth todo", collection.Todos[3].Text)
		assert.Equal(t, 4, collection.Todos[3].Position)
	})

	t.Run("preserves todo status when reordering", func(t *testing.T) {
		// Create store with mixed todo states
		store := testutil.CreateStoreWithSpecs(t, []testutil.TodoSpec{
			{Text: "Pending todo", Status: models.StatusPending},
			{Text: "Done todo", Status: models.StatusDone},
			{Text: "Another pending", Status: models.StatusPending},
		})

		// Set positions with gaps
		collection, _ := store.Load()
		collection.Todos[0].Position = 2
		collection.Todos[1].Position = 5
		collection.Todos[2].Position = 8
		if err := store.Save(collection); err != nil {
			t.Fatalf("failed to save collection: %v", err)
		}

		// Execute reorder command
		opts := reorder.Options{CollectionPath: store.Path()}
		result, err := reorder.Execute(opts)

		// Verify status preservation
		testutil.AssertNoError(t, err)
		assert.Equal(t, 3, result.ReorderedCount)

		collection, err = store.Load()
		testutil.AssertNoError(t, err)

		// First position: pending todo
		assert.Equal(t, 1, collection.Todos[0].Position)
		testutil.AssertTodoHasStatus(t, collection.Todos[0], models.StatusPending)
		assert.Equal(t, "Pending todo", collection.Todos[0].Text)

		// Second position: done todo
		assert.Equal(t, 2, collection.Todos[1].Position)
		testutil.AssertTodoHasStatus(t, collection.Todos[1], models.StatusDone)
		assert.Equal(t, "Done todo", collection.Todos[1].Text)

		// Third position: another pending
		assert.Equal(t, 3, collection.Todos[2].Position)
		testutil.AssertTodoHasStatus(t, collection.Todos[2], models.StatusPending)
		assert.Equal(t, "Another pending", collection.Todos[2].Text)
	})

	t.Run("handles empty collection", func(t *testing.T) {
		// Create empty store
		store := testutil.CreatePopulatedStore(t) // Creates empty if no args

		// Execute reorder command
		opts := reorder.Options{CollectionPath: store.Path()}
		result, err := reorder.Execute(opts)

		// Verify result
		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 0, result.ReorderedCount)
		assert.Len(t, result.Todos, 0)
	})

	t.Run("handles single todo", func(t *testing.T) {
		// Create store with single todo
		store := testutil.CreatePopulatedStore(t, "Single todo")

		// Set non-1 position
		collection, _ := store.Load()
		collection.Todos[0].Position = 5
		if err := store.Save(collection); err != nil {
			t.Fatalf("failed to save collection: %v", err)
		}

		// Execute reorder command
		opts := reorder.Options{CollectionPath: store.Path()}
		result, err := reorder.Execute(opts)

		// Verify result
		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.ReorderedCount) // Position changed from 5 to 1
		assert.Len(t, result.Todos, 1)

		// Verify new position
		collection, err = store.Load()
		testutil.AssertNoError(t, err)
		assert.Equal(t, 1, collection.Todos[0].Position)
		assert.Equal(t, "Single todo", collection.Todos[0].Text)
	})

	t.Run("handles duplicate positions correctly", func(t *testing.T) {
		// Create store with todos
		store := testutil.CreatePopulatedStore(t,
			"Todo A",
			"Todo B",
			"Todo C",
			"Todo D",
		)

		// Set duplicate positions
		collection, _ := store.Load()
		collection.Todos[0].Position = 3 // A
		collection.Todos[1].Position = 1 // B
		collection.Todos[2].Position = 3 // C (duplicate)
		collection.Todos[3].Position = 2 // D
		if err := store.Save(collection); err != nil {
			t.Fatalf("failed to save collection: %v", err)
		}

		// Execute reorder command
		opts := reorder.Options{CollectionPath: store.Path()}
		result, err := reorder.Execute(opts)

		// Verify result
		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.ReorderedCount) // Only C needs to change from position 3 to 4

		// Verify todos are sorted and have unique sequential positions
		collection, err = store.Load()
		testutil.AssertNoError(t, err)

		assert.Len(t, collection.Todos, 4)
		// Order should be: B(1), D(2), A(3), C(3) -> B(1), D(2), A(3), C(4)
		assert.Equal(t, "Todo B", collection.Todos[0].Text)
		assert.Equal(t, 1, collection.Todos[0].Position)
		assert.Equal(t, "Todo D", collection.Todos[1].Text)
		assert.Equal(t, 2, collection.Todos[1].Position)
		assert.Equal(t, "Todo A", collection.Todos[2].Text)
		assert.Equal(t, 3, collection.Todos[2].Position)
		assert.Equal(t, "Todo C", collection.Todos[3].Text)
		assert.Equal(t, 4, collection.Todos[3].Position)
	})

	t.Run("preserves all todo fields during reorder", func(t *testing.T) {
		// Create store with a todo
		store := testutil.CreateStoreWithSpecs(t, []testutil.TodoSpec{
			{Text: "Complex todo #project @context", Status: models.StatusPending},
		})

		// Set high position
		collection, _ := store.Load()
		collection.Todos[0].Position = 10
		if err := store.Save(collection); err != nil {
			t.Fatalf("failed to save collection: %v", err)
		}

		// Get the original todo to compare fields
		originalTodo := collection.Todos[0]

		// Execute reorder command
		opts := reorder.Options{CollectionPath: store.Path()}
		result, err := reorder.Execute(opts)

		// Verify result
		testutil.AssertNoError(t, err)
		assert.Equal(t, 1, result.ReorderedCount)

		// Verify all fields are preserved except position
		collection, err = store.Load()
		testutil.AssertNoError(t, err)

		reorderedTodo := collection.Todos[0]
		assert.Equal(t, originalTodo.ID, reorderedTodo.ID)
		assert.Equal(t, originalTodo.Text, reorderedTodo.Text)
		assert.Equal(t, originalTodo.Status, reorderedTodo.Status)
		assert.Equal(t, 1, reorderedTodo.Position) // Only position should change
		assert.Equal(t, originalTodo.Modified, reorderedTodo.Modified)
	})
}
