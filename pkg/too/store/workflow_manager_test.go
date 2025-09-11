package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/arthur-debert/too/pkg/too/config"
	"github.com/arthur-debert/too/pkg/too/models"
)

func TestWorkflowManager_BasicFunctionality(t *testing.T) {
	// Create a test collection
	collection := models.NewCollection()
	
	// Add a simple root todo (no children for now to avoid recursion)
	todo1, err := collection.CreateTodo("Test todo 1", "")
	if err != nil {
		t.Fatalf("Failed to create test todo: %v", err)
	}
	// Ensure ParentID is explicitly empty for root-level todos
	todo1.ParentID = ""
	
	// Create workflow manager - need to enable workflow for auto-transitions to work
	// Create a temporary config file with workflow enabled
	tempDir, err := os.MkdirTemp("", "workflow-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	configPath := filepath.Join(tempDir, "test.workflow.json")
	// Use default config (enabled workflow with todo preset) for basic functionality test
	testConfig := config.DefaultWorkflowConfig()
	err = config.SaveWorkflowConfig(testConfig, configPath)
	if err != nil {
		t.Fatalf("Failed to save test config: %v", err)
	}
	
	// Create workflow manager with the test path
	testCollectionPath := filepath.Join(tempDir, "test.todos")
	wm, err := NewWorkflowManager(collection, testCollectionPath)
	if err != nil {
		t.Fatalf("Failed to create workflow manager: %v", err)
	}
	
	t.Run("ResolvePositionPathInContext", func(t *testing.T) {
		// Resolve position path in active context (should only show pending items)
		uid, err := wm.ResolvePositionPathInContext(RootScope, "1", "active")
		if err != nil {
			t.Fatalf("ResolvePositionPathInContext failed: %v", err)
		}
		
		if uid != todo1.ID {
			t.Errorf("Expected UID %s, got %s", todo1.ID, uid)
		}
	})
	
	t.Run("SetStatus", func(t *testing.T) {
		// Test setting status on a simple root todo (no auto-transitions)
		
		// Set status using workflow manager
		err := wm.SetStatus(todo1.ID, "completion", "done")
		if err != nil {
			t.Fatalf("SetStatus failed: %v", err)
		}
		
		// Verify status was set
		status, err := wm.GetStatus(todo1.ID, "completion")
		if err != nil {
			t.Fatalf("GetStatus failed: %v", err)
		}
		
		if status != "done" {
			t.Errorf("Expected status 'done', got '%s'", status)
		}
		
		// Verify legacy status field was updated for backward compatibility
		updatedTodo := wm.GetCollection().FindItemByID(todo1.ID)
		if updatedTodo.GetStatus() != models.StatusDone {
			t.Errorf("Expected legacy status to be 'done', got '%s'", updatedTodo.GetStatus())
		}
	})
	
	t.Run("BuildResult", func(t *testing.T) {
		// Test building result for short mode
		result, err := wm.BuildResult(todo1.ID, "short", "pending")
		if err != nil {
			t.Fatalf("BuildResult failed: %v", err)
		}
		
		if result.Todo.ID != todo1.ID {
			t.Errorf("Expected todo ID %s, got %s", todo1.ID, result.Todo.ID)
		}
		
		if result.OldStatus != "pending" {
			t.Errorf("Expected old status 'pending', got '%s'", result.OldStatus)
		}
		
		if result.NewStatus != "done" {
			t.Errorf("Expected new status 'done', got '%s'", result.NewStatus)
		}
		
		// Test building result for long mode
		result, err = wm.BuildResult(todo1.ID, "long", "pending")
		if err != nil {
			t.Fatalf("BuildResult for long mode failed: %v", err)
		}
		
		// AllTodos might be empty if all todos are done, but should not be nil
		if result.AllTodos == nil {
			t.Error("Expected AllTodos to be initialized in long mode (even if empty)")
		}
		
		// TotalCount should be 1 (one todo created)
		if result.TotalCount != 1 {
			t.Errorf("Expected TotalCount to be 1, got %d", result.TotalCount)
		}
		
		// DoneCount should be 1 (todo is done)
		if result.DoneCount != 1 {
			t.Errorf("Expected DoneCount to be 1, got %d", result.DoneCount)
		}
	})
}

func TestWorkflowManager_BackwardCompatibility(t *testing.T) {
	// Create collection with legacy todos (no workflow statuses)
	collection := models.NewCollection()
	
	// Create a todo with only legacy status
	todo := &models.Todo{
		ID:       "test-id",
		Text:     "Test todo",
		Statuses: nil, // Explicitly nil to simulate old data
		Items:    []*models.Todo{},
	}
	collection.Todos = []*models.Todo{todo}
	
	// Create workflow manager (using temp directory for consistency)
	tempDir, err := os.MkdirTemp("", "workflow-backward-compat-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	testCollectionPath := filepath.Join(tempDir, "test.todos")
	wm, err := NewWorkflowManager(collection, testCollectionPath)
	if err != nil {
		t.Fatalf("Failed to create workflow manager: %v", err)
	}
	
	// Test that getting status works with legacy data
	status, err := wm.GetStatus(todo.ID, "completion")
	if err != nil {
		t.Fatalf("GetStatus failed for legacy data: %v", err)
	}
	
	if status != "pending" {
		t.Errorf("Expected status 'pending', got '%s'", status)
	}
	
	// Test that setting status works and initializes workflow data
	err = wm.SetStatus(todo.ID, "completion", "done")
	if err != nil {
		t.Fatalf("SetStatus failed: %v", err)
	}
	
	// Verify both workflow and legacy status are updated
	updatedTodo := wm.GetCollection().FindItemByID(todo.ID)
	if updatedTodo.GetStatus() != models.StatusDone {
		t.Errorf("Expected legacy status 'done', got '%s'", updatedTodo.GetStatus())
	}
	
	if updatedTodo.Statuses == nil {
		t.Error("Expected Statuses map to be initialized")
	} else if updatedTodo.Statuses["completion"] != "done" {
		t.Errorf("Expected workflow status 'done', got '%s'", updatedTodo.Statuses["completion"])
	}
}