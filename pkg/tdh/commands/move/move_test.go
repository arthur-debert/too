package move_test

import (
	"testing"

	"github.com/arthur-debert/tdh/pkg/tdh/commands/move"
	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/arthur-debert/tdh/pkg/tdh/store"
	"github.com/arthur-debert/tdh/pkg/tdh/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMove(t *testing.T) {
	t.Run("should move todo from root to nested position", func(t *testing.T) {
		// Create store with some todos
		dir := testutil.TempDir(t)
		dbPath := dir + "/test.json"
		s := store.NewStore(dbPath)

		err := s.Update(func(collection *models.Collection) error {
			// Create structure:
			// 1. Root todo 1
			// 2. Parent todo
			//    2.1 Child 1
			// 3. Root todo 3
			_, _ = collection.CreateTodo("Root todo 1", "")
			parent, _ := collection.CreateTodo("Parent todo", "")
			_, _ = collection.CreateTodo("Child 1", parent.ID)
			_, _ = collection.CreateTodo("Root todo 3", "")
			return nil
		})
		testutil.AssertNoError(t, err)

		// Move root todo 3 to be child of parent todo
		result, err := move.Execute("3", "2", move.Options{
			CollectionPath: s.Path(),
		})
		testutil.AssertNoError(t, err)

		assert.Equal(t, "Root todo 3", result.Todo.Text)
		assert.Equal(t, "3", result.OldPath)
		assert.Equal(t, "2.2", result.NewPath) // Should be second child of parent

		// Verify the structure
		collection, err := s.Load()
		testutil.AssertNoError(t, err)
		assert.Equal(t, 2, len(collection.Todos)) // Only 2 root todos now
		assert.Equal(t, "Root todo 1", collection.Todos[0].Text)
		assert.Equal(t, "Parent todo", collection.Todos[1].Text)
		assert.Equal(t, 2, len(collection.Todos[1].Items)) // Parent now has 2 children
		assert.Equal(t, "Child 1", collection.Todos[1].Items[0].Text)
		assert.Equal(t, "Root todo 3", collection.Todos[1].Items[1].Text)
	})

	t.Run("should move nested todo to root", func(t *testing.T) {
		// Create nested structure
		s := testutil.CreateNestedStore(t)

		// Move sub-task 1.2 to root
		result, err := move.Execute("1.2", "", move.Options{
			CollectionPath: s.Path(),
		})
		testutil.AssertNoError(t, err)

		assert.Equal(t, "Sub-task 1.2", result.Todo.Text)
		assert.Equal(t, "1.2", result.OldPath)
		assert.Equal(t, "3", result.NewPath) // Should be position 3 at root (after Parent and Another top-level)

		// Verify the structure
		collection, err := s.Load()
		testutil.AssertNoError(t, err)
		assert.Equal(t, 3, len(collection.Todos))
		assert.Equal(t, "Parent todo", collection.Todos[0].Text)
		assert.Equal(t, 1, len(collection.Todos[0].Items)) // Parent now has only 1 child
		assert.Equal(t, "Another top-level todo", collection.Todos[1].Text)
		assert.Equal(t, "Sub-task 1.2", collection.Todos[2].Text)
		assert.Equal(t, "", collection.Todos[2].ParentID) // Should have no parent
	})

	t.Run("should move todo with children", func(t *testing.T) {
		// Create nested structure
		s := testutil.CreateNestedStore(t)

		// No need to add another parent - we already have "Another top-level todo" at position 2

		// Move sub-task 1.2 (which has grandchild) to the new parent
		result, err := move.Execute("1.2", "2", move.Options{
			CollectionPath: s.Path(),
		})
		testutil.AssertNoError(t, err)

		assert.Equal(t, "Sub-task 1.2", result.Todo.Text)
		assert.Equal(t, "1.2", result.OldPath)
		assert.Equal(t, "2.1", result.NewPath)

		// Verify children moved with parent
		collection, err := s.Load()
		testutil.AssertNoError(t, err)
		newParent := collection.Todos[1]
		assert.Equal(t, "Another top-level todo", newParent.Text)
		assert.Equal(t, 1, len(newParent.Items))
		movedTodo := newParent.Items[0]
		assert.Equal(t, "Sub-task 1.2", movedTodo.Text)
		assert.Equal(t, 1, len(movedTodo.Items)) // Still has its grandchild
		assert.Equal(t, "Grandchild 1.2.1", movedTodo.Items[0].Text)
	})

	t.Run("should reorder siblings after move", func(t *testing.T) {
		dir := testutil.TempDir(t)
		dbPath := dir + "/test.json"
		s := store.NewStore(dbPath)

		err := s.Update(func(collection *models.Collection) error {
			// Create structure with multiple siblings
			parent, _ := collection.CreateTodo("Parent", "")
			_, _ = collection.CreateTodo("Child 1", parent.ID)
			_, _ = collection.CreateTodo("Child 2", parent.ID)
			_, _ = collection.CreateTodo("Child 3", parent.ID)
			_, _ = collection.CreateTodo("Child 4", parent.ID)
			return nil
		})
		testutil.AssertNoError(t, err)

		// Move child 2 to root
		_, err = move.Execute("1.2", "", move.Options{
			CollectionPath: s.Path(),
		})
		testutil.AssertNoError(t, err)

		// Verify remaining children are reordered
		collection, err := s.Load()
		testutil.AssertNoError(t, err)
		parent := collection.Todos[0]
		assert.Equal(t, 3, len(parent.Items))
		assert.Equal(t, "Child 1", parent.Items[0].Text)
		assert.Equal(t, 1, parent.Items[0].Position)
		assert.Equal(t, "Child 3", parent.Items[1].Text)
		assert.Equal(t, 2, parent.Items[1].Position)
		assert.Equal(t, "Child 4", parent.Items[2].Text)
		assert.Equal(t, 3, parent.Items[2].Position)
	})

	t.Run("should prevent moving parent into its own descendant", func(t *testing.T) {
		// Create nested structure
		s := testutil.CreateNestedStore(t)

		// Try to move parent into its own grandchild
		_, err := move.Execute("1", "1.2.1", move.Options{
			CollectionPath: s.Path(),
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot move a parent into its own descendant")
	})

	t.Run("should handle invalid source path", func(t *testing.T) {
		s := testutil.CreateNestedStore(t)

		// Try to move non-existent todo
		_, err := move.Execute("99", "1", move.Options{
			CollectionPath: s.Path(),
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "todo not found")
	})

	t.Run("should handle invalid destination path", func(t *testing.T) {
		s := testutil.CreateNestedStore(t)

		// Try to move to non-existent parent
		_, err := move.Execute("1.1", "99", move.Options{
			CollectionPath: s.Path(),
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "destination parent not found")
	})

	t.Run("should handle moving between deeply nested levels", func(t *testing.T) {
		dir := testutil.TempDir(t)
		dbPath := dir + "/test.json"
		s := store.NewStore(dbPath)

		err := s.Update(func(collection *models.Collection) error {
			// Create deep structure
			level1, _ := collection.CreateTodo("Level 1", "")
			level2, _ := collection.CreateTodo("Level 2", level1.ID)
			level3, _ := collection.CreateTodo("Level 3", level2.ID)
			_, _ = collection.CreateTodo("Level 4", level3.ID)

			// Create another branch
			another1, _ := collection.CreateTodo("Another 1", "")
			another2, _ := collection.CreateTodo("Another 2", another1.ID)
			_, _ = collection.CreateTodo("Another 3", another2.ID)

			return nil
		})
		testutil.AssertNoError(t, err)

		// Move Level 3 (with Level 4) to Another 2
		result, err := move.Execute("1.1.1", "2.1", move.Options{
			CollectionPath: s.Path(),
		})
		testutil.AssertNoError(t, err)

		assert.Equal(t, "Level 3", result.Todo.Text)
		assert.Equal(t, "1.1.1", result.OldPath)
		assert.Equal(t, "2.1.2", result.NewPath) // Second child of Another 2

		// Verify the structure
		collection, err := s.Load()
		testutil.AssertNoError(t, err)

		// Check original branch
		level1 := collection.Todos[0]
		assert.Equal(t, 1, len(level1.Items))
		level2 := level1.Items[0]
		assert.Equal(t, 0, len(level2.Items)) // Level 3 moved away

		// Check destination branch
		another1 := collection.Todos[1]
		another2 := another1.Items[0]
		assert.Equal(t, 2, len(another2.Items)) // Now has 2 children
		assert.Equal(t, "Another 3", another2.Items[0].Text)
		assert.Equal(t, "Level 3", another2.Items[1].Text)
		assert.Equal(t, 1, len(another2.Items[1].Items)) // Level 3 still has Level 4
		assert.Equal(t, "Level 4", another2.Items[1].Items[0].Text)
	})

	t.Run("should maintain todo properties after move", func(t *testing.T) {
		dir := testutil.TempDir(t)
		dbPath := dir + "/test.json"
		s := store.NewStore(dbPath)

		var originalID string
		err := s.Update(func(collection *models.Collection) error {
			// Create a done todo with specific properties
			parent, _ := collection.CreateTodo("Parent", "")
			todo, _ := collection.CreateTodo("Done todo", parent.ID)
			todo.Status = models.StatusDone
			originalID = todo.ID
			return nil
		})
		testutil.AssertNoError(t, err)

		// Move done todo to root
		result, err := move.Execute("1.1", "", move.Options{
			CollectionPath: s.Path(),
		})
		testutil.AssertNoError(t, err)

		// Verify properties maintained
		assert.Equal(t, originalID, result.Todo.ID)
		assert.Equal(t, models.StatusDone, result.Todo.Status)
		assert.Equal(t, "Done todo", result.Todo.Text)
	})
}
