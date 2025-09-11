# IDM API Reference

This document provides a reference for the public API of the `idm` package, separated into its two main components: the Core Registry and the Convenience Manager.

---

## 1. Core IDM API

These components are focused on stateless, read-only ID resolution.

### `idm.StoreAdapter`

This is the interface you must implement to connect the `Registry` to your data store.

```go
type StoreAdapter interface {
	GetChildren(parentUID string) ([]string, error)
	GetScopes() ([]string, error)
	GetAllUIDs() ([]string, error)
}
```

-   `GetChildren(parentUID string) ([]string, error)`: Should return an ordered list of child UIDs for a given parent UID (scope).
-   `GetScopes() ([]string, error)`: Should return a slice of all UIDs that act as parents (i.e., all possible scopes).
-   `GetAllUIDs() ([]string, error)`: Should return a slice containing every UID in the data store.

### `idm.Registry`

The `Registry` is the core, in-memory engine for mapping HIDs to UIDs.

```go
// Create a new, empty registry.
func NewRegistry() *Registry

// Populates a scope with UIDs by querying the adapter.
func (r *Registry) RebuildScope(adapter StoreAdapter, scope string) error

// Adds a UID to a scope and returns its new HID. Does not persist data.
func (r *Registry) Add(scope, uid string) uint

// Removes a UID from a scope. Does not persist data.
func (r *Registry) Remove(scope, uid string)

// Resolves a single 1-based HID within a scope to its corresponding UID.
func (r *Registry) ResolveHID(scope string, hid uint) (string, error)

// Resolves a dot-notation path (e.g., "1.2.1") into a UID by traversing nested scopes.
// Note: This method is on the Resolver, not the Registry itself.
func (resolver *Resolver) Resolve(startScope, path string) (string, error)

// Translates multiple dot-notation paths into UIDs.
func (r *Registry) ResolvePositionPaths(startScope string, paths []string) (map[string]string, error)
```

### `idm.Resolver`

A helper for resolving hierarchical, dot-notation paths.

```go
// Creates a new resolver instance tied to a registry.
func NewResolver(registry *Registry) *Resolver

// Resolves a full path like "1.2.1" starting from a given scope.
func (resolver *Resolver) Resolve(startScope, path string) (string, error)
```

---

## 2. Convenience Layer API

These components provide a stateful, high-level API for managing collections.

### `idm.ManagedStoreAdapter`

The expanded interface you must implement to use the `Manager`. It includes all methods from `idm.StoreAdapter` plus the following write methods.

```go
type ManagedStoreAdapter interface {
	StoreAdapter // Inherits GetChildren, GetScopes, GetAllUIDs

	AddItem(parentUID string) (string, error)
	RemoveItem(uid string) error
	MoveItem(uid, newParentUID string) error
	SetStatus(uid, status string) error
	SetPinned(uid string, isPinned bool) error
}
```

-   `AddItem(parentUID string) (string, error)`: Should create a new item under the given parent and return the new item's UID.
-   `RemoveItem(uid string) error`: Should permanently delete the item with the given UID.
-   `MoveItem(uid, newParentUID string) error`: Should change the parent of the item.
-   `SetStatus(uid, status string) error`: Should update the status of the item (e.g., for soft deletes).
-   `SetPinned(uid string, isPinned bool) error`: Should set the pinned state of the item.

### `idm.Manager`

The `Manager` is the primary entry point for the convenience layer.

```go
// Creates and initializes a new Manager, fully populating its internal registry.
func NewManager(adapter ManagedStoreAdapter) (*Manager, error)

// Returns the internal registry for direct, read-only operations like ID resolution.
func (m *Manager) Registry() *Registry

// --- Tree/Hierarchy Methods ---

// Creates a new item, persists it, and updates the registry.
func (m *Manager) Add(parentUID string) (newUID string, newHID uint, err error)

// Moves an item to a new parent, persists it, and updates the registry.
func (m *Manager) Move(uid, oldParentUID, newParentUID string) error

// --- Soft Delete Methods ---

// Marks an item as "deleted" and updates the registry.
func (m *Manager) SoftDelete(uid, parentUID string) error

// Marks an item as "active" and updates the registry.
func (m *Manager) Restore(uid, parentUID string) error

// Permanently deletes an item from the data store.
func (m *Manager) Purge(uid string) error

// --- Pinned Items Methods ---

// Adds an item to the "pinned" scope.
func (m *Manager) Pin(uid string) error

// Removes an item from the "pinned" scope.
func (m *Manager) Unpin(uid string) error
```
