package idm

import "fmt"

// Registry manages the mapping of stable Unique Identifiers (UIDs) to
// human-friendly, sequential IDs (HIDs) within different contexts, called "scopes".
type Registry struct {
	// scopes maps a scope name to an ordered list of UIDs.
	// The index in the slice corresponds to the HID (index + 1).
	scopes map[string][]string
}

// NewRegistry creates and initializes a new IDM Registry.
func NewRegistry() *Registry {
	return &Registry{
		scopes: make(map[string][]string),
	}
}

// RebuildScope clears and repopulates a scope using data from the store adapter.
// This is the primary way to sync the registry with the source of truth.
func (r *Registry) RebuildScope(adapter StoreAdapter, scope string) error {
	uids, err := adapter.GetChildren(scope)
	if err != nil {
		return fmt.Errorf("could not get children for scope '%s': %w", scope, err)
	}
	r.scopes[scope] = uids
	return nil
}

// Add adds a UID to a scope and returns its new HID.
// It assumes the UID is not already present.
func (r *Registry) Add(scope, uid string) uint {
	r.scopes[scope] = append(r.scopes[scope], uid)
	return uint(len(r.scopes[scope]))
}

// Remove removes a UID from a scope. If the scope or UID does not exist,
// it does nothing.
func (r *Registry) Remove(scope, uid string) {
	uids, ok := r.scopes[scope]
	if !ok {
		return
	}

	newUIDs := make([]string, 0, len(uids))
	for _, id := range uids {
		if id != uid {
			newUIDs = append(newUIDs, id)
		}
	}
	r.scopes[scope] = newUIDs
}

// ResolveHID converts a 1-based HID within a given scope to its corresponding UID.
func (r *Registry) ResolveHID(scope string, hid uint) (string, error) {
	uids, ok := r.scopes[scope]
	if !ok {
		return "", fmt.Errorf("scope '%s' not found", scope)
	}
	if hid < 1 || int(hid) > len(uids) {
		return "", fmt.Errorf("invalid HID %d in scope '%s' (contains %d items)", hid, scope, len(uids))
	}
	// HIDs are 1-based, slice indices are 0-based.
	return uids[hid-1], nil
}

// GetUIDs returns a concatenated list of UIDs from one or more scopes.
// The order is preserved within each scope, and scopes are processed in the order they are provided.
func (r *Registry) GetUIDs(scopes ...string) []string {
	var allUIDs []string
	for _, scope := range scopes {
		if uids, ok := r.scopes[scope]; ok {
			allUIDs = append(allUIDs, uids...)
		}
	}
	return allUIDs
}
