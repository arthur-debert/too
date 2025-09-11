package workflow

import (
	"fmt"
	
	"github.com/arthur-debert/too/pkg/idm"
)

// WorkflowStoreAdapter extends the basic IDM interfaces with workflow-specific operations.
// It provides methods for managing multi-dimensional status and context-aware queries.
type WorkflowStoreAdapter interface {
	// Inherit basic IDM functionality
	idm.ManagedStoreAdapter
	
	// Status management methods
	
	// SetItemStatus sets a specific status dimension for an item.
	SetItemStatus(uid, dimension, value string) error
	
	// GetItemStatus gets the status value for a specific dimension of an item.
	GetItemStatus(uid, dimension string) (string, error)
	
	// GetItemStatuses gets all status dimensions and their values for an item.
	GetItemStatuses(uid string) (map[string]string, error)
	
	// SetMultipleStatuses efficiently sets multiple status dimensions for an item.
	SetMultipleStatuses(uid string, statuses map[string]string) error
	
	// Context-aware query methods
	
	// GetChildrenInContext returns children of a parent that are visible in the given context.
	// This is used to filter items based on their status values and visibility rules.
	GetChildrenInContext(parentUID, context string, visibilityRules []VisibilityRule) ([]string, error)
	
	// GetAllItemsInContext returns all items (across all levels) that are visible in the given context.
	GetAllItemsInContext(context string, visibilityRules []VisibilityRule) ([]string, error)
	
	// Bulk operations for performance
	
	// GetStatusesBulk efficiently retrieves status information for multiple items.
	GetStatusesBulk(uids []string) (map[string]map[string]string, error)
	
	// SetStatusesBulk efficiently sets status information for multiple items.
	// The input map is uid -> dimension -> value.
	SetStatusesBulk(updates map[string]map[string]string) error
	
	// Lifecycle and validation hooks
	
	// OnStatusChange is called after a status change has been successfully applied,
	// allowing for side effects outside of the core workflow system.
	// 
	// This hook is intended for:
	// - Sending notifications (email, webhooks, etc.)
	// - Logging audit trails for compliance
	// - Triggering external integrations
	// - Updating derived data or caches
	// - Publishing events to message queues
	//
	// Note: Auto-transitions are handled separately by the StatusManager.
	// This hook should NOT be used for workflow logic that affects other items' statuses.
	OnStatusChange(uid, dimension, oldValue, newValue string) error
	
	// ValidateStatusChange is called before a status change to validate the operation.
	ValidateStatusChange(uid, dimension, oldValue, newValue string) error
}

// MockWorkflowAdapter is a test implementation of WorkflowStoreAdapter.
// It stores all data in memory and is useful for testing workflow logic
// without requiring a real persistence layer.
type MockWorkflowAdapter struct {
	items    map[string]*MockItem        // uid -> item
	statuses map[string]map[string]string // uid -> dimension -> value
	nextUID  int
}

// MockItem represents an item in the mock adapter.
type MockItem struct {
	UID      string
	ParentID string
	Children []string
}

// NewMockWorkflowAdapter creates a new mock adapter for testing.
func NewMockWorkflowAdapter() *MockWorkflowAdapter {
	return &MockWorkflowAdapter{
		items:    make(map[string]*MockItem),
		statuses: make(map[string]map[string]string),
		nextUID:  1,
	}
}

// Helper method to generate new UIDs
func (m *MockWorkflowAdapter) newUID() string {
	uid := fmt.Sprintf("item_%d", m.nextUID)
	m.nextUID++
	return uid
}

// Implementation of idm.StoreAdapter interface

func (m *MockWorkflowAdapter) GetChildren(parentUID string) ([]string, error) {
	if item, exists := m.items[parentUID]; exists {
		return item.Children, nil
	}
	if parentUID == "root" {
		// Return root-level items
		var rootItems []string
		for uid, item := range m.items {
			if item.ParentID == "" || item.ParentID == "root" {
				rootItems = append(rootItems, uid)
			}
		}
		return rootItems, nil
	}
	return []string{}, nil
}

func (m *MockWorkflowAdapter) GetScopes() ([]string, error) {
	scopes := []string{"root"}
	for uid, item := range m.items {
		if len(item.Children) > 0 {
			scopes = append(scopes, uid)
		}
	}
	return scopes, nil
}

func (m *MockWorkflowAdapter) GetAllUIDs() ([]string, error) {
	uids := make([]string, 0, len(m.items))
	for uid := range m.items {
		uids = append(uids, uid)
	}
	return uids, nil
}

// Implementation of idm.ManagedStoreAdapter interface

func (m *MockWorkflowAdapter) AddItem(parentUID string) (string, error) {
	uid := m.newUID()
	
	parentID := parentUID
	if parentUID == "root" {
		parentID = ""
	}
	
	item := &MockItem{
		UID:      uid,
		ParentID: parentID,
		Children: []string{},
	}
	
	m.items[uid] = item
	m.statuses[uid] = make(map[string]string)
	
	// Add to parent's children list
	if parentUID != "root" {
		if parent, exists := m.items[parentUID]; exists {
			parent.Children = append(parent.Children, uid)
		}
	}
	
	return uid, nil
}

func (m *MockWorkflowAdapter) RemoveItem(uid string) error {
	item, exists := m.items[uid]
	if !exists {
		return fmt.Errorf("item %s not found", uid)
	}
	
	// Remove from parent's children list
	if item.ParentID != "" {
		if parent, exists := m.items[item.ParentID]; exists {
			for i, childUID := range parent.Children {
				if childUID == uid {
					parent.Children = append(parent.Children[:i], parent.Children[i+1:]...)
					break
				}
			}
		}
	}
	
	// Remove the item and its statuses
	delete(m.items, uid)
	delete(m.statuses, uid)
	
	return nil
}

func (m *MockWorkflowAdapter) MoveItem(uid, newParentUID string) error {
	item, exists := m.items[uid]
	if !exists {
		return fmt.Errorf("item %s not found", uid)
	}
	
	// Remove from old parent
	if item.ParentID != "" {
		if oldParent, exists := m.items[item.ParentID]; exists {
			for i, childUID := range oldParent.Children {
				if childUID == uid {
					oldParent.Children = append(oldParent.Children[:i], oldParent.Children[i+1:]...)
					break
				}
			}
		}
	}
	
	// Update parent ID
	newParentID := newParentUID
	if newParentUID == "root" {
		newParentID = ""
	}
	item.ParentID = newParentID
	
	// Add to new parent
	if newParentUID != "root" {
		if newParent, exists := m.items[newParentUID]; exists {
			newParent.Children = append(newParent.Children, uid)
		}
	}
	
	return nil
}

func (m *MockWorkflowAdapter) SetStatus(uid, status string) error {
	// This is the legacy single-status method from ManagedStoreAdapter
	// We map it to a "legacy" dimension for compatibility
	return m.SetItemStatus(uid, "legacy", status)
}

func (m *MockWorkflowAdapter) SetPinned(uid string, isPinned bool) error {
	value := "false"
	if isPinned {
		value = "true"
	}
	return m.SetItemStatus(uid, "pinned", value)
}

func (m *MockWorkflowAdapter) GetParent(uid string) (string, error) {
	// Handle root scope specially
	if uid == "root" {
		return "", nil // Root has no parent
	}
	
	item, exists := m.items[uid]
	if !exists {
		return "", fmt.Errorf("item %s not found", uid)
	}
	
	// If ParentID is empty, this item is a direct child of root
	if item.ParentID == "" {
		return "root", nil
	}
	
	return item.ParentID, nil
}

// Implementation of WorkflowStoreAdapter interface

func (m *MockWorkflowAdapter) SetItemStatus(uid, dimension, value string) error {
	// Handle root scope specially
	if uid == "root" {
		if m.statuses[uid] == nil {
			m.statuses[uid] = make(map[string]string)
		}
		m.statuses[uid][dimension] = value
		return nil
	}
	
	if _, exists := m.items[uid]; !exists {
		return fmt.Errorf("item %s not found", uid)
	}
	
	if m.statuses[uid] == nil {
		m.statuses[uid] = make(map[string]string)
	}
	
	m.statuses[uid][dimension] = value
	return nil
}

func (m *MockWorkflowAdapter) GetItemStatus(uid, dimension string) (string, error) {
	// Handle root scope specially
	if uid == "root" {
		if statuses, exists := m.statuses[uid]; exists {
			if value, exists := statuses[dimension]; exists {
				return value, nil
			}
		}
		return "", fmt.Errorf("status %s not found for root", dimension)
	}
	
	if _, exists := m.items[uid]; !exists {
		return "", fmt.Errorf("item %s not found", uid)
	}
	
	if statuses, exists := m.statuses[uid]; exists {
		if value, exists := statuses[dimension]; exists {
			return value, nil
		}
	}
	
	return "", fmt.Errorf("dimension %s not found for item %s", dimension, uid)
}

func (m *MockWorkflowAdapter) GetItemStatuses(uid string) (map[string]string, error) {
	// Handle root scope specially
	if uid == "root" {
		if statuses, exists := m.statuses[uid]; exists {
			// Return a copy to prevent external modification
			result := make(map[string]string)
			for k, v := range statuses {
				result[k] = v
			}
			return result, nil
		}
		return make(map[string]string), nil // Return empty map for root if no statuses
	}
	
	if _, exists := m.items[uid]; !exists {
		return nil, fmt.Errorf("item %s not found", uid)
	}
	
	if statuses, exists := m.statuses[uid]; exists {
		// Return a copy to prevent external modification
		result := make(map[string]string)
		for k, v := range statuses {
			result[k] = v
		}
		return result, nil
	}
	
	return make(map[string]string), nil
}

func (m *MockWorkflowAdapter) SetMultipleStatuses(uid string, statuses map[string]string) error {
	if _, exists := m.items[uid]; !exists {
		return fmt.Errorf("item %s not found", uid)
	}
	
	if m.statuses[uid] == nil {
		m.statuses[uid] = make(map[string]string)
	}
	
	for dimension, value := range statuses {
		m.statuses[uid][dimension] = value
	}
	
	return nil
}

func (m *MockWorkflowAdapter) GetChildrenInContext(parentUID, context string, visibilityRules []VisibilityRule) ([]string, error) {
	// Get all children first
	allChildren, err := m.GetChildren(parentUID)
	if err != nil {
		return nil, err
	}
	
	// Filter based on visibility rules for the given context
	var visibleChildren []string
	for _, childUID := range allChildren {
		statuses, err := m.GetItemStatuses(childUID)
		if err != nil {
			continue // Skip items with errors
		}
		
		// Check if this item is visible in the context
		visible := false
		for _, rule := range visibilityRules {
			if rule.Matches(context, statuses) {
				visible = true
				break
			}
		}
		
		if visible {
			visibleChildren = append(visibleChildren, childUID)
		}
	}
	
	return visibleChildren, nil
}

func (m *MockWorkflowAdapter) GetAllItemsInContext(context string, visibilityRules []VisibilityRule) ([]string, error) {
	var visibleItems []string
	
	for uid := range m.items {
		statuses, err := m.GetItemStatuses(uid)
		if err != nil {
			continue
		}
		
		// Check if this item is visible in the context
		visible := false
		for _, rule := range visibilityRules {
			if rule.Matches(context, statuses) {
				visible = true
				break
			}
		}
		
		if visible {
			visibleItems = append(visibleItems, uid)
		}
	}
	
	return visibleItems, nil
}

func (m *MockWorkflowAdapter) GetStatusesBulk(uids []string) (map[string]map[string]string, error) {
	result := make(map[string]map[string]string)
	
	for _, uid := range uids {
		if statuses, err := m.GetItemStatuses(uid); err == nil {
			result[uid] = statuses
		}
	}
	
	return result, nil
}

func (m *MockWorkflowAdapter) SetStatusesBulk(updates map[string]map[string]string) error {
	for uid, statuses := range updates {
		if err := m.SetMultipleStatuses(uid, statuses); err != nil {
			return err
		}
	}
	return nil
}

func (m *MockWorkflowAdapter) OnStatusChange(uid, dimension, oldValue, newValue string) error {
	// Mock implementation - in a real implementation, this could:
	// - Log audit trail: fmt.Printf("Item %s: %s changed from %s to %s\n", uid, dimension, oldValue, newValue)
	// - Send notifications: emailService.NotifyStatusChange(uid, dimension, newValue)
	// - Update search indexes: searchIndex.UpdateItem(uid, map[string]string{dimension: newValue})
	// - Publish events: eventBus.Publish("status.changed", StatusChangeEvent{...})
	return nil
}

func (m *MockWorkflowAdapter) ValidateStatusChange(uid, dimension, oldValue, newValue string) error {
	// Mock implementation - always allows changes
	return nil
}