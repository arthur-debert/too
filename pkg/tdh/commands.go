package tdh

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// InitOptions contains options for the init command
type InitOptions struct {
	DBPath string
}

// InitResult contains the result of the init command
type InitResult struct {
	DBPath  string
	Created bool
	Message string
}

// Init initializes a new todo collection
func Init(opts InitOptions) (*InitResult, error) {
	dbPath := opts.DBPath
	if dbPath == "" {
		dbPath = GetDBPath()
		if dbPath == "" {
			// Use default home directory path
			home, err := os.UserHomeDir()
			if err != nil {
				return nil, fmt.Errorf("failed to get home directory: %w", err)
			}
			dbPath = filepath.Join(home, ".todos.json")
		}
	}

	created, err := CreateStoreFileIfNeeded(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create store file: %w", err)
	}

	result := &InitResult{
		DBPath:  dbPath,
		Created: created,
	}

	if created {
		result.Message = fmt.Sprintf("Initialized empty tdh collection in %s", dbPath)
	} else {
		result.Message = fmt.Sprintf("Reinitialized existing tdh collection in %s", dbPath)
	}

	return result, nil
}

// AddOptions contains options for the add command
type AddOptions struct {
	CollectionPath string
}

// AddResult contains the result of the add command
type AddResult struct {
	Todo *Todo
}

// Add adds a new todo to the collection
func Add(text string, opts AddOptions) (*AddResult, error) {
	if text == "" {
		return nil, fmt.Errorf("todo text cannot be empty")
	}

	collection, err := loadCollection(opts.CollectionPath)
	if err != nil {
		return nil, err
	}

	todo := collection.CreateTodo(text)
	if err := collection.Save(); err != nil {
		return nil, fmt.Errorf("failed to save collection: %w", err)
	}

	return &AddResult{Todo: todo}, nil
}

// ModifyOptions contains options for the modify command
type ModifyOptions struct {
	CollectionPath string
}

// ModifyResult contains the result of the modify command
type ModifyResult struct {
	Todo    *Todo
	OldText string
	NewText string
}

// Modify modifies the text of an existing todo
func Modify(id int, newText string, opts ModifyOptions) (*ModifyResult, error) {
	if newText == "" {
		return nil, fmt.Errorf("new todo text cannot be empty")
	}

	collection, err := loadCollection(opts.CollectionPath)
	if err != nil {
		return nil, err
	}

	todo, err := collection.Find(id)
	if err != nil {
		return nil, fmt.Errorf("todo not found: %w", err)
	}

	oldText := todo.Text
	todo.Text = newText

	if err := collection.Save(); err != nil {
		return nil, fmt.Errorf("failed to save collection: %w", err)
	}

	return &ModifyResult{
		Todo:    todo,
		OldText: oldText,
		NewText: newText,
	}, nil
}

// ToggleOptions contains options for the toggle command
type ToggleOptions struct {
	CollectionPath string
}

// ToggleResult contains the result of the toggle command
type ToggleResult struct {
	Todo      *Todo
	OldStatus string
	NewStatus string
}

// Toggle toggles the status of a todo
func Toggle(id int, opts ToggleOptions) (*ToggleResult, error) {
	collection, err := loadCollection(opts.CollectionPath)
	if err != nil {
		return nil, err
	}

	todo, err := collection.Find(id)
	if err != nil {
		return nil, fmt.Errorf("todo not found: %w", err)
	}

	oldStatus := todo.Status
	todo.Toggle()
	newStatus := todo.Status

	if err := collection.Save(); err != nil {
		return nil, fmt.Errorf("failed to save collection: %w", err)
	}

	return &ToggleResult{
		Todo:      todo,
		OldStatus: oldStatus,
		NewStatus: newStatus,
	}, nil
}

// CleanOptions contains options for the clean command
type CleanOptions struct {
	CollectionPath string
}

// CleanResult contains the result of the clean command
type CleanResult struct {
	RemovedCount int
	RemovedTodos []*Todo
	ActiveCount  int
}

// Clean removes finished todos from the collection
func Clean(opts CleanOptions) (*CleanResult, error) {
	collection, err := loadCollection(opts.CollectionPath)
	if err != nil {
		return nil, err
	}

	// Collect finished todos before cleaning
	var removedTodos []*Todo
	for _, todo := range collection.Todos {
		if todo.Status == "done" {
			removedTodos = append(removedTodos, todo)
		}
	}

	activeCount := collection.RemoveFinishedTodos()

	if err := collection.Save(); err != nil {
		return nil, fmt.Errorf("failed to save collection: %w", err)
	}

	return &CleanResult{
		RemovedCount: len(removedTodos),
		RemovedTodos: removedTodos,
		ActiveCount:  activeCount,
	}, nil
}

// ReorderOptions contains options for the reorder command
type ReorderOptions struct {
	CollectionPath string
}

// ReorderResult contains the result of the reorder command
type ReorderResult struct {
	TodoA *Todo
	TodoB *Todo
}

// Reorder swaps the position of two todos
func Reorder(idA, idB int, opts ReorderOptions) (*ReorderResult, error) {
	collection, err := loadCollection(opts.CollectionPath)
	if err != nil {
		return nil, err
	}

	if err := collection.Swap(idA, idB); err != nil {
		return nil, fmt.Errorf("failed to swap todos: %w", err)
	}

	if err := collection.Save(); err != nil {
		return nil, fmt.Errorf("failed to save collection: %w", err)
	}

	todoA, _ := collection.Find(idA)
	todoB, _ := collection.Find(idB)

	return &ReorderResult{
		TodoA: todoA,
		TodoB: todoB,
	}, nil
}

// SearchOptions contains options for the search command
type SearchOptions struct {
	CollectionPath string
	CaseSensitive  bool
}

// SearchResult contains the result of the search command
type SearchResult struct {
	Query        string
	MatchedTodos []*Todo
	TotalCount   int
}

// Search searches for todos containing the query string
func Search(query string, opts SearchOptions) (*SearchResult, error) {
	if query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	collection, err := loadCollection(opts.CollectionPath)
	if err != nil {
		return nil, err
	}

	var matchedTodos []*Todo
	searchQuery := query
	if !opts.CaseSensitive {
		searchQuery = strings.ToLower(query)
	}

	for _, todo := range collection.Todos {
		todoText := todo.Text
		if !opts.CaseSensitive {
			todoText = strings.ToLower(todoText)
		}
		if strings.Contains(todoText, searchQuery) {
			matchedTodos = append(matchedTodos, todo)
		}
	}

	return &SearchResult{
		Query:        query,
		MatchedTodos: matchedTodos,
		TotalCount:   len(collection.Todos),
	}, nil
}

// ListOptions contains options for listing todos
type ListOptions struct {
	CollectionPath string
}

// ListResult contains the result of listing todos
type ListResult struct {
	Todos      []*Todo
	TotalCount int
	DoneCount  int
}

// List returns all todos in the collection
func List(opts ListOptions) (*ListResult, error) {
	collection, err := loadCollection(opts.CollectionPath)
	if err != nil {
		return nil, err
	}

	doneCount := 0
	for _, todo := range collection.Todos {
		if todo.Status == "done" {
			doneCount++
		}
	}

	return &ListResult{
		Todos:      collection.Todos,
		TotalCount: len(collection.Todos),
		DoneCount:  doneCount,
	}, nil
}

// loadCollection loads a collection from the specified path or default
func loadCollection(path string) (*Collection, error) {
	if path == "" {
		path = GetDBPath()
		if path == "" {
			// Use default home directory path
			home, err := os.UserHomeDir()
			if err != nil {
				return nil, fmt.Errorf("failed to get home directory: %w", err)
			}
			path = filepath.Join(home, ".todos.json")
		}
	}

	// Check if we're dealing with a directory or file
	if filepath.Ext(path) == "" {
		// If no extension, assume it's a directory and append the default filename
		path = filepath.Join(path, ".todos.json")
	}

	collection := NewCollection(path)
	if err := collection.Load(); err != nil {
		return nil, fmt.Errorf("failed to load collection: %w", err)
	}

	return collection, nil
}
