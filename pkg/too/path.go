package too

import (
	"os"
	"path/filepath"
)

// ResolveCollectionPath resolves the collection file path using the following order:
// 1. If explicitPath is provided, use it as-is
// 2. Search current directory and parent directories for .todos file (like git)
// 3. Check TODO_DB_PATH environment variable
// 4. Fall back to ~/.todos.json if it exists
// 5. Default to .todos in current directory
func ResolveCollectionPath(explicitPath string) string {
	if explicitPath != "" {
		return explicitPath
	}

	// Search upward for .todos file (like git does for .git)
	dir, err := os.Getwd()
	if err == nil {
		for {
			todosPath := filepath.Join(dir, ".todos")
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

	// Check TODO_DB_PATH environment variable
	if envPath := os.Getenv("TODO_DB_PATH"); envPath != "" {
		return envPath
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