package store

import (
	"testing"

	"github.com/arthur-debert/too/pkg/idm/workflow"
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/store/internal"
)

func TestWorkflowTodoAdapter_BasicFunctionality(t *testing.T) {
	// Create a memory store with some test data
	store := internal.NewMemoryStore()
	collection := models.NewCollection()
	
	// Add some test todos
	todo1, err := collection.CreateTodo("Test todo 1", "")
	if err != nil {
		t.Fatalf("Failed to create test todo: %v", err)
	}
	
	todo2, err := collection.CreateTodo("Test todo 2", "")
	if err != nil {
		t.Fatalf("Failed to create test todo: %v", err)
	}
	
	// Save the collection
	if err := store.Save(collection); err != nil {
		t.Fatalf("Failed to save collection: %v", err)
	}
	
	// Create workflow adapter
	adapter, err := NewWorkflowTodoAdapter(store)
	if err != nil {
		t.Fatalf("Failed to create workflow adapter: %v", err)
	}
	
	// Test basic IDM interface methods
	t.Run("GetChildren", func(t *testing.T) {
		children, err := adapter.GetChildren(RootScope)
		if err != nil {
			t.Fatalf("GetChildren failed: %v", err)
		}
		
		if len(children) != 2 {
			t.Errorf("Expected 2 children, got %d", len(children))
		}
		
		expectedUIDs := []string{todo1.ID, todo2.ID}
		for i, child := range children {
			if child != expectedUIDs[i] {
				t.Errorf("Expected child %d to be %s, got %s", i, expectedUIDs[i], child)
			}
		}
	})
	
	t.Run("GetScopes", func(t *testing.T) {
		scopes, err := adapter.GetScopes()
		if err != nil {
			t.Fatalf("GetScopes failed: %v", err)
		}
		
		// Should have at least root scope
		if len(scopes) < 1 {
			t.Errorf("Expected at least 1 scope, got %d", len(scopes))
		}
		
		if scopes[0] != RootScope {
			t.Errorf("Expected first scope to be %s, got %s", RootScope, scopes[0])
		}
	})
	
	t.Run("GetAllUIDs", func(t *testing.T) {
		uids, err := adapter.GetAllUIDs()
		if err != nil {
			t.Fatalf("GetAllUIDs failed: %v", err)
		}
		
		if len(uids) != 2 {
			t.Errorf("Expected 2 UIDs, got %d", len(uids))
		}
	})
	
	// Test workflow-specific methods
	t.Run("SetItemStatus", func(t *testing.T) {
		err := adapter.SetItemStatus(todo1.ID, "completion", "done")
		if err != nil {
			t.Fatalf("SetItemStatus failed: %v", err)
		}
		
		// Verify the status was set
		status, err := adapter.GetItemStatus(todo1.ID, "completion")
		if err != nil {
			t.Fatalf("GetItemStatus failed: %v", err)
		}
		
		if status != "done" {
			t.Errorf("Expected status 'done', got '%s'", status)
		}
		
		// Verify backward compatibility - legacy status should be updated
		todo := adapter.Collection().FindItemByID(todo1.ID)
		if todo.GetStatus() != models.StatusDone {
			t.Errorf("Expected legacy status to be updated to 'done', got '%s'", todo.GetStatus())
		}
	})
	
	t.Run("SetItemStatus_Priority", func(t *testing.T) {
		err := adapter.SetItemStatus(todo1.ID, "priority", "high")
		if err != nil {
			t.Fatalf("SetItemStatus failed: %v", err)
		}
		
		// Verify the status was set
		status, err := adapter.GetItemStatus(todo1.ID, "priority")
		if err != nil {
			t.Fatalf("GetItemStatus failed: %v", err)
		}
		
		if status != "high" {
			t.Errorf("Expected priority 'high', got '%s'", status)
		}
	})
	
	t.Run("GetItemStatuses", func(t *testing.T) {
		statuses, err := adapter.GetItemStatuses(todo1.ID)
		if err != nil {
			t.Fatalf("GetItemStatuses failed: %v", err)
		}
		
		if statuses["completion"] != "done" {
			t.Errorf("Expected completion status 'done', got '%s'", statuses["completion"])
		}
		
		if statuses["priority"] != "high" {
			t.Errorf("Expected priority 'high', got '%s'", statuses["priority"])
		}
	})
	
	t.Run("GetChildrenInContext", func(t *testing.T) {
		// Create visibility rules for "active" context (only pending items)
		rules := []workflow.VisibilityRule{
			{
				Context:   "active",
				Dimension: "completion",
				Include:   []string{"pending"},
			},
		}
		
		// todo1 is done, todo2 is pending
		children, err := adapter.GetChildrenInContext(RootScope, "active", rules)
		if err != nil {
			t.Fatalf("GetChildrenInContext failed: %v", err)
		}
		
		// Should only return todo2 (pending)
		if len(children) != 1 {
			t.Errorf("Expected 1 child in active context, got %d", len(children))
		}
		
		if len(children) > 0 && children[0] != todo2.ID {
			t.Errorf("Expected child to be %s, got %s", todo2.ID, children[0])
		}
	})
}

func TestWorkflowTodoAdapter_BackwardCompatibility(t *testing.T) {
	// Test that existing todos without Statuses map work correctly
	store := internal.NewMemoryStore()
	collection := models.NewCollection()
	
	// Create a todo with only legacy status (simulating existing data)
	todo := &models.Todo{
		ID:       "test-id",
		Text:     "Test todo",
		Statuses: nil, // Explicitly nil to simulate old data
		Items:    []*models.Todo{},
	}
	collection.Todos = []*models.Todo{todo}
	
	if err := store.Save(collection); err != nil {
		t.Fatalf("Failed to save collection: %v", err)
	}
	
	// Create workflow adapter
	adapter, err := NewWorkflowTodoAdapter(store)
	if err != nil {
		t.Fatalf("Failed to create workflow adapter: %v", err)
	}
	
	// Test that GetItemStatus works with legacy data
	status, err := adapter.GetItemStatus(todo.ID, "completion")
	if err != nil {
		t.Fatalf("GetItemStatus failed for legacy data: %v", err)
	}
	
	if status != "pending" {
		t.Errorf("Expected status 'pending', got '%s'", status)
	}
	
	// Test that setting status initializes the Statuses map
	err = adapter.SetItemStatus(todo.ID, "priority", "medium")
	if err != nil {
		t.Fatalf("SetItemStatus failed: %v", err)
	}
	
	// Verify the Statuses map was initialized
	updatedTodo := adapter.Collection().FindItemByID(todo.ID)
	if updatedTodo.Statuses == nil {
		t.Error("Expected Statuses map to be initialized")
	}
	
	if updatedTodo.Statuses["completion"] != "pending" {
		t.Errorf("Expected completion status to be migrated to 'pending', got '%s'", updatedTodo.Statuses["completion"])
	}
	
	if updatedTodo.Statuses["priority"] != "medium" {
		t.Errorf("Expected priority 'medium', got '%s'", updatedTodo.Statuses["priority"])
	}
}