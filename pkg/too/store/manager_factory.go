package store

import (
	"fmt"
	"os"

	"github.com/arthur-debert/too/pkg/too/config"
	"github.com/arthur-debert/too/pkg/too/models"
)

// WorkflowManager defines the common interface for workflow managers.
// Both DirectWorkflowManager and PureIDMManager implement this interface.
type WorkflowManager interface {
	// Core operations
	Add(parentUID, text string) (string, error)
	SetStatus(uid, dimension, value string) error
	GetStatus(uid, dimension string) (string, error)
	ResolvePositionPath(scope, path string) (string, error)
	Save() error
	Move(uid, oldParentUID, newParentUID string) error
	GetPositionPath(scope, uid string) (string, error)
	
	// Query operations
	ListActive() interface{} // Returns []*models.Todo or []*models.IDMTodo
	ListArchived() interface{} // Returns []*models.Todo or []*models.IDMTodo
	ListAll() interface{} // Returns []*models.Todo or []*models.IDMTodo
	GetTodoByID(uid string) interface{} // Returns *models.Todo or *models.IDMTodo
	GetTodoByShortID(shortID string) (interface{}, error) // Returns *models.Todo or *models.IDMTodo
	CountTodos() (totalCount, doneCount int)
	CleanFinishedTodos() (interface{}, int, error) // Returns cleaned todos and count
	
	// Type detection
	IsPureIDM() bool // Returns true if this is a PureIDMManager
}

// CreateWorkflowManager creates the appropriate workflow manager based on configuration.
// It automatically detects whether to use pure IDM or traditional hierarchical storage.
func CreateWorkflowManager(collectionPath string) (WorkflowManager, error) {
	// Load IDM configuration
	idmConfig, err := config.LoadIDMConfig(collectionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load IDM config: %w", err)
	}

	// Check if the file exists and what format it's in
	if _, err := os.Stat(collectionPath); err == nil {
		// File exists - detect format
		data, err := os.ReadFile(collectionPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read collection file: %w", err)
		}

		// Auto-detect format
		if len(data) > 0 {
			// Check for "items" field (pure IDM) vs "todos" field (hierarchical)
			if hasItemsField(data) && !hasTodosField(data) {
				// Pure IDM format detected
				idmStore := NewIDMStore(collectionPath)
				return NewPureIDMManager(idmStore, collectionPath)
			} else if hasTodosField(data) {
				// Traditional hierarchical format detected
				store := NewStore(collectionPath)
				return NewDirectWorkflowManager(store, collectionPath)
			}
		}
	}

	// Use configuration preference or default
	if idmConfig.UsePureIDM {
		// Use pure IDM storage
		idmStore := NewIDMStore(collectionPath)
		return NewPureIDMManager(idmStore, collectionPath)
	} else {
		// Use traditional hierarchical storage with DirectWorkflowManager
		store := NewStore(collectionPath)
		return NewDirectWorkflowManager(store, collectionPath)
	}
}

// hasItemsField checks if JSON data contains "items" field (IDM format).
func hasItemsField(data []byte) bool {
	// Simple check - could be made more robust
	return len(data) > 10 && contains(data, []byte(`"items"`))
}

// hasTodosField checks if JSON data contains "todos" field (hierarchical format).
func hasTodosField(data []byte) bool {
	// Handle empty array case (legacy format)
	if len(data) > 0 && data[0] == '[' {
		return true // Arrays are legacy format
	}
	// Check for "todos" field
	return len(data) > 10 && contains(data, []byte(`"todos"`))
}

// contains checks if data contains the needle bytes.
func contains(data, needle []byte) bool {
	if len(needle) > len(data) {
		return false
	}
	for i := 0; i <= len(data)-len(needle); i++ {
		if bytesEqual(data[i:i+len(needle)], needle) {
			return true
		}
	}
	return false
}

// bytesEqual compares two byte slices.
func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// EnablePureIDM updates the configuration to use pure IDM storage.
func EnablePureIDM(collectionPath string) error {
	// First migrate the data if needed
	store := NewStore(collectionPath)
	if store.Exists() {
		// Load existing data
		collection, err := store.Load()
		if err != nil {
			return fmt.Errorf("failed to load existing collection: %w", err)
		}

		// Create backup
		backupPath := collectionPath + ".backup"
		data, err := os.ReadFile(collectionPath)
		if err != nil {
			return fmt.Errorf("failed to read original file: %w", err)
		}
		if err := os.WriteFile(backupPath, data, 0644); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}

		// Save as pure IDM
		idmStore := NewIDMStore(collectionPath) 
		idmCollection := models.MigrateToIDM(collection)
		if err := idmStore.SaveIDM(idmCollection); err != nil {
			// Restore backup on failure
			os.Rename(backupPath, collectionPath)
			return fmt.Errorf("failed to save as IDM: %w", err)
		}
	}

	// Update configuration
	idmConfig := &config.IDMConfig{UsePureIDM: true}
	return config.SaveIDMConfig(collectionPath, idmConfig)
}