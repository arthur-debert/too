package clean

import (
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/store"
)

// Options contains options for the clean command
type Options struct {
	CollectionPath string
}

// Result contains the result of the clean command
type Result struct {
	RemovedCount int
	RemovedTodos []*models.Todo
	ActiveCount  int
}

// Execute removes finished todos from the collection
func Execute(opts Options) (*Result, error) {
	s := store.NewStore(opts.CollectionPath)

	// Create direct workflow manager instead of using store.Update
	manager, err := store.NewDirectWorkflowManager(s, opts.CollectionPath)
	if err != nil {
		return nil, err
	}

	// Use the manager's integrated clean operation
	removedTodos, activeCount, err := manager.CleanFinishedTodos()
	if err != nil {
		return nil, err
	}

	// Save the changes through the manager
	err = manager.Save()
	if err != nil {
		return nil, err
	}

	return &Result{
		RemovedCount: len(removedTodos),
		RemovedTodos: removedTodos,
		ActiveCount:  activeCount,
	}, nil
}

