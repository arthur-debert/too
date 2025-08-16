package init

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/arthur-debert/tdh/pkg/tdh/store"
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
		storePath = ".todos"
	}

	s := store.NewStore(storePath)

	if !s.Exists() {
		// Create an empty collection to initialize the file
		if err := s.Save(models.NewCollection()); err != nil {
			return nil, fmt.Errorf("failed to create store file: %w", err)
		}
		return &Result{
			DBPath:  s.Path(),
			Created: true,
			Message: fmt.Sprintf("Initialized empty tdh collection in %s", s.Path()),
		}, nil
	}

	return &Result{
		DBPath:  s.Path(),
		Created: false,
		Message: fmt.Sprintf("Reinitialized existing tdh collection in %s", s.Path()),
	}, nil
}
