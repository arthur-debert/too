package init

import (
	"fmt"

	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/arthur-debert/tdh/pkg/tdh/store"
)

// Options contains options for the init command
type Options struct {
	DBPath string
}

// Result contains the result of the init command
type Result struct {
	DBPath  string
	Created bool
	Message string
}

// Execute initializes a new todo collection
func Execute(opts Options) (*Result, error) {
	s := store.NewStore(opts.DBPath)

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
