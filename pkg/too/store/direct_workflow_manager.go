package store

import (
	"fmt"

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

	// Keep order as-is from the collection structure
	// IDM will assign HIDs based on this order

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

// ListActive returns only active (pending) todos using IDM ordering.
// This replaces collection.ListActive() with IDM-aware filtering.
func (m *DirectWorkflowManager) ListActive() []*models.Todo {
	// Get all UIDs from root scope (IDM maintains proper ordering)
	uids := m.registry.GetUIDs(RootScope)
	
	var activeTodos []*models.Todo
	for _, uid := range uids {
		todo := m.collection.FindItemByID(uid)
		if todo != nil && todo.GetStatus() == models.StatusPending {
			// Clone the todo and recursively add active children
			clonedTodo := todo.Clone()
			clonedTodo.Items = m.getActiveChildren(todo)
			activeTodos = append(activeTodos, clonedTodo)
		}
	}
	
	return activeTodos
}

// ListArchived returns only archived (done) todos using collection traversal
// since done todos are not registered in IDM (only active todos get HIDs).
// This replaces collection.ListArchived() with proper behavioral propagation.
func (m *DirectWorkflowManager) ListArchived() []*models.Todo {
	return m.filterTodos(m.collection.Todos, func(t *models.Todo) bool {
		return t.GetStatus() == models.StatusDone
	}, false) // Don't recurse into done items (behavioral propagation)
}

// ListAll returns all todos regardless of status using collection traversal.
// This preserves the complete tree structure including any inconsistent states.
// This replaces collection.ListAll() with the same behavior.
func (m *DirectWorkflowManager) ListAll() []*models.Todo {
	return m.cloneTodos(m.collection.Todos)
}

// cloneTodos creates a deep copy of a slice of todos
func (m *DirectWorkflowManager) cloneTodos(todos []*models.Todo) []*models.Todo {
	cloned := make([]*models.Todo, len(todos))
	for i, todo := range todos {
		cloned[i] = todo.Clone()
	}
	return cloned
}

// GetTodoByID finds a todo by its ID without exposing the collection.
// This replaces direct collection.FindItemByID() calls.
func (m *DirectWorkflowManager) GetTodoByID(uid string) *models.Todo {
	return m.collection.FindItemByID(uid)
}

// GetTodoByShortID finds a todo by its short ID without exposing the collection.
// This replaces direct collection.FindItemByShortID() calls.
func (m *DirectWorkflowManager) GetTodoByShortID(shortID string) (*models.Todo, error) {
	return m.collection.FindItemByShortID(shortID)
}

// CountTodos returns the total count and done count of all todos without exposing the collection.
// This replaces direct access to collection.Todos for counting.
func (m *DirectWorkflowManager) CountTodos() (totalCount, doneCount int) {
	return m.countAllTodos(m.collection.Todos)
}

// countAllTodos recursively counts total and done todos
func (m *DirectWorkflowManager) countAllTodos(todos []*models.Todo) (total int, done int) {
	for _, todo := range todos {
		total++
		if todo.GetStatus() == models.StatusDone {
			done++
		}
		// Recursively count children
		childTotal, childDone := m.countAllTodos(todo.Items)
		total += childTotal
		done += childDone
	}
	return total, done
}

// getActiveChildren recursively returns only active children of a todo.
func (m *DirectWorkflowManager) getActiveChildren(parent *models.Todo) []*models.Todo {
	var activeChildren []*models.Todo
	
	for _, child := range parent.Items {
		if child.GetStatus() == models.StatusPending {
			clonedChild := child.Clone()
			clonedChild.Items = m.getActiveChildren(child)
			activeChildren = append(activeChildren, clonedChild)
		}
	}
	
	return activeChildren
}

// CleanFinishedTodos removes all done todos and their descendants from the collection and IDM.
// Returns the removed todos for reporting purposes.
func (m *DirectWorkflowManager) CleanFinishedTodos() ([]*models.Todo, int, error) {
	// Find all done items before removing them
	removedTodos := m.findDoneItems(m.collection.Todos)
	
	// Remove done todos from the collection
	m.collection.Todos = m.removeFinishedTodosRecursive(m.collection.Todos)
	
	// Remove done todos from IDM registry
	for _, removedTodo := range removedTodos {
		// Remove from parent scope (either root or parent todo)
		parentScope := RootScope
		if removedTodo.ParentID != "" {
			parentScope = removedTodo.ParentID
		}
		m.registry.Remove(parentScope, removedTodo.ID)
		
		// If the removed todo had children, remove its scope entirely
		if len(removedTodo.Items) > 0 {
			m.registry.RemoveScope(removedTodo.ID)
		}
	}
	
	// Count remaining active todos
	activeCount := m.countActiveTodos(m.collection.Todos)
	
	return removedTodos, activeCount, nil
}

// findDoneItems finds all done todos (not including their pending descendants)
func (m *DirectWorkflowManager) findDoneItems(todos []*models.Todo) []*models.Todo {
	var doneItems []*models.Todo
	for _, todo := range todos {
		if todo.GetStatus() == models.StatusDone {
			doneItems = append(doneItems, todo.Clone())
		}
		// Always recurse, as a pending parent can have done children
		doneItems = append(doneItems, m.findDoneItems(todo.Items)...)
	}
	return doneItems
}

// removeFinishedTodosRecursive removes done todos and their descendants
func (m *DirectWorkflowManager) removeFinishedTodosRecursive(todos []*models.Todo) []*models.Todo {
	var activeTodos []*models.Todo

	for _, todo := range todos {
		if todo.GetStatus() != models.StatusDone {
			// Keep this todo but recursively clean its children
			todoCopy := *todo
			todoCopy.Items = m.removeFinishedTodosRecursive(todo.Items)
			activeTodos = append(activeTodos, &todoCopy)
		}
		// If done, skip this todo and all its descendants
	}

	return activeTodos
}

// countActiveTodos recursively counts all active (non-done) todos
func (m *DirectWorkflowManager) countActiveTodos(todos []*models.Todo) int {
	count := 0
	for _, todo := range todos {
		if todo.GetStatus() != models.StatusDone {
			count++
			count += m.countActiveTodos(todo.Items)
		}
	}
	return count
}

// filterTodos recursively filters todos based on a predicate function.
// If recurseIntoDone is false, it stops recursion at done items (behavioral propagation).
// This is similar to the collection method but operates within DirectWorkflowManager.
func (m *DirectWorkflowManager) filterTodos(todos []*models.Todo, predicate func(*models.Todo) bool, recurseIntoDone bool) []*models.Todo {
	var filtered []*models.Todo

	for _, todo := range todos {
		if predicate(todo) {
			// Clone the todo to avoid modifying the original
			filteredTodo := &models.Todo{
				ID:       todo.ID,
				ParentID: todo.ParentID,
				Text:     todo.Text,
				Statuses: make(map[string]string),
				Modified: todo.Modified,
				Items:    []*models.Todo{},
			}

			// Clone statuses map
			for k, v := range todo.Statuses {
				filteredTodo.Statuses[k] = v
			}

			// If this todo is done and we're not recursing into done items,
			// stop here (behavioral propagation)
			if todo.GetStatus() == models.StatusDone && !recurseIntoDone {
				filtered = append(filtered, filteredTodo)
			} else {
				// Recursively filter children
				filteredTodo.Items = m.filterTodos(todo.Items, predicate, recurseIntoDone)
				filtered = append(filtered, filteredTodo)
			}
		}
	}

	return filtered
}