package datapath

import (
	"fmt"
	"os"
	"path/filepath"
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

	// Search upward for .todos.json file (like git does for .git)
	dir, err := os.Getwd()
	if err == nil {
		for {
			todosPath := filepath.Join(dir, ".todos.json")
			if _, err := os.Stat(todosPath); err == nil {
				return todosPath
			}
			
			// Check for old .db file and migrate if found
			oldPath := filepath.Join(dir, ".todos.db")
			if _, err := os.Stat(oldPath); err == nil {
				// Try to rename the file
				if err := os.Rename(oldPath, todosPath); err == nil {
					os.Stderr.WriteString(fmt.Sprintf("Migrated %s to %s\n", oldPath, todosPath))
					return todosPath
				}
				// If rename fails (e.g., permission issues), use old path
				return oldPath
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
		
		// Check for old ~/.todos.db and migrate
		oldHomeDefault := filepath.Join(home, ".todos.db")
		if _, err := os.Stat(oldHomeDefault); err == nil {
			if err := os.Rename(oldHomeDefault, homeDefault); err == nil {
				os.Stderr.WriteString(fmt.Sprintf("Migrated %s to %s\n", oldHomeDefault, homeDefault))
				return homeDefault
			}
			return oldHomeDefault
		}
	}

	// Default to current directory
	return ".todos.json"
}

// Options holds the options for the datapath command
type Options struct {
	CollectionPath string
}

// Result represents the result of the datapath command
type Result struct {
	Path string
}

// Execute shows the path to the data file
func Execute(opts Options) (*Result, error) {
	// Use the unified path resolution function
	storePath := ResolveCollectionPath(opts.CollectionPath)

	// Get the absolute path
	absPath, err := filepath.Abs(storePath)
	if err != nil {
		// If we can't get absolute path, just return the path as is
		absPath = storePath
	}

	return &Result{
		Path: absPath,
	}, nil
}
