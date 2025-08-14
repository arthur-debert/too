package add_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/arthur-debert/tdh/pkg/tdh/commands/add"
	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/arthur-debert/tdh/pkg/tdh/testutil"
	"github.com/stretchr/testify/assert"
)

func TestAddCommand(t *testing.T) {
	t.Run("adds todo to empty collection", func(t *testing.T) {
		// Use testutil to create a clean store
		store := testutil.CreatePopulatedStore(t) // Empty store

		opts := add.Options{CollectionPath: store.Path()}
		result, err := add.Execute("My first todo", opts)

		// Use testutil assertions
		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Todo)
		assert.Equal(t, "My first todo", result.Todo.Text)
		assert.Equal(t, int64(1), result.Todo.ID)
		assert.Equal(t, models.StatusPending, result.Todo.Status)

		// Verify it was saved using testutil
		collection, err := store.Load()
		testutil.AssertNoError(t, err)

		testutil.AssertCollectionSize(t, collection, 1)
		testutil.AssertTodoInList(t, collection.Todos, "My first todo")
	})

	t.Run("adds todo to existing collection", func(t *testing.T) {
		// Create store with existing todos
		store := testutil.CreatePopulatedStore(t, "Existing 1", "Existing 2")

		opts := add.Options{CollectionPath: store.Path()}
		result, err := add.Execute("New todo", opts)

		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "New todo", result.Todo.Text)
		assert.Equal(t, int64(3), result.Todo.ID) // Should be 3 after two existing

		// Verify all todos are present
		collection, err := store.Load()
		testutil.AssertNoError(t, err)

		testutil.AssertCollectionSize(t, collection, 3)
		testutil.AssertTodoInList(t, collection.Todos, "Existing 1")
		testutil.AssertTodoInList(t, collection.Todos, "Existing 2")
		testutil.AssertTodoInList(t, collection.Todos, "New todo")
	})

	t.Run("returns error for empty text", func(t *testing.T) {
		store := testutil.CreatePopulatedStore(t)

		opts := add.Options{CollectionPath: store.Path()}
		result, err := add.Execute("", opts)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "todo text cannot be empty")

		// Verify nothing was added
		collection, err := store.Load()
		testutil.AssertNoError(t, err)
		testutil.AssertCollectionSize(t, collection, 0)
	})

	t.Run("handles very long text", func(t *testing.T) {
		store := testutil.CreatePopulatedStore(t)
		longText := string(make([]byte, 1000))
		for i := range longText {
			longText = longText[:i] + "a" + longText[i+1:]
		}

		opts := add.Options{CollectionPath: store.Path()}
		result, err := add.Execute(longText, opts)

		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, longText, result.Todo.Text)

		// Verify it was saved
		collection, err := store.Load()
		testutil.AssertNoError(t, err)
		testutil.AssertCollectionSize(t, collection, 1)
		assert.Equal(t, longText, collection.Todos[0].Text)
	})

	t.Run("handles special characters in text", func(t *testing.T) {
		store := testutil.CreatePopulatedStore(t)
		specialText := `Special chars: !@#$%^&*()_+-={}[]|\:";'<>?,./` + "\n\t"

		opts := add.Options{CollectionPath: store.Path()}
		result, err := add.Execute(specialText, opts)

		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, specialText, result.Todo.Text)

		// Verify it was saved correctly
		collection, err := store.Load()
		testutil.AssertNoError(t, err)
		testutil.AssertTodoInList(t, collection.Todos, specialText)
	})

	t.Run("handles unicode text", func(t *testing.T) {
		store := testutil.CreatePopulatedStore(t)
		unicodeText := "Unicode: 你好世界 🌍 émojis 🚀"

		opts := add.Options{CollectionPath: store.Path()}
		result, err := add.Execute(unicodeText, opts)

		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, unicodeText, result.Todo.Text)

		// Verify it was saved correctly
		collection, err := store.Load()
		testutil.AssertNoError(t, err)
		testutil.AssertTodoInList(t, collection.Todos, unicodeText)
	})

	t.Run("returns wrapped error when store update fails", func(t *testing.T) {
		// Create a read-only directory to force a store error
		dir := t.TempDir()
		dbPath := filepath.Join(dir, "todos.json")

		// Create the file first
		err := os.WriteFile(dbPath, []byte("[]"), 0644)
		assert.NoError(t, err)

		// Make the directory read-only
		err = os.Chmod(dir, 0555)
		assert.NoError(t, err)
		defer func() { _ = os.Chmod(dir, 0755) }() // Restore permissions for cleanup

		opts := add.Options{CollectionPath: dbPath}
		result, err := add.Execute("Will fail", opts)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to add todo")
	})

	t.Run("adds multiple todos sequentially", func(t *testing.T) {
		store := testutil.CreatePopulatedStore(t)

		opts := add.Options{CollectionPath: store.Path()}

		// Add first todo
		result1, err := add.Execute("First", opts)
		testutil.AssertNoError(t, err)
		assert.Equal(t, int64(1), result1.Todo.ID)

		// Add second todo
		result2, err := add.Execute("Second", opts)
		testutil.AssertNoError(t, err)
		assert.Equal(t, int64(2), result2.Todo.ID)

		// Add third todo
		result3, err := add.Execute("Third", opts)
		testutil.AssertNoError(t, err)
		assert.Equal(t, int64(3), result3.Todo.ID)

		// Verify all are saved
		collection, err := store.Load()
		testutil.AssertNoError(t, err)
		testutil.AssertCollectionSize(t, collection, 3)
		assert.Equal(t, "First", collection.Todos[0].Text)
		assert.Equal(t, "Second", collection.Todos[1].Text)
		assert.Equal(t, "Third", collection.Todos[2].Text)
	})

	t.Run("handles whitespace-only text as empty", func(t *testing.T) {
		store := testutil.CreatePopulatedStore(t)

		testCases := []string{
			" ",
			"  ",
			"\t",
			"\n",
			" \t\n ",
		}

		for _, tc := range testCases {
			opts := add.Options{CollectionPath: store.Path()}
			result, err := add.Execute(tc, opts)

			// Should succeed - we only check for empty string, not whitespace
			testutil.AssertNoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tc, result.Todo.Text)
		}
	})

	t.Run("handles non-existent store path", func(t *testing.T) {
		// Create the parent directory first
		dir := filepath.Join(os.TempDir(), "test-add-dir")
		err := os.MkdirAll(dir, 0755)
		assert.NoError(t, err)
		defer func() { _ = os.RemoveAll(dir) }()

		nonExistentPath := filepath.Join(dir, "todos.json")

		opts := add.Options{CollectionPath: nonExistentPath}
		result, err := add.Execute("Test", opts)

		// Should succeed - store creates the file if it doesn't exist
		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Test", result.Todo.Text)

		// Verify file was created
		_, err = os.Stat(nonExistentPath)
		assert.NoError(t, err)
	})
}
