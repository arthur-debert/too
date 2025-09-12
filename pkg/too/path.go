package too

import (
	"os"
	"path/filepath"
)

// ResolveCollectionPath resolves the collection file path using the same logic as init command.
// If explicitPath is provided, it's used as-is.
// Otherwise, defaults to .todos in current directory, or ~/.todos.json if it exists.
func ResolveCollectionPath(explicitPath string) string {
	if explicitPath != "" {
		return explicitPath
	}

	// Check if ~/.todos.json exists (home directory default)
	home, err := os.UserHomeDir()
	if err == nil {
		homeDefault := filepath.Join(home, ".todos.json")
		if _, err := os.Stat(homeDefault); err == nil {
			return homeDefault
		}
	}

	// Default to current directory
	return ".todos"
}