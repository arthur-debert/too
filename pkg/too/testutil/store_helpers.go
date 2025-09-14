package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/store"
)

// CreateTestStore creates a temporary nanostore for testing
func CreateTestStore(t *testing.T) (*store.NanoStoreAdapter, string) {
	t.Helper()
	
	// Create temp directory
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	
	// Create store
	adapter, err := store.NewNanoStoreAdapter(dbPath)
	if err != nil {
		t.Fatalf("failed to create test store: %v", err)
	}
	
	return adapter, dbPath
}

// CreatePopulatedStore creates a store with some test data
func CreatePopulatedStore(t *testing.T) (*store.NanoStoreAdapter, string) {
	t.Helper()
	
	adapter, path := CreateTestStore(t)
	
	// Add some test todos
	todo1, err := adapter.Add("First todo", nil)
	if err != nil {
		t.Fatalf("failed to add first todo: %v", err)
	}
	
	todo2, err := adapter.Add("Second todo", nil)
	if err != nil {
		t.Fatalf("failed to add second todo: %v", err)
	}
	
	// Add a child todo using position path
	_, err = adapter.Add("Child of first", &todo1.PositionPath)
	if err != nil {
		t.Fatalf("failed to add child todo: %v", err)
	}
	
	// Complete the second todo
	if err := adapter.Complete(todo2.PositionPath); err != nil {
		t.Fatalf("failed to complete todo: %v", err)
	}
	
	return adapter, path
}

// TodoSpec defines a todo for test creation
type TodoSpec struct {
	Text      string
	ParentPos string // Position path of parent (e.g., "1", "1.2")
	Complete  bool
}

// CreateStoreWithSpecs creates a store with specific todos
func CreateStoreWithSpecs(t *testing.T, specs ...TodoSpec) (*store.NanoStoreAdapter, string) {
	t.Helper()
	
	adapter, path := CreateTestStore(t)
	
	// Keep track of position paths to UIDs for parent resolution
	posToUID := make(map[string]string)
	
	for _, spec := range specs {
		// Resolve parent if specified
		var parentPos *string
		if spec.ParentPos != "" {
			// Verify the parent exists by checking our map
			if _, ok := posToUID[spec.ParentPos]; ok {
				parentPos = &spec.ParentPos
			} else {
				t.Fatalf("parent position %s not found", spec.ParentPos)
			}
		}
		
		// Add the todo
		todo, err := adapter.Add(spec.Text, parentPos)
		if err != nil {
			t.Fatalf("failed to add todo %s: %v", spec.Text, err)
		}
		
		// Store position to UID mapping
		posToUID[todo.PositionPath] = todo.UID
		
		// Complete if requested
		if spec.Complete {
			if err := adapter.Complete(todo.PositionPath); err != nil {
				t.Fatalf("failed to complete todo %s: %v", spec.Text, err)
			}
		}
	}
	
	return adapter, path
}

// LoadTodos loads all todos from a store (wrapper for List)
func LoadTodos(t *testing.T, adapter *store.NanoStoreAdapter, showAll bool) []*models.Todo {
	t.Helper()
	
	todos, err := adapter.List(showAll)
	if err != nil {
		t.Fatalf("failed to load todos: %v", err)
	}
	
	return todos
}

// CreateTempDBPath creates a temporary database path for testing
func CreateTempDBPath(t *testing.T) string {
	t.Helper()
	
	tmpDir := t.TempDir()
	return filepath.Join(tmpDir, "test.db")
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}