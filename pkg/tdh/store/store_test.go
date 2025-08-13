package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/arthur-debert/tdh/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONFileStore_Load(t *testing.T) {
	t.Run("should load existing collection", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "tdh-test")
		require.NoError(t, err)
		defer func() {
			err := os.RemoveAll(dir)
			require.NoError(t, err)
		}()

		dbPath := filepath.Join(dir, "test.json")
		err = os.WriteFile(dbPath, []byte(`[{"id": 1, "text": "Test Todo", "status": "pending"}]`), 0600)
		require.NoError(t, err)

		store := NewJSONFileStore(dbPath)
		collection, err := store.Load()

		require.NoError(t, err)
		assert.NotNil(t, collection)
		assert.Len(t, collection.Todos, 1)
		assert.Equal(t, int64(1), collection.Todos[0].ID)
	})

	t.Run("should return empty collection if file does not exist", func(t *testing.T) {
		store := NewJSONFileStore("non-existent-file.json")
		collection, err := store.Load()

		require.NoError(t, err)
		assert.NotNil(t, collection)
		assert.Empty(t, collection.Todos)
	})

	t.Run("should return empty collection for empty file", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "tdh-test")
		require.NoError(t, err)
		defer func() {
			err := os.RemoveAll(dir)
			require.NoError(t, err)
		}()

		dbPath := filepath.Join(dir, "empty.json")
		err = os.WriteFile(dbPath, []byte(""), 0600)
		require.NoError(t, err)

		store := NewJSONFileStore(dbPath)
		collection, err := store.Load()

		require.NoError(t, err)
		assert.NotNil(t, collection)
		assert.Empty(t, collection.Todos)
	})
}

func TestJSONFileStore_Save(t *testing.T) {
	t.Run("should save collection atomically", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "tdh-test")
		require.NoError(t, err)
		defer func() {
			err := os.RemoveAll(dir)
			require.NoError(t, err)
		}()

		dbPath := filepath.Join(dir, "test.json")
		store := NewJSONFileStore(dbPath)
		collection := models.NewCollection(dbPath)
		collection.CreateTodo("My new todo")

		err = store.Save(collection)
		require.NoError(t, err)

		data, err := os.ReadFile(dbPath)
		require.NoError(t, err)
		assert.Contains(t, string(data), "My new todo")
	})
}

func TestJSONFileStore_Update(t *testing.T) {
	t.Run("should perform update transaction", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "tdh-test")
		require.NoError(t, err)
		defer func() {
			err := os.RemoveAll(dir)
			require.NoError(t, err)
		}()

		dbPath := filepath.Join(dir, "test.json")
		store := NewJSONFileStore(dbPath)

		// Initial save
		err = store.Save(models.NewCollection(dbPath))
		require.NoError(t, err)

		var todo *models.Todo
		err = store.Update(func(collection *models.Collection) error {
			todo = collection.CreateTodo("Updated from transaction")
			return nil
		})

		require.NoError(t, err)
		assert.NotNil(t, todo)
		assert.Equal(t, "Updated from transaction", todo.Text)

		// Verify file content
		finalCollection, err := store.Load()
		require.NoError(t, err)
		assert.Len(t, finalCollection.Todos, 1)
		assert.Equal(t, "Updated from transaction", finalCollection.Todos[0].Text)
	})
}

func TestMemoryStore(t *testing.T) {
	t.Run("should load and save correctly", func(t *testing.T) {
		store := NewMemoryStore()
		collection, err := store.Load()
		require.NoError(t, err)
		assert.Empty(t, collection.Todos)

		collection.CreateTodo("In-memory todo")
		err = store.Save(collection)
		require.NoError(t, err)

		loaded, err := store.Load()
		require.NoError(t, err)
		assert.Len(t, loaded.Todos, 1)
		assert.Equal(t, "In-memory todo", loaded.Todos[0].Text)
	})

	t.Run("should handle update correctly", func(t *testing.T) {
		store := NewMemoryStore()
		var todo *models.Todo
		err := store.Update(func(collection *models.Collection) error {
			todo = collection.CreateTodo("Updated in-memory")
			return nil
		})

		require.NoError(t, err)
		assert.NotNil(t, todo)

		loaded, err := store.Load()
		require.NoError(t, err)
		assert.Len(t, loaded.Todos, 1)
		assert.Equal(t, "Updated in-memory", loaded.Todos[0].Text)
	})

	t.Run("should simulate failure", func(t *testing.T) {
		store := NewMemoryStore()
		store.ShouldFail = true

		_, err := store.Load()
		assert.Error(t, err)

		err = store.Save(models.NewCollection(""))
		assert.Error(t, err)

		err = store.Update(func(c *models.Collection) error { return nil })
		assert.Error(t, err)
	})
}
