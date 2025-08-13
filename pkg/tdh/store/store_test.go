package store

import (
	"errors"
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

func TestJSONFileStore_ErrorHandling(t *testing.T) {
	t.Run("should return descriptive error when file cannot be opened", func(t *testing.T) {
		// Create a file with no read permissions
		file, err := os.CreateTemp("", "tdh-perm-test")
		require.NoError(t, err)
		defer func() { _ = os.Remove(file.Name()) }()

		_ = file.Close()
		// Remove read permissions
		err = os.Chmod(file.Name(), 0200)
		require.NoError(t, err)

		store := NewJSONFileStore(file.Name())
		_, err = store.Load()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to open store file")
		assert.Contains(t, err.Error(), file.Name())
	})

	t.Run("should return descriptive error for invalid JSON", func(t *testing.T) {
		// Create a file with invalid JSON
		file, err := os.CreateTemp("", "tdh-json-test")
		require.NoError(t, err)
		defer func() { _ = os.Remove(file.Name()) }()

		_, err = file.WriteString("{ invalid json }")
		require.NoError(t, err)
		_ = file.Close()

		store := NewJSONFileStore(file.Name())
		_, err = store.Load()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode JSON")
		assert.Contains(t, err.Error(), file.Name())
	})

	t.Run("should return descriptive error when save fails", func(t *testing.T) {
		// Use a non-existent directory for the store path
		store := NewJSONFileStore("/non-existent-dir/todos.json")
		collection := models.NewCollection("")

		err := store.Save(collection)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create temp file")
		assert.Contains(t, err.Error(), "/non-existent-dir")
	})
}

func TestStore_TransactionRollback(t *testing.T) {
	t.Run("JSONFileStore should rollback on error", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "tdh-rollback-test")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(dir) }()

		dbPath := filepath.Join(dir, "test.json")
		store := NewJSONFileStore(dbPath)

		// Create initial collection with one todo
		collection := models.NewCollection(dbPath)
		todo1 := collection.CreateTodo("Original todo")
		err = store.Save(collection)
		require.NoError(t, err)

		// Attempt an update that fails
		err = store.Update(func(c *models.Collection) error {
			c.CreateTodo("This should be rolled back")
			c.Todos[0].Text = "Modified text"
			return errors.New("simulated error")
		})

		assert.Error(t, err)
		assert.Equal(t, "simulated error", err.Error())

		// Verify the collection is unchanged
		loaded, err := store.Load()
		require.NoError(t, err)
		assert.Len(t, loaded.Todos, 1)
		assert.Equal(t, "Original todo", loaded.Todos[0].Text)
		assert.Equal(t, todo1.ID, loaded.Todos[0].ID)
	})

	t.Run("MemoryStore should rollback on error", func(t *testing.T) {
		store := NewMemoryStore()

		// Create initial collection with one todo
		collection := models.NewCollection("")
		todo1 := collection.CreateTodo("Original todo")
		err := store.Save(collection)
		require.NoError(t, err)

		// Attempt an update that fails
		err = store.Update(func(c *models.Collection) error {
			c.CreateTodo("This should be rolled back")
			c.Todos[0].Text = "Modified text"
			return errors.New("simulated error")
		})

		assert.Error(t, err)
		assert.Equal(t, "simulated error", err.Error())

		// Verify the collection is unchanged
		loaded, err := store.Load()
		require.NoError(t, err)
		assert.Len(t, loaded.Todos, 1)
		assert.Equal(t, "Original todo", loaded.Todos[0].Text)
		assert.Equal(t, todo1.ID, loaded.Todos[0].ID)
	})

	t.Run("successful update should persist changes", func(t *testing.T) {
		store := NewMemoryStore()

		// Create initial collection with one todo
		collection := models.NewCollection("")
		collection.CreateTodo("Original todo")
		err := store.Save(collection)
		require.NoError(t, err)

		// Perform a successful update
		var newTodo *models.Todo
		err = store.Update(func(c *models.Collection) error {
			newTodo = c.CreateTodo("New todo")
			c.Todos[0].Text = "Modified text"
			return nil
		})

		require.NoError(t, err)
		assert.NotNil(t, newTodo)

		// Verify the changes persisted
		loaded, err := store.Load()
		require.NoError(t, err)
		assert.Len(t, loaded.Todos, 2)
		assert.Equal(t, "Modified text", loaded.Todos[0].Text)
		assert.Equal(t, "New todo", loaded.Todos[1].Text)
	})
}
