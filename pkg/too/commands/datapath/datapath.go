package datapath

import (
	"os"
	"path/filepath"
	
	"github.com/arthur-debert/too/pkg/lipbalm"
)

// ResolveCollectionPath resolves the collection file path using the following order:
// 1. If explicitPath is provided, use it as-is
// 2. Search current directory and parent directories for .todos.json file (like git)
// 3. Check TODO_DB_PATH environment variable
// 4. Fall back to ~/.todos.json if it exists
// 5. Default to .todos.json in current directory
func ResolveCollectionPath(explicitPath string) string {
	if explicitPath != "" {
		return explicitPath
	}

	// Check TODO_DB_PATH environment variable
	if envPath := os.Getenv("TODO_DB_PATH"); envPath != "" {
		return envPath
	}

	// Search upward for .todos.json file (like git does for .git)
	dir, err := os.Getwd()
	if err == nil {
		for {
			todosPath := filepath.Join(dir, ".todos.json")
			if _, err := os.Stat(todosPath); err == nil {
				return todosPath
			}
			
			parent := filepath.Dir(dir)
			if parent == dir {
				// Reached root directory
				break
			}
			dir = parent
		}
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
	return ".todos.json"
}

// Options holds the options for the datapath command
type Options struct {
	CollectionPath string
}

// Execute shows the path to the data file
func Execute(opts Options) (*lipbalm.Message, error) {
	// Use the unified path resolution function
	storePath := ResolveCollectionPath(opts.CollectionPath)

	// Get the absolute path
	absPath, err := filepath.Abs(storePath)
	if err != nil {
		// If we can't get absolute path, just return the path as is
		absPath = storePath
	}

	// Return the path as a plain message for proper rendering
	return lipbalm.NewPlainMessage(absPath), nil
}
