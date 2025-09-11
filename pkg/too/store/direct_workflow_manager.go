package store

import (
	"fmt"
	"sort"

	"github.com/arthur-debert/too/pkg/idm"
	"github.com/arthur-debert/too/pkg/idm/workflow"
	"github.com/arthur-debert/too/pkg/too/config"
	"github.com/arthur-debert/too/pkg/too/models"
)

// RootScope is a special constant used to identify the root of the todo tree
// in the IDM registry.
const RootScope = "root"

// DirectWorkflowManager implements workflow operations directly on collections
// without intermediate adapter layers.
type DirectWorkflowManager struct {
	collection               *models.Collection
	store                    Store
	registry                 *idm.Registry
	statusManager            *workflow.StatusManager
	autoTransitionInProgress map[string]bool
}

// NewDirectWorkflowManager creates a workflow manager that works directly with collections.
func NewDirectWorkflowManager(store Store, collectionPath string) (*DirectWorkflowManager, error) {
	collection, err := store.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load collection: %w", err)
	}

	// Create a minimal adapter inline for IDM registry
	adapter := &directAdapter{collection: collection}
	
	// Create IDM registry
	registry := idm.NewRegistry()

	// Build the registry with all scopes
	if err := registry.RebuildScope(adapter, RootScope); err != nil {
		return nil, fmt.Errorf("failed to rebuild root scope: %w", err)
	}
	
	// Rebuild all child scopes
	collection.Walk(func(todo *models.Todo) {
		if len(todo.Items) > 0 {
			// This todo is a scope (has children)
			if err := registry.RebuildScope(adapter, todo.ID); err != nil {
				// Log error but continue
				return
			}
		}
	})

	// Load workflow configuration
	workflowConfigPath := config.GetWorkflowConfigPath(collectionPath)
	workflowConfig, err := config.LoadWorkflowConfig(workflowConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load workflow config: %w", err)
	}

	effectiveConfig, err := workflowConfig.GetWorkflowConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get effective workflow config: %w", err)
	}

	// Create workflow adapter inline
	workflowAdapter := &directWorkflowAdapter{collection: collection}

	// Create IDM manager with the adapter
	manager, err := idm.NewManager(adapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create IDM manager: %w", err)
	}

	// Create status manager
	statusManager, err := workflow.NewStatusManager(
		registry,
		manager,
		workflowAdapter,
		effectiveConfig,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create status manager: %w", err)
	}

	return &DirectWorkflowManager{
		collection:               collection,
		store:                    store,
		registry:                 registry,
		statusManager:            statusManager,
		autoTransitionInProgress: make(map[string]bool),
	}, nil
}

// Add creates a new todo item under the given parent.
func (m *DirectWorkflowManager) Add(parentUID, text string) (string, error) {
	// Create the todo directly in the collection
	var parentID string
	if parentUID != RootScope {
		parentID = parentUID
	}

	todo, err := m.collection.CreateTodo(text, parentID)
	if err != nil {
		return "", err
	}

	// Register with IDM
	if err := m.registry.RebuildScope(&directAdapter{collection: m.collection}, parentUID); err != nil {
		return "", fmt.Errorf("failed to rebuild scope after add: %w", err)
	}

	return todo.ID, nil
}

// SetStatus sets a status dimension for an item with recursion guard.
func (m *DirectWorkflowManager) SetStatus(uid, dimension, value string) error {
	if m.autoTransitionInProgress[uid] {
		return nil
	}

	m.autoTransitionInProgress[uid] = true
	defer delete(m.autoTransitionInProgress, uid)

	// Set status directly on todo
	todo := m.collection.FindItemByID(uid)
	if todo == nil {
		return fmt.Errorf("todo with ID %s not found", uid)
	}

	todo.EnsureStatuses()
	todo.Statuses[dimension] = value
	todo.SetModified()

	// Trigger workflow transitions
	return m.statusManager.SetStatus(uid, dimension, value)
}

// GetStatus gets a status dimension value.
func (m *DirectWorkflowManager) GetStatus(uid, dimension string) (string, error) {
	todo := m.collection.FindItemByID(uid)
	if todo == nil {
		return "", fmt.Errorf("todo with ID %s not found", uid)
	}

	value, _ := todo.GetWorkflowStatus(dimension)
	return value, nil
}

// ResolvePositionPath resolves a position path to a UID.
func (m *DirectWorkflowManager) ResolvePositionPath(scope, path string) (string, error) {
	return m.registry.ResolvePositionPath(scope, path)
}

// Save persists the current collection state.
func (m *DirectWorkflowManager) Save() error {
	return m.store.Save(m.collection)
}

// GetCollection returns the underlying collection.
func (m *DirectWorkflowManager) GetCollection() *models.Collection {
	return m.collection
}

// GetRegistry returns the IDM registry.
func (m *DirectWorkflowManager) GetRegistry() *idm.Registry {
	return m.registry
}

// Move moves a todo from one parent to another.
func (m *DirectWorkflowManager) Move(uid, oldParentUID, newParentUID string) error {
	// Find the todo
	todo := m.collection.FindItemByID(uid)
	if todo == nil {
		return fmt.Errorf("todo with ID %s not found", uid)
	}

	// Remove from old parent
	if oldParentUID == RootScope {
		// Remove from root collection
		for i, t := range m.collection.Todos {
			if t.ID == uid {
				m.collection.Todos = append(m.collection.Todos[:i], m.collection.Todos[i+1:]...)
				break
			}
		}
	} else {
		oldParent := m.collection.FindItemByID(oldParentUID)
		if oldParent != nil {
			for i, t := range oldParent.Items {
				if t.ID == uid {
					oldParent.Items = append(oldParent.Items[:i], oldParent.Items[i+1:]...)
					break
				}
			}
		}
	}

	// Update parent ID
	if newParentUID == RootScope {
		todo.ParentID = ""
	} else {
		todo.ParentID = newParentUID
	}

	// Add to new parent
	if newParentUID == RootScope {
		m.collection.Todos = append(m.collection.Todos, todo)
		models.ReorderTodos(m.collection.Todos)
	} else {
		newParent := m.collection.FindItemByID(newParentUID)
		if newParent == nil {
			return fmt.Errorf("new parent with ID %s not found", newParentUID)
		}
		newParent.Items = append(newParent.Items, todo)
		models.ReorderTodos(newParent.Items)
	}

	// Rebuild registries for affected scopes
	adapter := &directAdapter{collection: m.collection}
	if err := m.registry.RebuildScope(adapter, oldParentUID); err != nil {
		return err
	}
	if err := m.registry.RebuildScope(adapter, newParentUID); err != nil {
		return err
	}

	todo.SetModified()
	return nil
}

// GetPositionPath returns the position path for a given UID.
func (m *DirectWorkflowManager) GetPositionPath(scope, uid string) (string, error) {
	adapter := &directAdapter{collection: m.collection}
	return m.registry.GetPositionPath(scope, uid, adapter)
}

// directAdapter is a minimal inline adapter for IDM operations.
type directAdapter struct {
	collection *models.Collection
}

func (a *directAdapter) GetChildren(parentUID string) ([]string, error) {
	var todos []*models.Todo
	if parentUID == RootScope {
		todos = a.collection.Todos
	} else {
		parent := a.collection.FindItemByID(parentUID)
		if parent == nil {
			return []string{}, nil
		}
		todos = parent.Items
	}

	// Only return active items for HID assignment
	var children []string
	for _, todo := range todos {
		if todo.GetStatus() == models.StatusPending {
			children = append(children, todo.ID)
		}
	}

	// Sort by position
	sort.Slice(children, func(i, j int) bool {
		todoI := a.collection.FindItemByID(children[i])
		todoJ := a.collection.FindItemByID(children[j])
		if todoI == nil || todoJ == nil {
			return false
		}
		return todoI.Position < todoJ.Position
	})

	return children, nil
}

func (a *directAdapter) GetParent(uid string) (string, error) {
	todo := a.collection.FindItemByID(uid)
	if todo == nil {
		return "", fmt.Errorf("todo with UID %s not found", uid)
	}
	if todo.ParentID == "" {
		return RootScope, nil
	}
	return todo.ParentID, nil
}

func (a *directAdapter) GetScopes() ([]string, error) {
	var scopes []string
	scopes = append(scopes, RootScope)
	a.collection.Walk(func(todo *models.Todo) {
		scopes = append(scopes, todo.ID)
	})
	return scopes, nil
}

func (a *directAdapter) GetAllUIDs() ([]string, error) {
	var uids []string
	a.collection.Walk(func(todo *models.Todo) {
		uids = append(uids, todo.ID)
	})
	return uids, nil
}

// ManagedStoreAdapter methods
func (a *directAdapter) AddItem(parentUID string) (string, error) {
	var parentID string
	if parentUID != RootScope {
		parentID = parentUID
	}
	todo, err := a.collection.CreateTodo("", parentID)
	if err != nil {
		return "", err
	}
	return todo.ID, nil
}

func (a *directAdapter) RemoveItem(uid string) error {
	return a.collection.RemoveTodo(uid)
}

func (a *directAdapter) MoveItem(uid, newParentUID string) error {
	todo := a.collection.FindItemByID(uid)
	if todo == nil {
		return fmt.Errorf("todo with UID %s not found", uid)
	}
	// TODO: Implement move logic
	return fmt.Errorf("move not implemented")
}

func (a *directAdapter) SetStatus(uid, status string) error {
	todo := a.collection.FindItemByID(uid)
	if todo == nil {
		return fmt.Errorf("todo with UID %s not found", uid)
	}
	todo.EnsureStatuses()
	todo.Statuses["completion"] = status
	return nil
}

func (a *directAdapter) SetPinned(uid string, isPinned bool) error {
	// Not implemented for todos
	return nil
}

// directWorkflowAdapter implements workflow operations directly.
type directWorkflowAdapter struct {
	collection *models.Collection
}

func (a *directWorkflowAdapter) GetChildren(parentUID string) ([]string, error) {
	var todos []*models.Todo
	if parentUID == RootScope {
		todos = a.collection.Todos
	} else {
		parent := a.collection.FindItemByID(parentUID)
		if parent == nil {
			return []string{}, nil
		}
		todos = parent.Items
	}

	var children []string
	for _, todo := range todos {
		children = append(children, todo.ID)
	}
	return children, nil
}

func (a *directWorkflowAdapter) GetParent(uid string) (string, error) {
	todo := a.collection.FindItemByID(uid)
	if todo == nil {
		return "", fmt.Errorf("todo with UID %s not found", uid)
	}
	if todo.ParentID == "" {
		return RootScope, nil
	}
	return todo.ParentID, nil
}

func (a *directWorkflowAdapter) SetItemStatus(uid, dimension, value string) error {
	todo := a.collection.FindItemByID(uid)
	if todo == nil {
		return fmt.Errorf("todo with ID %s not found", uid)
	}
	
	todo.EnsureStatuses()
	todo.Statuses[dimension] = value
	todo.SetModified()
	return nil
}

func (a *directWorkflowAdapter) GetItemStatus(uid, dimension string) (string, error) {
	if uid == RootScope {
		return "", nil
	}
	
	todo := a.collection.FindItemByID(uid)
	if todo == nil {
		return "", fmt.Errorf("todo with ID %s not found", uid)
	}
	
	todo.EnsureStatuses()
	if value, exists := todo.Statuses[dimension]; exists {
		return value, nil
	}
	
	return "", fmt.Errorf("dimension %s not found for todo %s", dimension, uid)
}

func (a *directWorkflowAdapter) GetItemStatuses(uid string) (map[string]string, error) {
	todo := a.collection.FindItemByID(uid)
	if todo == nil {
		return nil, fmt.Errorf("todo with ID %s not found", uid)
	}
	
	todo.EnsureStatuses()
	result := make(map[string]string)
	for k, v := range todo.Statuses {
		result[k] = v
	}
	
	if _, exists := result["completion"]; !exists {
		result["completion"] = string(models.StatusPending)
	}
	
	return result, nil
}

func (a *directWorkflowAdapter) SetMultipleStatuses(uid string, statuses map[string]string) error {
	todo := a.collection.FindItemByID(uid)
	if todo == nil {
		return fmt.Errorf("todo with ID %s not found", uid)
	}
	
	todo.EnsureStatuses()
	for dimension, value := range statuses {
		todo.Statuses[dimension] = value
	}
	
	return nil
}

func (a *directWorkflowAdapter) GetChildrenInContext(parentUID, context string, rules []workflow.VisibilityRule) ([]string, error) {
	allChildren, err := a.GetChildren(parentUID)
	if err != nil {
		return nil, err
	}
	
	var visibleChildren []string
	for _, childUID := range allChildren {
		statuses, err := a.GetItemStatuses(childUID)
		if err != nil {
			continue
		}
		
		visible := true
		for _, rule := range rules {
			if !rule.Matches(context, statuses) {
				visible = false
				break
			}
		}
		
		if visible {
			visibleChildren = append(visibleChildren, childUID)
		}
	}
	
	return visibleChildren, nil
}

func (a *directWorkflowAdapter) GetAllItemsInContext(context string, rules []workflow.VisibilityRule) ([]string, error) {
	var visibleItems []string
	
	// Walk the entire collection
	a.collection.Walk(func(todo *models.Todo) {
		todo.EnsureStatuses()
		statuses := make(map[string]string)
		for k, v := range todo.Statuses {
			statuses[k] = v
		}
		
		visible := true
		for _, rule := range rules {
			if !rule.Matches(context, statuses) {
				visible = false
				break
			}
		}
		
		if visible {
			visibleItems = append(visibleItems, todo.ID)
		}
	})
	
	return visibleItems, nil
}

func (a *directWorkflowAdapter) GetStatusesBulk(uids []string) (map[string]map[string]string, error) {
	result := make(map[string]map[string]string)
	
	for _, uid := range uids {
		if statuses, err := a.GetItemStatuses(uid); err == nil {
			result[uid] = statuses
		}
	}
	
	return result, nil
}

func (a *directWorkflowAdapter) GetAllUIDs() ([]string, error) {
	var uids []string
	a.collection.Walk(func(todo *models.Todo) {
		uids = append(uids, todo.ID)
	})
	return uids, nil
}

// Implement ManagedStoreAdapter methods
func (a *directWorkflowAdapter) GetScopes() ([]string, error) {
	var scopes []string
	scopes = append(scopes, RootScope)
	a.collection.Walk(func(todo *models.Todo) {
		scopes = append(scopes, todo.ID)
	})
	return scopes, nil
}

func (a *directWorkflowAdapter) AddItem(parentUID string) (string, error) {
	var parentID string
	if parentUID != RootScope {
		parentID = parentUID
	}
	todo, err := a.collection.CreateTodo("", parentID)
	if err != nil {
		return "", err
	}
	return todo.ID, nil
}

func (a *directWorkflowAdapter) RemoveItem(uid string) error {
	return a.collection.RemoveTodo(uid)
}

func (a *directWorkflowAdapter) MoveItem(uid, newParentUID string) error {
	todo := a.collection.FindItemByID(uid)
	if todo == nil {
		return fmt.Errorf("todo with UID %s not found", uid)
	}
	// TODO: Implement move logic
	return fmt.Errorf("move not implemented")
}

func (a *directWorkflowAdapter) SetStatus(uid, status string) error {
	todo := a.collection.FindItemByID(uid)
	if todo == nil {
		return fmt.Errorf("todo with UID %s not found", uid)
	}
	todo.EnsureStatuses()
	todo.Statuses["completion"] = status
	return nil
}

func (a *directWorkflowAdapter) SetPinned(uid string, isPinned bool) error {
	// Not implemented for todos
	return nil
}

// Implement WorkflowStoreAdapter lifecycle hooks

func (a *directWorkflowAdapter) OnStatusChange(uid, dimension, oldValue, newValue string) error {
	// No-op for direct workflow manager - side effects are handled at a higher level
	return nil
}

func (a *directWorkflowAdapter) ValidateStatusChange(uid, dimension, oldValue, newValue string) error {
	// No validation needed for direct workflow manager
	return nil
}

func (a *directWorkflowAdapter) SetStatusesBulk(updates map[string]map[string]string) error {
	for uid, statuses := range updates {
		if err := a.SetMultipleStatuses(uid, statuses); err != nil {
			return err
		}
	}
	return nil
}