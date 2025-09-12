package search

import (
	"fmt"
	"sort"
	"strconv"
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
	
	// Sort by position path to match list command order
	sort.Slice(matchedIDMTodos, func(i, j int) bool {
		return comparePositionPaths(matchedIDMTodos[i].PositionPath, matchedIDMTodos[j].PositionPath)
	})
	
	return &Result{
		Query:        query,
		MatchedTodos: matchedIDMTodos,
		TotalCount:   len(allTodos), // Total count is all todos, not just matches
	}, nil
}

// comparePositionPaths compares two position paths (e.g., "1.1" vs "2.3") for sorting.
// Returns true if path1 should come before path2.
func comparePositionPaths(path1, path2 string) bool {
	// Handle empty paths (UIDs) - put them at the end
	if path1 == "" && path2 == "" {
		return false
	}
	if path1 == "" {
		return false // path1 is empty, comes after path2
	}
	if path2 == "" {
		return true // path2 is empty, path1 comes before
	}
	
	// Split paths into segments
	parts1 := strings.Split(path1, ".")
	parts2 := strings.Split(path2, ".")
	
	// Compare segment by segment
	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}
	
	for i := 0; i < maxLen; i++ {
		// If one path is shorter, it comes first (e.g., "1" before "1.1")
		if i >= len(parts1) {
			return true // path1 is shorter, comes first
		}
		if i >= len(parts2) {
			return false // path2 is shorter, comes first
		}
		
		// Convert to integers for proper numeric comparison
		num1, err1 := strconv.Atoi(parts1[i])
		num2, err2 := strconv.Atoi(parts2[i])
		
		// If either can't be parsed as number, fall back to string comparison
		if err1 != nil || err2 != nil {
			if parts1[i] < parts2[i] {
				return true
			}
			if parts1[i] > parts2[i] {
				return false
			}
			continue
		}
		
		// Numeric comparison
		if num1 < num2 {
			return true
		}
		if num1 > num2 {
			return false
		}
		// If equal, continue to next segment
	}
	
	// All compared segments are equal
	return false
}