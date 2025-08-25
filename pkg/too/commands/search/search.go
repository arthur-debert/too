package search

import (
	"fmt"

	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/store"
)

// Options contains options for the search command
type Options struct {
	CollectionPath string
	CaseSensitive  bool
}

// Result contains the result of the search command
type Result struct {
	Query        string
	MatchedTodos []*models.Todo
	TotalCount   int
}

// Execute searches for todos containing the query string
func Execute(query string, opts Options) (*Result, error) {
	if query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	s := store.NewStore(opts.CollectionPath)

	// Build query for Find API
	q := store.Query{
		TextContains:  &query,
		CaseSensitive: opts.CaseSensitive,
	}

	// Get matching todos and counts using Find
	findResult, err := s.Find(q)
	if err != nil {
		return nil, err
	}

	return &Result{
		Query:        query,
		MatchedTodos: findResult.Todos,
		TotalCount:   findResult.TotalCount,
	}, nil
}
