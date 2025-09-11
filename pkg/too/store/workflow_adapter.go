package store

import (
	"fmt"
	"sort"

	"github.com/arthur-debert/too/pkg/idm/workflow"
	"github.com/arthur-debert/too/pkg/too/models"
)

// WorkflowTodoAdapter implements the workflow.WorkflowStoreAdapter interface
// to bridge the gap between the workflow system and too's data model.
type WorkflowTodoAdapter struct {
	store      Store
	collection *models.Collection
}

// NewWorkflowTodoAdapter creates a new workflow adapter for the todo system.
func NewWorkflowTodoAdapter(store Store) (*WorkflowTodoAdapter, error) {
	collection, err := store.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load collection: %w", err)
	}
	return &WorkflowTodoAdapter{
		store:      store,
		collection: collection,
	}, nil
}

// Implementation of idm.StoreAdapter interface

// GetChildren returns child UIDs for a given parent UID.
func (a *WorkflowTodoAdapter) GetChildren(parentUID string) ([]string, error) {
	var targetTodos []*models.Todo

	if parentUID == RootScope {
		targetTodos = a.collection.Todos
	} else {
		parent := a.collection.FindItemByID(parentUID)
		if parent == nil {
			return []string{}, nil
		}
		targetTodos = parent.Items
	}

	// Filter based on workflow visibility (active items only by default)
	var children []string
	for _, todo := range targetTodos {
		// In workflow mode, we include all items and let context filtering handle visibility
		children = append(children, todo.ID)
	}

	// Sort by position for consistent ordering
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

// GetScopes returns all available scopes (parent UIDs that have children).
func (a *WorkflowTodoAdapter) GetScopes() ([]string, error) {
	scopes := []string{RootScope}
	
	// Find all todos that have children
	for _, todo := range a.collection.AllTodos() {
		if len(todo.Items) > 0 {
			scopes = append(scopes, todo.ID)
		}
	}
	
	return scopes, nil
}

// GetAllUIDs returns all UIDs in the collection.
func (a *WorkflowTodoAdapter) GetAllUIDs() ([]string, error) {
	var uids []string
	for _, todo := range a.collection.AllTodos() {
		uids = append(uids, todo.ID)
	}
	return uids, nil
}

// Implementation of idm.ManagedStoreAdapter interface

// AddItem creates a new todo item under the specified parent.
func (a *WorkflowTodoAdapter) AddItem(parentUID string) (string, error) {
	// Create a basic todo with placeholder text
	todo, err := a.collection.CreateTodo("", parentUID)
	if err != nil {
		return "", fmt.Errorf("failed to create todo: %w", err)
	}
	return todo.ID, nil
}

// RemoveItem removes a todo item from the collection.
func (a *WorkflowTodoAdapter) RemoveItem(uid string) error {
	todo := a.collection.FindItemByID(uid)
	if todo == nil {
		return fmt.Errorf("todo with ID %s not found", uid)
	}
	
	return a.collection.RemoveTodo(todo.ID)
}

// MoveItem moves a todo from one parent to another.
func (a *WorkflowTodoAdapter) MoveItem(uid, newParentUID string) error {
	todo := a.collection.FindItemByID(uid)
	if todo == nil {
		return fmt.Errorf("todo with ID %s not found", uid)
	}
	
	// Update parent ID
	oldParentID := todo.ParentID
	todo.ParentID = newParentUID
	
	// Remove from old parent's items list
	if oldParentID == "" {
		// Remove from root level
		for i, rootTodo := range a.collection.Todos {
			if rootTodo.ID == uid {
				a.collection.Todos = append(a.collection.Todos[:i], a.collection.Todos[i+1:]...)
				break
			}
		}
	} else {
		// Remove from old parent's items
		oldParent := a.collection.FindItemByID(oldParentID)
		if oldParent != nil {
			for i, child := range oldParent.Items {
				if child.ID == uid {
					oldParent.Items = append(oldParent.Items[:i], oldParent.Items[i+1:]...)
					break
				}
			}
		}
	}
	
	// Add to new parent's items list
	if newParentUID == RootScope || newParentUID == "" {
		todo.ParentID = ""
		todo.Position = a.collection.FindHighestPosition(a.collection.Todos) + 1
		a.collection.Todos = append(a.collection.Todos, todo)
	} else {
		newParent := a.collection.FindItemByID(newParentUID)
		if newParent == nil {
			return fmt.Errorf("new parent with ID %s not found", newParentUID)
		}
		todo.Position = a.collection.FindHighestPosition(newParent.Items) + 1
		newParent.Items = append(newParent.Items, todo)
	}
	
	return nil
}

// SetStatus sets the completion status through the workflow system.
func (a *WorkflowTodoAdapter) SetStatus(uid, status string) error {
	// Delegate to SetItemStatus with completion dimension
	return a.SetItemStatus(uid, "completion", status)
}

// SetPinned is not applicable to the todo system, but we implement it for interface compliance.
func (a *WorkflowTodoAdapter) SetPinned(uid string, isPinned bool) error {
	// In the todo system, we could use this for a "starred" or "favorite" feature
	// For now, we'll store this in the workflow statuses
	return a.SetItemStatus(uid, "pinned", fmt.Sprintf("%t", isPinned))
}

// GetParent returns the parent UID of the given item.
func (a *WorkflowTodoAdapter) GetParent(uid string) (string, error) {
	if uid == RootScope {
		return "", nil // Root has no parent
	}
	
	todo := a.collection.FindItemByID(uid)
	if todo == nil {
		return "", fmt.Errorf("todo with ID %s not found", uid)
	}
	
	if todo.ParentID == "" {
		return RootScope, nil
	}
	
	return todo.ParentID, nil
}

// Implementation of workflow.WorkflowStoreAdapter interface

// SetItemStatus sets a status dimension for a todo item.
func (a *WorkflowTodoAdapter) SetItemStatus(uid, dimension, value string) error {
	// Skip setting status on root scope as it's not a real todo
	if uid == RootScope {
		return nil
	}
	
	todo := a.collection.FindItemByID(uid)
	if todo == nil {
		return fmt.Errorf("todo with ID %s not found", uid)
	}
	
	// Initialize statuses map if needed
	todo.EnsureStatuses()
	
	// Set the status dimension
	todo.Statuses[dimension] = value
	
	// Update modified timestamp when status changes
	todo.SetModified()
	
	return nil
}

// GetItemStatus gets a status dimension for a todo item.
func (a *WorkflowTodoAdapter) GetItemStatus(uid, dimension string) (string, error) {
	// Root scope doesn't have status - return empty
	if uid == RootScope {
		return "", nil
	}
	
	todo := a.collection.FindItemByID(uid)
	if todo == nil {
		return "", fmt.Errorf("todo with ID %s not found", uid)
	}
	
	// Initialize statuses if needed
	todo.EnsureStatuses()
	
	// Check if the dimension exists
	if value, exists := todo.Statuses[dimension]; exists {
		return value, nil
	}
	
	return "", fmt.Errorf("dimension %s not found for todo %s", dimension, uid)
}

// GetItemStatuses gets all status dimensions for a todo item.
func (a *WorkflowTodoAdapter) GetItemStatuses(uid string) (map[string]string, error) {
	todo := a.collection.FindItemByID(uid)
	if todo == nil {
		return nil, fmt.Errorf("todo with ID %s not found", uid)
	}
	
	// Initialize statuses if needed
	todo.EnsureStatuses()
	
	// Return a copy to prevent external modification
	result := make(map[string]string)
	for k, v := range todo.Statuses {
		result[k] = v
	}
	
	// Ensure completion dimension is always available
	if _, exists := result["completion"]; !exists {
		result["completion"] = string(models.StatusPending)
	}
	
	return result, nil
}

// SetMultipleStatuses sets multiple status dimensions for a todo item.
func (a *WorkflowTodoAdapter) SetMultipleStatuses(uid string, statuses map[string]string) error {
	todo := a.collection.FindItemByID(uid)
	if todo == nil {
		return fmt.Errorf("todo with ID %s not found", uid)
	}
	
	// Initialize statuses if needed
	todo.EnsureStatuses()
	
	// Set all the status dimensions
	for dimension, value := range statuses {
		todo.Statuses[dimension] = value
	}
	
	return nil
}

// GetChildrenInContext returns children that are visible in the given context.
func (a *WorkflowTodoAdapter) GetChildrenInContext(parentUID, context string, visibilityRules []workflow.VisibilityRule) ([]string, error) {
	// Get all children first
	allChildren, err := a.GetChildren(parentUID)
	if err != nil {
		return nil, err
	}
	
	// Filter based on visibility rules for the given context
	var visibleChildren []string
	for _, childUID := range allChildren {
		statuses, err := a.GetItemStatuses(childUID)
		if err != nil {
			continue // Skip items with errors
		}
		
		// Check if this item is visible in the context (AND logic - all rules must match)
		visible := true
		for _, rule := range visibilityRules {
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

// GetAllItemsInContext returns all items that are visible in the given context.
func (a *WorkflowTodoAdapter) GetAllItemsInContext(context string, visibilityRules []workflow.VisibilityRule) ([]string, error) {
	allUIDs, err := a.GetAllUIDs()
	if err != nil {
		return nil, err
	}
	
	var visibleItems []string
	for _, uid := range allUIDs {
		statuses, err := a.GetItemStatuses(uid)
		if err != nil {
			continue
		}
		
		// Check if this item is visible in the context (AND logic - all rules must match)
		visible := true
		for _, rule := range visibilityRules {
			if !rule.Matches(context, statuses) {
				visible = false
				break
			}
		}
		
		if visible {
			visibleItems = append(visibleItems, uid)
		}
	}
	
	return visibleItems, nil
}

// GetStatusesBulk efficiently retrieves status information for multiple items.
func (a *WorkflowTodoAdapter) GetStatusesBulk(uids []string) (map[string]map[string]string, error) {
	result := make(map[string]map[string]string)
	
	for _, uid := range uids {
		if statuses, err := a.GetItemStatuses(uid); err == nil {
			result[uid] = statuses
		}
	}
	
	return result, nil
}

// SetStatusesBulk efficiently sets status information for multiple items.
func (a *WorkflowTodoAdapter) SetStatusesBulk(updates map[string]map[string]string) error {
	for uid, statuses := range updates {
		if err := a.SetMultipleStatuses(uid, statuses); err != nil {
			return err
		}
	}
	return nil
}

// OnStatusChange is called after a status change has been successfully applied.
// This is where we can add side effects like notifications, audit logging, etc.
func (a *WorkflowTodoAdapter) OnStatusChange(uid, dimension, oldValue, newValue string) error {
	// For now, this is a no-op, but we could add:
	// - Audit logging
	// - Notifications
	// - External system updates
	// - Metrics collection
	return nil
}

// ValidateStatusChange validates a status change before it's applied.
func (a *WorkflowTodoAdapter) ValidateStatusChange(uid, dimension, oldValue, newValue string) error {
	// Skip validation for root scope as it's not a real todo
	if uid == RootScope {
		return nil
	}
	
	// Basic validation - ensure the todo exists
	todo := a.collection.FindItemByID(uid)
	if todo == nil {
		return fmt.Errorf("todo with ID %s not found", uid)
	}
	
	// Additional validation could be added here:
	// - Business rules
	// - Permissions
	// - Data constraints
	
	return nil
}

// Save persists the current state to the store.
func (a *WorkflowTodoAdapter) Save() error {
	return a.store.Save(a.collection)
}

// Collection returns the current collection for direct access.
func (a *WorkflowTodoAdapter) Collection() *models.Collection {
	return a.collection
}