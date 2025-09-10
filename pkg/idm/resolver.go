package idm

import (
	"fmt"
	"strconv"
	"strings"
)

// ResolvePositionPath traverses a hierarchical structure to find a UID
// based on a dot-notation position path (e.g., "1.2.3").
// It resolves each HID in the path sequentially, using the resulting UID as the
// scope for the next part of the path.
func (r *Registry) ResolvePositionPath(startScope, path string) (string, error) {
	parts := strings.Split(path, ".")
	currentScope := startScope
	var currentUID string

	for i, part := range parts {
		pos, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return "", fmt.Errorf("invalid position '%s' in path: %w", part, err)
		}
		if pos < 1 {
			return "", fmt.Errorf("position must be >= 1, got %d", pos)
		}

		uid, err := r.ResolveHID(currentScope, uint(pos))
		if err != nil {
			// Provide a more user-friendly error
			errorPath := strings.Join(parts[:i+1], ".")
			return "", fmt.Errorf("no item found at position '%s'", errorPath)
		}

		currentUID = uid
		currentScope = uid // The resolved UID becomes the scope for the next level.
	}

	return currentUID, nil
}
