package datapath

import (
	"os"
	"path/filepath"
)

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
	var storePath string

	if opts.CollectionPath != "" {
		// Use explicit path if provided
		storePath = opts.CollectionPath
	} else {
		// Search upward for .todos file (like git does for .git)
		dir, err := os.Getwd()
		if err == nil {
			found := false
			for {
				todosPath := filepath.Join(dir, ".todos.db")
				if _, err := os.Stat(todosPath); err == nil {
					storePath = todosPath
					found = true
					break
				}
				
				parent := filepath.Dir(dir)
				if parent == dir {
					// Reached root directory
					break
				}
				dir = parent
			}
			
			if !found {
				// Check TODO_DB_PATH environment variable
				if envPath := os.Getenv("TODO_DB_PATH"); envPath != "" {
					storePath = envPath
				} else {
					// Check if ~/.todos.db exists (home directory default)
					home, err := os.UserHomeDir()
					if err == nil {
						homeDefault := filepath.Join(home, ".todos.db")
						if _, err := os.Stat(homeDefault); err == nil {
							storePath = homeDefault
						} else {
							// Default to current directory
							storePath = ".todos.db"
						}
					} else {
						// Fallback to current directory if can't get home
						storePath = ".todos.db"
					}
				}
			}
		} else {
			// Fallback to current directory if can't get working directory
			storePath = ".todos.db"
		}
	}

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
