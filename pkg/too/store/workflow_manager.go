package store

import (
	"fmt"

	"github.com/arthur-debert/too/pkg/idm"
	"github.com/arthur-debert/too/pkg/idm/workflow"
	"github.com/arthur-debert/too/pkg/too/config"
	"github.com/arthur-debert/too/pkg/too/models"
)

// WorkflowManager wraps the IDM workflow system for too's use cases.
type WorkflowManager struct {
	statusManager             *workflow.StatusManager
	adapter                   *WorkflowTodoAdapter
	registry                  *idm.Registry
	hierarchyMgr              *idm.Manager
	autoTransitionInProgress  map[string]bool // Guard against infinite recursion
}

// NewWorkflowManager creates a new workflow manager for the given collection.
func NewWorkflowManager(collection *models.Collection, collectionPath string) (*WorkflowManager, error) {
	// Create workflow adapter
	store := NewMemoryStoreFromCollection(collection)
	adapter, err := NewWorkflowTodoAdapter(store)
	if err != nil {
		return nil, fmt.Errorf("failed to create workflow adapter: %w", err)
	}

	// Create IDM manager for hierarchy operations
	hierarchyMgr, err := NewManagerFromCollection(collection)
	if err != nil {
		return nil, fmt.Errorf("failed to create IDM manager: %w", err)
	}

	// Load workflow configuration
	workflowConfigPath := config.GetWorkflowConfigPath(collectionPath)
	workflowConfig, err := config.LoadWorkflowConfig(workflowConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load workflow config: %w", err)
	}

	// Get effective workflow configuration
	effectiveConfig, err := workflowConfig.GetWorkflowConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get effective workflow config: %w", err)
	}

	// Create status manager
	statusManager, err := workflow.NewStatusManager(
		hierarchyMgr.Registry(),
		hierarchyMgr,
		adapter,
		effectiveConfig,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create status manager: %w", err)
	}

	return &WorkflowManager{
		statusManager:            statusManager,
		adapter:                  adapter,
		registry:                 hierarchyMgr.Registry(),
		hierarchyMgr:             hierarchyMgr,
		autoTransitionInProgress: make(map[string]bool),
	}, nil
}

// NewMemoryStoreFromCollection creates a memory store with the given collection.
func NewMemoryStoreFromCollection(collection *models.Collection) Store {
	// We need to create a memory store that contains the collection
	// This is a temporary approach for the migration
	return &memoryStoreWrapper{collection: collection}
}

// memoryStoreWrapper wraps a collection to implement the Store interface
type memoryStoreWrapper struct {
	collection *models.Collection
}

func (m *memoryStoreWrapper) Load() (*models.Collection, error) {
	return m.collection, nil
}

func (m *memoryStoreWrapper) Save(collection *models.Collection) error {
	// In memory store, we update the wrapped collection
	m.collection = collection
	return nil
}

func (m *memoryStoreWrapper) Update(fn func(*models.Collection) error) error {
	return fn(m.collection)
}

func (m *memoryStoreWrapper) Exists() bool {
	return m.collection != nil
}

func (m *memoryStoreWrapper) Find(query Query) (*FindResult, error) {
	// For workflow purposes, we don't need complex find operations
	// This is a minimal implementation
	return &FindResult{}, nil
}

func (m *memoryStoreWrapper) Path() string {
	return "memory://workflow"
}

// ResolvePositionPathInContext resolves a position path within the active context.
func (wm *WorkflowManager) ResolvePositionPathInContext(scope, path, context string) (string, error) {
	return wm.statusManager.ResolvePositionPathInContext(scope, path, context)
}

// SetStatus sets a status dimension for an item and triggers auto-transitions.
// Includes recursion guard to prevent infinite loops in auto-transitions.
func (wm *WorkflowManager) SetStatus(uid, dimension, value string) error {
	// Check if this item is already being processed to prevent infinite recursion
	if wm.autoTransitionInProgress[uid] {
		return nil // Skip to prevent recursion
	}
	
	// Mark this item as being processed
	wm.autoTransitionInProgress[uid] = true
	defer delete(wm.autoTransitionInProgress, uid)
	
	return wm.statusManager.SetStatus(uid, dimension, value)
}

// GetStatus gets a status dimension for an item.
func (wm *WorkflowManager) GetStatus(uid, dimension string) (string, error) {
	return wm.statusManager.GetStatus(uid, dimension)
}

// Transition performs a validated status transition.
func (wm *WorkflowManager) Transition(uid, dimension, newValue string) error {
	return wm.statusManager.Transition(uid, dimension, newValue)
}

// GetCollection returns the underlying collection for result building.
func (wm *WorkflowManager) GetCollection() *models.Collection {
	return wm.adapter.Collection()
}

// SaveCollection saves the current state of the collection.
func (wm *WorkflowManager) SaveCollection() error {
	return wm.adapter.Save()
}

// WorkflowResult contains the result of a workflow operation.
type WorkflowResult struct {
	Todo       *models.Todo
	OldStatus  string
	NewStatus  string
	AllTodos   []*models.Todo
	TotalCount int
	DoneCount  int
}

// BuildResult creates a result object for the given todo and mode.
func (wm *WorkflowManager) BuildResult(uid, mode, oldStatus string) (*WorkflowResult, error) {
	collection := wm.GetCollection()
	todo := collection.FindItemByID(uid)
	if todo == nil {
		return nil, fmt.Errorf("todo with ID '%s' not found", uid)
	}

	result := &WorkflowResult{
		Todo:      todo,
		OldStatus: oldStatus,
	}

	// Get current status for the result
	if status, err := wm.GetStatus(uid, "completion"); err == nil {
		result.NewStatus = status
	}
	
	// For mode-specific data
	if mode == "long" {
		result.AllTodos = collection.ListActive()
		// Ensure AllTodos is never nil
		if result.AllTodos == nil {
			result.AllTodos = []*models.Todo{}
		}
		result.TotalCount, result.DoneCount = countTodos(collection.Todos)
	}

	return result, nil
}

// countTodos counts total and done todos (helper function from complete command)
func countTodos(todos []*models.Todo) (total, done int) {
	for _, todo := range todos {
		total++
		if todo.Status == models.StatusDone {
			done++
		}
		subTotal, subDone := countTodos(todo.Items)
		total += subTotal
		done += subDone
	}
	return total, done
}