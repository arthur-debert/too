package models

import (
	"fmt"
	"testing"
	"time"
)

func TestIDMTodoBasics(t *testing.T) {
	// Test creating a new IDM todo
	todo := NewIDMTodo("Test task", "")
	
	if todo.Text != "Test task" {
		t.Errorf("Expected text 'Test task', got '%s'", todo.Text)
	}
	
	if todo.ParentID != "" {
		t.Errorf("Expected empty parent ID for root todo, got '%s'", todo.ParentID)
	}
	
	if todo.UID == "" {
		t.Error("Expected non-empty UID")
	}
	
	if todo.GetStatus() != StatusPending {
		t.Errorf("Expected status 'pending', got '%s'", todo.GetStatus())
	}
}

func TestIDMCollectionOperations(t *testing.T) {
	collection := NewIDMCollection()
	
	// Test adding items
	todo1 := NewIDMTodo("Task 1", "")
	todo2 := NewIDMTodo("Task 2", todo1.UID)
	
	collection.AddItem(todo1)
	collection.AddItem(todo2)
	
	if collection.Count() != 2 {
		t.Errorf("Expected 2 items, got %d", collection.Count())
	}
	
	// Test finding by UID
	found := collection.FindByUID(todo1.UID)
	if found == nil {
		t.Error("Expected to find todo1 by UID")
	}
	if found.Text != "Task 1" {
		t.Errorf("Expected 'Task 1', got '%s'", found.Text)
	}
	
	// Test getting children
	children := collection.GetChildren(todo1.UID)
	if len(children) != 1 {
		t.Errorf("Expected 1 child of todo1, got %d", len(children))
	}
	if children[0].UID != todo2.UID {
		t.Error("Expected child to be todo2")
	}
	
	// Test removing items
	if !collection.RemoveItem(todo2.UID) {
		t.Error("Expected to successfully remove todo2")
	}
	if collection.Count() != 1 {
		t.Errorf("Expected 1 item after removal, got %d", collection.Count())
	}
}

func TestMigrationToIDM(t *testing.T) {
	// Create a hierarchical collection
	collection := NewCollection()
	
	// Add root level todo
	rootTodo, err := collection.CreateTodo("Root task", "")
	if err != nil {
		t.Fatalf("Failed to create root todo: %v", err)
	}
	
	// Add child todo
	childTodo, err := collection.CreateTodo("Child task", rootTodo.ID)
	if err != nil {
		t.Fatalf("Failed to create child todo: %v", err)
	}
	
	// Add grandchild todo
	grandchildTodo, err := collection.CreateTodo("Grandchild task", childTodo.ID)
	if err != nil {
		t.Fatalf("Failed to create grandchild todo: %v", err)
	}
	
	// Migrate to IDM
	idmCollection := MigrateToIDM(collection)
	
	// Verify all items are present in flat structure
	if idmCollection.Count() != 3 {
		t.Errorf("Expected 3 items in IDM collection, got %d", idmCollection.Count())
	}
	
	// Verify root item
	idmRoot := idmCollection.FindByUID(rootTodo.ID)
	if idmRoot == nil {
		t.Error("Root todo not found in IDM collection")
	} else {
		if idmRoot.Text != "Root task" {
			t.Errorf("Expected root text 'Root task', got '%s'", idmRoot.Text)
		}
		if idmRoot.ParentID != "" {
			t.Errorf("Expected empty parent ID for root, got '%s'", idmRoot.ParentID)
		}
	}
	
	// Verify child item
	idmChild := idmCollection.FindByUID(childTodo.ID)
	if idmChild == nil {
		t.Error("Child todo not found in IDM collection")
	} else {
		if idmChild.ParentID != rootTodo.ID {
			t.Errorf("Expected child parent ID '%s', got '%s'", rootTodo.ID, idmChild.ParentID)
		}
	}
	
	// Verify grandchild item
	idmGrandchild := idmCollection.FindByUID(grandchildTodo.ID)
	if idmGrandchild == nil {
		t.Error("Grandchild todo not found in IDM collection")
	} else {
		if idmGrandchild.ParentID != childTodo.ID {
			t.Errorf("Expected grandchild parent ID '%s', got '%s'", childTodo.ID, idmGrandchild.ParentID)
		}
	}
	
	// Test hierarchy retrieval
	children := idmCollection.GetChildren(rootTodo.ID)
	if len(children) != 1 {
		t.Errorf("Expected 1 child of root, got %d", len(children))
	}
	
	descendants := idmCollection.GetDescendants(rootTodo.ID)
	if len(descendants) != 2 {
		t.Errorf("Expected 2 descendants of root, got %d", len(descendants))
	}
}

func TestMigrationFromIDM(t *testing.T) {
	// Create a flat IDM collection
	idmCollection := NewIDMCollection()
	
	root := NewIDMTodo("Root task", "")
	child := NewIDMTodo("Child task", root.UID)
	grandchild := NewIDMTodo("Grandchild task", child.UID)
	
	idmCollection.AddItem(root)
	idmCollection.AddItem(child)
	idmCollection.AddItem(grandchild)
	
	// Migrate back to hierarchical
	collection := MigrateFromIDM(idmCollection)
	
	// Verify hierarchical structure is restored
	if len(collection.Todos) != 1 {
		t.Errorf("Expected 1 root todo, got %d", len(collection.Todos))
	}
	
	rootTodo := collection.Todos[0]
	if rootTodo.Text != "Root task" {
		t.Errorf("Expected root text 'Root task', got '%s'", rootTodo.Text)
	}
	
	if len(rootTodo.Items) != 1 {
		t.Errorf("Expected 1 child of root, got %d", len(rootTodo.Items))
	}
	
	childTodo := rootTodo.Items[0]
	if childTodo.Text != "Child task" {
		t.Errorf("Expected child text 'Child task', got '%s'", childTodo.Text)
	}
	
	if len(childTodo.Items) != 1 {
		t.Errorf("Expected 1 grandchild, got %d", len(childTodo.Items))
	}
	
	grandchildTodo := childTodo.Items[0]
	if grandchildTodo.Text != "Grandchild task" {
		t.Errorf("Expected grandchild text 'Grandchild task', got '%s'", grandchildTodo.Text)
	}
}

func TestIDMTodoClone(t *testing.T) {
	original := NewIDMTodo("Original task", "parent123")
	original.Statuses["priority"] = "high"
	original.Modified = time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	
	clone := original.Clone()
	
	// Verify clone has same values
	if clone.UID != original.UID {
		t.Error("Clone should have same UID")
	}
	if clone.Text != original.Text {
		t.Error("Clone should have same text")
	}
	if clone.ParentID != original.ParentID {
		t.Error("Clone should have same parent ID")
	}
	if !clone.Modified.Equal(original.Modified) {
		t.Error("Clone should have same modified time")
	}
	
	// Verify statuses are deeply cloned
	if clone.Statuses["priority"] != "high" {
		t.Error("Clone should have same status values")
	}
	
	// Modify clone and verify original is unchanged
	clone.Text = "Modified text"
	clone.Statuses["priority"] = "low"
	
	if original.Text == "Modified text" {
		t.Error("Original should not be affected by clone modifications")
	}
	if original.Statuses["priority"] == "low" {
		t.Error("Original status should not be affected by clone modifications")
	}
}

func TestValidateIDMCollection(t *testing.T) {
	collection := NewIDMCollection()
	
	// Add valid items
	todo1 := NewIDMTodo("Task 1", "")
	todo2 := NewIDMTodo("Task 2", todo1.UID)
	
	collection.AddItem(todo1)
	collection.AddItem(todo2)
	
	// Should have no validation errors
	errors := ValidateIDMCollection(collection)
	if len(errors) != 0 {
		t.Errorf("Expected no validation errors, got %d: %v", len(errors), errors)
	}
	
	// Add item with invalid parent reference
	todo3 := NewIDMTodo("Task 3", "nonexistent-parent")
	collection.AddItem(todo3)
	
	errors = ValidateIDMCollection(collection)
	if len(errors) == 0 {
		t.Error("Expected validation errors for invalid parent reference")
	}
	
	// Test duplicate UID (manually create to test validation)
	todo4 := &IDMTodo{
		UID:      todo1.UID, // Duplicate UID
		ParentID: "",
		Text:     "Duplicate",
		Statuses: map[string]string{"completion": "pending"},
		Modified: time.Now(),
	}
	collection.AddItem(todo4)
	
	errors = ValidateIDMCollection(collection)
	foundDuplicateError := false
	for _, err := range errors {
		if err.Error() == fmt.Sprintf("duplicate UID found: %s", todo1.UID) {
			foundDuplicateError = true
			break
		}
	}
	if !foundDuplicateError {
		t.Error("Expected validation error for duplicate UID")
	}
}