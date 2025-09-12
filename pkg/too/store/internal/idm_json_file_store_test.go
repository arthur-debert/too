package internal

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/arthur-debert/too/pkg/too/models"
)

func TestIDMJSONFileStore_SaveAndLoad(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()
	storePath := filepath.Join(tempDir, "test-idm.json")

	store := NewIDMJSONFileStore(storePath)

	// Create test data
	collection := models.NewIDMCollection()
	
	todo1 := models.NewIDMTodo("Task 1", "")
	todo2 := models.NewIDMTodo("Task 2", todo1.UID)
	todo3 := models.NewIDMTodo("Task 3", "")
	
	collection.AddItem(todo1)
	collection.AddItem(todo2)
	collection.AddItem(todo3)

	// Save collection
	err := store.SaveIDM(collection)
	if err != nil {
		t.Fatalf("Failed to save collection: %v", err)
	}

	// Verify file exists
	if !store.Exists() {
		t.Error("Store file should exist after save")
	}

	// Load collection
	loadedCollection, err := store.LoadIDM()
	if err != nil {
		t.Fatalf("Failed to load collection: %v", err)
	}

	// Verify data
	if loadedCollection.Count() != 3 {
		t.Errorf("Expected 3 items, got %d", loadedCollection.Count())
	}

	// Check specific todos
	item1 := loadedCollection.FindByUID(todo1.UID)
	if item1 == nil {
		t.Error("Todo 1 not found after load")
	} else if item1.Text != "Task 1" {
		t.Errorf("Expected text 'Task 1', got '%s'", item1.Text)
	}

	item2 := loadedCollection.FindByUID(todo2.UID)
	if item2 == nil {
		t.Error("Todo 2 not found after load")
	} else if item2.ParentID != todo1.UID {
		t.Errorf("Expected parent ID '%s', got '%s'", todo1.UID, item2.ParentID)
	}
}

func TestIDMJSONFileStore_EmptyCollection(t *testing.T) {
	tempDir := t.TempDir()
	storePath := filepath.Join(tempDir, "empty-idm.json")

	store := NewIDMJSONFileStore(storePath)

	// Load non-existent file should return empty collection
	collection, err := store.LoadIDM()
	if err != nil {
		t.Fatalf("Failed to load empty collection: %v", err)
	}

	if collection.Count() != 0 {
		t.Errorf("Expected 0 items, got %d", collection.Count())
	}

	// Save empty collection
	err = store.SaveIDM(collection)
	if err != nil {
		t.Fatalf("Failed to save empty collection: %v", err)
	}

	// Load again
	loadedCollection, err := store.LoadIDM()
	if err != nil {
		t.Fatalf("Failed to load saved empty collection: %v", err)
	}

	if loadedCollection.Count() != 0 {
		t.Errorf("Expected 0 items after load, got %d", loadedCollection.Count())
	}
}

func TestIDMJSONFileStore_LegacyMigration(t *testing.T) {
	tempDir := t.TempDir()
	storePath := filepath.Join(tempDir, "legacy.json")

	// Write legacy hierarchical format
	legacyJSON := `{
		"todos": [
			{
				"id": "123",
				"parentId": "",
				"text": "Parent",
				"statuses": {"completion": "pending"},
				"modified": "2024-01-01T00:00:00Z",
				"items": [
					{
						"id": "456",
						"parentId": "123",
						"text": "Child",
						"statuses": {"completion": "done"},
						"modified": "2024-01-02T00:00:00Z",
						"items": []
					}
				]
			}
		]
	}`

	err := os.WriteFile(storePath, []byte(legacyJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to write legacy file: %v", err)
	}

	// Load with IDM store
	store := NewIDMJSONFileStore(storePath)
	collection, err := store.LoadIDM()
	if err != nil {
		t.Fatalf("Failed to load legacy format: %v", err)
	}

	// Verify migration worked
	if collection.Count() != 2 {
		t.Errorf("Expected 2 items after migration, got %d", collection.Count())
	}

	parent := collection.FindByUID("123")
	if parent == nil {
		t.Error("Parent not found after migration")
	} else if parent.Text != "Parent" {
		t.Errorf("Expected parent text 'Parent', got '%s'", parent.Text)
	}

	child := collection.FindByUID("456")
	if child == nil {
		t.Error("Child not found after migration")
	} else {
		if child.Text != "Child" {
			t.Errorf("Expected child text 'Child', got '%s'", child.Text)
		}
		if child.ParentID != "123" {
			t.Errorf("Expected child parent ID '123', got '%s'", child.ParentID)
		}
		if child.GetStatus() != models.StatusDone {
			t.Errorf("Expected child status 'done', got '%s'", child.GetStatus())
		}
	}
}

func TestIDMJSONFileStore_UpdateIDM(t *testing.T) {
	tempDir := t.TempDir()
	storePath := filepath.Join(tempDir, "update-test.json")

	store := NewIDMJSONFileStore(storePath)

	// Create initial data
	collection := models.NewIDMCollection()
	todo := models.NewIDMTodo("Original text", "")
	collection.AddItem(todo)

	err := store.SaveIDM(collection)
	if err != nil {
		t.Fatalf("Failed to save initial collection: %v", err)
	}

	// Update using UpdateIDM
	err = store.UpdateIDM(func(c *models.IDMCollection) error {
		item := c.FindByUID(todo.UID)
		if item != nil {
			item.Text = "Updated text"
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Failed to update collection: %v", err)
	}

	// Load and verify
	loadedCollection, err := store.LoadIDM()
	if err != nil {
		t.Fatalf("Failed to load updated collection: %v", err)
	}

	updatedTodo := loadedCollection.FindByUID(todo.UID)
	if updatedTodo == nil {
		t.Error("Todo not found after update")
	} else if updatedTodo.Text != "Updated text" {
		t.Errorf("Expected text 'Updated text', got '%s'", updatedTodo.Text)
	}
}

func TestIDMJSONFileStore_FindByShortID(t *testing.T) {
	tempDir := t.TempDir()
	storePath := filepath.Join(tempDir, "find-test.json")

	store := NewIDMJSONFileStore(storePath)

	// Create test data with known UIDs
	collection := models.NewIDMCollection()
	
	// Create todos with predictable UIDs for testing
	todo1 := &models.IDMTodo{
		UID:      "1234567890abcdef",
		Text:     "Task 1",
		Statuses: map[string]string{"completion": "pending"},
	}
	todo2 := &models.IDMTodo{
		UID:      "1234567890fedcba",
		Text:     "Task 2",
		Statuses: map[string]string{"completion": "pending"},
	}
	todo3 := &models.IDMTodo{
		UID:      "abcdef1234567890",
		Text:     "Task 3",
		Statuses: map[string]string{"completion": "pending"},
	}

	collection.AddItem(todo1)
	collection.AddItem(todo2)
	collection.AddItem(todo3)

	err := store.SaveIDM(collection)
	if err != nil {
		t.Fatalf("Failed to save collection: %v", err)
	}

	// Test finding by short ID
	found, err := store.FindItemByShortID("abcdef")
	if err != nil {
		t.Fatalf("Failed to find by short ID: %v", err)
	}
	if found.UID != todo3.UID {
		t.Errorf("Expected to find todo3, got %s", found.UID)
	}

	// Test ambiguous short ID
	_, err = store.FindItemByShortID("1234567")
	if err == nil {
		t.Error("Expected error for ambiguous short ID")
	}

	// Test not found
	_, err = store.FindItemByShortID("zzz")
	if err == nil {
		t.Error("Expected error for non-existent short ID")
	}
}