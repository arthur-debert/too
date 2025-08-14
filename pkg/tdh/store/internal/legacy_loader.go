package internal

import (
	"encoding/json"
	"time"

	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/google/uuid"
)

// LegacyTodoFormat represents the old todo format with int64 ID
type LegacyTodoFormat struct {
	ID       int64             `json:"id"`
	Text     string            `json:"text"`
	Status   models.TodoStatus `json:"status"`
	Modified time.Time         `json:"modified"`
}

// LoadTodosWithMigration attempts to load todos, handling both old and new formats
func LoadTodosWithMigration(data []byte) ([]*models.Todo, error) {
	// Handle empty data
	if len(data) == 0 {
		return []*models.Todo{}, nil
	}

	// First, try to load as new format
	var newFormatTodos []*models.Todo
	if err := json.Unmarshal(data, &newFormatTodos); err == nil {
		// Check if this is actually new format (has UUIDs and positions)
		if len(newFormatTodos) > 0 && newFormatTodos[0].Position > 0 && len(newFormatTodos[0].ID) > 10 {
			return newFormatTodos, nil
		}
	}

	// Try to load as legacy format
	var legacyTodos []LegacyTodoFormat
	if err := json.Unmarshal(data, &legacyTodos); err != nil {
		// If we can't parse as either format, return error
		return nil, err
	}

	// Convert legacy format to new format
	todos := make([]*models.Todo, 0, len(legacyTodos))
	for _, legacy := range legacyTodos {
		todo := &models.Todo{
			ID:       uuid.New().String(),
			Position: int(legacy.ID), // Use old ID as position
			Text:     legacy.Text,
			Status:   legacy.Status,
			Modified: legacy.Modified,
		}
		todos = append(todos, todo)
	}

	return todos, nil
}
