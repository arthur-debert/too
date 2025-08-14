package init_test

import (
	"os"
	"path/filepath"
	"testing"

	cmdinit "github.com/arthur-debert/tdh/pkg/tdh/commands/init"
	"github.com/arthur-debert/tdh/pkg/tdh/store"
	"github.com/arthur-debert/tdh/pkg/tdh/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitCommand(t *testing.T) {
	t.Run("creates new todo file when none exists", func(t *testing.T) {
		// Setup temp directory
		dir := t.TempDir()
		dbPath := filepath.Join(dir, "todos.json")

		// Execute init command
		opts := cmdinit.Options{DBPath: dbPath}
		result, err := cmdinit.Execute(opts)

		// Verify results
		testutil.AssertNoError(t, err)
		assert.True(t, result.Created)
		assert.Equal(t, dbPath, result.DBPath)
		assert.Contains(t, result.Message, "Initialized empty tdh collection")

		// Verify file was created with empty collection
		s := store.NewStore(dbPath)
		collection, err := s.Load()
		testutil.AssertNoError(t, err)
		testutil.AssertCollectionSize(t, collection, 0)
	})

	t.Run("reinitializes when file already exists", func(t *testing.T) {
		// Create a populated store using testutil
		s := testutil.CreatePopulatedStore(t, "Existing todo 1", "Existing todo 2")
		dbPath := s.Path()

		// Execute init command on existing file
		opts := cmdinit.Options{DBPath: dbPath}
		result, err := cmdinit.Execute(opts)

		// Verify results
		testutil.AssertNoError(t, err)
		assert.False(t, result.Created)
		assert.Equal(t, dbPath, result.DBPath)
		assert.Contains(t, result.Message, "Reinitialized existing tdh collection")

		// Verify existing todos are preserved
		collection, err := s.Load()
		testutil.AssertNoError(t, err)
		testutil.AssertCollectionSize(t, collection, 2)
		testutil.AssertTodoInList(t, collection.Todos, "Existing todo 1")
		testutil.AssertTodoInList(t, collection.Todos, "Existing todo 2")
	})

	t.Run("handles write errors gracefully", func(t *testing.T) {
		// Create a directory with no write permissions
		dir := t.TempDir()
		dbPath := filepath.Join(dir, "readonly", "todos.json")

		// Create parent directory with read-only permissions
		err := os.Mkdir(filepath.Dir(dbPath), 0555)
		require.NoError(t, err)

		// Execute init command
		opts := cmdinit.Options{DBPath: dbPath}
		result, err := cmdinit.Execute(opts)

		// Verify error handling
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to create store file")
	})
}
