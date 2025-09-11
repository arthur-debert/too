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

// IDMStoreWrapper wraps a traditional Store to provide IDM interface compatibility.
// This enables gradual migration from hierarchical to flat data models.
type IDMStoreWrapper struct {
	store Store
}

// NewIDMStoreWrapper creates an IDM store wrapper around a traditional store.
func NewIDMStoreWrapper(store Store) IDMStore {
	return &IDMStoreWrapper{store: store}
}

// LoadIDM loads the collection and converts it to the flat IDM format.
func (w *IDMStoreWrapper) LoadIDM() (*models.IDMCollection, error) {
	collection, err := w.store.Load()
	if err != nil {
		return nil, err
	}
	
	// Convert hierarchical to flat IDM structure
	return models.MigrateToIDM(collection), nil
}

// SaveIDM converts the flat IDM collection to hierarchical format and saves it.
func (w *IDMStoreWrapper) SaveIDM(idmCollection *models.IDMCollection) error {
	// Convert flat IDM structure back to hierarchical for storage
	collection := models.MigrateFromIDM(idmCollection)
	return w.store.Save(collection)
}

// Exists checks if the underlying store exists.
func (w *IDMStoreWrapper) Exists() bool {
	return w.store.Exists()
}

// UpdateIDM performs an atomic update operation on the IDM collection.
func (w *IDMStoreWrapper) UpdateIDM(updateFn func(collection *models.IDMCollection) error) error {
	return w.store.Update(func(collection *models.Collection) error {
		// Convert to IDM format
		idmCollection := models.MigrateToIDM(collection)
		
		// Apply the update function
		if err := updateFn(idmCollection); err != nil {
			return err
		}
		
		// Convert back and update the original collection in-place
		updatedCollection := models.MigrateFromIDM(idmCollection)
		collection.Todos = updatedCollection.Todos
		
		return nil
	})
}

// FindItemByUID finds a todo item by its UID.
func (w *IDMStoreWrapper) FindItemByUID(uid string) (*models.IDMTodo, error) {
	idmCollection, err := w.LoadIDM()
	if err != nil {
		return nil, err
	}
	
	item := idmCollection.FindByUID(uid)
	if item == nil {
		return nil, fmt.Errorf("todo with UID %s not found", uid)
	}
	
	return item, nil
}

// FindItemByShortID finds a todo item by its short ID.
func (w *IDMStoreWrapper) FindItemByShortID(shortID string) (*models.IDMTodo, error) {
	idmCollection, err := w.LoadIDM()
	if err != nil {
		return nil, err
	}
	
	var found *models.IDMTodo
	var count int
	
	for _, item := range idmCollection.Items {
		if len(item.UID) >= len(shortID) && item.UID[:len(shortID)] == shortID {
			found = item
			count++
		}
	}
	
	if count == 0 {
		return nil, fmt.Errorf("no todo found with reference '%s'", shortID)
	}
	if count > 1 {
		return nil, fmt.Errorf("multiple todos found with ambiguous reference '%s'", shortID)
	}
	
	return found, nil
}

// Path returns the path where the store persists data.
func (w *IDMStoreWrapper) Path() string {
	return w.store.Path()
}

// NewIDMStore creates a new IDM store using the pure IDM JSON file store implementation.
func NewIDMStore(path string) IDMStore {
	return internal.NewIDMJSONFileStore(path)
}