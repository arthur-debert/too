package add

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/store"
	"github.com/arthur-debert/too/pkg/too/testutil"
)

func TestAddWithPureIDMStorage(t *testing.T) {
	// Create a temporary IDM store using the pure implementation
	tempPath := filepath.Join(testutil.TempDir(t), "pure-idm-test.json")
	
	opts := Options{
		CollectionPath: tempPath,
		Mode:           "short",
	}

	// Execute add command which now uses pure IDM storage
	result, err := ExecuteIDM("First task", opts)
	if err != nil {
		t.Fatalf("Failed to add first task: %v", err)
	}

	// Add a second task
	result2, err := ExecuteIDM("Second task", opts)
	if err != nil {
		t.Fatalf("Failed to add second task: %v", err)
	}

	// Read the file directly to verify it's in pure IDM format
	data, err := os.ReadFile(tempPath)
	if err != nil {
		t.Fatalf("Failed to read storage file: %v", err)
	}

	// Unmarshal as IDMCollection
	var collection models.IDMCollection
	if err := json.Unmarshal(data, &collection); err != nil {
		t.Fatalf("Failed to unmarshal as IDMCollection: %v", err)
	}

	// Verify structure
	if collection.Items == nil {
		t.Fatal("Expected Items field in IDMCollection")
	}

	if len(collection.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(collection.Items))
	}

	// Verify the data doesn't have hierarchical structure
	var rawData map[string]interface{}
	if err := json.Unmarshal(data, &rawData); err != nil {
		t.Fatalf("Failed to unmarshal raw data: %v", err)
	}

	// Should have "items" field, not "todos"
	if _, hasItems := rawData["items"]; !hasItems {
		t.Error("Expected 'items' field in JSON")
	}

	if _, hasTodos := rawData["todos"]; hasTodos {
		t.Error("Should not have 'todos' field in pure IDM format")
	}

	// Verify the tasks are in order
	if collection.Items[0].Text != "First task" {
		t.Errorf("Expected first item to be 'First task', got '%s'", collection.Items[0].Text)
	}

	if collection.Items[1].Text != "Second task" {
		t.Errorf("Expected second item to be 'Second task', got '%s'", collection.Items[1].Text)
	}

	// Both should have UIDs matching the results
	if collection.Items[0].UID != result.Todo.UID {
		t.Error("First item UID doesn't match result")
	}

	if collection.Items[1].UID != result2.Todo.UID {
		t.Error("Second item UID doesn't match result")
	}
}

func TestAddWithParentInPureIDMStorage(t *testing.T) {
	tempPath := filepath.Join(testutil.TempDir(t), "pure-idm-parent-test.json")
	
	// Add parent
	parentOpts := Options{
		CollectionPath: tempPath,
		Mode:           "short",
	}
	
	parentResult, err := ExecuteIDM("Parent task", parentOpts)
	if err != nil {
		t.Fatalf("Failed to add parent: %v", err)
	}

	// Add child
	childOpts := Options{
		CollectionPath: tempPath,
		ParentPath:     parentResult.PositionPath, // Should be "1"
		Mode:           "short",
	}

	childResult, err := ExecuteIDM("Child task", childOpts)
	if err != nil {
		t.Fatalf("Failed to add child: %v", err)
	}

	// Load and verify
	idmStore := store.NewIDMStore(tempPath)
	collection, err := idmStore.LoadIDM()
	if err != nil {
		t.Fatalf("Failed to load collection: %v", err)
	}

	if collection.Count() != 2 {
		t.Errorf("Expected 2 items, got %d", collection.Count())
	}

	// Find child and verify parent relationship
	child := collection.FindByUID(childResult.Todo.UID)
	if child == nil {
		t.Fatal("Child not found in collection")
	}

	if child.ParentID != parentResult.Todo.UID {
		t.Errorf("Expected child parent ID '%s', got '%s'", parentResult.Todo.UID, child.ParentID)
	}

	// Verify no hierarchical nesting in storage
	if len(collection.Items) != 2 {
		t.Error("All items should be at root level in flat storage")
	}
}

func TestMultipleChildrenOrderingInPureIDM(t *testing.T) {
	tempPath := filepath.Join(testutil.TempDir(t), "pure-idm-ordering-test.json")
	
	// Add parent
	parentOpts := Options{
		CollectionPath: tempPath,
		Mode:           "short",
	}
	
	parentResult, err := ExecuteIDM("Parent", parentOpts)
	if err != nil {
		t.Fatalf("Failed to add parent: %v", err)
	}

	// Add three children in order
	childOpts := Options{
		CollectionPath: tempPath,
		ParentPath:     parentResult.PositionPath,
		Mode:           "short",
	}

	_, err = ExecuteIDM("Child 1", childOpts)
	if err != nil {
		t.Fatalf("Failed to add child 1: %v", err)
	}

	_, err = ExecuteIDM("Child 2", childOpts)
	if err != nil {
		t.Fatalf("Failed to add child 2: %v", err)
	}

	_, err = ExecuteIDM("Child 3", childOpts)
	if err != nil {
		t.Fatalf("Failed to add child 3: %v", err)
	}

	// Create a new manager to test ordering
	idmStore := store.NewIDMStore(tempPath)
	manager, err := store.NewPureIDMManager(idmStore, tempPath)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Get the children UIDs from the registry
	registry := manager.GetRegistry()
	childUIDs := registry.GetUIDs(parentResult.Todo.UID)

	if len(childUIDs) != 3 {
		t.Errorf("Expected 3 children, got %d", len(childUIDs))
	}

	// Get the actual todos to check their text
	collection := manager.GetCollection()
	
	// The children should be in the order they were added
	expectedTexts := []string{"Child 1", "Child 2", "Child 3"}
	for i, uid := range childUIDs {
		child := collection.FindByUID(uid)
		if child == nil {
			t.Errorf("Child %d not found", i+1)
			continue
		}
		if child.Text != expectedTexts[i] {
			t.Errorf("Expected child %d to be '%s', got '%s'", i+1, expectedTexts[i], child.Text)
		}
	}
}