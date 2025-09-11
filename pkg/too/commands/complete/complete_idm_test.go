package complete

import (
	"path/filepath"
	"testing"

	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/store"
	"github.com/arthur-debert/too/pkg/too/testutil"
)

func TestExecuteIDM(t *testing.T) {
	// Create a temporary IDM store with a todo
	tempPath := filepath.Join(testutil.TempDir(t), "test.json")
	idmStore := store.NewIDMStore(tempPath)

	// Add a todo using the IDM manager
	manager, err := store.NewPureIDMManager(idmStore, tempPath)
	if err != nil {
		t.Fatalf("Failed to create pure IDM manager: %v", err)
	}

	// Add a test todo
	todoUID, err := manager.Add(store.RootScope, "Test task")
	if err != nil {
		t.Fatalf("Failed to add test todo: %v", err)
	}

	// Save the todo
	if err := manager.Save(); err != nil {
		t.Fatalf("Failed to save todo: %v", err)
	}

	// Get the position path for the todo
	positionPath, err := manager.GetPositionPath(store.RootScope, todoUID)
	if err != nil {
		t.Fatalf("Failed to get position path: %v", err)
	}

	// Test completing the todo
	opts := Options{
		CollectionPath: tempPath,
		Mode:           "short",
	}

	result, err := ExecuteIDM(positionPath, opts)
	if err != nil {
		t.Fatalf("Failed to complete todo: %v", err)
	}

	// Verify result
	if result.Todo == nil {
		t.Fatal("Expected todo in result")
	}

	if result.Todo.Text != "Test task" {
		t.Errorf("Expected todo text 'Test task', got '%s'", result.Todo.Text)
	}

	if result.OldStatus != "pending" {
		t.Errorf("Expected old status 'pending', got '%s'", result.OldStatus)
	}

	if result.NewStatus != "done" {
		t.Errorf("Expected new status 'done', got '%s'", result.NewStatus)
	}

	// Verify the todo was actually updated by loading a fresh manager
	freshManager, err := store.NewPureIDMManager(idmStore, tempPath)
	if err != nil {
		t.Fatalf("Failed to create fresh IDM manager: %v", err)
	}

	updatedTodo := freshManager.GetTodoByUID(todoUID)
	if updatedTodo == nil {
		t.Fatal("Todo not found after completion")
	}

	if updatedTodo.GetStatus() != models.StatusDone {
		t.Errorf("Expected todo status 'done', got '%s'", updatedTodo.GetStatus())
	}
}

func TestExecuteIDMLongMode(t *testing.T) {
	// Create a temporary IDM store with multiple todos
	tempPath := filepath.Join(testutil.TempDir(t), "test.json")
	idmStore := store.NewIDMStore(tempPath)

	manager, err := store.NewPureIDMManager(idmStore, tempPath)
	if err != nil {
		t.Fatalf("Failed to create pure IDM manager: %v", err)
	}

	// Add multiple test todos
	todoUID1, err := manager.Add(store.RootScope, "Task 1")
	if err != nil {
		t.Fatalf("Failed to add task 1: %v", err)
	}

	_, err = manager.Add(store.RootScope, "Task 2")
	if err != nil {
		t.Fatalf("Failed to add task 2: %v", err)
	}

	_, err = manager.Add(store.RootScope, "Task 3")
	if err != nil {
		t.Fatalf("Failed to add task 3: %v", err)
	}

	if err := manager.Save(); err != nil {
		t.Fatalf("Failed to save todos: %v", err)
	}

	// Get position path for first todo
	positionPath, err := manager.GetPositionPath(store.RootScope, todoUID1)
	if err != nil {
		t.Fatalf("Failed to get position path: %v", err)
	}

	// Test completing in long mode
	opts := Options{
		CollectionPath: tempPath,
		Mode:           "long",
	}

	result, err := ExecuteIDM(positionPath, opts)
	if err != nil {
		t.Fatalf("Failed to complete todo in long mode: %v", err)
	}

	// Verify long mode data is populated
	if result.AllTodos == nil {
		t.Error("Expected AllTodos to be populated in long mode")
	}

	if len(result.AllTodos) != 2 { // Should have 2 active todos after completing 1
		t.Errorf("Expected 2 active todos, got %d", len(result.AllTodos))
	}

	if result.TotalCount != 3 {
		t.Errorf("Expected TotalCount 3, got %d", result.TotalCount)
	}

	if result.DoneCount != 1 {
		t.Errorf("Expected DoneCount 1, got %d", result.DoneCount)
	}
}

func TestConvertIDMResultToResultComplete(t *testing.T) {
	// Create an IDM todo
	idmTodo := models.NewIDMTodo("Test task", "")
	idmTodo.Statuses["completion"] = "done"

	idmResult := &IDMResult{
		Todo:       idmTodo,
		OldStatus:  "pending",
		NewStatus:  "done",
		Mode:       "short",
		TotalCount: 5,
		DoneCount:  3,
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

	if result.OldStatus != "pending" {
		t.Error("Old status not correctly converted")
	}

	if result.NewStatus != "done" {
		t.Error("New status not correctly converted")
	}

	if result.TotalCount != 5 {
		t.Error("Total count not correctly converted")
	}

	if result.DoneCount != 3 {
		t.Error("Done count not correctly converted")
	}

	// Verify that Items is empty (hierarchy managed by IDM)
	if len(result.Todo.Items) != 0 {
		t.Error("Expected empty Items slice in converted todo")
	}
}