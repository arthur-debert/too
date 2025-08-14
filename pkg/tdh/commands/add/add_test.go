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
		assert.Equal(t, 1, result.Todo.Position)
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
		assert.Equal(t, 3, result.Todo.Position) // Should be 3 after two existing

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
		assert.Equal(t, 1, result1.Todo.Position)

		// Add second todo
		result2, err := add.Execute("Second", opts)
		testutil.AssertNoError(t, err)
		assert.Equal(t, 2, result2.Todo.Position)

		// Add third todo
		result3, err := add.Execute("Third", opts)
		testutil.AssertNoError(t, err)
		assert.Equal(t, 3, result3.Todo.Position)

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

	t.Run("adds todo with parent", func(t *testing.T) {
		// Create store with a parent todo
		store := testutil.CreatePopulatedStore(t, "Parent todo")

		opts := add.Options{
			CollectionPath: store.Path(),
			ParentPath:     "1",
		}
		result, err := add.Execute("Child todo", opts)

		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Child todo", result.Todo.Text)
		assert.Equal(t, 1, result.Todo.Position) // First child of parent

		// Verify the structure
		collection, err := store.Load()
		testutil.AssertNoError(t, err)
		testutil.AssertCollectionSize(t, collection, 1) // One top-level todo

		parent := collection.Todos[0]
		assert.Equal(t, "Parent todo", parent.Text)
		assert.Len(t, parent.Items, 1)
		assert.Equal(t, "Child todo", parent.Items[0].Text)
		assert.Equal(t, parent.ID, parent.Items[0].ParentID)
	})

	t.Run("adds todo with nested parent path", func(t *testing.T) {
		// Create a nested structure using file store
		store := testutil.CreatePopulatedStore(t) // Empty store

		// Create the initial structure using the add command itself
		// First, create the parent
		opts1 := add.Options{
			CollectionPath: store.Path(),
			ParentPath:     "",
		}
		result1, err := add.Execute("Level 1", opts1)
		testutil.AssertNoError(t, err)
		assert.NotNil(t, result1)

		// Then create a child
		opts2 := add.Options{
			CollectionPath: store.Path(),
			ParentPath:     "1",
		}
		result2, err := add.Execute("Level 2", opts2)
		testutil.AssertNoError(t, err)
		assert.NotNil(t, result2)

		// Verify the structure before adding grandchild
		collection, err := store.Load()
		testutil.AssertNoError(t, err)
		t.Logf("Total top-level todos: %d", len(collection.Todos))
		if len(collection.Todos) > 0 {
			t.Logf("Parent position: %d, has %d items", collection.Todos[0].Position, len(collection.Todos[0].Items))
			if len(collection.Todos[0].Items) > 0 {
				t.Logf("Child position: %d", collection.Todos[0].Items[0].Position)
			}
		}

		// Add a grandchild
		opts := add.Options{
			CollectionPath: store.Path(),
			ParentPath:     "1.1",
		}
		result, err := add.Execute("Level 3", opts)

		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Level 3", result.Todo.Text)

		// Verify the structure
		collection, err = store.Load()
		testutil.AssertNoError(t, err)

		assert.Len(t, collection.Todos[0].Items[0].Items, 1)
		grandchild := collection.Todos[0].Items[0].Items[0]
		assert.Equal(t, "Level 3", grandchild.Text)
		assert.Equal(t, collection.Todos[0].Items[0].ID, grandchild.ParentID)
	})

	t.Run("returns error for invalid parent path", func(t *testing.T) {
		store := testutil.CreatePopulatedStore(t, "Parent todo")

		invalidPaths := []string{
			"99",    // Non-existent position
			"1.99",  // Non-existent child
			"abc",   // Invalid format
			"0",     // Invalid position (must be >= 1)
			"-1",    // Negative position
			"1.2.3", // Path too deep for structure
		}

		for _, path := range invalidPaths {
			opts := add.Options{
				CollectionPath: store.Path(),
				ParentPath:     path,
			}
			result, err := add.Execute("Child todo", opts)

			assert.Error(t, err, "Expected error for path: %s", path)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), "invalid parent path")
		}
	})

	t.Run("adds multiple children to same parent", func(t *testing.T) {
		store := testutil.CreatePopulatedStore(t, "Parent todo")

		// Add first child
		opts := add.Options{
			CollectionPath: store.Path(),
			ParentPath:     "1",
		}
		result1, err := add.Execute("First child", opts)
		testutil.AssertNoError(t, err)
		assert.Equal(t, 1, result1.Todo.Position)

		// Add second child
		result2, err := add.Execute("Second child", opts)
		testutil.AssertNoError(t, err)
		assert.Equal(t, 2, result2.Todo.Position)

		// Add third child
		result3, err := add.Execute("Third child", opts)
		testutil.AssertNoError(t, err)
		assert.Equal(t, 3, result3.Todo.Position)

		// Verify the structure
		collection, err := store.Load()
		testutil.AssertNoError(t, err)

		parent := collection.Todos[0]
		assert.Len(t, parent.Items, 3)
		assert.Equal(t, "First child", parent.Items[0].Text)
		assert.Equal(t, "Second child", parent.Items[1].Text)
		assert.Equal(t, "Third child", parent.Items[2].Text)
	})

	t.Run("preserves existing structure when adding with parent", func(t *testing.T) {
		// Create complex initial structure
		store := testutil.CreatePopulatedStore(t) // Empty store

		// Create two top-level todos
		opts1 := add.Options{CollectionPath: store.Path()}
		_, err := add.Execute("Todo 1", opts1)
		testutil.AssertNoError(t, err)

		_, err = add.Execute("Todo 2", opts1)
		testutil.AssertNoError(t, err)

		// Add children to first todo
		opts2 := add.Options{
			CollectionPath: store.Path(),
			ParentPath:     "1",
		}
		_, err = add.Execute("Child 1.1", opts2)
		testutil.AssertNoError(t, err)

		_, err = add.Execute("Child 1.2", opts2)
		testutil.AssertNoError(t, err)

		// Add a new child to the second todo
		opts3 := add.Options{
			CollectionPath: store.Path(),
			ParentPath:     "2",
		}
		_, err = add.Execute("Child 2.1", opts3)
		testutil.AssertNoError(t, err)

		// Verify structure is preserved
		collection, err := store.Load()
		testutil.AssertNoError(t, err)

		// First todo should be unchanged
		assert.Len(t, collection.Todos[0].Items, 2)
		assert.Equal(t, "Child 1.1", collection.Todos[0].Items[0].Text)
		assert.Equal(t, "Child 1.2", collection.Todos[0].Items[1].Text)

		// Second todo should have new child
		assert.Len(t, collection.Todos[1].Items, 1)
		assert.Equal(t, "Child 2.1", collection.Todos[1].Items[0].Text)
	})

	t.Run("handles empty parent path as root level", func(t *testing.T) {
		store := testutil.CreatePopulatedStore(t, "Existing todo")

		opts := add.Options{
			CollectionPath: store.Path(),
			ParentPath:     "", // Empty parent path
		}
		result, err := add.Execute("New root todo", opts)

		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 2, result.Todo.Position) // Second root-level todo

		// Verify it's at root level
		collection, err := store.Load()
		testutil.AssertNoError(t, err)
		testutil.AssertCollectionSize(t, collection, 2)
		assert.Equal(t, "New root todo", collection.Todos[1].Text)
		assert.Empty(t, collection.Todos[1].ParentID)
	})
}
