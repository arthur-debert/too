package store

import (
	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/arthur-debert/tdh/pkg/tdh/store/internal"
)

// Query is an alias for the internal Query struct, exposing it to the public API.
type Query = internal.Query

// FindResult is an alias for the internal FindResult struct, exposing it to the public API.
type FindResult = internal.FindResult

// Store defines the interface for persistence operations.
type Store interface {
	Load() (*models.Collection, error)
	Save(*models.Collection) error
	Exists() bool
	Update(func(collection *models.Collection) error) error
	Find(query Query) (*FindResult, error)
	Path() string // Returns the path where the store persists data
}
