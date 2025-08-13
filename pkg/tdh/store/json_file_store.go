package store

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/arthur-debert/tdh/pkg/models"
)

// JSONFileStore implements the Store interface using a JSON file.
type JSONFileStore struct {
	Path string
}

// NewJSONFileStore creates a new JSONFileStore.
func NewJSONFileStore(path string) *JSONFileStore {
	return &JSONFileStore{Path: path}
}

// Load reads the collection from the JSON file.
func (s *JSONFileStore) Load() (*models.Collection, error) {
	collection := models.NewCollection(s.Path)

	file, err := os.OpenFile(s.Path, os.O_RDONLY, 0600)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, start with empty collection
			return collection, nil
		}
		return nil, err
	}
	defer func() { _ = file.Close() }()

	err = json.NewDecoder(file).Decode(&collection.Todos)
	if err != nil {
		// Handle empty file case
		if err.Error() == "EOF" {
			return collection, nil
		}
		return nil, err
	}

	// Ensure non-nil slice
	if collection.Todos == nil {
		collection.Todos = []*models.Todo{}
	}
	return collection, nil
}

// Save writes the collection to the JSON file atomically.
func (s *JSONFileStore) Save(collection *models.Collection) error {
	data, err := json.MarshalIndent(&collection.Todos, "", "  ")
	if err != nil {
		return err
	}

	// Atomic save: write to a temp file first
	tempFile, err := os.CreateTemp(filepath.Dir(s.Path), ".tdh-*.json.tmp")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(tempFile.Name()) }() // Clean up temp file

	if _, err := tempFile.Write(data); err != nil {
		_ = tempFile.Close()
		return err
	}

	if err := tempFile.Close(); err != nil {
		return err
	}

	// Atomically replace the original file with the new one
	return os.Rename(tempFile.Name(), s.Path)
}

// Exists checks if the store file exists.
func (s *JSONFileStore) Exists() bool {
	_, err := os.Stat(s.Path)
	return !os.IsNotExist(err)
}

// Update performs a transactional update on the collection.
func (s *JSONFileStore) Update(fn func(collection *models.Collection) error) error {
	collection, err := s.Load()
	if err != nil {
		return err
	}

	if err := fn(collection); err != nil {
		return err
	}

	return s.Save(collection)
}
