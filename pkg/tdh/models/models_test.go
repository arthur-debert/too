package models_test

import (
	"testing"
	"time"

	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCollection(t *testing.T) {
	t.Run("creates empty collection", func(t *testing.T) {
		collection := models.NewCollection()

		assert.NotNil(t, collection)
		assert.NotNil(t, collection.Todos)
		assert.Empty(t, collection.Todos)
	})
}

func TestCollection_CreateTodo(t *testing.T) {
	t.Run("creates todo with correct attributes", func(t *testing.T) {
		collection := models.NewCollection()
		beforeCreate := time.Now()

		todo, err := collection.CreateTodo("Test todo", "")
		afterCreate := time.Now()

		require.NoError(t, err)
		assert.NotNil(t, todo)
		assert.Equal(t, 1, todo.Position)
		assert.Equal(t, "Test todo", todo.Text)
		assert.Equal(t, models.StatusPending, todo.Status)
		assert.Empty(t, todo.ParentID)
		assert.NotNil(t, todo.Items)
		assert.True(t, todo.Modified.After(beforeCreate) || todo.Modified.Equal(beforeCreate))
		assert.True(t, todo.Modified.Before(afterCreate) || todo.Modified.Equal(afterCreate))
	})

	t.Run("adds todo to collection", func(t *testing.T) {
		collection := models.NewCollection()

		todo, err := collection.CreateTodo("Test todo", "")

		require.NoError(t, err)
		assert.Len(t, collection.Todos, 1)
		assert.Equal(t, todo, collection.Todos[0])
	})

	t.Run("assigns incremental IDs", func(t *testing.T) {
		collection := models.NewCollection()

		todo1, err := collection.CreateTodo("First", "")
		require.NoError(t, err)
		todo2, err := collection.CreateTodo("Second", "")
		require.NoError(t, err)
		todo3, err := collection.CreateTodo("Third", "")
		require.NoError(t, err)

		assert.NotEmpty(t, todo1.ID) // Should have UUID
		assert.NotEmpty(t, todo2.ID)
		assert.NotEmpty(t, todo3.ID)
		assert.NotEqual(t, todo1.ID, todo2.ID) // UUIDs should be unique
		assert.NotEqual(t, todo2.ID, todo3.ID)
		assert.Equal(t, 1, todo1.Position)
		assert.Equal(t, 2, todo2.Position)
		assert.Equal(t, 3, todo3.Position)
	})

	t.Run("handles gaps in IDs correctly", func(t *testing.T) {
		collection := models.NewCollection()

		// Manually add todos with non-sequential positions
		collection.Todos = []*models.Todo{
			{ID: "id-1", Position: 1, Text: "First", Status: models.StatusPending},
			{ID: "id-2", Position: 5, Text: "Fifth", Status: models.StatusPending},
			{ID: "id-3", Position: 3, Text: "Third", Status: models.StatusPending},
		}

		// Create new todo - should get Position 6 (highest + 1)
		newTodo, err := collection.CreateTodo("New todo", "")
		require.NoError(t, err)

		assert.Equal(t, 6, newTodo.Position)
	})

	t.Run("handles empty text", func(t *testing.T) {
		collection := models.NewCollection()

		todo, err := collection.CreateTodo("", "")
		require.NoError(t, err)

		assert.Equal(t, "", todo.Text)
		assert.Equal(t, 1, todo.Position)
	})

	t.Run("creates multiple todos with different timestamps", func(t *testing.T) {
		collection := models.NewCollection()

		todo1, err := collection.CreateTodo("First", "")
		require.NoError(t, err)
		time.Sleep(2 * time.Millisecond) // Small delay to ensure different timestamps
		todo2, err := collection.CreateTodo("Second", "")
		require.NoError(t, err)

		assert.NotEqual(t, todo1.Modified, todo2.Modified)
		assert.True(t, todo2.Modified.After(todo1.Modified))
	})

	t.Run("creates nested todo with parent", func(t *testing.T) {
		collection := models.NewCollection()

		// Create parent
		parent, err := collection.CreateTodo("Parent todo", "")
		require.NoError(t, err)

		// Create child
		child, err := collection.CreateTodo("Child todo", parent.ID)
		require.NoError(t, err)

		assert.NotNil(t, child)
		assert.Equal(t, parent.ID, child.ParentID)
		assert.Equal(t, 1, child.Position)
		assert.Equal(t, "Child todo", child.Text)
		assert.Len(t, parent.Items, 1)
		assert.Equal(t, child, parent.Items[0])
	})

	t.Run("creates multiple nested todos", func(t *testing.T) {
		collection := models.NewCollection()

		// Create parent
		parent, err := collection.CreateTodo("Parent", "")
		require.NoError(t, err)

		// Create children
		child1, err := collection.CreateTodo("Child 1", parent.ID)
		require.NoError(t, err)
		child2, err := collection.CreateTodo("Child 2", parent.ID)
		require.NoError(t, err)
		child3, err := collection.CreateTodo("Child 3", parent.ID)
		require.NoError(t, err)

		assert.Len(t, parent.Items, 3)
		assert.Equal(t, 1, child1.Position)
		assert.Equal(t, 2, child2.Position)
		assert.Equal(t, 3, child3.Position)
	})

	t.Run("returns error for non-existent parent", func(t *testing.T) {
		collection := models.NewCollection()

		_, err := collection.CreateTodo("Orphan todo", "non-existent-id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parent todo with ID non-existent-id not found")
	})
}

func TestTodo_Clone(t *testing.T) {
	t.Run("creates exact copy of todo", func(t *testing.T) {
		original := &models.Todo{
			ID:       "test-id-42",
			ParentID: "parent-id-1",
			Position: 42,
			Text:     "Original todo",
			Status:   models.StatusDone,
			Modified: time.Now().Add(-time.Hour),
			Items:    []*models.Todo{},
		}

		clone := original.Clone()

		assert.Equal(t, original.ID, clone.ID)
		assert.Equal(t, original.ParentID, clone.ParentID)
		assert.Equal(t, original.Text, clone.Text)
		assert.Equal(t, original.Status, clone.Status)
		assert.Equal(t, original.Modified, clone.Modified)
		assert.NotNil(t, clone.Items)
		assert.Empty(t, clone.Items)
	})

	t.Run("clone is independent of original", func(t *testing.T) {
		original := &models.Todo{
			ID:       "test-id-1",
			ParentID: "",
			Position: 1,
			Text:     "Original",
			Status:   models.StatusPending,
			Modified: time.Now(),
			Items:    []*models.Todo{},
		}

		clone := original.Clone()

		// Modify clone
		clone.Text = "Modified"
		clone.Status = models.StatusDone
		clone.ID = "test-id-99"
		clone.ParentID = "new-parent"

		// Original should be unchanged
		assert.Equal(t, "Original", original.Text)
		assert.Equal(t, models.StatusPending, original.Status)
		assert.Equal(t, "test-id-1", original.ID)
		assert.Equal(t, "", original.ParentID)
		assert.Equal(t, 1, original.Position)
	})

	t.Run("deep clones nested items", func(t *testing.T) {
		child1 := &models.Todo{
			ID:       "child-1",
			ParentID: "parent-1",
			Position: 1,
			Text:     "Child 1",
			Status:   models.StatusPending,
			Modified: time.Now(),
			Items:    []*models.Todo{},
		}

		child2 := &models.Todo{
			ID:       "child-2",
			ParentID: "parent-1",
			Position: 2,
			Text:     "Child 2",
			Status:   models.StatusDone,
			Modified: time.Now(),
			Items:    []*models.Todo{},
		}

		grandchild := &models.Todo{
			ID:       "grandchild-1",
			ParentID: "child-1",
			Position: 1,
			Text:     "Grandchild",
			Status:   models.StatusPending,
			Modified: time.Now(),
			Items:    []*models.Todo{},
		}

		child1.Items = []*models.Todo{grandchild}

		original := &models.Todo{
			ID:       "parent-1",
			ParentID: "",
			Position: 1,
			Text:     "Parent",
			Status:   models.StatusPending,
			Modified: time.Now(),
			Items:    []*models.Todo{child1, child2},
		}

		clone := original.Clone()

		// Verify structure is preserved
		assert.Len(t, clone.Items, 2)
		assert.Equal(t, "Child 1", clone.Items[0].Text)
		assert.Equal(t, "Child 2", clone.Items[1].Text)
		assert.Len(t, clone.Items[0].Items, 1)
		assert.Equal(t, "Grandchild", clone.Items[0].Items[0].Text)

		// Verify deep cloning - modify clone without affecting original
		clone.Items[0].Text = "Modified Child 1"
		clone.Items[0].Items[0].Text = "Modified Grandchild"

		assert.Equal(t, "Child 1", original.Items[0].Text)
		assert.Equal(t, "Grandchild", original.Items[0].Items[0].Text)

		// Verify pointers are different
		assert.NotSame(t, original.Items[0], clone.Items[0])
		assert.NotSame(t, original.Items[0].Items[0], clone.Items[0].Items[0])
	})
}

func TestCollection_Clone(t *testing.T) {
	t.Run("creates deep copy of empty collection", func(t *testing.T) {
		original := models.NewCollection()

		clone := original.Clone()

		assert.NotNil(t, clone)
		assert.NotNil(t, clone.Todos)
		assert.Empty(t, clone.Todos)
		assert.NotSame(t, &original.Todos, &clone.Todos)
	})

	t.Run("creates deep copy of collection with todos", func(t *testing.T) {
		original := models.NewCollection()
		todo1, _ := original.CreateTodo("First", "")
		todo2, _ := original.CreateTodo("Second", "")
		todo2.Status = models.StatusDone

		clone := original.Clone()

		assert.Len(t, clone.Todos, 2)
		assert.NotSame(t, &original.Todos, &clone.Todos)

		// Check todos are cloned
		assert.NotSame(t, original.Todos[0], clone.Todos[0])
		assert.NotSame(t, original.Todos[1], clone.Todos[1])

		// Check todo values are preserved
		assert.Equal(t, todo1.ID, clone.Todos[0].ID)
		assert.Equal(t, todo1.Text, clone.Todos[0].Text)
		assert.Equal(t, todo2.Status, clone.Todos[1].Status)
	})

	t.Run("clone is independent of original", func(t *testing.T) {
		original := models.NewCollection()
		_, _ = original.CreateTodo("First", "")
		_, _ = original.CreateTodo("Second", "")

		clone := original.Clone()

		// Modify clone
		_, _ = clone.CreateTodo("Third", "")
		clone.Todos[0].Text = "Modified"
		clone.Todos[1].Status = models.StatusDone
		clone.Todos[1].Modified = time.Now()

		// Original should be unchanged
		assert.Len(t, original.Todos, 2)
		assert.Equal(t, "First", original.Todos[0].Text)
		assert.Equal(t, models.StatusPending, original.Todos[1].Status)
	})

	t.Run("handles nil todos slice", func(t *testing.T) {
		original := &models.Collection{
			Todos: nil,
		}

		clone := original.Clone()

		assert.NotNil(t, clone)
		assert.NotNil(t, clone.Todos)
		assert.Empty(t, clone.Todos)
	})
}

func TestTodoStatus_Constants(t *testing.T) {
	t.Run("status constants have correct values", func(t *testing.T) {
		assert.Equal(t, models.TodoStatus("pending"), models.StatusPending)
		assert.Equal(t, models.TodoStatus("done"), models.StatusDone)
	})
}

func TestMigrateCollection(t *testing.T) {
	t.Run("assigns IDs to todos without IDs", func(t *testing.T) {
		collection := &models.Collection{
			Todos: []*models.Todo{
				{
					Position: 1,
					Text:     "First",
					Status:   models.StatusPending,
					Modified: time.Now(),
				},
				{
					ID:       "existing-id",
					Position: 2,
					Text:     "Second",
					Status:   models.StatusDone,
					Modified: time.Now(),
				},
			},
		}

		models.MigrateCollection(collection)

		// First todo should get a new ID
		assert.NotEmpty(t, collection.Todos[0].ID)
		assert.NotEqual(t, "existing-id", collection.Todos[0].ID)

		// Second todo should keep its existing ID
		assert.Equal(t, "existing-id", collection.Todos[1].ID)

		// Both should have empty ParentID (top-level)
		assert.Empty(t, collection.Todos[0].ParentID)
		assert.Empty(t, collection.Todos[1].ParentID)

		// Both should have Items initialized
		assert.NotNil(t, collection.Todos[0].Items)
		assert.NotNil(t, collection.Todos[1].Items)
	})

	t.Run("initializes Items field for todos without it", func(t *testing.T) {
		todo := &models.Todo{
			ID:       "test-id",
			Position: 1,
			Text:     "Test",
			Status:   models.StatusPending,
			Modified: time.Now(),
			Items:    nil,
		}
		collection := &models.Collection{
			Todos: []*models.Todo{todo},
		}

		models.MigrateCollection(collection)

		assert.NotNil(t, todo.Items)
		assert.Empty(t, todo.Items)
	})

	t.Run("handles nested todos correctly", func(t *testing.T) {
		grandchild := &models.Todo{
			Position: 1,
			Text:     "Grandchild",
			Status:   models.StatusPending,
			Modified: time.Now(),
		}

		child := &models.Todo{
			ID:       "child-id",
			Position: 1,
			Text:     "Child",
			Status:   models.StatusPending,
			Modified: time.Now(),
			Items:    []*models.Todo{grandchild},
		}

		parent := &models.Todo{
			Position: 1,
			Text:     "Parent",
			Status:   models.StatusPending,
			Modified: time.Now(),
			Items:    []*models.Todo{child},
		}

		collection := &models.Collection{
			Todos: []*models.Todo{parent},
		}

		models.MigrateCollection(collection)

		// Parent should get an ID
		assert.NotEmpty(t, parent.ID)
		assert.Empty(t, parent.ParentID)

		// Child should keep its ID and get parent's ID as ParentID
		assert.Equal(t, "child-id", child.ID)
		assert.Equal(t, parent.ID, child.ParentID)

		// Grandchild should get an ID and child's ID as ParentID
		assert.NotEmpty(t, grandchild.ID)
		assert.Equal(t, child.ID, grandchild.ParentID)

		// All should have Items initialized
		assert.NotNil(t, parent.Items)
		assert.NotNil(t, child.Items)
		assert.NotNil(t, grandchild.Items)
	})

	t.Run("handles empty collection", func(t *testing.T) {
		collection := models.NewCollection()

		// Should not panic
		models.MigrateCollection(collection)

		assert.Empty(t, collection.Todos)
	})

	t.Run("preserves existing ParentIDs", func(t *testing.T) {
		child := &models.Todo{
			ID:       "child-id",
			ParentID: "custom-parent-id",
			Position: 1,
			Text:     "Child",
			Status:   models.StatusPending,
			Modified: time.Now(),
			Items:    []*models.Todo{},
		}

		parent := &models.Todo{
			ID:       "parent-id",
			Position: 1,
			Text:     "Parent",
			Status:   models.StatusPending,
			Modified: time.Now(),
			Items:    []*models.Todo{child},
		}

		collection := &models.Collection{
			Todos: []*models.Todo{parent},
		}

		models.MigrateCollection(collection)

		// Child should keep its existing ParentID
		assert.Equal(t, "custom-parent-id", child.ParentID)
	})

	t.Run("handles very deep nesting", func(t *testing.T) {
		// Create a 5-level deep structure
		level5 := &models.Todo{Position: 1, Text: "Level 5", Status: models.StatusPending, Modified: time.Now()}
		level4 := &models.Todo{Position: 1, Text: "Level 4", Status: models.StatusPending, Modified: time.Now(), Items: []*models.Todo{level5}}
		level3 := &models.Todo{Position: 1, Text: "Level 3", Status: models.StatusPending, Modified: time.Now(), Items: []*models.Todo{level4}}
		level2 := &models.Todo{Position: 1, Text: "Level 2", Status: models.StatusPending, Modified: time.Now(), Items: []*models.Todo{level3}}
		level1 := &models.Todo{Position: 1, Text: "Level 1", Status: models.StatusPending, Modified: time.Now(), Items: []*models.Todo{level2}}

		collection := &models.Collection{
			Todos: []*models.Todo{level1},
		}

		models.MigrateCollection(collection)

		// All should have IDs
		assert.NotEmpty(t, level1.ID)
		assert.NotEmpty(t, level2.ID)
		assert.NotEmpty(t, level3.ID)
		assert.NotEmpty(t, level4.ID)
		assert.NotEmpty(t, level5.ID)

		// Verify parent-child relationships
		assert.Empty(t, level1.ParentID)
		assert.Equal(t, level1.ID, level2.ParentID)
		assert.Equal(t, level2.ID, level3.ParentID)
		assert.Equal(t, level3.ID, level4.ParentID)
		assert.Equal(t, level4.ID, level5.ParentID)
	})
}

func TestFindItemByPositionPath(t *testing.T) {
	// Helper to create a test collection with nested structure
	createTestCollection := func() *models.Collection {
		// Create a nested structure:
		// 1. First todo
		//    1.1. First child
		//    1.2. Second child
		//        1.2.1. Grandchild
		// 2. Second todo
		// 3. Third todo
		//    3.1. Another child

		grandchild := &models.Todo{
			ID:       "grandchild-1",
			ParentID: "child-1-2",
			Position: 1,
			Text:     "Grandchild 1.2.1",
			Status:   models.StatusPending,
			Modified: time.Now(),
			Items:    []*models.Todo{},
		}

		child11 := &models.Todo{
			ID:       "child-1-1",
			ParentID: "todo-1",
			Position: 1,
			Text:     "Child 1.1",
			Status:   models.StatusPending,
			Modified: time.Now(),
			Items:    []*models.Todo{},
		}

		child12 := &models.Todo{
			ID:       "child-1-2",
			ParentID: "todo-1",
			Position: 2,
			Text:     "Child 1.2",
			Status:   models.StatusPending,
			Modified: time.Now(),
			Items:    []*models.Todo{grandchild},
		}

		child31 := &models.Todo{
			ID:       "child-3-1",
			ParentID: "todo-3",
			Position: 1,
			Text:     "Child 3.1",
			Status:   models.StatusPending,
			Modified: time.Now(),
			Items:    []*models.Todo{},
		}

		todo1 := &models.Todo{
			ID:       "todo-1",
			Position: 1,
			Text:     "First todo",
			Status:   models.StatusPending,
			Modified: time.Now(),
			Items:    []*models.Todo{child11, child12},
		}

		todo2 := &models.Todo{
			ID:       "todo-2",
			Position: 2,
			Text:     "Second todo",
			Status:   models.StatusPending,
			Modified: time.Now(),
			Items:    []*models.Todo{},
		}

		todo3 := &models.Todo{
			ID:       "todo-3",
			Position: 3,
			Text:     "Third todo",
			Status:   models.StatusPending,
			Modified: time.Now(),
			Items:    []*models.Todo{child31},
		}

		return &models.Collection{
			Todos: []*models.Todo{todo1, todo2, todo3},
		}
	}

	t.Run("finds top-level items", func(t *testing.T) {
		collection := createTestCollection()

		item, err := collection.FindItemByPositionPath("1")
		assert.NoError(t, err)
		assert.Equal(t, "First todo", item.Text)

		item, err = collection.FindItemByPositionPath("2")
		assert.NoError(t, err)
		assert.Equal(t, "Second todo", item.Text)

		item, err = collection.FindItemByPositionPath("3")
		assert.NoError(t, err)
		assert.Equal(t, "Third todo", item.Text)
	})

	t.Run("finds nested items", func(t *testing.T) {
		collection := createTestCollection()

		item, err := collection.FindItemByPositionPath("1.1")
		assert.NoError(t, err)
		assert.Equal(t, "Child 1.1", item.Text)

		item, err = collection.FindItemByPositionPath("1.2")
		assert.NoError(t, err)
		assert.Equal(t, "Child 1.2", item.Text)

		item, err = collection.FindItemByPositionPath("3.1")
		assert.NoError(t, err)
		assert.Equal(t, "Child 3.1", item.Text)
	})

	t.Run("finds deeply nested items", func(t *testing.T) {
		collection := createTestCollection()

		item, err := collection.FindItemByPositionPath("1.2.1")
		assert.NoError(t, err)
		assert.Equal(t, "Grandchild 1.2.1", item.Text)
	})

	t.Run("returns error for non-existent paths", func(t *testing.T) {
		collection := createTestCollection()

		_, err := collection.FindItemByPositionPath("4")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no item found at position 4")

		_, err = collection.FindItemByPositionPath("1.3")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no item found at position 3")

		_, err = collection.FindItemByPositionPath("2.1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no item found at position 1")
	})

	t.Run("handles empty path", func(t *testing.T) {
		collection := createTestCollection()

		_, err := collection.FindItemByPositionPath("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty path")
	})

	t.Run("handles invalid paths", func(t *testing.T) {
		collection := createTestCollection()

		_, err := collection.FindItemByPositionPath("abc")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid position")

		_, err = collection.FindItemByPositionPath("1.abc.3")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid position")

		_, err = collection.FindItemByPositionPath("0")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "position must be >= 1")

		_, err = collection.FindItemByPositionPath("1.-1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "position must be >= 1")
	})

	t.Run("handles paths with spaces", func(t *testing.T) {
		collection := createTestCollection()

		item, err := collection.FindItemByPositionPath(" 1 . 2 ")
		assert.NoError(t, err)
		assert.Equal(t, "Child 1.2", item.Text)
	})

	t.Run("works with empty collection", func(t *testing.T) {
		collection := models.NewCollection()

		_, err := collection.FindItemByPositionPath("1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no item found at position 1")
	})

	t.Run("works with unordered positions", func(t *testing.T) {
		// Create collection with non-sequential positions
		collection := &models.Collection{
			Todos: []*models.Todo{
				{
					ID:       "todo-5",
					Position: 5,
					Text:     "Position 5",
					Status:   models.StatusPending,
					Modified: time.Now(),
					Items: []*models.Todo{
						{
							ID:       "child-5-10",
							ParentID: "todo-5",
							Position: 10,
							Text:     "Child at position 10",
							Status:   models.StatusPending,
							Modified: time.Now(),
							Items:    []*models.Todo{},
						},
					},
				},
			},
		}

		item, err := collection.FindItemByPositionPath("5")
		assert.NoError(t, err)
		assert.Equal(t, "Position 5", item.Text)

		item, err = collection.FindItemByPositionPath("5.10")
		assert.NoError(t, err)
		assert.Equal(t, "Child at position 10", item.Text)
	})
}

func TestFindItemByID(t *testing.T) {
	// Helper to create a test collection with nested structure
	createTestCollection := func() *models.Collection {
		collection := models.NewCollection()

		// Create parent
		parent, _ := collection.CreateTodo("Parent", "")

		// Create children
		child1, _ := collection.CreateTodo("Child 1", parent.ID)
		_, _ = collection.CreateTodo("Child 2", parent.ID)

		// Create grandchild
		_, _ = collection.CreateTodo("Grandchild", child1.ID)

		return collection
	}

	t.Run("finds top-level todo by ID", func(t *testing.T) {
		collection := createTestCollection()
		parent := collection.Todos[0]

		found := collection.FindItemByID(parent.ID)
		assert.NotNil(t, found)
		assert.Equal(t, parent.ID, found.ID)
		assert.Equal(t, "Parent", found.Text)
	})

	t.Run("finds nested todo by ID", func(t *testing.T) {
		collection := createTestCollection()
		parent := collection.Todos[0]
		child1 := parent.Items[0]

		found := collection.FindItemByID(child1.ID)
		assert.NotNil(t, found)
		assert.Equal(t, child1.ID, found.ID)
		assert.Equal(t, "Child 1", found.Text)
	})

	t.Run("finds deeply nested todo by ID", func(t *testing.T) {
		collection := createTestCollection()
		parent := collection.Todos[0]
		child1 := parent.Items[0]
		grandchild := child1.Items[0]

		found := collection.FindItemByID(grandchild.ID)
		assert.NotNil(t, found)
		assert.Equal(t, grandchild.ID, found.ID)
		assert.Equal(t, "Grandchild", found.Text)
	})

	t.Run("returns nil for non-existent ID", func(t *testing.T) {
		collection := createTestCollection()

		found := collection.FindItemByID("non-existent-id")
		assert.Nil(t, found)
	})

	t.Run("handles empty collection", func(t *testing.T) {
		collection := models.NewCollection()

		found := collection.FindItemByID("any-id")
		assert.Nil(t, found)
	})
}
