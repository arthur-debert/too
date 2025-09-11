package add

import (
	"path/filepath"
	"testing"

	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/store"
	"github.com/arthur-debert/too/pkg/too/testutil"
)

func TestRunIDM(t *testing.T) {
	// Create a temporary IDM store
	tempPath := filepath.Join(testutil.TempDir(t), "test.json")
	idmStore := store.NewIDMStore(tempPath)

	opts := Options{
		CollectionPath: idmStore.Path(),
		Mode:           "short",
	}

	// Test adding a root todo
	result, err := RunIDM(idmStore, "Test task", opts)
	if err != nil {
		t.Fatalf("Failed to add todo: %v", err)
	}

	if result.Todo == nil {
		t.Fatal("Expected todo in result")
	}

	if result.Todo.Text != "Test task" {
		t.Errorf("Expected todo text 'Test task', got '%s'", result.Todo.Text)
	}

	if result.Todo.ParentID != "" {
		t.Errorf("Expected empty parent ID for root todo, got '%s'", result.Todo.ParentID)
	}

	if result.PositionPath == "" {
		t.Error("Expected non-empty position path")
	}

	// Verify the todo was saved to the store
	collection, err := idmStore.LoadIDM()
	if err != nil {
		t.Fatalf("Failed to load collection: %v", err)
	}

	if collection.Count() != 1 {
		t.Errorf("Expected 1 todo in collection, got %d", collection.Count())
	}

	saved := collection.FindByUID(result.Todo.UID)
	if saved == nil {
		t.Error("Todo not found in saved collection")
	} else if saved.Text != "Test task" {
		t.Errorf("Expected saved todo text 'Test task', got '%s'", saved.Text)
	}
}

func TestRunIDMWithParent(t *testing.T) {
	// Create a temporary IDM store with existing data
	tempPath := filepath.Join(testutil.TempDir(t), "test.json")
	idmStore := store.NewIDMStore(tempPath)

	// First add a parent todo
	parentResult, err := RunIDM(idmStore, "Parent task", Options{
		CollectionPath: idmStore.Path(),
		Mode:           "short",
	})
	if err != nil {
		t.Fatalf("Failed to add parent todo: %v", err)
	}

	// Now add a child todo using position path
	opts := Options{
		CollectionPath: idmStore.Path(),
		ParentPath:     parentResult.PositionPath, // Should be "1"
		Mode:           "short",
	}

	childResult, err := RunIDM(idmStore, "Child task", opts)
	if err != nil {
		t.Fatalf("Failed to add child todo: %v", err)
	}

	if childResult.Todo.ParentID != parentResult.Todo.UID {
		t.Errorf("Expected child parent ID '%s', got '%s'", parentResult.Todo.UID, childResult.Todo.ParentID)
	}

	// Verify the collection has 2 todos
	collection, err := idmStore.LoadIDM()
	if err != nil {
		t.Fatalf("Failed to load collection: %v", err)
	}

	if collection.Count() != 2 {
		t.Errorf("Expected 2 todos in collection, got %d", collection.Count())
	}

	// Verify parent-child relationship
	children := collection.GetChildren(parentResult.Todo.UID)
	if len(children) != 1 {
		t.Errorf("Expected 1 child of parent, got %d", len(children))
	}
	if children[0].UID != childResult.Todo.UID {
		t.Error("Child UID does not match expected")
	}
}

func TestRunIDMLongMode(t *testing.T) {
	// Create a temporary IDM store
	tempPath := filepath.Join(testutil.TempDir(t), "test.json")
	idmStore := store.NewIDMStore(tempPath)

	// Add a few todos first
	_, err := RunIDM(idmStore, "Task 1", Options{CollectionPath: idmStore.Path(), Mode: "short"})
	if err != nil {
		t.Fatalf("Failed to add task 1: %v", err)
	}

	_, err = RunIDM(idmStore, "Task 2", Options{CollectionPath: idmStore.Path(), Mode: "short"})
	if err != nil {
		t.Fatalf("Failed to add task 2: %v", err)
	}

	// Add a todo in long mode
	opts := Options{
		CollectionPath: idmStore.Path(),
		Mode:           "long",
	}

	result, err := RunIDM(idmStore, "Task 3", opts)
	if err != nil {
		t.Fatalf("Failed to add todo in long mode: %v", err)
	}

	// Verify long mode data is populated
	if result.AllTodos == nil {
		t.Error("Expected AllTodos to be populated in long mode")
	}

	if len(result.AllTodos) != 3 {
		t.Errorf("Expected 3 todos in AllTodos, got %d", len(result.AllTodos))
	}

	if result.TotalCount != 3 {
		t.Errorf("Expected TotalCount 3, got %d", result.TotalCount)
	}

	if result.DoneCount != 0 {
		t.Errorf("Expected DoneCount 0, got %d", result.DoneCount)
	}
}

func TestConvertIDMResultToResult(t *testing.T) {
	// Create an IDM todo
	idmTodo := models.NewIDMTodo("Test task", "parent123")
	idmTodo.Statuses["priority"] = "high"

	idmResult := &IDMResult{
		Todo:         idmTodo,
		PositionPath: "1.2",
		Mode:         "short",
		TotalCount:   5,
		DoneCount:    2,
	}

	// Convert to traditional result
	result := ConvertIDMResultToResult(idmResult)

	// Verify conversion
	if result.Todo.ID != idmTodo.UID {
		t.Error("Todo UID not correctly converted to ID")
	}

	if result.Todo.Text != idmTodo.Text {
		t.Error("Todo text not correctly converted")
	}

	if result.Todo.ParentID != idmTodo.ParentID {
		t.Error("Todo parent ID not correctly converted")
	}

	if result.Todo.Statuses["priority"] != "high" {
		t.Error("Todo statuses not correctly converted")
	}

	if result.PositionPath != "1.2" {
		t.Error("Position path not correctly converted")
	}

	if result.TotalCount != 5 {
		t.Error("Total count not correctly converted")
	}

	if result.DoneCount != 2 {
		t.Error("Done count not correctly converted")
	}

	// Verify that Items is empty (hierarchy managed by IDM)
	if len(result.Todo.Items) != 0 {
		t.Error("Expected empty Items slice in converted todo")
	}
}