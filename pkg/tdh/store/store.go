package store

import "github.com/arthur-debert/tdh/pkg/models"

// Store defines the interface for persistence operations.
type Store interface {
	Load() (*models.Collection, error)
	Save(*models.Collection) error
	Exists() bool
	Update(func(collection *models.Collection) error) error
	Path() string // Returns the path where the store persists data
}
