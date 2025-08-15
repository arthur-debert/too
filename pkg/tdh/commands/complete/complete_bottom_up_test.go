package complete_test

import (
	"testing"

	"github.com/arthur-debert/tdh/pkg/tdh/commands/complete"
	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/arthur-debert/tdh/pkg/tdh/store"
	"github.com/arthur-debert/tdh/pkg/tdh/testutil"
	"github.com/stretchr/testify/assert"
)

func TestExecute_BottomUpCompletion(t *testing.T) {
	t.Run("should complete parent when all children are completed", func(t *testing.T) {
		// Create a store with parent and two children
		s := testutil.CreateNestedStore(t)

		// Load collection to get the structure
		collection, err := s.Load()
		testutil.AssertNoError(t, err)

		// Check collection structure
		assert.NotNil(t, collection)

		// Find the parent with two children (position 1)
		parent := collection.Todos[0]
		assert.Equal(t, "Parent todo", parent.Text)
		assert.Equal(t, 2, len(parent.Items))

		// Complete the first child (1.1)
		result, err := complete.Execute("1.1", complete.Options{
			CollectionPath: s.Path(),
		})
		testutil.AssertNoError(t, err)
		assert.Equal(t, "Sub-task 1.1", result.Todo.Text)
		assert.Equal(t, string(models.StatusDone), result.NewStatus)

		// Verify parent is still pending (not all children complete)
		collection, err = s.Load()
		testutil.AssertNoError(t, err)
		parent = collection.Todos[0]
		assert.Equal(t, models.StatusPending, parent.Status)

		// Complete the second child (1.2)
		result, err = complete.Execute("1.2", complete.Options{
			CollectionPath: s.Path(),
		})
		testutil.AssertNoError(t, err)
		assert.Equal(t, "Sub-task 1.2", result.Todo.Text)

		// Now verify parent is automatically completed
		collection, err = s.Load()
		testutil.AssertNoError(t, err)
		parent = collection.Todos[0]
		assert.Equal(t, models.StatusDone, parent.Status)
	})

	t.Run("should not complete parent if it has no children", func(t *testing.T) {
		// Create a simple store
		s := testutil.CreatePopulatedStore(t, "Todo without children")

		// Complete the todo
		result, err := complete.Execute("1", complete.Options{
			CollectionPath: s.Path(),
		})
		testutil.AssertNoError(t, err)
		assert.Equal(t, "Todo without children", result.Todo.Text)
		assert.Equal(t, string(models.StatusDone), result.NewStatus)

		// Just verify it completes normally
		collection, err := s.Load()
		testutil.AssertNoError(t, err)
		assert.Equal(t, models.StatusDone, collection.Todos[0].Status)
	})

	t.Run("should handle multi-level bottom-up completion", func(t *testing.T) {
		// Create a nested store with grandchildren
		s := testutil.CreateNestedStore(t)

		// We have:
		// 1. Parent todo
		//    1.1 Sub-task 1.1
		//    1.2 Sub-task 1.2
		//        1.2.1 Grandchild 1.2.1

		// Complete sub-task 1.1
		_, err := complete.Execute("1.1", complete.Options{
			CollectionPath: s.Path(),
		})
		testutil.AssertNoError(t, err)

		// Complete grandchild 1.2.1
		_, err = complete.Execute("1.2.1", complete.Options{
			CollectionPath: s.Path(),
		})
		testutil.AssertNoError(t, err)

		// Verify sub-task 1.2 was auto-completed
		collection, err := s.Load()
		testutil.AssertNoError(t, err)
		parent := collection.Todos[0]
		subTask2 := parent.Items[1]
		assert.Equal(t, "Sub-task 1.2", subTask2.Text)
		assert.Equal(t, models.StatusDone, subTask2.Status)

		// And verify parent was also auto-completed
		assert.Equal(t, models.StatusDone, parent.Status)
	})

	t.Run("should not complete parent if some children are still pending", func(t *testing.T) {
		// Create a store with custom nested structure
		dir := testutil.TempDir(t)
		dbPath := dir + "/test.json"
		s := store.NewStore(dbPath)

		// Create parent with three children
		err := s.Update(func(collection *models.Collection) error {
			parent, _ := collection.CreateTodo("Parent with three children", "")
			_, _ = collection.CreateTodo("Child 1", parent.ID)
			_, _ = collection.CreateTodo("Child 2", parent.ID)
			_, _ = collection.CreateTodo("Child 3", parent.ID)
			return nil
		})
		testutil.AssertNoError(t, err)

		// Complete first two children
		_, err = complete.Execute("1.1", complete.Options{
			CollectionPath: s.Path(),
		})
		testutil.AssertNoError(t, err)

		_, err = complete.Execute("1.2", complete.Options{
			CollectionPath: s.Path(),
		})
		testutil.AssertNoError(t, err)

		// Verify parent is still pending
		collection, err := s.Load()
		testutil.AssertNoError(t, err)
		parent := collection.Todos[0]
		assert.Equal(t, models.StatusPending, parent.Status)

		// Complete the third child
		_, err = complete.Execute("1.3", complete.Options{
			CollectionPath: s.Path(),
		})
		testutil.AssertNoError(t, err)

		// Now parent should be complete
		collection, err = s.Load()
		testutil.AssertNoError(t, err)
		parent = collection.Todos[0]
		assert.Equal(t, models.StatusDone, parent.Status)
	})

	t.Run("should handle complex nested hierarchy", func(t *testing.T) {
		// Create a complex hierarchy
		dir := testutil.TempDir(t)
		dbPath := dir + "/test.json"
		s := store.NewStore(dbPath)

		err := s.Update(func(collection *models.Collection) error {
			// Create structure:
			// 1. Project A
			//    1.1 Phase 1
			//        1.1.1 Task A
			//        1.1.2 Task B
			//    1.2 Phase 2
			//        1.2.1 Task C
			project, _ := collection.CreateTodo("Project A", "")
			phase1, _ := collection.CreateTodo("Phase 1", project.ID)
			_, _ = collection.CreateTodo("Task A", phase1.ID)
			_, _ = collection.CreateTodo("Task B", phase1.ID)
			phase2, _ := collection.CreateTodo("Phase 2", project.ID)
			_, _ = collection.CreateTodo("Task C", phase2.ID)
			return nil
		})
		testutil.AssertNoError(t, err)

		// Complete all tasks
		_, err = complete.Execute("1.1.1", complete.Options{CollectionPath: s.Path()})
		testutil.AssertNoError(t, err)

		_, err = complete.Execute("1.1.2", complete.Options{CollectionPath: s.Path()})
		testutil.AssertNoError(t, err)

		// Phase 1 should be auto-completed
		collection, err := s.Load()
		testutil.AssertNoError(t, err)
		project := collection.Todos[0]
		phase1 := project.Items[0]
		assert.Equal(t, models.StatusDone, phase1.Status)

		// But project should still be pending (Phase 2 not complete)
		assert.Equal(t, models.StatusPending, project.Status)

		// Complete Task C
		_, err = complete.Execute("1.2.1", complete.Options{CollectionPath: s.Path()})
		testutil.AssertNoError(t, err)

		// Now everything should be complete
		collection, err = s.Load()
		testutil.AssertNoError(t, err)
		project = collection.Todos[0]
		assert.Equal(t, models.StatusDone, project.Status)
		assert.Equal(t, models.StatusDone, project.Items[0].Status) // Phase 1
		assert.Equal(t, models.StatusDone, project.Items[1].Status) // Phase 2
	})

	t.Run("should not auto-complete childless parent when sibling completes", func(t *testing.T) {
		// This test verifies the business rule that childless parents are not auto-completed
		dir := testutil.TempDir(t)
		dbPath := dir + "/test.json"
		s := store.NewStore(dbPath)

		err := s.Update(func(collection *models.Collection) error {
			// Create a parent with two children
			parent, _ := collection.CreateTodo("Parent", "")
			_, _ = collection.CreateTodo("Child with grandchildren", parent.ID)
			childless, _ := collection.CreateTodo("Childless child", parent.ID)

			// Give the first child some grandchildren
			firstChild := parent.Items[0]
			_, _ = collection.CreateTodo("Grandchild 1", firstChild.ID)
			_, _ = collection.CreateTodo("Grandchild 2", firstChild.ID)

			// Verify childless has no children
			assert.Equal(t, 0, len(childless.Items))
			return nil
		})
		testutil.AssertNoError(t, err)

		// Complete the childless child
		_, err = complete.Execute("1.2", complete.Options{
			CollectionPath: s.Path(),
		})
		testutil.AssertNoError(t, err)

		// Parent should still be pending (other child not complete)
		collection, err := s.Load()
		testutil.AssertNoError(t, err)
		parent := collection.Todos[0]
		assert.Equal(t, models.StatusPending, parent.Status)

		// Complete grandchildren
		_, err = complete.Execute("1.1.1", complete.Options{
			CollectionPath: s.Path(),
		})
		testutil.AssertNoError(t, err)

		_, err = complete.Execute("1.1.2", complete.Options{
			CollectionPath: s.Path(),
		})
		testutil.AssertNoError(t, err)

		// Now parent should be complete (all children complete)
		collection, err = s.Load()
		testutil.AssertNoError(t, err)
		parent = collection.Todos[0]
		assert.Equal(t, models.StatusDone, parent.Status)

		// Verify the childless child is still childless and complete
		childless := parent.Items[1]
		assert.Equal(t, "Childless child", childless.Text)
		assert.Equal(t, 0, len(childless.Items))
		assert.Equal(t, models.StatusDone, childless.Status)
	})

	t.Run("should handle root level items without panic", func(t *testing.T) {
		// This test ensures completing root items (with no parent) doesn't cause issues
		dir := testutil.TempDir(t)
		dbPath := dir + "/test.json"
		s := store.NewStore(dbPath)

		// Create a mix of root level items and nested items
		err := s.Update(func(collection *models.Collection) error {
			// Root level todos
			_, _ = collection.CreateTodo("Root todo 1", "")
			rootWithChildren, _ := collection.CreateTodo("Root with children", "")
			_, _ = collection.CreateTodo("Child 1", rootWithChildren.ID)
			_, _ = collection.CreateTodo("Child 2", rootWithChildren.ID)
			_, _ = collection.CreateTodo("Root todo 3", "")
			return nil
		})
		testutil.AssertNoError(t, err)

		// Complete a root level item without children - should work fine
		result, err := complete.Execute("1", complete.Options{
			CollectionPath: s.Path(),
		})
		testutil.AssertNoError(t, err)
		assert.Equal(t, "Root todo 1", result.Todo.Text)
		assert.Equal(t, "", result.Todo.ParentID) // Verify it has no parent

		// Complete children to trigger bottom-up on a root item
		_, err = complete.Execute("2.1", complete.Options{
			CollectionPath: s.Path(),
		})
		testutil.AssertNoError(t, err)

		_, err = complete.Execute("2.2", complete.Options{
			CollectionPath: s.Path(),
		})
		testutil.AssertNoError(t, err)

		// Verify root item with children was auto-completed
		collection, err := s.Load()
		testutil.AssertNoError(t, err)
		rootWithChildren := collection.Todos[1]
		assert.Equal(t, "Root with children", rootWithChildren.Text)
		assert.Equal(t, models.StatusDone, rootWithChildren.Status)
		assert.Equal(t, "", rootWithChildren.ParentID) // Verify it's still at root
	})
}
