package list

import (
	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/arthur-debert/tdh/pkg/tdh/store"
)

// Options contains options for listing todos
type Options struct {
	CollectionPath string
	ShowDone       bool
	ShowAll        bool
}

// Result contains the result of listing todos
type Result struct {
	Todos      []*models.Todo
	TotalCount int
	DoneCount  int
}

// Execute returns todos from the collection with optional filtering
func Execute(opts Options) (*Result, error) {
	s := store.NewStore(opts.CollectionPath)

	// Build query based on options
	var query store.Query
	if !opts.ShowAll {
		status := string(models.StatusPending)
		if opts.ShowDone {
			status = string(models.StatusDone)
		}
		query.Status = &status
	}

	// Get filtered todos and counts using Find
	findResult, err := s.Find(query)
	if err != nil {
		return nil, err
	}

	return &Result{
		Todos:      findResult.Todos,
		TotalCount: findResult.TotalCount,
		DoneCount:  findResult.DoneCount,
	}, nil
}
