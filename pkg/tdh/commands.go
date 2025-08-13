package tdh

import (
	"fmt"

	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/arthur-debert/tdh/pkg/tdh/store"
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
	s := store.NewStore(opts.DBPath)

	if !s.Exists() {
		// Create an empty collection to initialize the file
		if err := s.Save(models.NewCollection(s.Path())); err != nil {
			return nil, fmt.Errorf("failed to create store file: %w", err)
		}
		return &InitResult{
			DBPath:  s.Path(),
			Created: true,
			Message: fmt.Sprintf("Initialized empty tdh collection in %s", s.Path()),
		}, nil
	}

	return &InitResult{
		DBPath:  s.Path(),
		Created: false,
		Message: fmt.Sprintf("Reinitialized existing tdh collection in %s", s.Path()),
	}, nil
}

// AddOptions contains options for the add command
type AddOptions struct {
	CollectionPath string
}

// AddResult contains the result of the add command
type AddResult struct {
	Todo *models.Todo
}

// Add adds a new todo to the collection
func Add(text string, opts AddOptions) (*AddResult, error) {
	if text == "" {
		return nil, fmt.Errorf("todo text cannot be empty")
	}

	s := store.NewStore(opts.CollectionPath)
	var todo *models.Todo

	err := s.Update(func(collection *models.Collection) error {
		todo = collection.CreateTodo(text)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to add todo: %w", err)
	}

	return &AddResult{Todo: todo}, nil
}

// ModifyOptions contains options for the modify command
type ModifyOptions struct {
	CollectionPath string
}

// ModifyResult contains the result of the modify command
type ModifyResult struct {
	Todo    *models.Todo
	OldText string
	NewText string
}

// Modify modifies the text of an existing todo
func Modify(id int, newText string, opts ModifyOptions) (*ModifyResult, error) {
	if newText == "" {
		return nil, fmt.Errorf("new todo text cannot be empty")
	}

	s := store.NewStore(opts.CollectionPath)
	var todo *models.Todo
	var oldText string

	err := s.Update(func(collection *models.Collection) error {
		var err error
		todo, err = Find(collection, id)
		if err != nil {
			return fmt.Errorf("todo not found: %w", err)
		}
		oldText = todo.Text
		todo.Text = newText
		return nil
	})

	if err != nil {
		return nil, err
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
	Todo      *models.Todo
	OldStatus string
	NewStatus string
}

// Toggle toggles the status of a todo
func Toggle(id int, opts ToggleOptions) (*ToggleResult, error) {
	s := store.NewStore(opts.CollectionPath)
	var todo *models.Todo
	var oldStatus string
	var newStatus string

	err := s.Update(func(collection *models.Collection) error {
		var err error
		todo, err = Find(collection, id)
		if err != nil {
			return fmt.Errorf("todo not found: %w", err)
		}
		oldStatus = todo.Status
		todo.Toggle()
		newStatus = todo.Status
		return nil
	})

	if err != nil {
		return nil, err
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
	RemovedTodos []*models.Todo
	ActiveCount  int
}

// Clean removes finished todos from the collection
func Clean(opts CleanOptions) (*CleanResult, error) {
	s := store.NewStore(opts.CollectionPath)
	var removedTodos []*models.Todo
	var activeCount int

	err := s.Update(func(collection *models.Collection) error {
		for _, todo := range collection.Todos {
			if todo.Status == "done" {
				removedTodos = append(removedTodos, todo)
			}
		}
		activeCount = RemoveFinishedTodos(collection)
		return nil
	})

	if err != nil {
		return nil, err
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
	TodoA *models.Todo
	TodoB *models.Todo
}

// Reorder swaps the position of two todos
func Reorder(idA, idB int, opts ReorderOptions) (*ReorderResult, error) {
	s := store.NewStore(opts.CollectionPath)
	var todoA, todoB *models.Todo

	err := s.Update(func(collection *models.Collection) error {
		if err := Swap(collection, idA, idB); err != nil {
			return fmt.Errorf("failed to swap todos: %w", err)
		}
		todoA, _ = Find(collection, idA)
		todoB, _ = Find(collection, idB)
		return nil
	})

	if err != nil {
		return nil, err
	}

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
	MatchedTodos []*models.Todo
	TotalCount   int
}

// Search searches for todos containing the query string
func Search(query string, opts SearchOptions) (*SearchResult, error) {
	if query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	s := store.NewStore(opts.CollectionPath)

	// Build query for Find API
	q := store.Query{
		TextContains:  &query,
		CaseSensitive: opts.CaseSensitive,
	}

	// Get matching todos using Find
	matchedTodos, err := s.Find(q)
	if err != nil {
		return nil, err
	}

	// Still need total count from full collection
	collection, err := s.Load()
	if err != nil {
		return nil, err
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
	ShowDone       bool
	ShowAll        bool
}

// ListResult contains the result of listing todos
type ListResult struct {
	Todos      []*models.Todo
	TotalCount int
	DoneCount  int
}

// List returns todos from the collection with optional filtering
func List(opts ListOptions) (*ListResult, error) {
	s := store.NewStore(opts.CollectionPath)
	collection, err := s.Load()
	if err != nil {
		return nil, err
	}

	todos := collection.Todos
	doneCount := 0

	// Count done todos
	for _, todo := range collection.Todos {
		if todo.Status == "done" {
			doneCount++
		}
	}

	// Apply filtering if not showing all
	if !opts.ShowAll {
		var filteredTodos []*models.Todo
		for _, todo := range collection.Todos {
			if opts.ShowDone && todo.Status == "done" {
				filteredTodos = append(filteredTodos, todo)
			} else if !opts.ShowDone && todo.Status != "done" {
				filteredTodos = append(filteredTodos, todo)
			}
		}
		todos = filteredTodos
	}

	return &ListResult{
		Todos:      todos,
		TotalCount: len(collection.Todos),
		DoneCount:  doneCount,
	}, nil
}
