package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/arthur-debert/too/pkg/too/models"
)

// JSONFileStore implements the Store interface using a JSON file.
type JSONFileStore struct {
	PathValue string
}

// NewJSONFileStore creates a new JSONFileStore.
func NewJSONFileStore(path string) *JSONFileStore {
	return &JSONFileStore{PathValue: path}
}

// Load reads the collection from the JSON file.
func (s *JSONFileStore) Load() (*models.Collection, error) {
	collection := models.NewCollection()

	// Read the entire file
	data, err := os.ReadFile(s.PathValue)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, start with empty collection
			return collection, nil
		}
		return nil, fmt.Errorf("failed to read store file %s: %w", s.PathValue, err)
	}

	// Use the new loader that handles both formats
	todos, err := LoadTodosWithMigration(data)
	if err != nil {
		// Handle empty file case
		if errors.Is(err, io.EOF) {
			return collection, nil
		}
		return nil, fmt.Errorf("failed to decode JSON from %s: %w", s.PathValue, err)
	}

	collection.Todos = todos

	// Ensure non-nil slice
	if collection.Todos == nil {
		collection.Todos = []*models.Todo{}
	}

	// Check if migration is needed by looking for todos without Items field initialized
	needsMigration := false
	for _, todo := range collection.Todos {
		if todo.Items == nil {
			needsMigration = true
			break
		}
	}

	// Migrate collection to support nested lists
	models.MigrateCollection(collection)

	// If migration was needed, save the migrated data back
	if needsMigration {
		if err := s.Save(collection); err != nil {
			// Log the error but don't fail the load
			// The migration will be attempted again next time
			// This prevents data loss if save fails
			fmt.Fprintf(os.Stderr, "Warning: Failed to save migrated data: %v\n", err)
		}
	}

	return collection, nil
}

// Save writes the collection to the JSON file atomically.
func (s *JSONFileStore) Save(collection *models.Collection) error {
	data, err := json.MarshalIndent(&collection.Todos, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal todos to JSON: %w", err)
	}

	// Atomic save: write to a temp file first
	dir := filepath.Dir(s.PathValue)
	tempFile, err := os.CreateTemp(dir, ".todos-*.json.tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file in %s: %w", dir, err)
	}
	tempPath := tempFile.Name()
	defer func() {
		// Clean up temp file if it still exists
		if _, err := os.Stat(tempPath); err == nil {
			_ = os.Remove(tempPath)
		}
	}()

	if _, err := tempFile.Write(data); err != nil {
		_ = tempFile.Close()
		return fmt.Errorf("failed to write data to temp file %s: %w", tempPath, err)
	}

	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file %s: %w", tempPath, err)
	}

	// Atomically replace the original file with the new one
	if err := os.Rename(tempPath, s.PathValue); err != nil {
		return fmt.Errorf("failed to atomically save file %s: %w", s.PathValue, err)
	}
	return nil
}

// Exists checks if the store file exists.
func (s *JSONFileStore) Exists() bool {
	_, err := os.Stat(s.PathValue)
	return !os.IsNotExist(err)
}

// Update performs a transactional update on the collection.
// If the update function returns an error, the collection is not modified.
func (s *JSONFileStore) Update(fn func(collection *models.Collection) error) error {
	collection, err := s.Load()
	if err != nil {
		return err
	}

	// Create a deep copy to ensure rollback safety
	clone := collection.Clone()

	if err := fn(clone); err != nil {
		// On error, discard the clone and return the error
		return err
	}

	// Only save if the update function succeeded
	return s.Save(clone)
}

// Path returns the file path where the store persists data.
func (s *JSONFileStore) Path() string {
	return s.PathValue
}

// Find retrieves todos based on the provided query.
// This implementation uses O(n) linear search through all todos, which is
// acceptable for typical todo list sizes. Future store implementations
// (e.g., SQLite) can optimize this with proper indexing.
func (s *JSONFileStore) Find(query Query) (*FindResult, error) {
	collection, err := s.Load()
	if err != nil {
		return nil, err
	}

	// Calculate counts from the full collection
	totalCount, doneCount := CountTodos(collection.Todos)

	return &FindResult{
		Todos:      query.FilterTodos(collection.Todos),
		TotalCount: totalCount,
		DoneCount:  doneCount,
	}, nil
}
