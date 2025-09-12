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

// RemoveScope completely removes a scope and all its UIDs from the registry.
func (r *Registry) RemoveScope(scope string) {
	delete(r.scopes, scope)
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

// GetHID returns the HID for a given UID within a scope, or 0 if not found.
func (r *Registry) GetHID(scope, uid string) uint {
	uids, ok := r.scopes[scope]
	if !ok {
		return 0
	}
	for i, id := range uids {
		if id == uid {
			return uint(i + 1) // HIDs are 1-based
		}
	}
	return 0
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

// ResolvePositionPaths translates multiple dot-notation paths into UIDs.
// It returns a map of path -> UID and the first error encountered.
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

// GetPositionPath returns the position path for a given UID by traversing up
// the hierarchy to build the dot-notation path.
func (r *Registry) GetPositionPath(startScope, targetUID string, adapter StoreAdapter) (string, error) {
	// Check if targetUID is directly in the startScope
	if hid := r.GetHID(startScope, targetUID); hid > 0 {
		return fmt.Sprintf("%d", hid), nil
	}

	// Search all scopes for the targetUID and build path recursively
	for scope := range r.scopes {
		if hid := r.GetHID(scope, targetUID); hid > 0 {
			// Found it in this scope, now build the path
			if scope == startScope {
				return fmt.Sprintf("%d", hid), nil
			}
			
			// Get the path to the scope, then append this HID
			parentPath, err := r.GetPositionPath(startScope, scope, adapter)
			if err != nil {
				continue // Try other scopes
			}
			return fmt.Sprintf("%s.%d", parentPath, hid), nil
		}
	}

	return "", fmt.Errorf("UID '%s' not found in any scope accessible from '%s'", targetUID, startScope)
}
