package idm

// StoreAdapter defines the core, read-only methods the Registry needs to
// interact with the underlying data store.
type StoreAdapter interface {
	// GetChildren returns an ordered list of UIDs for a given parent UID (scope).
	GetChildren(parentUID string) ([]string, error)

	// GetScopes returns all possible scopes that the registry might need to manage.
	GetScopes() ([]string, error)

	// GetAllUIDs returns all UIDs in the collection, regardless of status.
	// This is used for UID resolution.
	GetAllUIDs() ([]string, error)
}

// ManagedStoreAdapter is the new, richer adapter interface that an
// application must implement to use the Manager. It includes methods for
// writing and changing data.
type ManagedStoreAdapter interface {
	// Inherit the read-only methods from the core adapter.
	StoreAdapter

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