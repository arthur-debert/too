package internal

import (
	"os"

	"github.com/arthur-debert/too/pkg/too/models"
)

// MemoryStore implements the Store interface for testing purposes.
type MemoryStore struct {
	Collection *models.Collection
	ShouldFail bool // Flag to simulate save/load errors
}

// NewMemoryStore creates a new MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		Collection: models.NewCollection(),
	}
}

// Load returns the in-memory collection.
func (s *MemoryStore) Load() (*models.Collection, error) {
	if s.ShouldFail {
		return nil, os.ErrNotExist
	}
	return s.Collection, nil
}

// Save updates the in-memory collection.
func (s *MemoryStore) Save(collection *models.Collection) error {
	if s.ShouldFail {
		return os.ErrPermission
	}
	s.Collection = collection
	return nil
}

// Exists always returns true for the memory store.
func (s *MemoryStore) Exists() bool {
	return true
}

// Update performs a transactional update on the in-memory collection.
// If the update function returns an error, the collection is not modified.
func (s *MemoryStore) Update(fn func(collection *models.Collection) error) error {
	collection, err := s.Load()
	if err != nil {
		return err
	}

	// Create a deep copy to ensure rollback safety
	clone := collection.Clone()

	if err := fn(clone); err != nil {
		// On error, discard the clone and return the error
		return err
	}

	// Only save if the update function succeeded
	return s.Save(clone)
}

// Path returns a mock path for the memory store.
func (s *MemoryStore) Path() string {
	return "memory://todos"
}

// Find retrieves todos based on the provided query.
// This implementation uses O(n) linear search through all todos, which is
// acceptable for typical todo list sizes. Future store implementations
// (e.g., SQLite) can optimize this with proper indexing.
func (s *MemoryStore) Find(query Query) (*FindResult, error) {
	if s.ShouldFail {
		return nil, os.ErrNotExist
	}

	// Calculate counts from the full collection
	totalCount, doneCount := CountTodos(s.Collection.Todos)

	return &FindResult{
		Todos:      query.FilterTodos(s.Collection.Todos),
		TotalCount: totalCount,
		DoneCount:  doneCount,
	}, nil
}
