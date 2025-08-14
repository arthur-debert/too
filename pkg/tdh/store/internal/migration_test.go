package internal

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrateToUUIDAndPosition(t *testing.T) {
	t.Run("migrates legacy todos to new format", func(t *testing.T) {
		dir := t.TempDir()
		dbPath := filepath.Join(dir, "test.json")

		// Create legacy format file
		legacyData := []map[string]interface{}{
			{
				"id":       1,
				"text":     "First todo",
				"status":   "pending",
				"modified": time.Now().Format(time.RFC3339),
			},
			{
				"id":       5,
				"text":     "Second todo",
				"status":   "done",
				"modified": time.Now().Format(time.RFC3339),
			},
		}

		data, err := json.Marshal(legacyData)
		require.NoError(t, err)
		err = os.WriteFile(dbPath, data, 0600)
		require.NoError(t, err)

		// Run migration
		err = MigrateToUUIDAndPosition(dbPath)
		require.NoError(t, err)

		// Load and verify
		store := NewJSONFileStore(dbPath)
		collection, err := store.Load()
		require.NoError(t, err)

		assert.Len(t, collection.Todos, 2)

		// First todo
		assert.NotEmpty(t, collection.Todos[0].ID)
		assert.Equal(t, 1, collection.Todos[0].Position)
		assert.Equal(t, "First todo", collection.Todos[0].Text)
		assert.Equal(t, models.StatusPending, collection.Todos[0].Status)

		// Second todo
		assert.NotEmpty(t, collection.Todos[1].ID)
		assert.Equal(t, 5, collection.Todos[1].Position)
		assert.Equal(t, "Second todo", collection.Todos[1].Text)
		assert.Equal(t, models.StatusDone, collection.Todos[1].Status)
	})

	t.Run("skips migration if already migrated", func(t *testing.T) {
		dir := t.TempDir()
		dbPath := filepath.Join(dir, "test.json")

		// Create new format file
		collection := &models.Collection{
			Todos: []*models.Todo{
				{
					ID:       "uuid-1",
					Position: 1,
					Text:     "Already migrated",
					Status:   models.StatusPending,
					Modified: time.Now(),
				},
			},
		}

		data, err := json.Marshal(collection.Todos)
		require.NoError(t, err)
		err = os.WriteFile(dbPath, data, 0600)
		require.NoError(t, err)

		originalData, _ := os.ReadFile(dbPath)

		// Run migration
		err = MigrateToUUIDAndPosition(dbPath)
		require.NoError(t, err)

		// Verify file unchanged
		newData, _ := os.ReadFile(dbPath)
		assert.Equal(t, originalData, newData)
	})

	t.Run("handles empty file", func(t *testing.T) {
		dir := t.TempDir()
		dbPath := filepath.Join(dir, "empty.json")

		err := os.WriteFile(dbPath, []byte(""), 0600)
		require.NoError(t, err)

		err = MigrateToUUIDAndPosition(dbPath)
		require.NoError(t, err)
	})

	t.Run("handles non-existent file", func(t *testing.T) {
		err := MigrateToUUIDAndPosition("/non/existent/file.json")
		require.NoError(t, err)
	})
}
