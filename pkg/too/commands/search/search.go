package search

import (
	"fmt"
	"strings"

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
	MatchedTodos []*models.IDMTodo
	TotalCount   int
}

// Execute searches for todos containing the query string
func Execute(query string, opts Options) (*Result, error) {
	if query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	// Create IDM store and manager
	idmStore := store.NewIDMStore(opts.CollectionPath)
	manager, err := store.NewPureIDMManager(idmStore, opts.CollectionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create manager: %w", err)
	}

	// Get all todos (both pending and done)
	allTodos := manager.ListAll()
	
	// Search through todos
	var matchedIDMTodos []*models.IDMTodo
	searchQuery := query
	if !opts.CaseSensitive {
		searchQuery = strings.ToLower(query)
	}
	
	for _, todo := range allTodos {
		text := todo.Text
		if !opts.CaseSensitive {
			text = strings.ToLower(text)
		}
		
		if strings.Contains(text, searchQuery) {
			matchedIDMTodos = append(matchedIDMTodos, todo)
		}
	}
	
	// CRITICAL: Attach IDM position paths for consistent display
	manager.AttachPositionPaths(matchedIDMTodos)
	
	return &Result{
		Query:        query,
		MatchedTodos: matchedIDMTodos,
		TotalCount:   len(allTodos), // Total count is all todos, not just matches
	}, nil
}