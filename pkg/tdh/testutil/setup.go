package testutil

import (
	"testing"

	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/arthur-debert/tdh/pkg/tdh/store"
)

// TodoSpec describes a todo to be created in tests
type TodoSpec struct {
	Text   string
	Status models.TodoStatus
}

// CreatePopulatedStore creates an in-memory store populated with todos.
// All todos are created with pending status.
func CreatePopulatedStore(t *testing.T, texts ...string) store.Store {
	t.Helper()

	s := store.NewMemoryStore()
	collection := models.NewCollection()

	for _, text := range texts {
		collection.CreateTodo(text)
	}

	if err := s.Save(collection); err != nil {
		t.Fatalf("failed to save collection: %v", err)
	}

	return s
}

// CreateStoreWithSpecs creates an in-memory store with todos matching the given specifications.
// This allows creating todos with specific statuses.
func CreateStoreWithSpecs(t *testing.T, specs []TodoSpec) store.Store {
	t.Helper()

	s := store.NewMemoryStore()
	collection := models.NewCollection()

	for _, spec := range specs {
		todo := collection.CreateTodo(spec.Text)
		todo.Status = spec.Status
	}

	if err := s.Save(collection); err != nil {
		t.Fatalf("failed to save collection: %v", err)
	}

	return s
}

// NewTestCollection creates a new collection with the given todos for testing.
func NewTestCollection(todos ...*models.Todo) *models.Collection {
	collection := models.NewCollection()
	collection.Todos = todos
	return collection
}

// TempDir creates a temporary directory for the test and returns its path.
// The directory is automatically removed when the test completes.
func TempDir(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}
