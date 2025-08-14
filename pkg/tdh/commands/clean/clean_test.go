package clean_test

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

func TestCleanCommand(t *testing.T) {
	dbPath, cleanup := setupTestStore(t)
	defer cleanup()

	// Add some todos
	_, err := tdh.Add("Pending todo", tdh.AddOptions{CollectionPath: dbPath})
	require.NoError(t, err)
	addResult, err := tdh.Add("Done todo", tdh.AddOptions{CollectionPath: dbPath})
	require.NoError(t, err)
	_, err = tdh.Toggle(int(addResult.Todo.ID), tdh.ToggleOptions{CollectionPath: dbPath})
	require.NoError(t, err)

	cleanOpts := tdh.CleanOptions{CollectionPath: dbPath}
	cleanResult, err := tdh.Clean(cleanOpts)

	require.NoError(t, err)
	assert.Equal(t, 1, cleanResult.RemovedCount)
	assert.Equal(t, 1, cleanResult.ActiveCount)

	// Verify
	s := store.NewStore(dbPath)
	collection, err := s.Load()
	require.NoError(t, err)
	assert.Len(t, collection.Todos, 1)
	assert.Equal(t, "Pending todo", collection.Todos[0].Text)
}
