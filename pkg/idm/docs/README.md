# IDM: A Dual ID Management Library

`idm` is a Go library designed to solve a common problem in applications that manage collections of items: the need for both stable, machine-readable identifiers and simple, human-friendly identifiers.

It provides a decoupled, two-layer architecture for managing the mapping between permanent Unique Identifiers (UIDs) and ephemeral, sequential Human-friendly IDs (HIDs).

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
