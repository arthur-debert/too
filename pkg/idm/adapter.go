package idm

// StoreAdapter defines the methods the Registry needs to interact with
// the underlying data store. It provides the registry with the necessary
// structural information about items without exposing implementation details.
type StoreAdapter interface {
	// GetChildren returns an ordered list of UIDs for a given parent UID.
	// For root items, the parent UID can be a special constant (e.g., "root").
	GetChildren(parentUID string) ([]string, error)

	// GetScopes returns all possible scopes that the registry might need to manage.
	// In a hierarchical system like too, this would be all UIDs that serve as parents,
	// plus a "root" scope.
	GetScopes() ([]string, error)
}
