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

// --- Soft Delete Methods ---

const (
	// StatusActive is the status for items that are currently active.
	StatusActive = "active"
	// StatusDeleted is the status for items that have been soft-deleted.
	StatusDeleted = "deleted"
)

// SoftDelete moves an item to the "deleted" state.
// It tells the adapter to change the item's status and then rebuilds the
// parent scope, which should no longer include the item.
func (m *Manager) SoftDelete(uid, parentUID string) error {
	// 1. Tell the adapter to change the item's status.
	if err := m.adapter.SetStatus(uid, StatusDeleted); err != nil {
		return fmt.Errorf("adapter failed to set status to deleted: %w", err)
	}

	// 2. Rebuild the scope from which the item was removed.
	// The adapter's GetChildren(parentUID) should no longer return this UID.
	if err := m.reg.RebuildScope(m.adapter, parentUID); err != nil {
		return fmt.Errorf("failed to rebuild parent scope '%s' after soft delete: %w", parentUID, err)
	}
	return nil
}

// Restore moves an item back to the "active" state from a soft-deleted state.
func (m *Manager) Restore(uid, parentUID string) error {
	// 1. Tell the adapter to change the item's status.
	if err := m.adapter.SetStatus(uid, StatusActive); err != nil {
		return fmt.Errorf("adapter failed to set status to active: %w", err)
	}

	// 2. Rebuild the scope to which the item is returning.
	if err := m.reg.RebuildScope(m.adapter, parentUID); err != nil {
		return fmt.Errorf("failed to rebuild parent scope '%s' after restore: %w", parentUID, err)
	}
	return nil
}

// Purge permanently deletes a soft-deleted item via the adapter.
// Note: This does not update the registry, as the item should already be
// gone from any active scopes.
func (m *Manager) Purge(uid string) error {
	if err := m.adapter.RemoveItem(uid); err != nil {
		return fmt.Errorf("adapter failed to remove item: %w", err)
	}
	return nil
}
