package datapath

import (
	"os"
	"path/filepath"

	"github.com/arthur-debert/too/pkg/too/store"
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
		// Use default path logic similar to init command
		// Check if ~/.todos.json exists (home directory default)
		home, err := os.UserHomeDir()
		if err == nil {
			homeDefault := filepath.Join(home, ".todos.json")
			if _, err := os.Stat(homeDefault); err == nil {
				storePath = homeDefault
			} else {
				// Default to current directory
				storePath = ".todos"
			}
		} else {
			// Fallback to current directory if can't get home
			storePath = ".todos"
		}
	}

	// Create an IDM store with the resolved path
	s := store.NewIDMStore(storePath)

	// Get the absolute path
	absPath, err := filepath.Abs(s.Path())
	if err != nil {
		// If we can't get absolute path, just return the path as is
		absPath = s.Path()
	}

	return &Result{
		Path: absPath,
	}, nil
}
