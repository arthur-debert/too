package reorder

import (
	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/arthur-debert/tdh/pkg/tdh/store"
)

// Options contains options for the reorder command
type Options struct {
	CollectionPath string
}

// Result contains the result of the reorder command
type Result struct {
	ReorderedCount int
	Todos          []*models.Todo
}

// Execute reorders todos by sorting them by their current position and reassigning sequential positions
func Execute(opts Options) (*Result, error) {
	s := store.NewStore(opts.CollectionPath)
	var finalCollection *models.Collection
	var originalPositions map[string]int

	err := s.Update(func(collection *models.Collection) error {
		// Store original positions by ID to calculate changes
		originalPositions = make(map[string]int)
		for _, todo := range collection.Todos {
			originalPositions[todo.ID] = todo.Position
		}

		// Use the collection's Reorder method
		collection.Reorder()

		// Store reference to the collection (safe because Update works on a clone)
		finalCollection = collection
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Calculate how many todos had their position changed
	count := 0
	for _, todo := range finalCollection.Todos {
		if originalPos, exists := originalPositions[todo.ID]; exists && originalPos != todo.Position {
			count++
		}
	}

	return &Result{
		ReorderedCount: count,
		Todos:          finalCollection.Todos,
	}, nil
}
