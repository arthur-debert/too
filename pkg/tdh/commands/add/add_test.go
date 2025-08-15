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
}

func TestAddCommandWithParent(t *testing.T) {
	t.Run("adds sub-todo to parent", func(t *testing.T) {
		// Create store with parent todo
		store := testutil.CreatePopulatedStore(t, "Parent task")

		// Add sub-task
		opts := add.Options{
			CollectionPath: store.Path(),
			ParentPath:     "1",
		}
		result, err := add.Execute("Sub-task", opts)

		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Sub-task", result.Todo.Text)
		assert.Equal(t, 1, result.Todo.Position) // First child

		// Verify structure
		collection, err := store.Load()
		testutil.AssertNoError(t, err)

		parent := collection.Todos[0]
		assert.Len(t, parent.Items, 1)
		assert.Equal(t, "Sub-task", parent.Items[0].Text)
		assert.Equal(t, parent.ID, parent.Items[0].ParentID)
	})

	t.Run("adds multiple sub-todos to same parent", func(t *testing.T) {
		store := testutil.CreatePopulatedStore(t, "Parent task")

		// Add first child
		opts := add.Options{
			CollectionPath: store.Path(),
			ParentPath:     "1",
		}
		_, err := add.Execute("Child 1", opts)
		testutil.AssertNoError(t, err)

		// Add second child
		result2, err := add.Execute("Child 2", opts)
		testutil.AssertNoError(t, err)
		assert.Equal(t, 2, result2.Todo.Position) // Second child

		// Add third child
		result3, err := add.Execute("Child 3", opts)
		testutil.AssertNoError(t, err)
		assert.Equal(t, 3, result3.Todo.Position) // Third child

		// Verify structure
		collection, err := store.Load()
		testutil.AssertNoError(t, err)

		parent := collection.Todos[0]
		assert.Len(t, parent.Items, 3)
		assert.Equal(t, "Child 1", parent.Items[0].Text)
		assert.Equal(t, "Child 2", parent.Items[1].Text)
		assert.Equal(t, "Child 3", parent.Items[2].Text)
	})

	t.Run("adds nested sub-todo using position path", func(t *testing.T) {
		// Create nested structure
		store := testutil.CreateNestedStore(t)

		// Add to nested position 1.2 (second child of first parent)
		opts := add.Options{
			CollectionPath: store.Path(),
			ParentPath:     "1.2",
		}
		result, err := add.Execute("New grandchild", opts)

		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 2, result.Todo.Position) // Second grandchild

		// Verify structure
		collection, err := store.Load()
		testutil.AssertNoError(t, err)

		parent := collection.Todos[0]
		subTask := parent.Items[1]      // Position 2 = second child
		assert.Len(t, subTask.Items, 2) // Had 1, now has 2
		assert.Equal(t, "New grandchild", subTask.Items[1].Text)
	})

	t.Run("returns error for non-existent parent", func(t *testing.T) {
		store := testutil.CreatePopulatedStore(t, "Only todo")

		opts := add.Options{
			CollectionPath: store.Path(),
			ParentPath:     "99",
		}
		result, err := add.Execute("Orphan task", opts)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "parent todo not found")

		// Verify nothing was added
		collection, err := store.Load()
		testutil.AssertNoError(t, err)
		testutil.AssertCollectionSize(t, collection, 1)
	})

	t.Run("returns error for invalid parent path", func(t *testing.T) {
		store := testutil.CreatePopulatedStore(t, "Parent")

		testCases := []struct {
			name       string
			parentPath string
			errMsg     string
		}{
			{"invalid format", "abc", "invalid position"},
			{"negative position", "-1", "position must be >= 1"},
			{"zero position", "0", "position must be >= 1"},
			{"trailing dot", "1.", "invalid position"},
			{"leading dot", ".1", "invalid position"},
			{"empty path", "", ""}, // Empty is valid (no parent)
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				opts := add.Options{
					CollectionPath: store.Path(),
					ParentPath:     tc.parentPath,
				}
				result, err := add.Execute("Test task", opts)

				if tc.errMsg != "" {
					assert.Error(t, err)
					assert.Nil(t, result)
					assert.Contains(t, err.Error(), tc.errMsg)
				} else {
					// Empty parent path should succeed
					testutil.AssertNoError(t, err)
					assert.NotNil(t, result)
				}
			})
		}
	})

	t.Run("deeply nested parent paths work", func(t *testing.T) {
		// Create a deeply nested structure manually
		store := testutil.CreatePopulatedStore(t)

		// Build: 1 -> 1.1 -> 1.1.1 -> 1.1.1.1
		collection := models.NewCollection()
		parent, _ := collection.CreateTodo("Level 1", "")
		child, _ := collection.CreateTodo("Level 2", parent.ID)
		grandchild, _ := collection.CreateTodo("Level 3", child.ID)
		_, _ = collection.CreateTodo("Level 4", grandchild.ID)

		err := store.Save(collection)
		testutil.AssertNoError(t, err)

		// Add to the deepest level
		opts := add.Options{
			CollectionPath: store.Path(),
			ParentPath:     "1.1.1.1",
		}
		result, err := add.Execute("Level 5", opts)

		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)

		// Verify it was added at the right place
		collection, err = store.Load()
		testutil.AssertNoError(t, err)

		// Navigate to level 4
		level1 := collection.Todos[0]
		level2 := level1.Items[0]
		level3 := level2.Items[0]
		level4 := level3.Items[0]

		assert.Len(t, level4.Items, 1)
		assert.Equal(t, "Level 5", level4.Items[0].Text)
	})

	t.Run("parent path with gaps returns error", func(t *testing.T) {
		store := testutil.CreateNestedStore(t)

		// Try to add to 1.99 (parent 1 doesn't have 99 children)
		opts := add.Options{
			CollectionPath: store.Path(),
			ParentPath:     "1.99",
		}
		result, err := add.Execute("Invalid child", opts)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "no item found at position 99")
	})
}
