package init

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/arthur-debert/too/pkg/too/store"
)

// Options contains options for the init command
type Options struct {
	DBPath     string
	UseHomeDir bool // Create .todos file in home directory instead of current directory
}

// Result contains the result of the init command
type Result struct {
	DBPath  string
	Created bool
	Message string
}

// Execute initializes a new todo collection
func Execute(opts Options) (*Result, error) {
	var storePath string

	if opts.DBPath != "" {
		// Use explicit path if provided
		storePath = opts.DBPath
	} else if opts.UseHomeDir {
		// Use home directory
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		storePath = filepath.Join(home, ".todos.json")
	} else {
		// Use current directory (default)
		storePath = ".todos.json"
	}

	// Check if database already exists
	exists := false
	if _, err := os.Stat(storePath); err == nil {
		exists = true
	}

	// Create the nanostore database
	adapter, err := store.NewNanoStoreAdapter(storePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create store: %w", err)
	}
	defer adapter.Close()

	if !exists {
		return &Result{
			DBPath:  storePath,
			Created: true,
			Message: fmt.Sprintf("Initialized empty too collection in %s", storePath),
		}, nil
	}

	return &Result{
		DBPath:  storePath,
		Created: false,
		Message: fmt.Sprintf("Reinitialized existing too collection in %s", storePath),
	}, nil
}
