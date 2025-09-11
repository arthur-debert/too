package models

import (
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

func TestIDMTodoStatusMethods(t *testing.T) {
	todo := NewIDMTodo("Test task", "")
	
	// Test initial status
	if !todo.IsPending() {
		t.Error("New todo should be pending")
	}
	if todo.IsComplete() {
		t.Error("New todo should not be complete")
	}
	
	// Test setting to done
	todo.Statuses["completion"] = string(StatusDone)
	
	if todo.IsPending() {
		t.Error("Todo should not be pending after setting to done")
	}
	if !todo.IsComplete() {
		t.Error("Todo should be complete after setting to done")
	}
}

func TestIDMTodoShortID(t *testing.T) {
	todo := NewIDMTodo("Test", "")
	
	shortID := todo.GetShortID()
	if len(shortID) != 7 {
		t.Errorf("Expected short ID length of 7, got %d", len(shortID))
	}
	
	// Test with manually set short UID
	todo.UID = "abc"
	shortID = todo.GetShortID()
	if shortID != "abc" {
		t.Errorf("Expected short ID 'abc', got '%s'", shortID)
	}
}

func TestIDMCollectionGetDescendants(t *testing.T) {
	collection := NewIDMCollection()
	
	// Create a hierarchy
	root := NewIDMTodo("Root", "")
	child1 := NewIDMTodo("Child 1", root.UID)
	child2 := NewIDMTodo("Child 2", root.UID)
	grandchild1 := NewIDMTodo("Grandchild 1", child1.UID)
	grandchild2 := NewIDMTodo("Grandchild 2", child1.UID)
	
	collection.AddItem(root)
	collection.AddItem(child1)
	collection.AddItem(child2)
	collection.AddItem(grandchild1)
	collection.AddItem(grandchild2)
	
	// Test getting all descendants
	descendants := collection.GetDescendants(root.UID)
	if len(descendants) != 4 {
		t.Errorf("Expected 4 descendants of root, got %d", len(descendants))
	}
	
	// Test getting descendants of child1
	child1Descendants := collection.GetDescendants(child1.UID)
	if len(child1Descendants) != 2 {
		t.Errorf("Expected 2 descendants of child1, got %d", len(child1Descendants))
	}
	
	// Test getting descendants of leaf node
	leafDescendants := collection.GetDescendants(grandchild1.UID)
	if len(leafDescendants) != 0 {
		t.Errorf("Expected 0 descendants of leaf node, got %d", len(leafDescendants))
	}
}