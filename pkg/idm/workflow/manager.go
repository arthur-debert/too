package workflow

import (
	"fmt"
	
	"github.com/arthur-debert/too/pkg/idm"
)

// StatusManager provides a high-level interface for managing hierarchical items
// with multi-dimensional status and context-aware visibility rules.
type StatusManager struct {
	registry     *idm.Registry
	hierarchyMgr *idm.Manager
	adapter      WorkflowStoreAdapter
	config       WorkflowConfig
	
	// Cache for performance optimization
	visibilityCache map[string]map[string]bool // context -> uid -> visible
}

// NewStatusManager creates a new StatusManager with the given configuration.
func NewStatusManager(
	registry *idm.Registry,
	hierarchyMgr *idm.Manager,
	adapter WorkflowStoreAdapter,
	config WorkflowConfig,
) (*StatusManager, error) {
	// Validate the configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid workflow config: %w", err)
	}
	
	sm := &StatusManager{
		registry:        registry,
		hierarchyMgr:    hierarchyMgr,
		adapter:         adapter,
		config:          config,
		visibilityCache: make(map[string]map[string]bool),
	}
	
	return sm, nil
}

// GetConfig returns the workflow configuration.
func (sm *StatusManager) GetConfig() WorkflowConfig {
	return sm.config
}

// Registry returns the underlying IDM registry for direct access.
func (sm *StatusManager) Registry() *idm.Registry {
	return sm.registry
}

// HierarchyManager returns the underlying IDM hierarchy manager for direct access.
func (sm *StatusManager) HierarchyManager() *idm.Manager {
	return sm.hierarchyMgr
}

// --- Core Status Operations ---

// GetStatus returns the status value for a specific dimension of an item.
func (sm *StatusManager) GetStatus(uid, dimension string) (string, error) {
	// Check if dimension exists
	if sm.config.GetDimension(dimension) == nil {
		return "", fmt.Errorf("unknown dimension: %s", dimension)
	}
	
	return sm.adapter.GetItemStatus(uid, dimension)
}

// SetStatus sets the status value for a specific dimension of an item.
// This method bypasses transition validation - use Transition() for validated changes.
func (sm *StatusManager) SetStatus(uid, dimension, value string) error {
	// Validate dimension and value
	dim := sm.config.GetDimension(dimension)
	if dim == nil {
		return fmt.Errorf("unknown dimension: %s", dimension)
	}
	if !dim.HasValue(value) {
		return fmt.Errorf("invalid value '%s' for dimension '%s'", value, dimension)
	}
	
	// Get old value for hooks
	oldValue, _ := sm.adapter.GetItemStatus(uid, dimension)
	
	// Validate the change
	if err := sm.adapter.ValidateStatusChange(uid, dimension, oldValue, value); err != nil {
		return fmt.Errorf("status change validation failed: %w", err)
	}
	
	// Set the new status
	if err := sm.adapter.SetItemStatus(uid, dimension, value); err != nil {
		return err
	}
	
	// Clear visibility cache since status changed
	sm.clearVisibilityCache()
	
	// Call post-change hook
	if err := sm.adapter.OnStatusChange(uid, dimension, oldValue, value); err != nil {
		return fmt.Errorf("post-change hook failed: %w", err)
	}
	
	// Trigger auto-transitions on this item and its parent
	if err := sm.TriggerAutoTransitions("status_change", uid); err != nil {
		return fmt.Errorf("auto-transitions failed: %w", err)
	}
	
	// Also trigger auto-transitions on parent (for bottom-up completion)
	if err := sm.triggerParentAutoTransitions(uid); err != nil {
		return fmt.Errorf("parent auto-transitions failed: %w", err)
	}
	
	return nil
}

// GetAllStatuses returns all status dimensions and their values for an item.
func (sm *StatusManager) GetAllStatuses(uid string) (map[string]string, error) {
	return sm.adapter.GetItemStatuses(uid)
}

// SetMultipleStatuses sets multiple status dimensions for an item at once.
func (sm *StatusManager) SetMultipleStatuses(uid string, statuses map[string]string) error {
	// Validate all dimensions and values first
	for dimension, value := range statuses {
		dim := sm.config.GetDimension(dimension)
		if dim == nil {
			return fmt.Errorf("unknown dimension: %s", dimension)
		}
		if !dim.HasValue(value) {
			return fmt.Errorf("invalid value '%s' for dimension '%s'", value, dimension)
		}
	}
	
	// Get old values for hooks
	oldStatuses, _ := sm.adapter.GetItemStatuses(uid)
	
	// Validate all changes
	for dimension, value := range statuses {
		oldValue := oldStatuses[dimension]
		if err := sm.adapter.ValidateStatusChange(uid, dimension, oldValue, value); err != nil {
			return fmt.Errorf("status change validation failed for %s: %w", dimension, err)
		}
	}
	
	// Apply all changes
	if err := sm.adapter.SetMultipleStatuses(uid, statuses); err != nil {
		return err
	}
	
	// Clear visibility cache
	sm.clearVisibilityCache()
	
	// Call post-change hooks for each changed dimension
	for dimension, value := range statuses {
		oldValue := oldStatuses[dimension]
		if oldValue != value {
			if err := sm.adapter.OnStatusChange(uid, dimension, oldValue, value); err != nil {
				return fmt.Errorf("post-change hook failed for %s: %w", dimension, err)
			}
		}
	}
	
	// Trigger auto-transitions on this item and its parent
	if err := sm.TriggerAutoTransitions("status_change", uid); err != nil {
		return fmt.Errorf("auto-transitions failed: %w", err)
	}
	
	// Also trigger auto-transitions on parent (for bottom-up completion)
	if err := sm.triggerParentAutoTransitions(uid); err != nil {
		return fmt.Errorf("parent auto-transitions failed: %w", err)
	}
	
	return nil
}

// --- Transition Management ---

// CanTransition checks if a status transition is allowed according to the workflow rules.
func (sm *StatusManager) CanTransition(uid, dimension, newValue string) error {
	// Check if dimension exists
	dim := sm.config.GetDimension(dimension)
	if dim == nil {
		return fmt.Errorf("unknown dimension: %s", dimension)
	}
	
	// Check if new value is valid
	if !dim.HasValue(newValue) {
		return fmt.Errorf("invalid value '%s' for dimension '%s'", newValue, dimension)
	}
	
	// Get current value
	currentValue, err := sm.adapter.GetItemStatus(uid, dimension)
	if err != nil {
		// If no current value, check if we can set to new value from default
		if dim.DefaultValue != "" {
			currentValue = dim.DefaultValue
		} else {
			return fmt.Errorf("no current value and no default for dimension '%s'", dimension)
		}
	}
	
	// Check transition rules
	rules, exists := sm.config.Transitions[dimension]
	if !exists {
		// No rules means any transition is allowed
		return nil
	}
	
	for _, rule := range rules {
		if rule.CanTransition(currentValue, newValue) {
			// Found a matching rule, check custom validator if present
			if rule.Validator != nil {
				return rule.Validator(uid, sm.adapter)
			}
			return nil
		}
	}
	
	return fmt.Errorf("transition from '%s' to '%s' not allowed in dimension '%s'", currentValue, newValue, dimension)
}

// Transition performs a validated status transition.
func (sm *StatusManager) Transition(uid, dimension, newValue string) error {
	// Check if transition is allowed
	if err := sm.CanTransition(uid, dimension, newValue); err != nil {
		return err
	}
	
	// Perform the transition
	return sm.SetStatus(uid, dimension, newValue)
}

// --- Context-Aware Operations ---

// IsVisibleInContext checks if an item is visible in the given context.
func (sm *StatusManager) IsVisibleInContext(uid, context string) (bool, error) {
	// Check cache first
	if contextCache, exists := sm.visibilityCache[context]; exists {
		if visible, cached := contextCache[uid]; cached {
			return visible, nil
		}
	}
	
	// Get item statuses
	statuses, err := sm.adapter.GetItemStatuses(uid)
	if err != nil {
		return false, err
	}
	
	// Check visibility rules for this context
	rules, exists := sm.config.Visibility[context]
	if !exists {
		// No rules for this context means everything is visible
		return true, nil
	}
	
	// Item is visible if it matches any rule for this context
	for _, rule := range rules {
		if rule.Matches(context, statuses) {
			sm.cacheVisibility(context, uid, true)
			return true, nil
		}
	}
	
	sm.cacheVisibility(context, uid, false)
	return false, nil
}

// GetChildrenInContext returns children of a parent that are visible in the given context.
func (sm *StatusManager) GetChildrenInContext(parentUID, context string) ([]string, error) {
	// Get visibility rules for this context
	rules, exists := sm.config.Visibility[context]
	if !exists {
		// No rules means use standard GetChildren
		return sm.adapter.GetChildren(parentUID)
	}
	
	return sm.adapter.GetChildrenInContext(parentUID, context, rules)
}

// ResolvePositionPathInContext resolves a position path within a specific context.
// Only items visible in the context are considered for position numbering.
// This method creates a temporary context-aware registry and uses the core IDM resolver.
func (sm *StatusManager) ResolvePositionPathInContext(startScope, path, context string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("empty position path")
	}
	
	// Create a context-aware adapter that only returns visible children
	contextAdapter := &contextAwareAdapter{
		base:    sm.adapter,
		context: context,
		manager: sm,
	}
	
	// Create a temporary registry and populate it with context-visible children
	tempRegistry := idm.NewRegistry()
	
	// Build all relevant scopes that might be needed for path resolution
	scopes, err := sm.adapter.GetScopes()
	if err != nil {
		return "", fmt.Errorf("failed to get scopes: %w", err)
	}
	
	for _, scope := range scopes {
		if err := tempRegistry.RebuildScope(contextAdapter, scope); err != nil {
			continue // Skip scopes with errors, but don't fail completely
		}
	}
	
	// Use the core IDM resolver for path resolution
	return tempRegistry.ResolvePositionPath(startScope, path)
}

// GetPositionPathInContext returns the position path for an item within a specific context.
// This method creates a temporary context-aware registry and uses the core IDM GetPositionPath.
func (sm *StatusManager) GetPositionPathInContext(startScope, targetUID, context string) (string, error) {
	// Check if the target is visible in this context
	visible, err := sm.IsVisibleInContext(targetUID, context)
	if err != nil {
		return "", err
	}
	if !visible {
		return "", fmt.Errorf("item '%s' not visible in context '%s'", targetUID, context)
	}
	
	// Create a context-aware adapter that only returns visible children
	contextAdapter := &contextAwareAdapter{
		base:    sm.adapter,
		context: context,
		manager: sm,
	}
	
	// Create a temporary registry and build all context-visible scopes
	tempRegistry := idm.NewRegistry()
	scopes, err := sm.adapter.GetScopes()
	if err != nil {
		return "", err
	}
	
	for _, scope := range scopes {
		if err := tempRegistry.RebuildScope(contextAdapter, scope); err != nil {
			continue // Skip scopes with errors
		}
	}
	
	// Use the core IDM GetPositionPath method
	return tempRegistry.GetPositionPath(startScope, targetUID, contextAdapter)
}

// --- Auto-Transitions ---

// TriggerAutoTransitions executes automatic status transitions based on the trigger type.
func (sm *StatusManager) TriggerAutoTransitions(triggerType, targetUID string) error {
	for _, rule := range sm.config.AutoTransitions {
		if rule.Trigger == triggerType || (rule.Trigger == "status_change" && triggerType == "child_status_change") {
			if err := sm.executeAutoTransition(rule, targetUID); err != nil {
				return fmt.Errorf("failed to execute auto-transition rule: %w", err)
			}
		}
	}
	return nil
}

// executeAutoTransition executes a single auto-transition rule.
func (sm *StatusManager) executeAutoTransition(rule AutoTransitionRule, targetUID string) error {
	switch rule.Condition {
	case "all_children_status_equals":
		return sm.executeAllChildrenStatusEquals(rule, targetUID)
	default:
		// Unknown condition type, skip silently
		return nil
	}
}

// executeAllChildrenStatusEquals implements the "all_children_status_equals" condition.
func (sm *StatusManager) executeAllChildrenStatusEquals(rule AutoTransitionRule, targetUID string) error {
	// Get all children of the target item
	allChildren, err := sm.adapter.GetChildren(targetUID)
	if err != nil {
		return err
	}
	
	// If no children, skip this rule
	if len(allChildren) == 0 {
		return nil
	}
	
	// Check if all children have the required status
	for _, childUID := range allChildren {
		childStatus, err := sm.adapter.GetItemStatus(childUID, rule.TargetDimension)
		if err != nil {
			// If child doesn't have this dimension, condition fails
			return nil
		}
		if childStatus != rule.ConditionValue {
			// At least one child doesn't match, condition fails
			return nil
		}
	}
	
	// Check if the target is already in the desired state (prevent infinite recursion)
	currentStatus, err := sm.adapter.GetItemStatus(targetUID, rule.TargetDimension)
	if err == nil && currentStatus == rule.ActionValue {
		// Already in desired state, nothing to do
		return nil
	}
	
	// All children match the condition, execute the action
	switch rule.Action {
	case "set_status":
		return sm.SetStatus(targetUID, rule.TargetDimension, rule.ActionValue)
	default:
		// Unknown action type, skip silently
		return nil
	}
}

// --- Metrics and Aggregation ---

// GetMetrics returns aggregated metrics for items in a given scope.
func (sm *StatusManager) GetMetrics(scope string) (*StatusMetrics, error) {
	metrics := NewStatusMetrics()
	
	// Get all items in the scope (recursively)
	allItems, err := sm.getAllItemsInScope(scope)
	if err != nil {
		return nil, err
	}
	
	// Collect statistics
	for _, uid := range allItems {
		statuses, err := sm.adapter.GetItemStatuses(uid)
		if err != nil {
			continue // Skip items with errors
		}
		metrics.AddItem(statuses)
	}
	
	// Add context-specific counts
	for context := range sm.config.Visibility {
		visibleItems, err := sm.GetAllItemsInContext(context)
		if err != nil {
			continue
		}
		metrics.SetContextCount(context, len(visibleItems))
	}
	
	return metrics, nil
}

// GetAllItemsInContext returns all items that are visible in the given context.
func (sm *StatusManager) GetAllItemsInContext(context string) ([]string, error) {
	rules, exists := sm.config.Visibility[context]
	if !exists {
		// No rules means all items are visible
		return sm.adapter.GetAllUIDs()
	}
	
	return sm.adapter.GetAllItemsInContext(context, rules)
}

// --- Context-Aware Adapter ---

// contextAwareAdapter wraps a WorkflowStoreAdapter to provide context-filtered children
// for use with the core IDM resolver. It implements idm.StoreAdapter.
type contextAwareAdapter struct {
	base    WorkflowStoreAdapter
	context string
	manager *StatusManager
}

// GetChildren returns only children that are visible in the specified context
func (ca *contextAwareAdapter) GetChildren(scope string) ([]string, error) {
	return ca.manager.GetChildrenInContext(scope, ca.context)
}

// GetScopes delegates to the base adapter
func (ca *contextAwareAdapter) GetScopes() ([]string, error) {
	return ca.base.GetScopes()
}

// GetAllUIDs delegates to the base adapter
func (ca *contextAwareAdapter) GetAllUIDs() ([]string, error) {
	return ca.base.GetAllUIDs()
}

// --- Helper Methods ---

// getAllItemsInScope recursively gets all items under a given scope.
func (sm *StatusManager) getAllItemsInScope(scope string) ([]string, error) {
	var allItems []string
	
	children, err := sm.adapter.GetChildren(scope)
	if err != nil {
		return nil, err
	}
	
	for _, child := range children {
		allItems = append(allItems, child)
		
		// Recursively get children of this child
		grandChildren, err := sm.getAllItemsInScope(child)
		if err != nil {
			continue // Skip subtrees with errors
		}
		allItems = append(allItems, grandChildren...)
	}
	
	return allItems, nil
}

// cacheVisibility stores a visibility result in the cache.
func (sm *StatusManager) cacheVisibility(context, uid string, visible bool) {
	if sm.visibilityCache[context] == nil {
		sm.visibilityCache[context] = make(map[string]bool)
	}
	sm.visibilityCache[context][uid] = visible
}

// clearVisibilityCache clears the visibility cache (called when statuses change).
func (sm *StatusManager) clearVisibilityCache() {
	sm.visibilityCache = make(map[string]map[string]bool)
}

// InitializeItemWithDefaults sets default status values for a new item.
func (sm *StatusManager) InitializeItemWithDefaults(uid string) error {
	defaults := make(map[string]string)
	
	for _, dimension := range sm.config.Dimensions {
		if dimension.DefaultValue != "" {
			defaults[dimension.Name] = dimension.DefaultValue
		}
	}
	
	if len(defaults) > 0 {
		return sm.adapter.SetMultipleStatuses(uid, defaults)
	}
	
	return nil
}

// triggerParentAutoTransitions triggers auto-transitions on the parent of the given item.
func (sm *StatusManager) triggerParentAutoTransitions(uid string) error {
	// Use the new GetParent method for efficient parent discovery
	parentUID, err := sm.adapter.GetParent(uid)
	if err != nil {
		return err
	}
	
	// If there's no parent (empty string), this is a root item
	if parentUID == "" {
		return nil
	}
	
	// Trigger auto-transitions on the parent
	return sm.TriggerAutoTransitions("child_status_change", parentUID)
}