package add_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/arthur-debert/tdh/pkg/tdh"
	"github.com/arthur-debert/tdh/pkg/tdh/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestStore creates a temporary directory and a JSONFileStore for testing.
func setupTestStore(t *testing.T) (string, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "tdh-cmd-test")
	require.NoError(t, err)
	dbPath := filepath.Join(dir, "test.json")
	return dbPath, func() {
		err := os.RemoveAll(dir)
		require.NoError(t, err)
	}
}

func TestAddCommand(t *testing.T) {
	dbPath, cleanup := setupTestStore(t)
	defer cleanup()

	opts := tdh.AddOptions{CollectionPath: dbPath}
	result, err := tdh.Add("My first todo", opts)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "My first todo", result.Todo.Text)
	assert.Equal(t, int64(1), result.Todo.ID)

	// Verify it was saved
	s := store.NewStore(dbPath)
	collection, err := s.Load()
	require.NoError(t, err)
	assert.Len(t, collection.Todos, 1)
}
