# IDM Usage Guide

This guide provides a detailed walkthrough of how to integrate and use the `idm` library, covering both the core components and the high-level `Manager`.

> **Note**: This guide reflects real-world usage patterns discovered during large-scale production integration. Examples are based on successful production deployments.

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
    if err != nil {
        log.Printf("Failed to get scopes: %v", err)
        return
    }

    // Populate the registry by rebuilding each scope
    for _, scope := range scopes {
        err := reg.RebuildScope(myAdapter, scope)
        if err != nil {
            log.Printf("Failed to rebuild scope %s: %v", scope, err)
            return
        }
    }

    // Now, resolve the user's input using the Registry directly
    // Note: ResolvePositionPath is available directly on Registry (no separate resolver needed)
    uid, err := reg.ResolvePositionPath("root", userInput)
    if err != nil {
        log.Printf("Position path %s not found: %v", userInput, err)
        return
    }

    // Use the stable UID to perform actions in your application
    fmt.Printf("User wants to act on item with stable UID: %s\n", uid)
}
```

### Important Registry Usage Notes

**Performance Tip**: Registry creation and population is fast but not free. In high-performance scenarios, consider:
- Creating one Registry per request/transaction scope
- Avoiding repeated rebuilds of the same scope within a single operation
- Using the Manager layer for long-lived operations

**Error Handling**: Always handle `ResolvePositionPath` errors - they indicate invalid user input or stale data.

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

The `Manager` can be used in two ways:
1. **Long-lived Manager**: Created once at startup for simple applications
2. **Transaction-aware Manager**: Created per transaction for database consistency

**Long-lived Manager Example:**
```go
// main.go - Simple applications
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
```

**Transaction-aware Manager Example (Recommended for databases):**
```go
// For applications using database transactions
func executeInTransaction(db *sql.DB, operation func(*idm.Manager) error) error {
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback() // Safe to call even after commit

    // Create adapter that works with this specific transaction
    adapter := &MyTransactionAdapter{tx: tx}
    
    // Create Manager for this transaction scope
    manager, err := idm.NewManager(adapter)
    if err != nil {
        return fmt.Errorf("failed to create transaction manager: %w", err)
    }

    // Perform operations within the transaction
    if err := operation(manager); err != nil {
        return err
    }

    return tx.Commit()
}

// Usage:
err := executeInTransaction(db, func(manager *idm.Manager) error {
    newUID, _, err := manager.Add("parent-uid")
    if err != nil {
        return err
    }
    
    return manager.Move("another-uid", "old-parent", newUID)
})
```

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
    // Use Registry's ResolvePositionPath directly (no separate resolver needed)
    uid, err := manager.Registry().ResolvePositionPath("root", path)
    if err != nil {
        log.Printf("Failed to resolve path %s: %v", path, err)
        return
    }
    fmt.Printf("Path %s resolves to UID: %s\n", path, uid)
}
```

## 3. Integration Patterns & Best Practices

Based on production usage, here are proven patterns for successful IDM integration:

### Pattern 1: Command-Based Architecture
```go
// Excellent for CLI applications or command-based systems
type Command struct {
    manager *idm.Manager
}

func (c *Command) Execute(userPath string) error {
    // Resolve user input to stable UID
    uid, err := c.manager.Registry().ResolvePositionPath("root", userPath)
    if err != nil {
        return fmt.Errorf("invalid position: %s", userPath)
    }
    
    // Perform business logic with stable UID
    return c.performOperation(uid)
}
```

### Pattern 2: Hybrid Approach for Status Management
```go
// For applications with complex visibility rules
func completeItem(manager *idm.Manager, path string) error {
    // Use Manager for ID resolution
    uid, err := manager.Registry().ResolvePositionPath("root", path)
    if err != nil {
        return err
    }
    
    // Use traditional methods for status changes if they have
    // different visibility requirements than IDM's soft-delete model
    return setItemStatus(uid, "completed")
}
```

### Pattern 3: Scope-based Organization
```go
// Leverage scopes for different views of the same data
func listActiveItems(manager *idm.Manager) []string {
    // Regular scope shows active items
    return getItemsInScope(manager, "root")
}

func listPinnedItems(manager *idm.Manager) []string {
    // Special scope for pinned items
    return getItemsInScope(manager, idm.ScopePinned)
}
```

### Common Pitfalls to Avoid

1. **Don't mix Manager instances**: Create one Manager per transaction/operation scope
2. **Handle position path errors**: Invalid paths are common with user input
3. **Understand soft-delete semantics**: Items marked as "deleted" won't appear in active scopes
4. **Test with realistic data**: Nested hierarchies reveal edge cases not visible in simple tests
```
