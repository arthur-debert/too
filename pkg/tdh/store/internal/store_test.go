package internal_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/arthur-debert/tdh/pkg/tdh/store"
	"github.com/arthur-debert/tdh/pkg/tdh/store/internal"
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

		store := internal.NewJSONFileStore(dbPath)
		collection, err := store.Load()

		require.NoError(t, err)
		assert.NotNil(t, collection)
		assert.Len(t, collection.Todos, 1)
		assert.Equal(t, 1, collection.Todos[0].Position)
		assert.NotEmpty(t, collection.Todos[0].ID) // Should have UUID
	})

	t.Run("should return empty collection if file does not exist", func(t *testing.T) {
		store := internal.NewJSONFileStore("non-existent-file.json")
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

		store := internal.NewJSONFileStore(dbPath)
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
		store := internal.NewJSONFileStore(dbPath)
		collection := models.NewCollection()
		_, err = collection.CreateTodo("My new todo", "")
		require.NoError(t, err)

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
		store := internal.NewJSONFileStore(dbPath)

		// Initial save
		err = store.Save(models.NewCollection())
		require.NoError(t, err)

		var todo *models.Todo
		err = store.Update(func(collection *models.Collection) error {
			todo, _ = collection.CreateTodo("Updated from transaction", "")
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
		store := internal.NewMemoryStore()
		collection, err := store.Load()
		require.NoError(t, err)
		assert.Empty(t, collection.Todos)

		_, err = collection.CreateTodo("In-memory todo", "")
		require.NoError(t, err)
		err = store.Save(collection)
		require.NoError(t, err)

		loaded, err := store.Load()
		require.NoError(t, err)
		assert.Len(t, loaded.Todos, 1)
		assert.Equal(t, "In-memory todo", loaded.Todos[0].Text)
	})

	t.Run("should handle update correctly", func(t *testing.T) {
		store := internal.NewMemoryStore()
		var todo *models.Todo
		err := store.Update(func(collection *models.Collection) error {
			todo, _ = collection.CreateTodo("Updated in-memory", "")
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
		store := internal.NewMemoryStore()
		store.ShouldFail = true

		_, err := store.Load()
		assert.Error(t, err)

		err = store.Save(models.NewCollection())
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

		store := internal.NewJSONFileStore(file.Name())
		_, err = store.Load()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read store file")
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

		store := internal.NewJSONFileStore(file.Name())
		_, err = store.Load()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode JSON")
		assert.Contains(t, err.Error(), file.Name())
	})

	t.Run("should return descriptive error when save fails", func(t *testing.T) {
		// Use a non-existent directory for the store path
		store := internal.NewJSONFileStore("/non-existent-dir/todos.json")
		collection := models.NewCollection()

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
		store := internal.NewJSONFileStore(dbPath)

		// Create initial collection with one todo
		collection := models.NewCollection()
		todo1, _ := collection.CreateTodo("Original todo", "")
		err = store.Save(collection)
		require.NoError(t, err)

		// Attempt an update that fails
		err = store.Update(func(c *models.Collection) error {
			_, _ = c.CreateTodo("This should be rolled back", "")
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
		store := internal.NewMemoryStore()

		// Create initial collection with one todo
		collection := models.NewCollection()
		todo1, _ := collection.CreateTodo("Original todo", "")
		err := store.Save(collection)
		require.NoError(t, err)

		// Attempt an update that fails
		err = store.Update(func(c *models.Collection) error {
			_, _ = c.CreateTodo("This should be rolled back", "")
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
		store := internal.NewMemoryStore()

		// Create initial collection with one todo
		collection := models.NewCollection()
		_, _ = collection.CreateTodo("Original todo", "")
		err := store.Save(collection)
		require.NoError(t, err)

		// Perform a successful update
		var newTodo *models.Todo
		err = store.Update(func(c *models.Collection) error {
			newTodo, _ = c.CreateTodo("New todo", "")
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

func TestStore_Exists(t *testing.T) {
	t.Run("JSONFileStore.Exists should return false for non-existent file", func(t *testing.T) {
		store := internal.NewJSONFileStore("/non-existent-path/file.json")
		assert.False(t, store.Exists())
	})

	t.Run("JSONFileStore.Exists should return true for existing file", func(t *testing.T) {
		file, err := os.CreateTemp("", "tdh-exists-test")
		require.NoError(t, err)
		defer func() { _ = os.Remove(file.Name()) }()
		_ = file.Close()

		store := internal.NewJSONFileStore(file.Name())
		assert.True(t, store.Exists())
	})

	t.Run("MemoryStore.Exists should always return true", func(t *testing.T) {
		store := internal.NewMemoryStore()
		assert.True(t, store.Exists())
	})
}

func TestStore_Path(t *testing.T) {
	t.Run("JSONFileStore.Path should return the file path", func(t *testing.T) {
		expectedPath := "/some/path/todos.json"
		store := internal.NewJSONFileStore(expectedPath)
		assert.Equal(t, expectedPath, store.Path())
	})

	t.Run("MemoryStore.Path should return memory URL", func(t *testing.T) {
		store := internal.NewMemoryStore()
		assert.Equal(t, "memory://todos", store.Path())
	})
}

func TestJSONFileStore_SaveEdgeCases(t *testing.T) {
	t.Run("should handle save with nil todos slice", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "tdh-save-test")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(dir) }()

		dbPath := filepath.Join(dir, "test.json")
		store := internal.NewJSONFileStore(dbPath)
		collection := &models.Collection{
			Todos: nil,
		}

		err = store.Save(collection)
		require.NoError(t, err)

		// Verify the file was created with null value
		data, err := os.ReadFile(dbPath)
		require.NoError(t, err)
		assert.Equal(t, "null", string(data))
	})

	t.Run("should cleanup temp file even if rename fails", func(t *testing.T) {
		// Create a read-only directory to cause rename to fail
		dir, err := os.MkdirTemp("", "tdh-readonly-test")
		require.NoError(t, err)
		defer func() {
			_ = os.Chmod(dir, 0755) // Restore permissions before cleanup
			_ = os.RemoveAll(dir)
		}()

		// Create the file first
		dbPath := filepath.Join(dir, "test.json")
		err = os.WriteFile(dbPath, []byte("[]"), 0644)
		require.NoError(t, err)

		// Make directory read-only
		err = os.Chmod(dir, 0555)
		require.NoError(t, err)

		store := internal.NewJSONFileStore(dbPath)
		collection := models.NewCollection()
		_, _ = collection.CreateTodo("Test", "")

		// This should fail during temp file creation
		err = store.Save(collection)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create temp file")
	})
}

func TestStore_Find(t *testing.T) {
	stores := map[string]store.Store{
		"JSONFileStore": newTestJSONFileStore(t),
		"MemoryStore":   internal.NewMemoryStore(),
	}

	for name, s := range stores {
		t.Run(name, func(t *testing.T) {
			// Setup initial data
			collection := models.NewCollection()
			_, _ = collection.CreateTodo("Buy milk", "")
			doneTodo, _ := collection.CreateTodo("Buy eggs", "")
			doneTodo.Toggle()
			_, _ = collection.CreateTodo("Buy bread and milk", "")
			err := s.Save(collection)
			require.NoError(t, err)

			t.Run("should find by status", func(t *testing.T) {
				doneStatus := string(models.StatusDone)
				query := store.Query{Status: &doneStatus}
				results, err := s.Find(query)
				require.NoError(t, err)
				assert.Len(t, results.Todos, 1)
				assert.Equal(t, "Buy eggs", results.Todos[0].Text)
				assert.Equal(t, 3, results.TotalCount)
				assert.Equal(t, 1, results.DoneCount)
			})

			t.Run("should find by text (case-insensitive)", func(t *testing.T) {
				text := "milk"
				query := store.Query{TextContains: &text}
				results, err := s.Find(query)
				require.NoError(t, err)
				assert.Len(t, results.Todos, 2)
			})

			t.Run("should find by text (case-sensitive)", func(t *testing.T) {
				text := "Milk"
				query := store.Query{TextContains: &text, CaseSensitive: true}
				results, err := s.Find(query)
				require.NoError(t, err)
				assert.Len(t, results.Todos, 0)
			})

			t.Run("should combine filters", func(t *testing.T) {
				text := "milk"
				pendingStatus := string(models.StatusPending)
				query := store.Query{TextContains: &text, Status: &pendingStatus}
				results, err := s.Find(query)
				require.NoError(t, err)
				assert.Len(t, results.Todos, 2)
			})

			t.Run("should return empty slice for no matches", func(t *testing.T) {
				text := "non-existent"
				query := store.Query{TextContains: &text}
				results, err := s.Find(query)
				require.NoError(t, err)
				assert.Len(t, results.Todos, 0)
			})
		})
	}
}

func newTestJSONFileStore(t *testing.T) store.Store {
	t.Helper()
	dir, err := os.MkdirTemp("", "tdh-find-test")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	return internal.NewJSONFileStore(filepath.Join(dir, "test.json"))
}

func TestJSONFileStore_NestedTodos(t *testing.T) {
	t.Run("should save and load nested todos", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "tdh-nested-test")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(dir) }()

		dbPath := filepath.Join(dir, "test.json")
		store := internal.NewJSONFileStore(dbPath)

		// Create a collection with nested todos
		collection := models.NewCollection()
		parent, _ := collection.CreateTodo("Parent task", "")

		child1 := &models.Todo{
			ID:       "child-1",
			ParentID: parent.ID,
			Position: 1,
			Text:     "Child task 1",
			Status:   models.StatusPending,
			Modified: parent.Modified,
			Items:    []*models.Todo{},
		}

		child2 := &models.Todo{
			ID:       "child-2",
			ParentID: parent.ID,
			Position: 2,
			Text:     "Child task 2",
			Status:   models.StatusDone,
			Modified: parent.Modified,
			Items:    []*models.Todo{},
		}

		grandchild := &models.Todo{
			ID:       "grandchild-1",
			ParentID: child1.ID,
			Position: 1,
			Text:     "Grandchild task",
			Status:   models.StatusPending,
			Modified: parent.Modified,
			Items:    []*models.Todo{},
		}

		child1.Items = []*models.Todo{grandchild}
		parent.Items = []*models.Todo{child1, child2}

		// Save the collection
		err = store.Save(collection)
		require.NoError(t, err)

		// Load it back
		loaded, err := store.Load()
		require.NoError(t, err)

		// Verify structure is preserved
		assert.Len(t, loaded.Todos, 1)
		assert.Equal(t, "Parent task", loaded.Todos[0].Text)
		assert.Len(t, loaded.Todos[0].Items, 2)

		// Verify children
		assert.Equal(t, "Child task 1", loaded.Todos[0].Items[0].Text)
		assert.Equal(t, "Child task 2", loaded.Todos[0].Items[1].Text)
		assert.Equal(t, models.StatusDone, loaded.Todos[0].Items[1].Status)

		// Verify grandchild
		assert.Len(t, loaded.Todos[0].Items[0].Items, 1)
		assert.Equal(t, "Grandchild task", loaded.Todos[0].Items[0].Items[0].Text)

		// Verify ParentIDs are preserved
		assert.Equal(t, parent.ID, loaded.Todos[0].Items[0].ParentID)
		assert.Equal(t, parent.ID, loaded.Todos[0].Items[1].ParentID)
		assert.Equal(t, child1.ID, loaded.Todos[0].Items[0].Items[0].ParentID)
	})

	t.Run("should migrate flat todos to nested structure", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "tdh-migrate-test")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(dir) }()

		dbPath := filepath.Join(dir, "test.json")

		// Write old format data without nested structure
		oldData := `[
			{"id": "existing-id-1", "position": 1, "text": "First todo", "status": "pending", "modified": "2024-01-01T00:00:00Z"},
			{"id": "existing-id-2", "position": 2, "text": "Second todo", "status": "done", "modified": "2024-01-01T00:00:00Z"}
		]`
		err = os.WriteFile(dbPath, []byte(oldData), 0600)
		require.NoError(t, err)

		store := internal.NewJSONFileStore(dbPath)
		collection, err := store.Load()
		require.NoError(t, err)

		// Verify migration happened
		assert.Len(t, collection.Todos, 2)

		// All todos should have Items initialized
		for _, todo := range collection.Todos {
			assert.NotNil(t, todo.Items)
			assert.Empty(t, todo.ParentID) // Top-level todos have empty ParentID
		}

		// Save and reload to ensure persistence
		err = store.Save(collection)
		require.NoError(t, err)

		reloaded, err := store.Load()
		require.NoError(t, err)
		assert.Len(t, reloaded.Todos, 2)
		for _, todo := range reloaded.Todos {
			assert.NotNil(t, todo.Items)
		}
	})

	t.Run("should handle deeply nested structures", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "tdh-deep-test")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(dir) }()

		dbPath := filepath.Join(dir, "test.json")
		store := internal.NewJSONFileStore(dbPath)

		// Create a 5-level deep structure
		collection := models.NewCollection()
		level1, _ := collection.CreateTodo("Level 1", "")

		level2 := &models.Todo{
			ID: "level-2", ParentID: level1.ID, Position: 1,
			Text: "Level 2", Status: models.StatusPending,
			Modified: level1.Modified, Items: []*models.Todo{},
		}

		level3 := &models.Todo{
			ID: "level-3", ParentID: level2.ID, Position: 1,
			Text: "Level 3", Status: models.StatusPending,
			Modified: level1.Modified, Items: []*models.Todo{},
		}

		level4 := &models.Todo{
			ID: "level-4", ParentID: level3.ID, Position: 1,
			Text: "Level 4", Status: models.StatusPending,
			Modified: level1.Modified, Items: []*models.Todo{},
		}

		level5 := &models.Todo{
			ID: "level-5", ParentID: level4.ID, Position: 1,
			Text: "Level 5", Status: models.StatusPending,
			Modified: level1.Modified, Items: []*models.Todo{},
		}

		level4.Items = []*models.Todo{level5}
		level3.Items = []*models.Todo{level4}
		level2.Items = []*models.Todo{level3}
		level1.Items = []*models.Todo{level2}

		// Save and reload
		err = store.Save(collection)
		require.NoError(t, err)

		loaded, err := store.Load()
		require.NoError(t, err)

		// Navigate to the deepest level
		current := loaded.Todos[0]
		assert.Equal(t, "Level 1", current.Text)

		current = current.Items[0]
		assert.Equal(t, "Level 2", current.Text)

		current = current.Items[0]
		assert.Equal(t, "Level 3", current.Text)

		current = current.Items[0]
		assert.Equal(t, "Level 4", current.Text)

		current = current.Items[0]
		assert.Equal(t, "Level 5", current.Text)
		assert.Empty(t, current.Items)
	})
}
