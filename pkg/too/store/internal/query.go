package internal

import (
	"strings"

	"github.com/arthur-debert/too/pkg/too/models"
)

// Query defines parameters for finding todos.
// All fields are optional - nil pointer fields are ignored during filtering.
// Multiple criteria are combined with AND logic (all must match).
//
// This struct is designed for easy extensibility - new query parameters
// can be added as needed without breaking existing code.
type Query struct {
	// Status filters todos by their status (e.g., "done", "pending").
	// Nil means no status filtering.
	Status *string

	// TextContains filters todos whose text contains this substring.
	// Nil means no text filtering.
	TextContains *string

	// CaseSensitive controls whether text search is case-sensitive.
	// Defaults to false (case-insensitive search).
	// Note: This enhancement was added beyond the original design to improve usability.
	CaseSensitive bool
}

// MatchesTodo checks if a todo matches the query criteria.
// This helper function encapsulates the common filtering logic used by all store implementations.
func (q Query) MatchesTodo(todo *models.Todo) bool {
	// Apply status filter
	if q.Status != nil && string(todo.GetStatus()) != *q.Status {
		return false
	}

	// Apply text containment filter
	if q.TextContains != nil {
		text := todo.Text
		searchText := *q.TextContains
		if !q.CaseSensitive {
			text = strings.ToLower(text)
			searchText = strings.ToLower(searchText)
		}
		if !strings.Contains(text, searchText) {
			return false
		}
	}

	return true
}

// FilterTodos applies the query to a slice of todos and returns matching results.
// This helper function provides consistent filtering behavior across all store implementations.
func (q Query) FilterTodos(todos []*models.Todo) []*models.Todo {
	var results []*models.Todo
	for _, todo := range todos {
		if q.MatchesTodo(todo) {
			results = append(results, todo)
		}
	}
	return results
}
