# IDM Usage Guide

This guide provides a detailed walkthrough of how to integrate and use the `idm` library, covering both the core components and the high-level `Manager`.

## 1. Core IDM: The Registry

The core of `idm` is the `Registry`, a stateless component designed for one primary purpose: resolving Human-friendly IDs (HIDs) to stable Unique Identifiers (UIDs). It is ideal for applications that need fast, in-memory lookups and prefer to handle their own data persistence logic.

### Concepts

-   **UID:** A permanent, unique identifier for an item (e.g., a UUID). This should never change.
-   **HID:** A temporary, 1-based sequential integer that the user sees. An item's HID can change frequently.
-   **Scope:** A named context in which HIDs are assigned. A scope is simply an ordered list of UIDs. For a tree structure, each parent's UID can be a scope for its children.
-   **StoreAdapter:** The bridge between the `Registry` and your data. You must implement this interface.

### How to Use the Registry

#### Step 1: Implement the `StoreAdapter`

The `Registry` knows nothing about your data models or storage. You must teach it how to access your data's structure by implementing the `idm.StoreAdapter` interface.

```go
// Your application's adapter
type MyAppAdapter struct {
    // ... a connection to your database, file, etc.
}

// GetChildren returns an ordered list of child UIDs for a given parent.
func (a *MyAppAdapter) GetChildren(parentUID string) ([]string, error) {
    // Your logic to fetch ordered child UIDs from your data store.
    // For example, query a database:
    // "SELECT id FROM items WHERE parent_id = ? ORDER BY position ASC"
}

// GetScopes returns all UIDs that act as parents.
func (a *MyAppAdapter) GetScopes() ([]string, error) {
    // Your logic to find all items that have children.
    // For example: "SELECT DISTINCT parent_id FROM items"
}

// GetAllUIDs returns every UID in the system.
func (a *MyAppAdapter) GetAllUIDs() ([]string, error) {
    // Your logic to get all UIDs.
    // For example: "SELECT id FROM items"
}
```

#### Step 2: Populate and Use the Registry

Before handling a user request, you create and populate the `Registry`.

```go
import "github.com/arthur-debert/too/pkg/idm"

func handleRequest(userInput string) { // e.g., userInput is "1.2"
    myAdapter := &MyAppAdapter{...}
    reg := idm.NewRegistry()

    // Get all scopes (e.g., "root", "parent_uid_1", "parent_uid_5")
    scopes, err := myAdapter.GetScopes()
    // ... handle error

    // Populate the registry by rebuilding each scope
    for _, scope := range scopes {
        err := reg.RebuildScope(myAdapter, scope)
        // ... handle error
    }

    // Now, resolve the user's input
    // The resolver is used for path-based lookups (e.g., "1.2.3")
    resolver := idm.NewResolver(reg)
    uid, err := resolver.Resolve("root", userInput)
    if err != nil {
        // Handle "not found" error
    }

    // Use the stable UID to perform actions in your application
    fmt.Printf("User wants to act on item with stable UID: %s\n", uid)
}
```

---

## 2. Convenience Layer: The Manager

The `Manager` is a stateful, high-level component that sits on top of the `Registry`. It is designed for applications that want to offload common operational logic, such as adding, moving, and managing the state of items.

### Concepts

The `Manager` uses a more powerful adapter, the `ManagedStoreAdapter`, which includes methods for both reading and writing data.

### How to Use the Manager

#### Step 1: Implement the `ManagedStoreAdapter`

You must implement methods for modifying your data, in addition to the read-only methods from the `StoreAdapter`.

```go
type MyAppManagedAdapter struct {
    // ...
}

// Implement all methods from StoreAdapter (GetChildren, etc.)...

// AddItem creates a new item in your data store.
func (a *MyAppManagedAdapter) AddItem(parentUID string) (string, error) {
    // Your logic to create a new item, generate a UID,
    // and persist it.
}

// MoveItem changes an item's parent in your data store.
func (a *MyAppManagedAdapter) MoveItem(uid, newParentUID string) error {
    // Your logic to update the parent_id of an item.
}

// SetStatus updates an item's status for soft deletes.
func (a *MyAppManagedAdapter) SetStatus(uid, status string) error {
    // Your logic to set a "status" field on your item model.
}

// SetPinned updates an item's pinned status.
func (a *MyAppManagedAdapter) SetPinned(uid string, isPinned bool) error {
    // Your logic to set a boolean "is_pinned" flag.
}

// RemoveItem permanently deletes an item.
func (a *MyAppManagedAdapter) RemoveItem(uid string) error {
    // Your logic to DELETE an item from your data store.
}
```

#### Step 2: Initialize and Use the Manager

The `Manager` is intended to be a long-lived object that you initialize once when your application starts.

```go
// main.go
func main() {
    myAdapter := &MyAppManagedAdapter{...}

    // NewManager automatically populates the internal registry
    manager, err := idm.NewManager(myAdapter)
    if err != nil {
        log.Fatalf("Failed to init IDM Manager: %v", err)
    }

    // Now, use the manager to handle operations...
    handleAddItem(manager, "root")
    handleMoveItem(manager, "uid3", "uid1", "uid2")
}

// ---

// Example Handlers ---

func handleAddItem(manager *idm.Manager, parentUID string) {
    newUID, newHID, err := manager.Add(parentUID)
    // ...
    fmt.Printf("Added item %s with HID %d\n", newUID, newHID)
}

func handleMoveItem(manager *idm.Manager, uid, oldParent, newParent string) {
    err := manager.Move(uid, oldParent, newParent)
    // ...
    fmt.Printf("Moved item %s to parent %s\n", uid, newParent)
}

func handleSoftDelete(manager *idm.Manager, uid, parentUID string) {
    err := manager.SoftDelete(uid, parentUID)
    // ...
    fmt.Printf("Soft-deleted item %s\n", uid)
}

// To resolve IDs, access the manager's internal registry
func resolvePath(manager *idm.Manager, path string) {
    uid, err := idm.NewResolver(manager.Registry()).Resolve("root", path)
    // ...
}
```
