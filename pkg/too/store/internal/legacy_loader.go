package internal

import (
	"encoding/json"
	"time"

	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/google/uuid"
)

// LegacyTodoFormat represents the old todo format with int64 ID
type LegacyTodoFormat struct {
	ID       int64             `json:"id"`
	Text     string            `json:"text"`
	Status   models.TodoStatus `json:"status"`
	Modified time.Time         `json:"modified"`
}

// MixedIDTodoFormat represents todos that might have string or numeric IDs
type MixedIDTodoFormat struct {
	ID       interface{}       `json:"id"`       // Can be string or number
	Position int               `json:"position"` // Optional position field
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

	// First, try to load as new format (with Items field)
	var newFormatTodos []*models.Todo
	if err := json.Unmarshal(data, &newFormatTodos); err == nil {
		// Check if this is actually new format (has Items field initialized)
		if len(newFormatTodos) > 0 && newFormatTodos[0].Items != nil {
			return newFormatTodos, nil
		}
	}

	// Try to load with mixed ID format (could be string or numeric IDs)
	var mixedTodos []MixedIDTodoFormat
	if err := json.Unmarshal(data, &mixedTodos); err == nil {
		// Convert mixed format to new format
		todos := make([]*models.Todo, 0, len(mixedTodos))
		for i, mixed := range mixedTodos {
			todo := &models.Todo{
				Text:     mixed.Text,
				Status:   mixed.Status,
				Modified: mixed.Modified,
				Position: mixed.Position,
			}

			// Handle different ID types
			switch id := mixed.ID.(type) {
			case string:
				todo.ID = id
			case float64:
				// JSON numbers come as float64
				todo.ID = uuid.New().String()
				if mixed.Position == 0 {
					todo.Position = int(id)
				}
			default:
				todo.ID = uuid.New().String()
			}

			// Ensure position is set
			if todo.Position == 0 {
				todo.Position = i + 1
			}

			// Ensure ID is set
			if todo.ID == "" {
				todo.ID = uuid.New().String()
			}

			todos = append(todos, todo)
		}
		return todos, nil
	}

	// Try to load as pure legacy format with int64 IDs
	var legacyTodos []LegacyTodoFormat
	if err := json.Unmarshal(data, &legacyTodos); err != nil {
		// If we can't parse as any format, return error
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
