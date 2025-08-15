package clean_test

import (
	"testing"

	"github.com/arthur-debert/tdh/pkg/tdh/commands/clean"
	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/arthur-debert/tdh/pkg/tdh/store"
	"github.com/arthur-debert/tdh/pkg/tdh/testutil"
	"github.com/stretchr/testify/assert"
)

func TestClean_HierarchyAware(t *testing.T) {
	t.Run("should remove done parent and all its descendants", func(t *testing.T) {
		// Create nested structure with done parent
		dir := testutil.TempDir(t)
		dbPath := dir + "/test.json"
		s := store.NewStore(dbPath)

		err := s.Update(func(collection *models.Collection) error {
			// Done parent with pending children
			parent1, _ := collection.CreateTodo("Done parent", "")
			parent1.Status = models.StatusDone
			child1, _ := collection.CreateTodo("Pending child 1", parent1.ID)
			child2, _ := collection.CreateTodo("Pending child 2", parent1.ID)
			_, _ = collection.CreateTodo("Grandchild 1", child1.ID)
			_, _ = collection.CreateTodo("Grandchild 2", child2.ID)

			// Pending parent
			_, _ = collection.CreateTodo("Pending parent", "")

			return nil
		})
		testutil.AssertNoError(t, err)

		// Execute clean
		result, err := clean.Execute(clean.Options{
			CollectionPath: s.Path(),
		})
		testutil.AssertNoError(t, err)

		// Should report only the done parent (not pending descendants)
		assert.Equal(t, 1, result.RemovedCount)
		assert.Equal(t, 1, result.ActiveCount) // Only "Pending parent" remains

		// Verify only done parent is reported as removed
		assert.Equal(t, 1, len(result.RemovedTodos))
		assert.Equal(t, "Done parent", result.RemovedTodos[0].Text)
		assert.Equal(t, models.StatusDone, result.RemovedTodos[0].Status)

		// Verify only pending parent remains
		collection, err := s.Load()
		testutil.AssertNoError(t, err)
		assert.Equal(t, 1, len(collection.Todos))
		assert.Equal(t, "Pending parent", collection.Todos[0].Text)
	})

	t.Run("should remove done children but keep pending parent", func(t *testing.T) {
		// Create structure with pending parent and done children
		dir := testutil.TempDir(t)
		dbPath := dir + "/test.json"
		s := store.NewStore(dbPath)

		err := s.Update(func(collection *models.Collection) error {
			parent, _ := collection.CreateTodo("Pending parent", "")

			// Mix of done and pending children
			child1, _ := collection.CreateTodo("Done child 1", parent.ID)
			child1.Status = models.StatusDone

			_, _ = collection.CreateTodo("Pending child", parent.ID)

			child3, _ := collection.CreateTodo("Done child 2", parent.ID)
			child3.Status = models.StatusDone

			// Grandchildren of done child
			_, _ = collection.CreateTodo("Grandchild of done", child1.ID)

			return nil
		})
		testutil.AssertNoError(t, err)

		// Execute clean
		result, err := clean.Execute(clean.Options{
			CollectionPath: s.Path(),
		})
		testutil.AssertNoError(t, err)

		// Should report only done children (not their pending descendants)
		assert.Equal(t, 2, result.RemovedCount) // Done child 1 + Done child 2
		assert.Equal(t, 2, result.ActiveCount)  // Pending parent + Pending child

		// Verify structure
		collection, err := s.Load()
		testutil.AssertNoError(t, err)
		assert.Equal(t, 1, len(collection.Todos))
		parent := collection.Todos[0]
		assert.Equal(t, "Pending parent", parent.Text)
		assert.Equal(t, 1, len(parent.Items))
		assert.Equal(t, "Pending child", parent.Items[0].Text)
	})

	t.Run("should handle deeply nested done branches", func(t *testing.T) {
		dir := testutil.TempDir(t)
		dbPath := dir + "/test.json"
		s := store.NewStore(dbPath)

		err := s.Update(func(collection *models.Collection) error {
			// Create deep structure
			root, _ := collection.CreateTodo("Root", "")
			level1, _ := collection.CreateTodo("Level 1", root.ID)
			level2, _ := collection.CreateTodo("Level 2 (done)", level1.ID)
			level2.Status = models.StatusDone
			level3, _ := collection.CreateTodo("Level 3", level2.ID)
			level4, _ := collection.CreateTodo("Level 4", level3.ID)
			_, _ = collection.CreateTodo("Level 5", level4.ID)

			// Add another branch at level 1
			_, _ = collection.CreateTodo("Level 1 sibling", root.ID)

			return nil
		})
		testutil.AssertNoError(t, err)

		// Execute clean
		result, err := clean.Execute(clean.Options{
			CollectionPath: s.Path(),
		})
		testutil.AssertNoError(t, err)

		// Should report only level 2 (the done item)
		assert.Equal(t, 1, result.RemovedCount) // Only Level 2
		assert.Equal(t, 3, result.ActiveCount)  // Root, Level 1, Level 1 sibling

		// Verify structure
		collection, err := s.Load()
		testutil.AssertNoError(t, err)
		assert.Equal(t, 1, len(collection.Todos))
		root := collection.Todos[0]
		assert.Equal(t, "Root", root.Text)
		assert.Equal(t, 2, len(root.Items))
		assert.Equal(t, "Level 1", root.Items[0].Text)
		assert.Equal(t, 0, len(root.Items[0].Items)) // Level 2 and descendants removed
		assert.Equal(t, "Level 1 sibling", root.Items[1].Text)
	})

	t.Run("should handle multiple done branches", func(t *testing.T) {
		// Use the standard nested store
		s := testutil.CreateNestedStore(t)

		// Mark both top-level todos as done
		err := s.Update(func(collection *models.Collection) error {
			collection.Todos[0].Status = models.StatusDone
			collection.Todos[1].Status = models.StatusDone
			return nil
		})
		testutil.AssertNoError(t, err)

		// Execute clean
		result, err := clean.Execute(clean.Options{
			CollectionPath: s.Path(),
		})
		testutil.AssertNoError(t, err)

		// Should report only the 2 done parents
		assert.Equal(t, 2, result.RemovedCount) // 2 parents marked as done
		assert.Equal(t, 0, result.ActiveCount)

		// Verify empty collection
		collection, err := s.Load()
		testutil.AssertNoError(t, err)
		assert.Equal(t, 0, len(collection.Todos))
	})

	t.Run("should preserve exact structure of removed items", func(t *testing.T) {
		// Create a simple nested structure
		dir := testutil.TempDir(t)
		dbPath := dir + "/test.json"
		s := store.NewStore(dbPath)

		err := s.Update(func(collection *models.Collection) error {
			parent, _ := collection.CreateTodo("Done parent", "")
			parent.Status = models.StatusDone
			child1, _ := collection.CreateTodo("Child 1", parent.ID)
			_, _ = collection.CreateTodo("Grandchild", child1.ID)
			_, _ = collection.CreateTodo("Child 2", parent.ID)
			return nil
		})
		testutil.AssertNoError(t, err)

		// Execute clean
		result, err := clean.Execute(clean.Options{
			CollectionPath: s.Path(),
		})
		testutil.AssertNoError(t, err)

		// Find the removed parent in results
		var removedParent *models.Todo
		for _, todo := range result.RemovedTodos {
			if todo.Text == "Done parent" && todo.ParentID == "" {
				removedParent = todo
				break
			}
		}

		assert.NotNil(t, removedParent)
		assert.Equal(t, 2, len(removedParent.Items))
		assert.Equal(t, "Child 1", removedParent.Items[0].Text)
		assert.Equal(t, 1, len(removedParent.Items[0].Items))
		assert.Equal(t, "Grandchild", removedParent.Items[0].Items[0].Text)
		assert.Equal(t, "Child 2", removedParent.Items[1].Text)
	})
}
