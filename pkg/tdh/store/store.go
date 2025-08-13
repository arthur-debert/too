package store

import (
	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/arthur-debert/tdh/pkg/tdh/store/internal"
)

// Query is an alias for the internal Query struct, exposing it to the public API.
type Query = internal.Query

// Store defines the interface for persistence operations.
type Store interface {
	Load() (*models.Collection, error)
	Save(*models.Collection) error
	Exists() bool
	Update(func(collection *models.Collection) error) error
	// Find retrieves todos matching the given query criteria.
	// All query fields are optional - nil fields are ignored.
	// Multiple criteria are combined with AND logic.
	// Returns an empty slice (not nil) when no todos match.
	Find(query Query) ([]*models.Todo, error)
	Path() string // Returns the path where the store persists data
}
