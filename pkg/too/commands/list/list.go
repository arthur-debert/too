package list

import (
	"fmt"

	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/store"
)

// Options contains options for listing todos
type Options struct {
	CollectionPath string
	ShowDone       bool
	ShowAll        bool
}

// Result contains the result of listing todos
type Result struct {
	Todos      []*models.IDMTodo
	TotalCount int
	DoneCount  int
}

// Execute returns todos from the collection using pure IDM.
func Execute(opts Options) (*Result, error) {
	// Create IDM store and manager
	idmStore := store.NewIDMStore(opts.CollectionPath)
	manager, err := store.NewPureIDMManager(idmStore, opts.CollectionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create manager: %w", err)
	}

	// Get todos based on options
	var idmTodos []*models.IDMTodo
	if opts.ShowAll {
		idmTodos = manager.ListAll()
	} else if opts.ShowDone {
		idmTodos = manager.ListArchived()
	} else {
		idmTodos = manager.ListActive()
	}

	// Get counts
	totalCount, doneCount := manager.CountTodos()

	// Build hierarchical structure for display compatibility
	// Convert IDMTodos to display format with hierarchy
	todos := make([]*models.IDMTodo, 0)
	if len(idmTodos) > 0 {
		// For now, just return flat structure
		// The display layer will handle hierarchy building if needed
		todos = idmTodos
	}

	return &Result{
		Todos:      todos,
		TotalCount: totalCount,
		DoneCount:  doneCount,
	}, nil
}

