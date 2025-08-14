package models_test

import (
	"testing"
	"time"

	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/stretchr/testify/assert"
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

		todo := collection.CreateTodo("Test todo")
		afterCreate := time.Now()

		assert.NotNil(t, todo)
		assert.Equal(t, int64(1), todo.ID)
		assert.Equal(t, "Test todo", todo.Text)
		assert.Equal(t, models.StatusPending, todo.Status)
		assert.True(t, todo.Modified.After(beforeCreate) || todo.Modified.Equal(beforeCreate))
		assert.True(t, todo.Modified.Before(afterCreate) || todo.Modified.Equal(afterCreate))
	})

	t.Run("adds todo to collection", func(t *testing.T) {
		collection := models.NewCollection()

		todo := collection.CreateTodo("Test todo")

		assert.Len(t, collection.Todos, 1)
		assert.Equal(t, todo, collection.Todos[0])
	})

	t.Run("assigns incremental IDs", func(t *testing.T) {
		collection := models.NewCollection()

		todo1 := collection.CreateTodo("First")
		todo2 := collection.CreateTodo("Second")
		todo3 := collection.CreateTodo("Third")

		assert.Equal(t, int64(1), todo1.ID)
		assert.Equal(t, int64(2), todo2.ID)
		assert.Equal(t, int64(3), todo3.ID)
	})

	t.Run("handles gaps in IDs correctly", func(t *testing.T) {
		collection := models.NewCollection()

		// Manually add todos with non-sequential IDs
		collection.Todos = []*models.Todo{
			{ID: 1, Text: "First", Status: models.StatusPending},
			{ID: 5, Text: "Fifth", Status: models.StatusPending},
			{ID: 3, Text: "Third", Status: models.StatusPending},
		}

		// Create new todo - should get ID 6 (highest + 1)
		newTodo := collection.CreateTodo("New todo")

		assert.Equal(t, int64(6), newTodo.ID)
	})

	t.Run("handles empty text", func(t *testing.T) {
		collection := models.NewCollection()

		todo := collection.CreateTodo("")

		assert.Equal(t, "", todo.Text)
		assert.Equal(t, int64(1), todo.ID)
	})

	t.Run("creates multiple todos with different timestamps", func(t *testing.T) {
		collection := models.NewCollection()

		todo1 := collection.CreateTodo("First")
		time.Sleep(2 * time.Millisecond) // Small delay to ensure different timestamps
		todo2 := collection.CreateTodo("Second")

		assert.NotEqual(t, todo1.Modified, todo2.Modified)
		assert.True(t, todo2.Modified.After(todo1.Modified))
	})
}

func TestTodo_Toggle(t *testing.T) {
	t.Run("toggles from pending to done", func(t *testing.T) {
		todo := &models.Todo{
			ID:       1,
			Text:     "Test",
			Status:   models.StatusPending,
			Modified: time.Now().Add(-time.Hour),
		}
		oldModified := todo.Modified

		time.Sleep(2 * time.Millisecond)
		todo.Toggle()

		assert.Equal(t, models.StatusDone, todo.Status)
		assert.True(t, todo.Modified.After(oldModified))
	})

	t.Run("toggles from done to pending", func(t *testing.T) {
		todo := &models.Todo{
			ID:       1,
			Text:     "Test",
			Status:   models.StatusDone,
			Modified: time.Now().Add(-time.Hour),
		}
		oldModified := todo.Modified

		time.Sleep(2 * time.Millisecond)
		todo.Toggle()

		assert.Equal(t, models.StatusPending, todo.Status)
		assert.True(t, todo.Modified.After(oldModified))
	})

	t.Run("updates modified time on toggle", func(t *testing.T) {
		todo := &models.Todo{
			ID:       1,
			Text:     "Test",
			Status:   models.StatusPending,
			Modified: time.Now().Add(-24 * time.Hour), // Yesterday
		}
		oldModified := todo.Modified
		beforeToggle := time.Now()

		todo.Toggle()
		afterToggle := time.Now()

		assert.True(t, todo.Modified.After(oldModified))
		assert.True(t, todo.Modified.After(beforeToggle) || todo.Modified.Equal(beforeToggle))
		assert.True(t, todo.Modified.Before(afterToggle) || todo.Modified.Equal(afterToggle))
	})
}

func TestTodo_Clone(t *testing.T) {
	t.Run("creates exact copy of todo", func(t *testing.T) {
		original := &models.Todo{
			ID:       42,
			Text:     "Original todo",
			Status:   models.StatusDone,
			Modified: time.Now().Add(-time.Hour),
		}

		clone := original.Clone()

		assert.Equal(t, original.ID, clone.ID)
		assert.Equal(t, original.Text, clone.Text)
		assert.Equal(t, original.Status, clone.Status)
		assert.Equal(t, original.Modified, clone.Modified)
	})

	t.Run("clone is independent of original", func(t *testing.T) {
		original := &models.Todo{
			ID:       1,
			Text:     "Original",
			Status:   models.StatusPending,
			Modified: time.Now(),
		}

		clone := original.Clone()

		// Modify clone
		clone.Text = "Modified"
		clone.Status = models.StatusDone
		clone.ID = 99

		// Original should be unchanged
		assert.Equal(t, "Original", original.Text)
		assert.Equal(t, models.StatusPending, original.Status)
		assert.Equal(t, int64(1), original.ID)
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
		todo1 := original.CreateTodo("First")
		todo2 := original.CreateTodo("Second")
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
		original.CreateTodo("First")
		original.CreateTodo("Second")

		clone := original.Clone()

		// Modify clone
		clone.CreateTodo("Third")
		clone.Todos[0].Text = "Modified"
		clone.Todos[1].Toggle()

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
