package testutil

import (
	"path/filepath"
	"testing"

	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/store"
)

// TodoSpec describes a todo to be created in tests
type TodoSpec struct {
	Text     string
	Status   models.TodoStatus
	Children []TodoSpec
}

// CreateIDMStore creates a new IDM store with test data for testing.
func CreateIDMStore(t *testing.T, texts ...string) store.IDMStore {
	t.Helper()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.json")
	idmStore := store.NewIDMStore(dbPath)

	// Create manager to add todos
	manager, err := store.NewPureIDMManager(idmStore, dbPath)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Add todos
	for _, text := range texts {
		if _, err := manager.Add(store.RootScope, text); err != nil {
			t.Fatalf("failed to add todo: %v", err)
		}
	}

	// Save
	if err := manager.Save(); err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	return idmStore
}

// CreateNestedIDMStore creates an IDM store with nested todos for testing.
func CreateNestedIDMStore(t *testing.T) store.IDMStore {
	t.Helper()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.json")
	idmStore := store.NewIDMStore(dbPath)

	// Create manager
	manager, err := store.NewPureIDMManager(idmStore, dbPath)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Create parent todo
	parentUID, err := manager.Add(store.RootScope, "Parent todo")
	if err != nil {
		t.Fatalf("failed to add parent: %v", err)
	}

	// Create sub-tasks
	_, err = manager.Add(parentUID, "Sub-task 1.1")
	if err != nil {
		t.Fatalf("failed to add sub-task 1.1: %v", err)
	}

	subTask2UID, err := manager.Add(parentUID, "Sub-task 1.2")
	if err != nil {
		t.Fatalf("failed to add sub-task 1.2: %v", err)
	}

	// Create grandchild
	_, err = manager.Add(subTask2UID, "Grandchild 1.2.1")
	if err != nil {
		t.Fatalf("failed to add grandchild: %v", err)
	}

	// Create another top-level todo
	_, err = manager.Add(store.RootScope, "Another top-level todo")
	if err != nil {
		t.Fatalf("failed to add another top-level todo: %v", err)
	}

	// Save
	if err := manager.Save(); err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	return idmStore
}

// CreateIDMStoreWithStatuses creates an IDM store with todos having specific statuses.
func CreateIDMStoreWithStatuses(t *testing.T, specs []TodoSpec) store.IDMStore {
	t.Helper()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.json")
	idmStore := store.NewIDMStore(dbPath)

	// Create manager
	manager, err := store.NewPureIDMManager(idmStore, dbPath)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	// Add todos with statuses
	for _, spec := range specs {
		uid, err := manager.Add(store.RootScope, spec.Text)
		if err != nil {
			t.Fatalf("failed to add todo: %v", err)
		}

		if spec.Status != "" {
			if err := manager.SetStatus(uid, "completion", string(spec.Status)); err != nil {
				t.Fatalf("failed to set status: %v", err)
			}
		}
	}

	// Save
	if err := manager.Save(); err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	return idmStore
}

// CreateTestIDMCollection creates a test IDM collection with given todos.
func CreateTestIDMCollection(texts ...string) *models.IDMCollection {
	collection := &models.IDMCollection{
		Items: make([]*models.IDMTodo, 0),
	}

	for _, text := range texts {
		todo := models.NewIDMTodo(text, "")
		collection.Items = append(collection.Items, todo)
	}

	return collection
}