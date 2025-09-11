package clean_test

import (
	"fmt"
	"testing"

	"github.com/arthur-debert/too/pkg/too/commands/clean"
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/store"
	"github.com/arthur-debert/too/pkg/too/testutil"
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
			parent1.Statuses = map[string]string{"completion": string(models.StatusDone)}
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
		assert.Equal(t, models.StatusDone, result.RemovedTodos[0].GetStatus())

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
			child1.Statuses = map[string]string{"completion": string(models.StatusDone)}

			_, _ = collection.CreateTodo("Pending child", parent.ID)

			child3, _ := collection.CreateTodo("Done child 2", parent.ID)
			child3.Statuses = map[string]string{"completion": string(models.StatusDone)}

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
			level2.Statuses = map[string]string{"completion": string(models.StatusDone)}
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
			collection.Todos[0].Statuses = map[string]string{"completion": string(models.StatusDone)}
			collection.Todos[1].Statuses = map[string]string{"completion": string(models.StatusDone)}
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
			parent.Statuses = map[string]string{"completion": string(models.StatusDone)}
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

	t.Run("should only report done items not pending descendants", func(t *testing.T) {
		dir := testutil.TempDir(t)
		dbPath := dir + "/test.json"
		s := store.NewStore(dbPath)

		err := s.Update(func(collection *models.Collection) error {
			// Done parent with mix of done and pending children
			parent, _ := collection.CreateTodo("Done parent", "")
			parent.Statuses = map[string]string{"completion": string(models.StatusDone)}

			pending1, _ := collection.CreateTodo("Pending child 1", parent.ID)
			_, _ = collection.CreateTodo("Pending grandchild", pending1.ID)

			done1, _ := collection.CreateTodo("Done child", parent.ID)
			done1.Statuses = map[string]string{"completion": string(models.StatusDone)}

			_, _ = collection.CreateTodo("Pending child 2", parent.ID)

			return nil
		})
		testutil.AssertNoError(t, err)

		// Execute clean
		result, err := clean.Execute(clean.Options{
			CollectionPath: s.Path(),
		})
		testutil.AssertNoError(t, err)

		// Should report only the 2 done items, not the 3 pending descendants
		assert.Equal(t, 2, result.RemovedCount)

		// Verify reported items are only done ones
		for _, todo := range result.RemovedTodos {
			assert.Equal(t, models.StatusDone, todo.GetStatus())
		}

		// Verify all items were actually removed
		collection, err := s.Load()
		testutil.AssertNoError(t, err)
		assert.Equal(t, 0, len(collection.Todos))
	})

	t.Run("should handle complex mixed hierarchy", func(t *testing.T) {
		dir := testutil.TempDir(t)
		dbPath := dir + "/test.json"
		s := store.NewStore(dbPath)

		err := s.Update(func(collection *models.Collection) error {
			// Pending root with done children
			root1, _ := collection.CreateTodo("Pending root 1", "")
			done1, _ := collection.CreateTodo("Done child 1", root1.ID)
			done1.Statuses = map[string]string{"completion": string(models.StatusDone)}
			_, _ = collection.CreateTodo("Pending grandchild 1", done1.ID)

			pending1, _ := collection.CreateTodo("Pending child 2", root1.ID)
			done2, _ := collection.CreateTodo("Done grandchild", pending1.ID)
			done2.Statuses = map[string]string{"completion": string(models.StatusDone)}

			// Done root with pending children
			root2, _ := collection.CreateTodo("Done root", "")
			root2.Statuses = map[string]string{"completion": string(models.StatusDone)}
			_, _ = collection.CreateTodo("Pending child under done", root2.ID)

			// Another pending root
			_, _ = collection.CreateTodo("Pending root 2", "")

			return nil
		})
		testutil.AssertNoError(t, err)

		// Execute clean
		result, err := clean.Execute(clean.Options{
			CollectionPath: s.Path(),
		})
		testutil.AssertNoError(t, err)

		// Should report 3 done items
		assert.Equal(t, 3, result.RemovedCount)

		// Should have 3 active items left
		assert.Equal(t, 3, result.ActiveCount)

		// Verify structure
		collection, err := s.Load()
		testutil.AssertNoError(t, err)
		assert.Equal(t, 2, len(collection.Todos))

		// First root should have one pending child left
		assert.Equal(t, "Pending root 1", collection.Todos[0].Text)
		assert.Equal(t, 1, len(collection.Todos[0].Items))
		assert.Equal(t, "Pending child 2", collection.Todos[0].Items[0].Text)
		assert.Equal(t, 0, len(collection.Todos[0].Items[0].Items)) // Grandchild removed

		// Second root unchanged
		assert.Equal(t, "Pending root 2", collection.Todos[1].Text)
	})

	t.Run("should handle all done children under pending parent", func(t *testing.T) {
		dir := testutil.TempDir(t)
		dbPath := dir + "/test.json"
		s := store.NewStore(dbPath)

		err := s.Update(func(collection *models.Collection) error {
			parent, _ := collection.CreateTodo("Pending parent", "")

			// All children are done
			for i := 1; i <= 3; i++ {
				child, _ := collection.CreateTodo(fmt.Sprintf("Done child %d", i), parent.ID)
				child.Statuses = map[string]string{"completion": string(models.StatusDone)}

				// Add some grandchildren
				for j := 1; j <= 2; j++ {
					_, _ = collection.CreateTodo(fmt.Sprintf("Grandchild %d.%d", i, j), child.ID)
				}
			}

			return nil
		})
		testutil.AssertNoError(t, err)

		// Execute clean
		result, err := clean.Execute(clean.Options{
			CollectionPath: s.Path(),
		})
		testutil.AssertNoError(t, err)

		// Should report only the 3 done children
		assert.Equal(t, 3, result.RemovedCount)
		assert.Equal(t, 1, result.ActiveCount) // Only parent remains

		// Verify only parent remains with no children
		collection, err := s.Load()
		testutil.AssertNoError(t, err)
		assert.Equal(t, 1, len(collection.Todos))
		assert.Equal(t, "Pending parent", collection.Todos[0].Text)
		assert.Equal(t, 0, len(collection.Todos[0].Items))
	})

	t.Run("edge case: empty collection", func(t *testing.T) {
		dir := testutil.TempDir(t)
		dbPath := dir + "/test.json"
		s := store.NewStore(dbPath)

		// Initialize empty collection
		err := s.Update(func(collection *models.Collection) error {
			return nil
		})
		testutil.AssertNoError(t, err)

		// Execute clean on empty collection
		result, err := clean.Execute(clean.Options{
			CollectionPath: s.Path(),
		})
		testutil.AssertNoError(t, err)

		assert.Equal(t, 0, result.RemovedCount)
		assert.Equal(t, 0, result.ActiveCount)
		assert.Equal(t, 0, len(result.RemovedTodos))
	})

	t.Run("edge case: no done items", func(t *testing.T) {
		dir := testutil.TempDir(t)
		dbPath := dir + "/test.json"
		s := store.NewStore(dbPath)

		err := s.Update(func(collection *models.Collection) error {
			parent, _ := collection.CreateTodo("Pending parent", "")
			_, _ = collection.CreateTodo("Pending child 1", parent.ID)
			_, _ = collection.CreateTodo("Pending child 2", parent.ID)
			return nil
		})
		testutil.AssertNoError(t, err)

		// Execute clean
		result, err := clean.Execute(clean.Options{
			CollectionPath: s.Path(),
		})
		testutil.AssertNoError(t, err)

		assert.Equal(t, 0, result.RemovedCount)
		assert.Equal(t, 3, result.ActiveCount)

		// Verify nothing changed
		collection, err := s.Load()
		testutil.AssertNoError(t, err)
		assert.Equal(t, 1, len(collection.Todos))
		assert.Equal(t, 2, len(collection.Todos[0].Items))
	})
}
