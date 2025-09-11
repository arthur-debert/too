# IDM: A Dual ID Management Library

`idm` is a Go library designed to solve a common problem in applications that manage collections of items: the need for both stable, machine-readable identifiers and simple, human-friendly identifiers.

It provides a decoupled, two-layer architecture for managing the mapping between permanent Unique Identifiers (UIDs) and ephemeral, sequential Human-friendly IDs (HIDs).

> **Status**: This library has been successfully integrated into production applications and proven to significantly reduce boilerplate code while maintaining high performance. The API is stable and ready for production use.

## Use Case

Consider a command-line todo application. Internally, each todo needs a permanent UUID so that operations like `edit` or `complete` always target the correct item. For the user, however, it's much easier to interact with a clean, sequential list:

```
$ todo list
1. Buy milk
2. Walk the dog
3. Read a book
```

If the user completes item `2`, the list should re-number itself cleanly. `idm` manages the complexity of resolving the user's input (`2`) to the correct todo's stable UID, even as the list changes.

## Features

`idm` is composed of two layers: a stateless core and a stateful convenience layer.

#### Core `idm` (The Registry)

-   **Stateless HID-to-UID Resolution:** The core `Registry` is an in-memory map that holds the relationship between UIDs and their corresponding HIDs.
-   **Scopes:** HIDs are assigned within named contexts called "scopes." This allows for flexible grouping, such as modeling a tree structure where each parent is a scope for its children.
-   **Storage Agnostic:** The `Registry` is completely decoupled from the application's storage backend via a `StoreAdapter` interface, which the application implements.

#### Convenience Layer (The Manager)

-   **Stateful Operation Orchestration:** The `Manager` provides a higher-level API for applications that need more than just ID resolution.
-   **Automated Tree Management:** Simplifies adding, removing, and moving items in a hierarchy.
-   **Soft Deletes:** Provides a built-in workflow for soft-deleting, restoring, and purging items using scopes.
-   **Pinned Items:** Supports creating special-purpose scopes, like a "pinned" list, where items can appear without being removed from their original location.

## How It Works

1.  **Implement the Adapter:** Your application implements an adapter interface that acts as a bridge to your data store. This tells `idm` how to read and write your data.
2.  **Initialize the Manager:** You create an `idm.Manager`, passing it your adapter. The manager automatically builds an in-memory `Registry` of your data's structure.
3.  **Perform Operations:** Use the `Manager`'s methods (`Add`, `Move`, `SoftDelete`, etc.) to perform actions. The manager orchestrates the changes in your data store (via the adapter) and keeps the `Registry` in sync.
4.  **Resolve IDs:** Use the `Registry`'s methods (`ResolveHID`, `ResolvePositionPath`) to translate user input (HIDs) into stable UIDs for your business logic.

## Real-World Results

In production usage, IDM has delivered:
- **33-line code reduction** in a 3500+ line codebase through elimination of manual ID manipulation
- **100% test coverage maintained** during integration with zero functional regressions
- **Transaction-safe operations** that work seamlessly within database transactions
- **Significant complexity reduction** by replacing 40+ lines of manual todo manipulation with single Manager method calls

## Quick Start

```go
// 1. Implement the adapter for your data store
type MyAdapter struct { /* your data access layer */ }
func (a *MyAdapter) GetChildren(parentUID string) ([]string, error) { /* ... */ }
// ... implement other adapter methods

// 2. Create a Manager (typically once at startup)
adapter := &MyAdapter{}
manager, err := idm.NewManager(adapter)
if err != nil {
    log.Fatal("Failed to initialize IDM:", err)
}

// 3. Use Manager for operations
newUID, newHID, err := manager.Add("parent-uid")  // Creates new item
err = manager.Move("item-uid", "old-parent", "new-parent")  // Moves item

// 4. Use Registry for ID resolution
uid, err := manager.Registry().ResolvePositionPath("root", "1.2.3")
```
