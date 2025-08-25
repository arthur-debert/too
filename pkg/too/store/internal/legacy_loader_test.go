package internal

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadTodosWithMigration(t *testing.T) {
	t.Run("loads new format correctly", func(t *testing.T) {
		// Create new format data
		todos := []*models.Todo{
			{
				ID:       "550e8400-e29b-41d4-a716-446655440000",
				Position: 1,
				Text:     "Test todo",
				Status:   models.StatusPending,
				Modified: time.Now(),
			},
		}
		data, err := json.Marshal(todos)
		require.NoError(t, err)

		// Load it
		loaded, err := LoadTodosWithMigration(data)
		require.NoError(t, err)

		assert.Len(t, loaded, 1)
		assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", loaded[0].ID)
		assert.Equal(t, 1, loaded[0].Position)
		assert.Equal(t, "Test todo", loaded[0].Text)
	})

	t.Run("migrates legacy format transparently", func(t *testing.T) {
		// Create legacy format data
		legacyData := []LegacyTodoFormat{
			{
				ID:       1,
				Text:     "First todo",
				Status:   models.StatusPending,
				Modified: time.Now(),
			},
			{
				ID:       5,
				Text:     "Second todo",
				Status:   models.StatusDone,
				Modified: time.Now(),
			},
		}
		data, err := json.Marshal(legacyData)
		require.NoError(t, err)

		// Load it
		loaded, err := LoadTodosWithMigration(data)
		require.NoError(t, err)

		assert.Len(t, loaded, 2)

		// First todo
		assert.NotEmpty(t, loaded[0].ID) // Should have UUID
		assert.Equal(t, 1, loaded[0].Position)
		assert.Equal(t, "First todo", loaded[0].Text)
		assert.Equal(t, models.StatusPending, loaded[0].Status)

		// Second todo
		assert.NotEmpty(t, loaded[1].ID) // Should have UUID
		assert.Equal(t, 5, loaded[1].Position)
		assert.Equal(t, "Second todo", loaded[1].Text)
		assert.Equal(t, models.StatusDone, loaded[1].Status)
	})

	t.Run("handles empty data", func(t *testing.T) {
		loaded, err := LoadTodosWithMigration([]byte{})
		require.NoError(t, err)
		assert.Empty(t, loaded)
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		loaded, err := LoadTodosWithMigration([]byte("invalid json"))
		assert.Error(t, err)
		assert.Nil(t, loaded)
	})

	t.Run("handles mixed format gracefully", func(t *testing.T) {
		// This tests the edge case where the data might be partially migrated
		// The loader should handle it correctly
		mixedData := []map[string]interface{}{
			{
				"id":       "550e8400-e29b-41d4-a716-446655440000",
				"position": 1,
				"text":     "New format",
				"status":   "pending",
				"modified": time.Now().Format(time.RFC3339),
			},
		}
		data, err := json.Marshal(mixedData)
		require.NoError(t, err)

		loaded, err := LoadTodosWithMigration(data)
		require.NoError(t, err)
		assert.Len(t, loaded, 1)
	})
}
