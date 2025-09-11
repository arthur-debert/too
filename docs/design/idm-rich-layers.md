# IDM Convenience Layer: The Manager

Now that the core, stateless `idm.Registry` is in place, we can build a stateful, high-level convenience layer on top of it to handle common use cases like hierarchy management, soft deletes, and pinned items.

This document outlines the design for this new layer, centered around an `idm.Manager` struct.

## 1. Core Design: The `Manager` and a Richer Adapter

The core `idm.Registry` will remain a stateless "query engine" for HIDs. We will introduce a new `idm.Manager` as the stateful "command engine." This requires expanding the `StoreAdapter` interface into a new `ManagedStoreAdapter` that includes methods for writing and changing data.

```go
// ManagedStoreAdapter is the new, richer adapter interface that an 
// application must implement to use the Manager.
type ManagedStoreAdapter interface {
    // Inherit the read-only methods from the core adapter.
    idm.StoreAdapter 

    // --- Tree/Hierarchy Methods ---
    // AddItem creates a new item with the given parent and returns its new UID.
    AddItem(parentUID string) (string, error) 
    // RemoveItem permanently deletes an item and all its descendants.
    RemoveItem(uid string) error
    // MoveItem changes an item's parent.
    MoveItem(uid, newParentUID string) error

    // --- Soft Delete Methods ---
    // SetStatus changes the status of an item (e.g., "active", "deleted").
    SetStatus(uid, status string) error

    // --- Pinned Items Methods ---
    // SetPinned marks an item as pinned or not.
    SetPinned(uid string, isPinned bool) error
}

// The new convenience layer Manager.
type Manager struct {
    reg     *idm.Registry
    adapter ManagedStoreAdapter
}

func NewManager(adapter ManagedStoreAdapter) (*Manager, error) {
    // ... logic to initialize the manager and fully populate the registry ...
}
```

## 2. Feature Implementation

### 2.1. Trees: Automated Hierarchy Management

The `Manager` will handle the logic of updating the registry after the adapter performs the actual data manipulation.

**API & Flow:**

```go
// Add creates a new item under a parent and updates the registry.
func (m *Manager) Add(parentUID string) (newUID string, newHID uint, err error) {
    // 1. Tell the adapter to create the item in the data store.
    newUID, err = m.adapter.AddItem(parentUID)
    if err != nil {
        return "", 0, err
    }

    // 2. Update the in-memory registry for the affected scope.
    newHID = m.reg.Add(parentUID, newUID)
    return newUID, newHID, nil
}

// Move changes an item's parent and updates all affected scopes.
func (m *Manager) Move(uid, oldParentUID, newParentUID string) error {
    // 1. Tell the adapter to move the item in the data store.
    if err := m.adapter.MoveItem(uid, newParentUID); err != nil {
        return err
    }
    
    // 2. Rebuild the affected scopes from the source of truth.
    if err := m.reg.RebuildScope(m.adapter, oldParentUID); err != nil {
        return err
    }
    if err := m.reg.RebuildScope(m.adapter, newParentUID); err != nil {
        return err
    }

    return nil
}
```

### 2.2. Soft Delete

This is a perfect use case for scopes. The `Manager` will abstract this entire workflow.

**API & Flow:**

```go
const (
    StatusActive  = "active"
    StatusDeleted = "deleted"
)

// SoftDelete moves an item to the "deleted" state.
func (m *Manager) SoftDelete(uid, parentUID string) error {
    // 1. Tell the adapter to change the item's status.
    if err := m.adapter.SetStatus(uid, StatusDeleted); err != nil {
        return err
    }

    // 2. Rebuild the scope from which the item was removed.
    return m.reg.RebuildScope(m.adapter, parentUID)
}

// Restore moves an item back to the "active" state.
func (m *Manager) Restore(uid, parentUID string) error {
    if err := m.adapter.SetStatus(uid, StatusActive); err != nil {
        return err
    }
    return m.reg.RebuildScope(m.adapter, parentUID)
}

// Purge permanently deletes a soft-deleted item.
func (m *Manager) Purge(uid string) error {
    return m.adapter.RemoveItem(uid)
}
```

### 2.3. Pinned Items

This demonstrates how an item's UID can exist in multiple scopes simultaneously.

**API & Flow:**

```go
const ScopePinned = "pinned"

// Pin adds an item's UID to the pinned scope.
func (m *Manager) Pin(uid string) error {
    // 1. Tell the adapter to mark the item as pinned.
    if err := m.adapter.SetPinned(uid, true); err != nil {
        return err
    }

    // 2. Rebuild the special "pinned" scope in the registry.
    return m.reg.RebuildScope(m.adapter, ScopePinned)
}

// Unpin removes an item's UID from the pinned scope.
func (m *Manager) Unpin(uid string) error {
    if err := m.adapter.SetPinned(uid, false); err != nil {
        return err
    }
    return m.reg.RebuildScope(m.adapter, ScopePinned)
}
```

### 2.4. Multi-ID Translation

This is a performance and convenience enhancement that belongs directly on the core `idm.Registry`.

**API & Flow:**

```go
// In pkg/idm/registry.go

// ResolvePositionPaths translates multiple dot-notation paths into UIDs.
func (r *Registry) ResolvePositionPaths(startScope string, paths []string) (map[string]string, error) {
    results := make(map[string]string, len(paths))
    for _, path := range paths {
        uid, err := r.ResolvePositionPath(startScope, path)
        if err != nil {
            return nil, err // Stop on first error
        }
        results[path] = uid
    }
    return results, nil
}
```
