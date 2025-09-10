# IDM: Abstracted ID Management

## 1. Motivation

Many command-line applications that manage collections of items face a common user interface challenge: the need for both stable, unique identifiers for internal logic and simple, human-friendly identifiers for user interaction.

-   **Unique Identifiers (UIDs):** These are permanent, machine-readable IDs (like UUIDs) that ensure operations always target the correct item, regardless of its state or the user's current view.
-   **Human-friendly Identifiers (HIDs):** These are simple, sequential integers (like `1`, `2`, `3`) that are easy for users to see, remember, and type.

In `too`, this was solved with a dual-ID system where HIDs were stored as a `Position` field on the `Todo` model. This `Position` was frequently recalculated and reset, leading to logic that was tightly coupled with the data model and storage layer. This pattern was identified as a recurring problem in other applications, prompting the need for a generic, reusable solution.

The `idm` (ID Manager) package was created to abstract this dual-ID management into a decoupled, reusable library.

## 2. Design

The core design of the `idm` package is based on decoupling the ID management logic from the application's specific data structures and storage implementation.

### Core Concepts

-   **Registry:** The central component of the `idm` package. It is an in-memory map that holds the relationship between UIDs and their corresponding HIDs. It does not store the application's actual data, only the UIDs.
-   **Scope:** A named context in which HIDs are assigned. A scope is simply an ordered list of UIDs. For `too`, a scope is the UID of a parent todo, with a special `"root"` scope for top-level items. This concept is flexible enough to model other scenarios, such as separate scopes for `"active"`, `"archived"`, or `"pinned"` items.
-   **Store Adapter:** A crucial interface that acts as a bridge between the generic `Registry` and the application's specific data store. The application must implement this interface, which allows the `Registry` to query for the structure of the data (e.g., "what are the child UIDs of this parent UID?") without needing to know anything about the underlying models or storage format (e.g., JSON, SQL, files).

### Architecture Flow

The interaction between the components follows this pattern:

1.  The application initializes its native data store (e.g., `too`'s `JSONFileStore`).
2.  It creates an `IDMStoreAdapter` that wraps the native store.
3.  It creates an `idm.Registry`.
4.  Before processing a command, the application populates the `Registry` by asking the `Adapter` for all available scopes and then rebuilding each scope.
5.  When a user provides a HID (e.g., `too edit 1.2`), the application uses the `Registry` to resolve the HID path into a stable UID.
6.  The application then uses this UID to perform operations on its data models.

This design ensures that all the complex logic for parsing position paths and managing HID-to-UID mapping is centralized and completely separate from the application's business logic.

## 3. API

The `idm` package exposes a simple and focused API.

### `StoreAdapter` Interface

This is the interface the application must implement to connect its data store to the registry.

```go
// StoreAdapter defines the methods the Registry needs to interact with
// the underlying data store.
type StoreAdapter interface {
	// GetChildren returns an ordered list of UIDs for a given parent UID.
	GetChildren(parentUID string) ([]string, error)

	// GetScopes returns all possible scopes that the registry might need to manage.
	GetScopes() ([]string, error)
}
```

### `Registry` API

The primary struct for all ID management operations. See the footnote for the full implementation [^1].

-   `NewRegistry() *Registry`: Creates a new, empty registry.
-   `RebuildScope(adapter StoreAdapter, scope string) error`: Populates a scope with UIDs by querying the adapter.
-   `ResolveHID(scope string, hid uint) (string, error)`: Resolves a single HID within a scope to its corresponding UID.
-   `ResolvePositionPath(startScope, path string) (string, error)`: Resolves a dot-notation path (e.g., `"1.2.1"`) into a UID by traversing nested scopes.

[^1]: **`registry.go` Implementation**
    ```go
    package idm

    import "fmt"

    type Registry struct {
    	scopes map[string][]string
    }

    func NewRegistry() *Registry {
    	return &Registry{
    		scopes: make(map[string][]string),
    	}
    }

    func (r *Registry) RebuildScope(adapter StoreAdapter, scope string) error {
    	uids, err := adapter.GetChildren(scope)
    	if err != nil {
    		return fmt.Errorf("could not get children for scope '%s': %w", scope, err)
    	}
    	r.scopes[scope] = uids
    	return nil
    }

    func (r *Registry) ResolveHID(scope string, hid uint) (string, error) {
    	uids, ok := r.scopes[scope]
    	if !ok {
    		return "", fmt.Errorf("scope '%s' not found", scope)
    	}
    	if hid < 1 || int(hid) > len(uids) {
    		return "", fmt.Errorf("invalid HID %d in scope '%s'", hid, scope)
    	}
    	return uids[hid-1], nil
    }
    ```
