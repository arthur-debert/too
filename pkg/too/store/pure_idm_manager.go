package store

import (
	"fmt"

	"github.com/arthur-debert/too/pkg/idm"
	"github.com/arthur-debert/too/pkg/idm/workflow"
	"github.com/arthur-debert/too/pkg/too/config"
	"github.com/arthur-debert/too/pkg/too/models"
)

// PureIDMManager implements workflow operations using only IDM and flat data structures.
// This represents the final evolution from hierarchical Collections to pure IDM management.
type PureIDMManager struct {
	collection               *models.IDMCollection
	store                    IDMStore
	registry                 *idm.Registry
	statusManager            *workflow.StatusManager
	autoTransitionInProgress map[string]bool
}

// NewPureIDMManager creates a workflow manager that operates entirely on flat IDM data structures.
func NewPureIDMManager(store IDMStore, collectionPath string) (*PureIDMManager, error) {
	// Load the flat IDM collection
	collection, err := store.LoadIDM()
	if err != nil {
		return nil, fmt.Errorf("failed to load IDM collection: %w", err)
	}

	// Create IDM adapter for the flat structure
	adapter := &pureIDMAdapter{collection: collection}
	
	// Create IDM registry
	registry := idm.NewRegistry()

	// Build the registry with all scopes
	if err := registry.RebuildScope(adapter, RootScope); err != nil {
		return nil, fmt.Errorf("failed to rebuild root scope: %w", err)
	}
	
	// Rebuild all child scopes (items that have children)
	for _, item := range collection.Items {
		if len(collection.GetChildren(item.UID)) > 0 {
			// This item is a scope (has children)
			if err := registry.RebuildScope(adapter, item.UID); err != nil {
				// Log error but continue
				continue
			}
		}
	}

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

	// Create workflow adapter
	workflowAdapter := &pureIDMWorkflowAdapter{collection: collection}

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

	return &PureIDMManager{
		collection:               collection,
		store:                    store,
		registry:                 registry,
		statusManager:            statusManager,
		autoTransitionInProgress: make(map[string]bool),
	}, nil
}

// Add creates a new todo item under the given parent.
func (m *PureIDMManager) Add(parentUID, text string) (string, error) {
	// Create the todo directly in the flat collection
	var parentID string
	if parentUID != RootScope {
		parentID = parentUID
	}

	todo := models.NewIDMTodo(text, parentID)
	m.collection.AddItem(todo)

	// Register with IDM
	if err := m.registry.RebuildScope(&pureIDMAdapter{collection: m.collection}, parentUID); err != nil {
		return "", fmt.Errorf("failed to rebuild scope after add: %w", err)
	}

	return todo.UID, nil
}

// SetStatus sets a status dimension for an item with recursion guard.
func (m *PureIDMManager) SetStatus(uid, dimension, value string) error {
	if m.autoTransitionInProgress[uid] {
		return nil
	}

	m.autoTransitionInProgress[uid] = true
	defer delete(m.autoTransitionInProgress, uid)

	// Set status directly on todo
	todo := m.collection.FindByUID(uid)
	if todo == nil {
		return fmt.Errorf("todo with UID %s not found", uid)
	}

	todo.EnsureStatuses()
	todo.Statuses[dimension] = value
	todo.SetModified()

	// Trigger workflow transitions
	return m.statusManager.SetStatus(uid, dimension, value)
}

// GetStatus gets a status dimension value.
func (m *PureIDMManager) GetStatus(uid, dimension string) (string, error) {
	todo := m.collection.FindByUID(uid)
	if todo == nil {
		return "", fmt.Errorf("todo with UID %s not found", uid)
	}

	value, _ := todo.GetWorkflowStatus(dimension)
	return value, nil
}

// ResolvePositionPath resolves a position path to a UID.
func (m *PureIDMManager) ResolvePositionPath(scope, path string) (string, error) {
	return m.registry.ResolvePositionPath(scope, path)
}

// Save persists the current collection state.
func (m *PureIDMManager) Save() error {
	return m.store.SaveIDM(m.collection)
}

// GetCollection returns the underlying IDM collection.
func (m *PureIDMManager) GetCollection() *models.IDMCollection {
	return m.collection
}

// GetRegistry returns the IDM registry.
func (m *PureIDMManager) GetRegistry() *idm.Registry {
	return m.registry
}

// Move moves a todo from one parent to another.
func (m *PureIDMManager) Move(uid, oldParentUID, newParentUID string) error {
	// Find the todo
	todo := m.collection.FindByUID(uid)
	if todo == nil {
		return fmt.Errorf("todo with UID %s not found", uid)
	}

	// Update parent ID in the flat structure
	if newParentUID == RootScope {
		todo.ParentID = ""
	} else {
		todo.ParentID = newParentUID
	}

	// Rebuild registries for affected scopes
	adapter := &pureIDMAdapter{collection: m.collection}
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
func (m *PureIDMManager) GetPositionPath(scope, uid string) (string, error) {
	adapter := &pureIDMAdapter{collection: m.collection}
	return m.registry.GetPositionPath(scope, uid, adapter)
}

// ListActive returns only active (pending) todos using IDM ordering.
func (m *PureIDMManager) ListActive() interface{} {
	// Get all UIDs from root scope (IDM maintains proper ordering)
	uids := m.registry.GetUIDs(RootScope)
	
	var activeTodos []*models.IDMTodo
	for _, uid := range uids {
		todo := m.collection.FindByUID(uid)
		if todo != nil && todo.GetStatus() == models.StatusPending {
			activeTodos = append(activeTodos, todo.Clone())
		}
	}
	
	return activeTodos
}

// ListArchived returns only archived (done) todos.
func (m *PureIDMManager) ListArchived() interface{} {
	var archivedTodos []*models.IDMTodo
	for _, todo := range m.collection.Items {
		if todo.GetStatus() == models.StatusDone {
			archivedTodos = append(archivedTodos, todo.Clone())
		}
	}
	return archivedTodos
}

// ListAll returns all todos regardless of status.
func (m *PureIDMManager) ListAll() interface{} {
	var allTodos []*models.IDMTodo
	for _, todo := range m.collection.Items {
		allTodos = append(allTodos, todo.Clone())
	}
	return allTodos
}

// GetTodoByUID finds a todo by its UID.
func (m *PureIDMManager) GetTodoByUID(uid string) *models.IDMTodo {
	return m.collection.FindByUID(uid)
}

// GetTodoByID finds a todo by its UID (alias for GetTodoByUID for interface compatibility).
func (m *PureIDMManager) GetTodoByID(uid string) interface{} {
	return m.collection.FindByUID(uid)
}

// GetTodoByShortID finds a todo by its short ID.
func (m *PureIDMManager) GetTodoByShortID(shortID string) (interface{}, error) {
	return m.store.FindItemByShortID(shortID)
}

// CountTodos returns the total count and done count of all todos.
func (m *PureIDMManager) CountTodos() (totalCount, doneCount int) {
	for _, todo := range m.collection.Items {
		totalCount++
		if todo.GetStatus() == models.StatusDone {
			doneCount++
		}
	}
	return totalCount, doneCount
}

// CleanFinishedTodos removes all done todos and their descendants from the collection and IDM.
func (m *PureIDMManager) CleanFinishedTodos() (interface{}, int, error) {
	var removedTodos []*models.IDMTodo
	var remainingItems []*models.IDMTodo
	
	// Find all done items and their descendants
	doneUIDs := make(map[string]bool)
	for _, todo := range m.collection.Items {
		if todo.GetStatus() == models.StatusDone {
			doneUIDs[todo.UID] = true
			// Mark all descendants as done too
			descendants := m.collection.GetDescendants(todo.UID)
			for _, desc := range descendants {
				doneUIDs[desc.UID] = true
			}
		}
	}
	
	// Separate removed vs remaining items
	for _, todo := range m.collection.Items {
		if doneUIDs[todo.UID] {
			removedTodos = append(removedTodos, todo.Clone())
		} else {
			remainingItems = append(remainingItems, todo)
		}
	}
	
	// Update collection to only contain remaining items
	m.collection.Items = remainingItems
	
	// Remove done todos from IDM registry
	for _, removedTodo := range removedTodos {
		// Remove from parent scope
		parentScope := RootScope
		if removedTodo.ParentID != "" {
			parentScope = removedTodo.ParentID
		}
		m.registry.Remove(parentScope, removedTodo.UID)
		
		// If the removed todo had children, remove its scope entirely
		if len(m.collection.GetChildren(removedTodo.UID)) > 0 {
			m.registry.RemoveScope(removedTodo.UID)
		}
	}
	
	// Count remaining active todos
	activeCount := 0
	for _, todo := range remainingItems {
		if todo.GetStatus() != models.StatusDone {
			activeCount++
		}
	}
	
	return removedTodos, activeCount, nil
}

// pureIDMAdapter implements IDM adapter interface for flat IDM collections.
type pureIDMAdapter struct {
	collection *models.IDMCollection
}

func (a *pureIDMAdapter) GetChildren(parentUID string) ([]string, error) {
	// Map RootScope to empty string for the data model
	var parentID string
	if parentUID == RootScope {
		parentID = ""
	} else {
		parentID = parentUID
	}
	
	children := a.collection.GetChildren(parentID)
	
	// Only return active items for HID assignment
	var activeChildren []string
	for _, child := range children {
		if child.GetStatus() == models.StatusPending {
			activeChildren = append(activeChildren, child.UID)
		}
	}
	
	return activeChildren, nil
}

func (a *pureIDMAdapter) GetParent(uid string) (string, error) {
	todo := a.collection.FindByUID(uid)
	if todo == nil {
		return "", fmt.Errorf("todo with UID %s not found", uid)
	}
	if todo.ParentID == "" {
		return RootScope, nil
	}
	return todo.ParentID, nil
}

func (a *pureIDMAdapter) GetScopes() ([]string, error) {
	var scopes []string
	scopes = append(scopes, RootScope)
	
	// Add all items as potential scopes
	for _, item := range a.collection.Items {
		scopes = append(scopes, item.UID)
	}
	
	return scopes, nil
}

func (a *pureIDMAdapter) GetAllUIDs() ([]string, error) {
	var uids []string
	for _, item := range a.collection.Items {
		uids = append(uids, item.UID)
	}
	return uids, nil
}

// ManagedStoreAdapter methods
func (a *pureIDMAdapter) AddItem(parentUID string) (string, error) {
	var parentID string
	if parentUID != RootScope {
		parentID = parentUID
	}
	todo := models.NewIDMTodo("", parentID)
	a.collection.AddItem(todo)
	return todo.UID, nil
}

func (a *pureIDMAdapter) RemoveItem(uid string) error {
	if a.collection.RemoveItem(uid) {
		return nil
	}
	return fmt.Errorf("todo with UID %s not found", uid)
}

func (a *pureIDMAdapter) MoveItem(uid, newParentUID string) error {
	todo := a.collection.FindByUID(uid)
	if todo == nil {
		return fmt.Errorf("todo with UID %s not found", uid)
	}
	
	if newParentUID == RootScope {
		todo.ParentID = ""
	} else {
		todo.ParentID = newParentUID
	}
	
	return nil
}

func (a *pureIDMAdapter) SetStatus(uid, status string) error {
	todo := a.collection.FindByUID(uid)
	if todo == nil {
		return fmt.Errorf("todo with UID %s not found", uid)
	}
	todo.EnsureStatuses()
	todo.Statuses["completion"] = status
	return nil
}

func (a *pureIDMAdapter) SetPinned(uid string, isPinned bool) error {
	// Not implemented for todos
	return nil
}

// pureIDMWorkflowAdapter implements workflow operations for flat IDM collections.
type pureIDMWorkflowAdapter struct {
	collection *models.IDMCollection
}

func (a *pureIDMWorkflowAdapter) GetChildren(parentUID string) ([]string, error) {
	// Map RootScope to empty string for the data model
	var parentID string
	if parentUID == RootScope {
		parentID = ""
	} else {
		parentID = parentUID
	}
	
	children := a.collection.GetChildren(parentID)
	var childUIDs []string
	for _, child := range children {
		childUIDs = append(childUIDs, child.UID)
	}
	return childUIDs, nil
}

func (a *pureIDMWorkflowAdapter) GetParent(uid string) (string, error) {
	todo := a.collection.FindByUID(uid)
	if todo == nil {
		return "", fmt.Errorf("todo with UID %s not found", uid)
	}
	if todo.ParentID == "" {
		return RootScope, nil
	}
	return todo.ParentID, nil
}

func (a *pureIDMWorkflowAdapter) SetItemStatus(uid, dimension, value string) error {
	todo := a.collection.FindByUID(uid)
	if todo == nil {
		return fmt.Errorf("todo with UID %s not found", uid)
	}
	
	todo.EnsureStatuses()
	todo.Statuses[dimension] = value
	todo.SetModified()
	return nil
}

func (a *pureIDMWorkflowAdapter) GetItemStatus(uid, dimension string) (string, error) {
	if uid == RootScope {
		return "", nil
	}
	
	todo := a.collection.FindByUID(uid)
	if todo == nil {
		return "", fmt.Errorf("todo with UID %s not found", uid)
	}
	
	todo.EnsureStatuses()
	if value, exists := todo.Statuses[dimension]; exists {
		return value, nil
	}
	
	return "", fmt.Errorf("dimension %s not found for todo %s", dimension, uid)
}

func (a *pureIDMWorkflowAdapter) GetItemStatuses(uid string) (map[string]string, error) {
	todo := a.collection.FindByUID(uid)
	if todo == nil {
		return nil, fmt.Errorf("todo with UID %s not found", uid)
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

func (a *pureIDMWorkflowAdapter) SetMultipleStatuses(uid string, statuses map[string]string) error {
	todo := a.collection.FindByUID(uid)
	if todo == nil {
		return fmt.Errorf("todo with UID %s not found", uid)
	}
	
	todo.EnsureStatuses()
	for dimension, value := range statuses {
		todo.Statuses[dimension] = value
	}
	
	return nil
}

func (a *pureIDMWorkflowAdapter) GetChildrenInContext(parentUID, context string, rules []workflow.VisibilityRule) ([]string, error) {
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

func (a *pureIDMWorkflowAdapter) GetAllItemsInContext(context string, rules []workflow.VisibilityRule) ([]string, error) {
	var visibleItems []string
	
	for _, item := range a.collection.Items {
		item.EnsureStatuses()
		statuses := make(map[string]string)
		for k, v := range item.Statuses {
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
			visibleItems = append(visibleItems, item.UID)
		}
	}
	
	return visibleItems, nil
}

func (a *pureIDMWorkflowAdapter) GetStatusesBulk(uids []string) (map[string]map[string]string, error) {
	result := make(map[string]map[string]string)
	
	for _, uid := range uids {
		if statuses, err := a.GetItemStatuses(uid); err == nil {
			result[uid] = statuses
		}
	}
	
	return result, nil
}

func (a *pureIDMWorkflowAdapter) GetAllUIDs() ([]string, error) {
	var uids []string
	for _, item := range a.collection.Items {
		uids = append(uids, item.UID)
	}
	return uids, nil
}

// Implement remaining ManagedStoreAdapter methods
func (a *pureIDMWorkflowAdapter) GetScopes() ([]string, error) {
	var scopes []string
	scopes = append(scopes, RootScope)
	for _, item := range a.collection.Items {
		scopes = append(scopes, item.UID)
	}
	return scopes, nil
}

func (a *pureIDMWorkflowAdapter) AddItem(parentUID string) (string, error) {
	var parentID string
	if parentUID != RootScope {
		parentID = parentUID
	}
	todo := models.NewIDMTodo("", parentID)
	a.collection.AddItem(todo)
	return todo.UID, nil
}

func (a *pureIDMWorkflowAdapter) RemoveItem(uid string) error {
	if a.collection.RemoveItem(uid) {
		return nil
	}
	return fmt.Errorf("todo with UID %s not found", uid)
}

func (a *pureIDMWorkflowAdapter) MoveItem(uid, newParentUID string) error {
	todo := a.collection.FindByUID(uid)
	if todo == nil {
		return fmt.Errorf("todo with UID %s not found", uid)
	}
	
	if newParentUID == RootScope {
		todo.ParentID = ""
	} else {
		todo.ParentID = newParentUID
	}
	
	return nil
}

func (a *pureIDMWorkflowAdapter) SetStatus(uid, status string) error {
	todo := a.collection.FindByUID(uid)
	if todo == nil {
		return fmt.Errorf("todo with UID %s not found", uid)
	}
	todo.EnsureStatuses()
	todo.Statuses["completion"] = status
	return nil
}

func (a *pureIDMWorkflowAdapter) SetPinned(uid string, isPinned bool) error {
	// Not implemented for todos
	return nil
}

// Implement WorkflowStoreAdapter lifecycle hooks
func (a *pureIDMWorkflowAdapter) OnStatusChange(uid, dimension, oldValue, newValue string) error {
	// No-op for pure IDM manager - side effects are handled at a higher level
	return nil
}

func (a *pureIDMWorkflowAdapter) ValidateStatusChange(uid, dimension, oldValue, newValue string) error {
	// No validation needed for pure IDM manager
	return nil
}

func (a *pureIDMWorkflowAdapter) SetStatusesBulk(updates map[string]map[string]string) error {
	for uid, statuses := range updates {
		if err := a.SetMultipleStatuses(uid, statuses); err != nil {
			return err
		}
	}
	return nil
}

// IsPureIDM returns true for PureIDMManager
func (m *PureIDMManager) IsPureIDM() bool {
	return true
}