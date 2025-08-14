package internal

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/google/uuid"
)

// LegacyTodo represents the old todo format with int64 ID
type LegacyTodo struct {
	ID       int64             `json:"id"`
	Text     string            `json:"text"`
	Status   models.TodoStatus `json:"status"`
	Modified json.RawMessage   `json:"modified"`
}

// MigrateToUUIDAndPosition reads a JSON file and migrates todos from old format to new format
func MigrateToUUIDAndPosition(path string) error {
	// Check if file exists first
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}

	// Read the raw JSON data
	data, err := os.ReadFile(path)
	if err != nil {
		return nil // Skip migration on read errors
	}

	// Handle empty file
	if len(data) == 0 {
		return nil
	}

	// Try to decode as new format first
	var newTodos []*models.Todo
	if err := json.Unmarshal(data, &newTodos); err == nil {
		// Check if already migrated by looking for Position field
		if len(newTodos) > 0 && newTodos[0].Position > 0 {
			// Already migrated
			return nil
		}
	}

	// Decode as legacy format
	var legacyTodos []json.RawMessage
	if err := json.Unmarshal(data, &legacyTodos); err != nil {
		return fmt.Errorf("failed to decode legacy todos: %w", err)
	}

	// Convert to new format
	newTodos = make([]*models.Todo, 0, len(legacyTodos))
	for _, rawTodo := range legacyTodos {
		// Try to decode into a map to handle flexible format
		var todoMap map[string]interface{}
		if err := json.Unmarshal(rawTodo, &todoMap); err != nil {
			return fmt.Errorf("failed to decode todo: %w", err)
		}

		// Extract ID as float64 (JSON numbers are float64)
		var position int
		if idVal, ok := todoMap["id"]; ok {
			if idFloat, ok := idVal.(float64); ok {
				position = int(idFloat)
			}
		}

		todo := &models.Todo{
			ID:       uuid.New().String(),
			Position: position,
			Text:     "",
			Status:   models.StatusPending,
		}

		// Extract text
		if textVal, ok := todoMap["text"]; ok {
			if text, ok := textVal.(string); ok {
				todo.Text = text
			}
		}

		// Extract status
		if statusVal, ok := todoMap["status"]; ok {
			if status, ok := statusVal.(string); ok {
				todo.Status = models.TodoStatus(status)
			}
		}

		// Extract modified time
		if modifiedVal, ok := todoMap["modified"]; ok {
			modifiedBytes, _ := json.Marshal(modifiedVal)
			_ = json.Unmarshal(modifiedBytes, &todo.Modified)
		}

		newTodos = append(newTodos, todo)
	}

	// Create collection and save using the store
	collection := &models.Collection{Todos: newTodos}
	store := &JSONFileStore{PathValue: path}

	return store.Save(collection)
}
