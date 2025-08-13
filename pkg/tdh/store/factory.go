package store

import (
	"os"
	"path/filepath"
)

// getDefaultPath returns the default path for the tdh store.
func getDefaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback if home dir is not available
		return ".tdh.json"
	}
	return filepath.Join(home, ".tdh.json")
}

// NewStore creates a new store based on the provided path.
// If path is empty, it uses the default path.
func NewStore(path string) Store {
	if path == "" {
		path = getDefaultPath()
	}
	return NewJSONFileStore(path)
}
