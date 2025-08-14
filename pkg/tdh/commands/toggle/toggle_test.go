package toggle_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/arthur-debert/tdh/pkg/tdh"
	"github.com/arthur-debert/tdh/pkg/tdh/models"
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

func TestToggleCommand(t *testing.T) {
	dbPath, cleanup := setupTestStore(t)
	defer cleanup()

	// Add a todo first
	addOpts := tdh.AddOptions{CollectionPath: dbPath}
	addResult, err := tdh.Add("Todo to toggle", addOpts)
	require.NoError(t, err)

	toggleOpts := tdh.ToggleOptions{CollectionPath: dbPath}
	toggleResult, err := tdh.Toggle(int(addResult.Todo.ID), toggleOpts)

	require.NoError(t, err)
	assert.Equal(t, string(models.StatusDone), toggleResult.NewStatus)

	// Verify it was saved
	s := store.NewStore(dbPath)
	collection, err := s.Load()
	require.NoError(t, err)

	// Find the todo by ID
	var found *models.Todo
	for _, todo := range collection.Todos {
		if todo.ID == addResult.Todo.ID {
			found = todo
			break
		}
	}
	require.NotNil(t, found, "todo not found")
	assert.Equal(t, models.StatusDone, found.Status)
}
