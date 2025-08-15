package list_test

import (
	"testing"

	"github.com/arthur-debert/tdh/pkg/tdh/commands/list"
	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/arthur-debert/tdh/pkg/tdh/store"
	"github.com/arthur-debert/tdh/pkg/tdh/testutil"
	"github.com/stretchr/testify/assert"
)

func TestList_BehavioralPropagation(t *testing.T) {
	t.Run("should not show children of done parent", func(t *testing.T) {
		// Create a nested structure with a done parent
		dir := testutil.TempDir(t)
		dbPath := dir + "/test.json"
		s := store.NewStore(dbPath)

		err := s.Update(func(collection *models.Collection) error {
			// Create parent (done) with pending children
			parent, _ := collection.CreateTodo("Done parent", "")
			parent.Status = models.StatusDone

			_, _ = collection.CreateTodo("Pending child 1", parent.ID)
			_, _ = collection.CreateTodo("Pending child 2", parent.ID)

			// Create another pending parent with children
			parent2, _ := collection.CreateTodo("Pending parent", "")
			_, _ = collection.CreateTodo("Child of pending", parent2.ID)

			return nil
		})
		testutil.AssertNoError(t, err)

		// Execute list command (default shows pending only)
		result, err := list.Execute(list.Options{
			CollectionPath: s.Path(),
			ShowDone:       false,
			ShowAll:        false,
		})
		testutil.AssertNoError(t, err)

		// Should only show the pending parent and its child
		assert.Equal(t, 1, len(result.Todos))
		assert.Equal(t, "Pending parent", result.Todos[0].Text)
		assert.Equal(t, 1, len(result.Todos[0].Items))
		assert.Equal(t, "Child of pending", result.Todos[0].Items[0].Text)

		// Totals should count everything
		assert.Equal(t, 5, result.TotalCount)
		assert.Equal(t, 1, result.DoneCount)
	})

	t.Run("should show done parent without children when showing done", func(t *testing.T) {
		// Create a nested structure
		s := testutil.CreateNestedStore(t)

		// Mark parent as done
		err := s.Update(func(collection *models.Collection) error {
			parent := collection.Todos[0]
			parent.Status = models.StatusDone
			return nil
		})
		testutil.AssertNoError(t, err)

		// Execute list command showing done items
		result, err := list.Execute(list.Options{
			CollectionPath: s.Path(),
			ShowDone:       true,
			ShowAll:        false,
		})
		testutil.AssertNoError(t, err)

		// Should show only the done parent, not its children
		assert.Equal(t, 1, len(result.Todos))
		assert.Equal(t, "Parent todo", result.Todos[0].Text)
		assert.Equal(t, models.StatusDone, result.Todos[0].Status)
		assert.Equal(t, 0, len(result.Todos[0].Items)) // No children shown
	})

	t.Run("should show everything when ShowAll is true", func(t *testing.T) {
		// Create a nested structure with mixed statuses
		dir := testutil.TempDir(t)
		dbPath := dir + "/test.json"
		s := store.NewStore(dbPath)

		err := s.Update(func(collection *models.Collection) error {
			// Done parent with pending children
			parent1, _ := collection.CreateTodo("Done parent", "")
			parent1.Status = models.StatusDone
			child1, _ := collection.CreateTodo("Pending child", parent1.ID)
			_, _ = collection.CreateTodo("Grandchild", child1.ID)

			// Pending parent with done child
			parent2, _ := collection.CreateTodo("Pending parent", "")
			doneChild, _ := collection.CreateTodo("Done child", parent2.ID)
			doneChild.Status = models.StatusDone

			return nil
		})
		testutil.AssertNoError(t, err)

		// Execute with ShowAll
		result, err := list.Execute(list.Options{
			CollectionPath: s.Path(),
			ShowDone:       false,
			ShowAll:        true,
		})
		testutil.AssertNoError(t, err)

		// Should show everything, including children of done parents
		assert.Equal(t, 2, len(result.Todos))

		// First parent (done) with all its children
		assert.Equal(t, "Done parent", result.Todos[0].Text)
		assert.Equal(t, models.StatusDone, result.Todos[0].Status)
		assert.Equal(t, 1, len(result.Todos[0].Items))
		assert.Equal(t, "Pending child", result.Todos[0].Items[0].Text)
		assert.Equal(t, 1, len(result.Todos[0].Items[0].Items))
		assert.Equal(t, "Grandchild", result.Todos[0].Items[0].Items[0].Text)

		// Second parent with its done child
		assert.Equal(t, "Pending parent", result.Todos[1].Text)
		assert.Equal(t, 1, len(result.Todos[1].Items))
		assert.Equal(t, "Done child", result.Todos[1].Items[0].Text)
	})

	t.Run("should handle deeply nested done branches", func(t *testing.T) {
		dir := testutil.TempDir(t)
		dbPath := dir + "/test.json"
		s := store.NewStore(dbPath)

		err := s.Update(func(collection *models.Collection) error {
			// Create deep structure
			root, _ := collection.CreateTodo("Root", "")
			level1, _ := collection.CreateTodo("Level 1", root.ID)
			level2, _ := collection.CreateTodo("Level 2", level1.ID)
			level2.Status = models.StatusDone // Mark middle level as done
			_, _ = collection.CreateTodo("Level 3", level2.ID)
			_, _ = collection.CreateTodo("Level 4", level2.ID)

			return nil
		})
		testutil.AssertNoError(t, err)

		// Execute default list
		result, err := list.Execute(list.Options{
			CollectionPath: s.Path(),
		})
		testutil.AssertNoError(t, err)

		// Should show root, level 1, but not level 2's children
		assert.Equal(t, 1, len(result.Todos))
		root := result.Todos[0]
		assert.Equal(t, "Root", root.Text)
		assert.Equal(t, 1, len(root.Items))

		level1 := root.Items[0]
		assert.Equal(t, "Level 1", level1.Text)
		assert.Equal(t, 0, len(level1.Items)) // Level 2 is done, so not shown with children
	})
}
