package move_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/arthur-debert/tdh/pkg/tdh"
	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/arthur-debert/tdh/pkg/tdh/store"
	"github.com/arthur-debert/tdh/pkg/tdh/testutil"
)

// createStoreWithDoneItems creates a store with properly positioned done items
func createStoreWithDoneItems(t *testing.T, specs []testutil.TodoSpec) store.Store {
	t.Helper()
	// Use CreateStoreWithNestedSpecs for nested structures
	store := testutil.CreateStoreWithNestedSpecs(t, specs)

	// Load and reset positions to ensure done items have position 0
	collection, err := store.Load()
	testutil.AssertNoError(t, err)

	// Reset positions at root level to ensure done items get position 0
	collection.ResetRootPositions()

	// Also reset positions for any nested items
	var resetNestedPositions func([]*models.Todo)
	resetNestedPositions = func(todos []*models.Todo) {
		for _, todo := range todos {
			if len(todo.Items) > 0 {
				models.ResetActivePositions(todo.Items)
				resetNestedPositions(todo.Items)
			}
		}
	}
	resetNestedPositions(collection.Todos)

	err = store.Save(collection)
	testutil.AssertNoError(t, err)

	return store
}

func TestMoveCommand(t *testing.T) {
	t.Run("moves a top-level item to be a child of another", func(t *testing.T) {
		store := testutil.CreateStoreWithNestedSpecs(t, []testutil.TodoSpec{
			{Text: "Parent 1"},
			{Text: "Item to move"},
		})

		opts := tdh.MoveOptions{CollectionPath: store.Path()}
		result, err := tdh.Move("2", "1", opts) // Move item at pos 2 to be child of item at pos 1

		testutil.AssertNoError(t, err)
		assert.Equal(t, "2", result.OldPath)
		assert.Equal(t, "1.1", result.NewPath)

		collection, _ := store.Load()
		assert.Len(t, collection.Todos, 1)
		assert.Equal(t, "Parent 1", collection.Todos[0].Text)
		assert.Len(t, collection.Todos[0].Items, 1)
		assert.Equal(t, "Item to move", collection.Todos[0].Items[0].Text)
	})

	t.Run("moves a nested item to the root", func(t *testing.T) {
		store := testutil.CreateStoreWithNestedSpecs(t, []testutil.TodoSpec{
			{Text: "Parent 1", Children: []testutil.TodoSpec{
				{Text: "Item to move"},
			}},
			{Text: "Parent 2"},
		})

		opts := tdh.MoveOptions{CollectionPath: store.Path()}
		result, err := tdh.Move("1.1", "", opts) // Move item at pos 1.1 to root

		testutil.AssertNoError(t, err)
		assert.Equal(t, "1.1", result.OldPath)
		assert.Equal(t, "3", result.NewPath) // Becomes the 3rd top-level item

		collection, _ := store.Load()
		assert.Len(t, collection.Todos, 3)
		assert.Len(t, collection.Todos[0].Items, 0) // Old parent is now empty
		assert.Equal(t, "Item to move", collection.Todos[2].Text)
	})

	t.Run("moves a deeply nested item between branches", func(t *testing.T) {
		store := testutil.CreateStoreWithNestedSpecs(t, []testutil.TodoSpec{
			{Text: "Branch A", Children: []testutil.TodoSpec{
				{Text: "Sub A", Children: []testutil.TodoSpec{
					{Text: "Item to move"},
				}},
			}},
			{Text: "Branch B"},
		})

		opts := tdh.MoveOptions{CollectionPath: store.Path()}
		result, err := tdh.Move("1.1.1", "2", opts) // Move 1.1.1 to be child of 2

		testutil.AssertNoError(t, err)
		assert.Equal(t, "1.1.1", result.OldPath)
		assert.Equal(t, "2.1", result.NewPath)

		collection, _ := store.Load()
		branchA := collection.Todos[0].Items[0]
		branchB := collection.Todos[1]
		assert.Len(t, branchA.Items, 0, "Original parent should be empty")
		assert.Len(t, branchB.Items, 1, "New parent should have one child")
		assert.Equal(t, "Item to move", branchB.Items[0].Text)
	})

	t.Run("fails to move a non-existent source", func(t *testing.T) {
		store := testutil.CreateStoreWithNestedSpecs(t, []testutil.TodoSpec{{Text: "Parent"}})
		opts := tdh.MoveOptions{CollectionPath: store.Path()}
		_, err := tdh.Move("99", "1", opts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "todo not found at position: 99")
	})

	t.Run("fails to move to a non-existent destination", func(t *testing.T) {
		store := testutil.CreateStoreWithNestedSpecs(t, []testutil.TodoSpec{{Text: "Item to move"}})
		opts := tdh.MoveOptions{CollectionPath: store.Path()}
		_, err := tdh.Move("1", "99", opts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "destination parent not found at position: 99")
	})

	t.Run("fails to move a parent into its own child", func(t *testing.T) {
		store := testutil.CreateStoreWithNestedSpecs(t, []testutil.TodoSpec{
			{Text: "Parent", Children: []testutil.TodoSpec{
				{Text: "Child"},
			}},
		})
		opts := tdh.MoveOptions{CollectionPath: store.Path()}
		_, err := tdh.Move("1", "1.1", opts) // Move "Parent" into "Child"
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot move a parent into its own descendant")
	})

	// New tests for position behavior with done items
	t.Run("maintains position sequence when moving with done items present", func(t *testing.T) {
		// Create a mix of active and done items
		store := createStoreWithDoneItems(t, []testutil.TodoSpec{
			{Text: "Active Parent", Status: models.StatusPending},
			{Text: "Done Item", Status: models.StatusDone},
			{Text: "Item to move", Status: models.StatusPending},
			{Text: "Another active", Status: models.StatusPending},
		})

		// Verify initial state: done item should be at position 0
		collection, _ := store.Load()
		var doneItem *models.Todo
		for _, todo := range collection.Todos {
			if todo.Status == models.StatusDone {
				doneItem = todo
				break
			}
		}
		assert.NotNil(t, doneItem)
		assert.Equal(t, 0, doneItem.Position, "Done item should have position 0")

		// Move active item at position 2 to be child of item at position 1
		opts := tdh.MoveOptions{CollectionPath: store.Path()}
		result, err := tdh.Move("2", "1", opts)

		testutil.AssertNoError(t, err)
		assert.Equal(t, "2", result.OldPath)
		assert.Equal(t, "1.1", result.NewPath)

		// Verify the remaining active items are reordered correctly
		collection, _ = store.Load()

		// Count active root items
		activeCount := 0
		for _, todo := range collection.Todos {
			if todo.Status == models.StatusPending {
				activeCount++
				// Active items should have sequential positions starting from 1
				assert.Greater(t, todo.Position, 0, "Active item '%s' should have position > 0", todo.Text)
			} else {
				// Done items should have position 0
				assert.Equal(t, 0, todo.Position, "Done item '%s' should have position 0", todo.Text)
			}
		}
		assert.Equal(t, 2, activeCount, "Should have 2 active root items after move")
	})

	t.Run("fails to move a done item", func(t *testing.T) {
		store := createStoreWithDoneItems(t, []testutil.TodoSpec{
			{Text: "Active Parent", Status: models.StatusPending},
			{Text: "Done Item", Status: models.StatusDone},
		})

		// Done items have position 0 and cannot be addressed by position path
		opts := tdh.MoveOptions{CollectionPath: store.Path()}
		_, err := tdh.Move("0", "1", opts)
		assert.Error(t, err)
		// Since position 0 items can't be found by position path, we expect a "not found" error
		assert.Contains(t, err.Error(), "todo not found at position: 0")
	})

	t.Run("reorders positions at source location after move", func(t *testing.T) {
		store := testutil.CreateStoreWithSpecs(t, []testutil.TodoSpec{
			{Text: "Item 1", Status: models.StatusPending},
			{Text: "Item 2", Status: models.StatusPending},
			{Text: "Item 3", Status: models.StatusPending},
			{Text: "Item 4", Status: models.StatusPending},
		})

		// Move item at position 2
		opts := tdh.MoveOptions{CollectionPath: store.Path()}
		_, err := tdh.Move("2", "1", opts)
		testutil.AssertNoError(t, err)

		// Verify remaining items are reordered to 1, 2, 3
		collection, _ := store.Load()
		rootActiveItems := []*models.Todo{}
		for _, todo := range collection.Todos {
			if todo.ParentID == "" && todo.Status == models.StatusPending {
				rootActiveItems = append(rootActiveItems, todo)
			}
		}

		assert.Len(t, rootActiveItems, 3, "Should have 3 root active items")
		// Sort by position to check sequence
		expectedPositions := []int{1, 2, 3}
		actualPositions := []int{}
		for _, item := range rootActiveItems {
			actualPositions = append(actualPositions, item.Position)
		}
		// Sort to compare
		assert.ElementsMatch(t, expectedPositions, actualPositions, "Positions should be sequential")
	})

	t.Run("assigns correct position at destination with existing children", func(t *testing.T) {
		store := testutil.CreateStoreWithNestedSpecs(t, []testutil.TodoSpec{
			{Text: "Parent", Status: models.StatusPending, Children: []testutil.TodoSpec{
				{Text: "Existing child 1", Status: models.StatusPending},
				{Text: "Existing child 2", Status: models.StatusPending},
			}},
			{Text: "Item to move", Status: models.StatusPending},
		})

		// Before move, parent should have 2 children
		collection, _ := store.Load()
		parent := collection.Todos[0]
		assert.Len(t, parent.Items, 2, "Parent should initially have 2 children")

		opts := tdh.MoveOptions{CollectionPath: store.Path()}
		result, err := tdh.Move("2", "1", opts)

		testutil.AssertNoError(t, err)
		// The result path tells us the actual position
		assert.NotEmpty(t, result.NewPath)

		// Verify the parent now has 3 children with correct positions
		collection, _ = store.Load()
		parent = collection.Todos[0]
		assert.Len(t, parent.Items, 3)

		// Check positions are sequential
		for i, child := range parent.Items {
			assert.Equal(t, i+1, child.Position, "Child %d should have position %d", i, i+1)
		}
	})

	t.Run("handles moving between parents with mixed status children", func(t *testing.T) {
		store := createStoreWithDoneItems(t, []testutil.TodoSpec{
			{Text: "Parent A", Status: models.StatusPending, Children: []testutil.TodoSpec{
				{Text: "A Active 1", Status: models.StatusPending},
				{Text: "A Done", Status: models.StatusDone},
				{Text: "A Active 2", Status: models.StatusPending},
			}},
			{Text: "Parent B", Status: models.StatusPending, Children: []testutil.TodoSpec{
				{Text: "B Done", Status: models.StatusDone},
				{Text: "B Active 1", Status: models.StatusPending},
			}},
		})

		// Move "A Active 1" from Parent A to Parent B
		opts := tdh.MoveOptions{CollectionPath: store.Path()}
		result, err := tdh.Move("1.1", "2", opts)

		testutil.AssertNoError(t, err)
		// Parent B should now have active items at positions 1 and 2
		assert.Equal(t, "2.2", result.NewPath, "Moved item should get position 2 in Parent B")

		collection, _ := store.Load()

		// Check Parent A - should have one active item at position 1
		parentA := collection.Todos[0]
		activeInA := 0
		for _, child := range parentA.Items {
			if child.Status == models.StatusPending {
				activeInA++
				assert.Equal(t, activeInA, child.Position, "Active items in Parent A should be sequentially numbered")
			} else {
				assert.Equal(t, 0, child.Position, "Done items should have position 0")
			}
		}
		assert.Equal(t, 1, activeInA, "Parent A should have 1 active child")

		// Check Parent B - should have two active items at positions 1 and 2
		parentB := collection.Todos[1]
		activeInB := 0
		for _, child := range parentB.Items {
			if child.Status == models.StatusPending {
				activeInB++
			}
		}
		assert.Equal(t, 2, activeInB, "Parent B should have 2 active children")
	})

	t.Run("maintains position invariants when moving nested items with done siblings", func(t *testing.T) {
		store := createStoreWithDoneItems(t, []testutil.TodoSpec{
			{Text: "Root 1", Status: models.StatusPending, Children: []testutil.TodoSpec{
				{Text: "Branch A", Status: models.StatusPending, Children: []testutil.TodoSpec{
					{Text: "Nested Active", Status: models.StatusPending},
					{Text: "Nested Done", Status: models.StatusDone},
				}},
				{Text: "Branch B", Status: models.StatusPending},
			}},
		})

		// Move the nested active item to Branch B
		opts := tdh.MoveOptions{CollectionPath: store.Path()}
		result, err := tdh.Move("1.1.1", "1.2", opts)

		testutil.AssertNoError(t, err)
		assert.Equal(t, "1.2.1", result.NewPath)

		collection, _ := store.Load()

		// Verify Branch A now only has the done item at position 0
		branchA := collection.Todos[0].Items[0]
		assert.Len(t, branchA.Items, 1)
		assert.Equal(t, "Nested Done", branchA.Items[0].Text)
		assert.Equal(t, 0, branchA.Items[0].Position)

		// Verify Branch B has the moved item at position 1
		branchB := collection.Todos[0].Items[1]
		assert.Len(t, branchB.Items, 1)
		assert.Equal(t, "Nested Active", branchB.Items[0].Text)
		assert.Equal(t, 1, branchB.Items[0].Position)
	})

	t.Run("handles moving the last active item from a parent", func(t *testing.T) {
		store := createStoreWithDoneItems(t, []testutil.TodoSpec{
			{Text: "Parent", Status: models.StatusPending, Children: []testutil.TodoSpec{
				{Text: "Done child", Status: models.StatusDone},
				{Text: "Last active", Status: models.StatusPending},
			}},
			{Text: "Destination", Status: models.StatusPending},
		})

		opts := tdh.MoveOptions{CollectionPath: store.Path()}
		_, err := tdh.Move("1.1", "2", opts)

		testutil.AssertNoError(t, err)

		collection, _ := store.Load()
		parent := collection.Todos[0]

		// Parent should only have the done item left
		assert.Len(t, parent.Items, 1)
		assert.Equal(t, "Done child", parent.Items[0].Text)
		assert.Equal(t, models.StatusDone, parent.Items[0].Status)
		assert.Equal(t, 0, parent.Items[0].Position)
	})

	t.Run("builds correct position paths excluding done items", func(t *testing.T) {
		store := createStoreWithDoneItems(t, []testutil.TodoSpec{
			{Text: "Done at root", Status: models.StatusDone},
			{Text: "Active 1", Status: models.StatusPending, Children: []testutil.TodoSpec{
				{Text: "Done child", Status: models.StatusDone},
				{Text: "Active child 1", Status: models.StatusPending},
				{Text: "Active child 2", Status: models.StatusPending},
			}},
			{Text: "Active 2", Status: models.StatusPending},
		})

		// Move "Active child 2" to "Active 2"
		opts := tdh.MoveOptions{CollectionPath: store.Path()}
		result, err := tdh.Move("1.2", "2", opts)

		testutil.AssertNoError(t, err)
		// The path should be based on active item positions only
		assert.Equal(t, "2.1", result.NewPath, "Path should reflect active-only positioning")
	})

	t.Run("verifies move command triggers position reordering", func(t *testing.T) {
		// Create a simple structure to test reordering
		store := testutil.CreateStoreWithSpecs(t, []testutil.TodoSpec{
			{Text: "Parent", Status: models.StatusPending},
			{Text: "Item A", Status: models.StatusPending},
			{Text: "Item B", Status: models.StatusPending},
			{Text: "Item C", Status: models.StatusPending},
		})

		// Move Item B into Parent
		opts := tdh.MoveOptions{CollectionPath: store.Path()}
		_, err := tdh.Move("3", "1", opts)
		testutil.AssertNoError(t, err)

		// Check that remaining root items were reordered
		collection, _ := store.Load()

		// Should have 3 root items: Parent, Item A, Item C
		assert.Len(t, collection.Todos, 3)

		// Verify positions are sequential 1, 2, 3
		positions := []int{}
		for _, todo := range collection.Todos {
			positions = append(positions, todo.Position)
		}
		assert.ElementsMatch(t, []int{1, 2, 3}, positions, "Root positions should be sequential after move")

		// Verify the moved item is in the parent
		parent := collection.Todos[0]
		assert.Len(t, parent.Items, 1)
		assert.Equal(t, "Item B", parent.Items[0].Text)
		assert.Equal(t, 1, parent.Items[0].Position)
	})
}
