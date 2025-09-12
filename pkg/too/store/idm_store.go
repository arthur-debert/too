package store

import (
	"fmt"
	
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/store/internal"
)

// IDMStore defines the interface for persistence operations using the pure IDM data model.
// This interface works with flat IDMCollection structures instead of hierarchical Collections.
type IDMStore interface {
	LoadIDM() (*models.IDMCollection, error)
	SaveIDM(*models.IDMCollection) error
	Exists() bool
	UpdateIDM(func(collection *models.IDMCollection) error) error
	FindItemByUID(uid string) (*models.IDMTodo, error)
	FindItemByShortID(shortID string) (*models.IDMTodo, error)
	Path() string
}

// idmStore is a concrete implementation using the internal IDMJSONFileStore.
type idmStore struct {
	*internal.IDMJSONFileStore
}

// NewIDMStore creates a new IDM store with the given path.
func NewIDMStore(path string) IDMStore {
	return &idmStore{
		IDMJSONFileStore: internal.NewIDMJSONFileStore(path),
	}
}

// FindItemByUID finds a todo item by its UID.
func (s *idmStore) FindItemByUID(uid string) (*models.IDMTodo, error) {
	collection, err := s.LoadIDM()
	if err != nil {
		return nil, err
	}
	
	item := collection.FindByUID(uid)
	if item == nil {
		return nil, fmt.Errorf("todo with UID %s not found", uid)
	}
	
	return item, nil
}

// FindItemByShortID finds a todo item by its short ID.
func (s *idmStore) FindItemByShortID(shortID string) (*models.IDMTodo, error) {
	collection, err := s.LoadIDM()
	if err != nil {
		return nil, err
	}
	
	var found *models.IDMTodo
	var count int
	
	for _, item := range collection.Items {
		if len(item.UID) >= len(shortID) && item.UID[:len(shortID)] == shortID {
			found = item
			count++
			if count > 1 {
				return nil, fmt.Errorf("short ID '%s' is ambiguous", shortID)
			}
		}
	}
	
	if found == nil {
		return nil, fmt.Errorf("no todo found with short ID '%s'", shortID)
	}
	
	return found, nil
}