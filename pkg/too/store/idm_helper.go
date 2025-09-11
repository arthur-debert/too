package store

import (
	"fmt"

	"github.com/arthur-debert/too/pkg/idm"
	"github.com/arthur-debert/too/pkg/too/models"
)

// NewManagerFromStore creates a new IDM Manager from a store.
// This is a convenience function to reduce boilerplate when working with the IDM layer.
func NewManagerFromStore(s Store) (*idm.Manager, error) {
	adapter, err := NewIDMStoreAdapter(s)
	if err != nil {
		return nil, fmt.Errorf("failed to create idm adapter: %w", err)
	}
	
	manager, err := idm.NewManager(adapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create idm manager: %w", err)
	}
	
	return manager, nil
}

// NewManagerFromCollection creates a new IDM Manager from a collection instance.
// This is useful when working within a transaction where you have a specific collection.
func NewManagerFromCollection(collection *models.Collection) (*idm.Manager, error) {
	adapter := &IDMStoreAdapter{
		store:      nil, // Not needed for collection-based operations
		collection: collection,
	}
	
	manager, err := idm.NewManager(adapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create idm manager: %w", err)
	}
	
	return manager, nil
}