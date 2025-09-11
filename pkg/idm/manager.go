package idm

import "fmt"

// Manager provides a stateful, high-level convenience layer for managing
// hierarchical and state-based collections on top of the core Registry.
// It orchestrates operations between a data store (via the ManagedStoreAdapter)
// and the in-memory ID registry.
type Manager struct {
	reg     *Registry
	adapter ManagedStoreAdapter
}

// NewManager creates and initializes a new Manager. It performs a full build
// of the registry to ensure it's in sync with the data store on startup.
func NewManager(adapter ManagedStoreAdapter) (*Manager, error) {
	reg := NewRegistry()

	scopes, err := adapter.GetScopes()
	if err != nil {
		return nil, fmt.Errorf("could not get scopes to build registry: %w", err)
	}

	for _, scope := range scopes {
		if err := reg.RebuildScope(adapter, scope); err != nil {
			return nil, fmt.Errorf("could not rebuild scope '%s': %w", scope, err)
		}
	}

	return &Manager{
		reg:     reg,
		adapter: adapter,
	}, nil
}

// Registry returns the internal registry, allowing for direct read-only
// operations like HID-to-UID resolution.
func (m *Manager) Registry() *Registry {
	return m.reg
}

// --- Tree/Hierarchy Methods ---

// Add creates a new item under a parent, persists it via the adapter,
// and updates the registry.
func (m *Manager) Add(parentUID string) (newUID string, newHID uint, err error) {
	// 1. Tell the adapter to create the item in the data store.
	newUID, err = m.adapter.AddItem(parentUID)
	if err != nil {
		return "", 0, fmt.Errorf("adapter failed to add item: %w", err)
	}

	// 2. Update the in-memory registry for the affected scope.
	// This is more efficient than a full rebuild.
	newHID = m.reg.Add(parentUID, newUID)

	return newUID, newHID, nil
}

// Move changes an item's parent, persists the change via the adapter,
// and updates all affected scopes in the registry.
func (m *Manager) Move(uid, oldParentUID, newParentUID string) error {
	// 1. Tell the adapter to move the item in the data store.
	if err := m.adapter.MoveItem(uid, newParentUID); err != nil {
		return fmt.Errorf("adapter failed to move item: %w", err)
	}

	// 2. Rebuild the affected scopes from the source of truth for consistency.
	if err := m.reg.RebuildScope(m.adapter, oldParentUID); err != nil {
		return fmt.Errorf("failed to rebuild old parent scope '%s': %w", oldParentUID, err)
	}
	if err := m.reg.RebuildScope(m.adapter, newParentUID); err != nil {
		return fmt.Errorf("failed to rebuild new parent scope '%s': %w", newParentUID, err)
	}

	return nil
}
