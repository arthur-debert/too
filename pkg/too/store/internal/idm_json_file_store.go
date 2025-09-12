package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/arthur-debert/too/pkg/too/models"
)

// IDMJSONFileStore implements pure IDM storage using a JSON file.
// This store saves and loads IDMCollection directly without any hierarchical conversion.
type IDMJSONFileStore struct {
	PathValue string
}

// legacyTodo represents the old hierarchical todo format
type legacyTodo struct {
	ID       string            `json:"id"`
	ParentID string            `json:"parentId"`
	Text     string            `json:"text"`
	Statuses map[string]string `json:"statuses"`
	Modified time.Time         `json:"modified"`
	Items    []legacyTodo      `json:"items"`
}

// legacyFormat represents the old collection format
type legacyFormat struct {
	Todos []legacyTodo `json:"todos"`
}

// migrateFromLegacy converts legacy hierarchical format to flat IDM format
func migrateFromLegacy(legacy legacyFormat) models.IDMCollection {
	collection := models.IDMCollection{
		Items: []*models.IDMTodo{},
	}
	
	// Recursively flatten the hierarchy
	var flatten func(todos []legacyTodo)
	flatten = func(todos []legacyTodo) {
		for _, todo := range todos {
			idmTodo := &models.IDMTodo{
				UID:      todo.ID,
				ParentID: todo.ParentID,
				Text:     todo.Text,
				Statuses: todo.Statuses,
				Modified: todo.Modified,
			}
			collection.Items = append(collection.Items, idmTodo)
			
			// Process children
			if len(todo.Items) > 0 {
				flatten(todo.Items)
			}
		}
	}
	
	flatten(legacy.Todos)
	return collection
}

// NewIDMJSONFileStore creates a new IDM JSON file store.
func NewIDMJSONFileStore(path string) *IDMJSONFileStore {
	return &IDMJSONFileStore{PathValue: path}
}

// LoadIDM reads the IDM collection from the JSON file.
func (s *IDMJSONFileStore) LoadIDM() (*models.IDMCollection, error) {
	// Check if file exists
	if _, err := os.Stat(s.PathValue); os.IsNotExist(err) {
		// Return empty collection if file doesn't exist
		return models.NewIDMCollection(), nil
	}

	// Read the file
	data, err := os.ReadFile(s.PathValue)
	if err != nil {
		return nil, fmt.Errorf("failed to read store file %s: %w", s.PathValue, err)
	}

	// Handle empty file
	if len(data) == 0 {
		return models.NewIDMCollection(), nil
	}

	// Try to unmarshal as IDMCollection directly
	var collection models.IDMCollection
	if err := json.Unmarshal(data, &collection); err != nil {
		// Try legacy format with hierarchical todos
		var legacyData legacyFormat
		if legacyErr := json.Unmarshal(data, &legacyData); legacyErr != nil {
			return nil, fmt.Errorf("failed to decode JSON from %s: %w (original error: %v)", s.PathValue, legacyErr, err)
		}
		// Migrate from legacy format
		collection = migrateFromLegacy(legacyData)
	} else {
		// Check if the collection has the old "todos" field instead of "items"
		// This indicates legacy format even if JSON unmarshal didn't fail
		if len(collection.Items) == 0 {
			// Try legacy format
			var legacyData legacyFormat
			if legacyErr := json.Unmarshal(data, &legacyData); legacyErr == nil && len(legacyData.Todos) > 0 {
				// Migrate from legacy format
				collection = migrateFromLegacy(legacyData)
			}
		}
	}

	// Ensure Items is not nil
	if collection.Items == nil {
		collection.Items = []*models.IDMTodo{}
	}

	return &collection, nil
}

// SaveIDM writes the IDM collection to the JSON file.
func (s *IDMJSONFileStore) SaveIDM(collection *models.IDMCollection) error {
	// Ensure directory exists
	dir := filepath.Dir(s.PathValue)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Marshal to JSON with indentation for readability
	data, err := json.MarshalIndent(collection, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal collection: %w", err)
	}

	// Write to temporary file first
	tempFile := s.PathValue + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary file: %w", err)
	}

	// Rename temporary file to actual file (atomic operation)
	if err := os.Rename(tempFile, s.PathValue); err != nil {
		// Clean up temp file if rename fails
		_ = os.Remove(tempFile) // Ignore error on cleanup
		return fmt.Errorf("failed to save file: %w", err)
	}

	return nil
}

// Exists checks if the store file exists.
func (s *IDMJSONFileStore) Exists() bool {
	_, err := os.Stat(s.PathValue)
	return err == nil
}

// Path returns the file path of the store.
func (s *IDMJSONFileStore) Path() string {
	return s.PathValue
}

// UpdateIDM performs an atomic update operation on the IDM collection.
func (s *IDMJSONFileStore) UpdateIDM(updateFn func(collection *models.IDMCollection) error) error {
	// Load current collection
	collection, err := s.LoadIDM()
	if err != nil {
		return fmt.Errorf("failed to load collection: %w", err)
	}

	// Apply update function
	if err := updateFn(collection); err != nil {
		return err
	}

	// Save updated collection
	if err := s.SaveIDM(collection); err != nil {
		return fmt.Errorf("failed to save collection: %w", err)
	}

	return nil
}

// FindItemByUID finds a todo item by its UID.
func (s *IDMJSONFileStore) FindItemByUID(uid string) (*models.IDMTodo, error) {
	collection, err := s.LoadIDM()
	if err != nil {
		return nil, err
	}

	item := collection.FindByUID(uid)
	if item == nil {
		return nil, fmt.Errorf("todo with UID %s not found", uid)
	}

	return item, nil
}

// FindItemByShortID finds a todo item by its short ID.
func (s *IDMJSONFileStore) FindItemByShortID(shortID string) (*models.IDMTodo, error) {
	collection, err := s.LoadIDM()
	if err != nil {
		return nil, err
	}

	var found *models.IDMTodo
	var count int

	for _, item := range collection.Items {
		if len(item.UID) >= len(shortID) && item.UID[:len(shortID)] == shortID {
			found = item
			count++
		}
	}

	if count == 0 {
		return nil, fmt.Errorf("no todo found with reference '%s'", shortID)
	}
	if count > 1 {
		return nil, fmt.Errorf("multiple todos found with ambiguous reference '%s'", shortID)
	}

	return found, nil
}