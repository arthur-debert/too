package list_test

import (
	"testing"

	"github.com/arthur-debert/tdh/pkg/tdh/commands/list"
	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/arthur-debert/tdh/pkg/tdh/testutil"
	"github.com/stretchr/testify/assert"
)

func TestListCommand(t *testing.T) {
	t.Run("lists only pending todos by default", func(t *testing.T) {
		// Create store with mixed todo states using testutil
		store := testutil.CreateStoreWithSpecs(t, []testutil.TodoSpec{
			{Text: "Pending task 1", Status: models.StatusPending},
			{Text: "Done task 1", Status: models.StatusDone},
			{Text: "Pending task 2", Status: models.StatusPending},
			{Text: "Done task 2", Status: models.StatusDone},
		})

		// Execute list command with default options
		opts := list.Options{
			CollectionPath: store.Path(),
			ShowDone:       false,
			ShowAll:        false,
		}
		result, err := list.Execute(opts)

		// Verify using testutil assertions
		testutil.AssertNoError(t, err)
		assert.Equal(t, 2, len(result.Todos))
		assert.Equal(t, 4, result.TotalCount)
		assert.Equal(t, 2, result.DoneCount)

		// Verify only pending todos are returned
		for _, todo := range result.Todos {
			testutil.AssertTodoHasStatus(t, todo, models.StatusPending)
		}
		testutil.AssertTodoInList(t, result.Todos, "Pending task 1")
		testutil.AssertTodoInList(t, result.Todos, "Pending task 2")
	})

	t.Run("lists only done todos with ShowDone flag", func(t *testing.T) {
		// Create store with mixed todo states
		store := testutil.CreateStoreWithSpecs(t, []testutil.TodoSpec{
			{Text: "Pending task", Status: models.StatusPending},
			{Text: "Done task 1", Status: models.StatusDone},
			{Text: "Done task 2", Status: models.StatusDone},
		})

		// Execute list command with ShowDone flag
		opts := list.Options{
			CollectionPath: store.Path(),
			ShowDone:       true,
			ShowAll:        false,
		}
		result, err := list.Execute(opts)

		// Verify results
		testutil.AssertNoError(t, err)
		assert.Equal(t, 2, len(result.Todos))
		assert.Equal(t, 3, result.TotalCount)
		assert.Equal(t, 2, result.DoneCount)

		// Verify only done todos are returned
		for _, todo := range result.Todos {
			testutil.AssertTodoHasStatus(t, todo, models.StatusDone)
		}
		testutil.AssertTodoInList(t, result.Todos, "Done task 1")
		testutil.AssertTodoInList(t, result.Todos, "Done task 2")
	})

	t.Run("lists all todos with ShowAll flag", func(t *testing.T) {
		// Create store with mixed todo states
		store := testutil.CreateStoreWithSpecs(t, []testutil.TodoSpec{
			{Text: "Pending task 1", Status: models.StatusPending},
			{Text: "Done task 1", Status: models.StatusDone},
			{Text: "Pending task 2", Status: models.StatusPending},
			{Text: "Done task 2", Status: models.StatusDone},
		})

		// Execute list command with ShowAll flag
		opts := list.Options{
			CollectionPath: store.Path(),
			ShowDone:       false, // This should be ignored when ShowAll is true
			ShowAll:        true,
		}
		result, err := list.Execute(opts)

		// Verify all todos are returned
		testutil.AssertNoError(t, err)
		assert.Equal(t, 4, len(result.Todos))
		assert.Equal(t, 4, result.TotalCount)
		assert.Equal(t, 2, result.DoneCount)

		// Verify all todos are present
		testutil.AssertTodoInList(t, result.Todos, "Pending task 1")
		testutil.AssertTodoInList(t, result.Todos, "Done task 1")
		testutil.AssertTodoInList(t, result.Todos, "Pending task 2")
		testutil.AssertTodoInList(t, result.Todos, "Done task 2")
	})

	t.Run("handles empty collection gracefully", func(t *testing.T) {
		// Create empty store
		store := testutil.CreatePopulatedStore(t) // Empty store

		// Execute list command
		opts := list.Options{
			CollectionPath: store.Path(),
			ShowDone:       false,
			ShowAll:        false,
		}
		result, err := list.Execute(opts)

		// Verify empty results
		testutil.AssertNoError(t, err)
		assert.Empty(t, result.Todos)
		assert.Equal(t, 0, result.TotalCount)
		assert.Equal(t, 0, result.DoneCount)
	})

	t.Run("returns counts even when filtering", func(t *testing.T) {
		// Create store with specific todo distribution
		store := testutil.CreateStoreWithSpecs(t, []testutil.TodoSpec{
			{Text: "Pending 1", Status: models.StatusPending},
			{Text: "Pending 2", Status: models.StatusPending},
			{Text: "Pending 3", Status: models.StatusPending},
			{Text: "Done 1", Status: models.StatusDone},
			{Text: "Done 2", Status: models.StatusDone},
		})

		// Test with pending filter
		opts := list.Options{
			CollectionPath: store.Path(),
			ShowDone:       false,
			ShowAll:        false,
		}
		result, err := list.Execute(opts)

		testutil.AssertNoError(t, err)
		assert.Equal(t, 3, len(result.Todos)) // Only pending todos returned
		assert.Equal(t, 5, result.TotalCount) // But total count includes all
		assert.Equal(t, 2, result.DoneCount)  // And done count is accurate
	})

	t.Run("lists nested todos with hierarchy preserved", func(t *testing.T) {
		// Create store with nested structure
		store := testutil.CreatePopulatedStore(t) // Empty store

		// Build nested structure
		err := store.Update(func(collection *models.Collection) error {
			// Create parent todos
			parent1, _ := collection.CreateTodo("Parent 1", "")
			parent2, _ := collection.CreateTodo("Parent 2", "")

			// Add children to parent1
			child11, _ := collection.CreateTodo("Child 1.1", parent1.ID)
			child12, _ := collection.CreateTodo("Child 1.2", parent1.ID)

			// Add grandchild
			_, _ = collection.CreateTodo("Grandchild 1.1.1", child11.ID)

			// Add child to parent2
			_, _ = collection.CreateTodo("Child 2.1", parent2.ID)

			// Mark some as done
			child12.Toggle()
			parent2.Toggle()

			return nil
		})
		testutil.AssertNoError(t, err)

		// Test ShowAll - should return full hierarchy
		opts := list.Options{
			CollectionPath: store.Path(),
			ShowAll:        true,
		}
		result, err := list.Execute(opts)

		testutil.AssertNoError(t, err)
		assert.Equal(t, 2, len(result.Todos)) // Two top-level todos
		assert.Equal(t, 6, result.TotalCount) // Total todos in tree
		assert.Equal(t, 2, result.DoneCount)  // Two done todos

		// Verify hierarchy
		parent1 := result.Todos[0]
		assert.Equal(t, "Parent 1", parent1.Text)
		assert.Len(t, parent1.Items, 2)

		child11 := parent1.Items[0]
		assert.Equal(t, "Child 1.1", child11.Text)
		assert.Len(t, child11.Items, 1)
		assert.Equal(t, "Grandchild 1.1.1", child11.Items[0].Text)

		parent2 := result.Todos[1]
		assert.Equal(t, "Parent 2", parent2.Text)
		assert.Equal(t, models.StatusDone, parent2.Status)
		assert.Len(t, parent2.Items, 1)
	})

	t.Run("filters nested todos while preserving hierarchy", func(t *testing.T) {
		// Create store with nested structure
		store := testutil.CreatePopulatedStore(t) // Empty store

		// Build nested structure with mixed statuses
		err := store.Update(func(collection *models.Collection) error {
			// Create done parent with pending child
			parent1, _ := collection.CreateTodo("Done Parent", "")
			parent1.Toggle() // Mark as done
			_, _ = collection.CreateTodo("Pending Child", parent1.ID)

			// Create pending parent with done child
			parent2, _ := collection.CreateTodo("Pending Parent", "")
			child2, _ := collection.CreateTodo("Done Child", parent2.ID)
			child2.Toggle() // Mark as done

			// Create all-done branch
			parent3, _ := collection.CreateTodo("All Done Parent", "")
			parent3.Toggle()
			child3, _ := collection.CreateTodo("All Done Child", parent3.ID)
			child3.Toggle()

			// Create all-pending branch
			parent4, _ := collection.CreateTodo("All Pending Parent", "")
			_, _ = collection.CreateTodo("All Pending Child", parent4.ID)

			return nil
		})
		testutil.AssertNoError(t, err)

		// Test ShowDone=false (default) - shows pending todos and parents with pending children
		opts := list.Options{
			CollectionPath: store.Path(),
			ShowDone:       false,
			ShowAll:        false,
		}
		result, err := list.Execute(opts)

		testutil.AssertNoError(t, err)
		assert.Equal(t, 3, len(result.Todos)) // Three branches have pending todos
		assert.Equal(t, 8, result.TotalCount) // Total todos in tree
		assert.Equal(t, 4, result.DoneCount)  // Four done todos

		// Verify correct branches are included
		var texts []string
		for _, todo := range result.Todos {
			texts = append(texts, todo.Text)
		}
		assert.Contains(t, texts, "Done Parent")        // Has pending child
		assert.Contains(t, texts, "Pending Parent")     // Is pending
		assert.Contains(t, texts, "All Pending Parent") // All pending
		assert.NotContains(t, texts, "All Done Parent") // No pending in branch

		// Verify filtered children
		for _, todo := range result.Todos {
			if todo.Text == "Done Parent" {
				assert.Len(t, todo.Items, 1)
				assert.Equal(t, "Pending Child", todo.Items[0].Text)
			}
			if todo.Text == "Pending Parent" {
				assert.Len(t, todo.Items, 0) // Done child filtered out
			}
		}
	})

	t.Run("counts nested todos correctly", func(t *testing.T) {
		// Create deeply nested structure
		store := testutil.CreatePopulatedStore(t)

		err := store.Update(func(collection *models.Collection) error {
			// Create a 5-level deep structure
			l1, _ := collection.CreateTodo("Level 1", "")
			l2, _ := collection.CreateTodo("Level 2", l1.ID)
			l3, _ := collection.CreateTodo("Level 3", l2.ID)
			l4, _ := collection.CreateTodo("Level 4", l3.ID)
			l5, _ := collection.CreateTodo("Level 5", l4.ID)

			// Mark alternating levels as done
			l1.Toggle()
			l3.Toggle()
			l5.Toggle()

			// Add a separate branch
			b1, _ := collection.CreateTodo("Branch 1", "")
			_, _ = collection.CreateTodo("Branch 1.1", b1.ID)

			return nil
		})
		testutil.AssertNoError(t, err)

		opts := list.Options{
			CollectionPath: store.Path(),
			ShowAll:        true,
		}
		result, err := list.Execute(opts)

		testutil.AssertNoError(t, err)
		assert.Equal(t, 2, len(result.Todos)) // Two top-level todos
		assert.Equal(t, 7, result.TotalCount) // Seven total todos
		assert.Equal(t, 3, result.DoneCount)  // Three marked as done
	})

	t.Run("handles empty nested structure", func(t *testing.T) {
		// Create store with parent that has no children
		store := testutil.CreatePopulatedStore(t, "Parent with no children")

		opts := list.Options{
			CollectionPath: store.Path(),
			ShowAll:        true,
		}
		result, err := list.Execute(opts)

		testutil.AssertNoError(t, err)
		assert.Equal(t, 1, len(result.Todos))
		assert.Equal(t, 1, result.TotalCount)
		assert.Equal(t, 0, result.DoneCount)

		// Verify empty Items array
		assert.NotNil(t, result.Todos[0].Items)
		assert.Empty(t, result.Todos[0].Items)
	})
}
