package move_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/arthur-debert/too/pkg/too"
	"github.com/arthur-debert/too/pkg/too/testutil"
)

func TestMoveCommand(t *testing.T) {
	t.Run("moves a top-level item to be a child of another", func(t *testing.T) {
		store := testutil.CreateStoreWithNestedSpecs(t, []testutil.TodoSpec{
			{Text: "Parent 1"},
			{Text: "Item to move"},
		})
		collection, _ := store.Load()
		collection.Reorder()
		testutil.AssertNoError(t, store.Save(collection))

		opts := too.MoveOptions{CollectionPath: store.Path()}
		result, err := too.Move("2", "1", opts) // Move item at pos 2 to be child of item at pos 1

		testutil.AssertNoError(t, err)
		if err != nil {
			return
		}
		assert.Equal(t, "2", result.OldPath)
		assert.NotEmpty(t, result.NewPath)

		collection, _ = store.Load()
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
		collection, _ := store.Load()
		collection.Reorder()
		testutil.AssertNoError(t, store.Save(collection))

		opts := too.MoveOptions{CollectionPath: store.Path()}
		result, err := too.Move("1.1", "", opts) // Move item at pos 1.1 to root

		testutil.AssertNoError(t, err)
		if err != nil {
			return
		}
		assert.Equal(t, "1.1", result.OldPath)
		assert.NotEmpty(t, result.NewPath) // Becomes a top-level item

		collection, _ = store.Load()
		assert.Len(t, collection.Todos, 3)
		assert.Len(t, collection.Todos[0].Items, 0) // Old parent is now empty
		// Moved item gets position=0 and is placed after existing active items
		assert.Equal(t, "Item to move", collection.Todos[1].Text)
		assert.Equal(t, "Parent 2", collection.Todos[2].Text)
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
		collection, _ := store.Load()
		collection.Reorder()
		testutil.AssertNoError(t, store.Save(collection))

		opts := too.MoveOptions{CollectionPath: store.Path()}
		result, err := too.Move("1.1.1", "2", opts) // Move 1.1.1 to be child of 2

		testutil.AssertNoError(t, err)
		if err != nil {
			return
		}
		assert.Equal(t, "1.1.1", result.OldPath)
		assert.NotEmpty(t, result.NewPath)

		collection, _ = store.Load()
		branchA := collection.Todos[0].Items[0]
		branchB := collection.Todos[1]
		assert.Len(t, branchA.Items, 0, "Original parent should be empty")
		assert.Len(t, branchB.Items, 1, "New parent should have one child")
		assert.Equal(t, "Item to move", branchB.Items[0].Text)
	})

	t.Run("fails to move a non-existent source", func(t *testing.T) {
		store := testutil.CreateStoreWithNestedSpecs(t, []testutil.TodoSpec{{Text: "Parent"}})
		opts := too.MoveOptions{CollectionPath: store.Path()}
		_, err := too.Move("99", "1", opts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "todo not found at position: 99")
	})

	t.Run("fails to move to a non-existent destination", func(t *testing.T) {
		store := testutil.CreateStoreWithNestedSpecs(t, []testutil.TodoSpec{{Text: "Item to move"}})
		collection, _ := store.Load()
		collection.Reorder()
		testutil.AssertNoError(t, store.Save(collection))
		opts := too.MoveOptions{CollectionPath: store.Path()}
		_, err := too.Move("1", "99", opts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "destination parent not found at position: 99")
	})

	t.Run("fails to move a parent into its own child", func(t *testing.T) {
		store := testutil.CreateStoreWithNestedSpecs(t, []testutil.TodoSpec{
			{Text: "Parent", Children: []testutil.TodoSpec{
				{Text: "Child"},
			}},
		})
		collection, _ := store.Load()
		collection.Reorder()
		testutil.AssertNoError(t, store.Save(collection))
		opts := too.MoveOptions{CollectionPath: store.Path()}
		_, err := too.Move("1", "1.1", opts) // Move "Parent" into "Child"
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot move a parent into its own descendant")
	})
}
