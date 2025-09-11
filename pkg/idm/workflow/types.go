package workflow

import "fmt"

// StatusDimension defines a named axis of status values that items can have.
// For example, a "completion" dimension might have values ["pending", "done"],
// while a "priority" dimension might have values ["low", "medium", "high"].
type StatusDimension struct {
	Name         string   // The dimension name (e.g., "completion", "priority")
	Values       []string // Valid values for this dimension
	DefaultValue string   // Default value when creating new items
}

// Validate checks if the dimension configuration is valid.
func (d StatusDimension) Validate() error {
	if d.Name == "" {
		return fmt.Errorf("dimension name cannot be empty")
	}
	if len(d.Values) == 0 {
		return fmt.Errorf("dimension must have at least one value")
	}
	if d.DefaultValue != "" {
		found := false
		for _, v := range d.Values {
			if v == d.DefaultValue {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("default value '%s' not found in dimension values", d.DefaultValue)
		}
	}
	return nil
}

// HasValue checks if a value is valid for this dimension.
func (d StatusDimension) HasValue(value string) bool {
	for _, v := range d.Values {
		if v == value {
			return true
		}
	}
	return false
}

// VisibilityRule defines which items are visible in a given context based on their status.
// For example, an "active" context might only show items with completion="pending",
// while an "all" context shows both "pending" and "done" items.
type VisibilityRule struct {
	Context   string   // The context name (e.g., "active", "all", "admin")
	Dimension string   // Which status dimension to check
	Include   []string // Status values to include (empty means include all)
	Exclude   []string // Status values to exclude (takes precedence over Include)
}

// Matches checks if an item with the given status values is visible in this rule's context.
func (r VisibilityRule) Matches(context string, statuses map[string]string) bool {
	if r.Context != context {
		return false
	}
	
	statusValue, exists := statuses[r.Dimension]
	if !exists {
		return false // Item doesn't have this dimension
	}
	
	// Check exclusions first (they take precedence)
	for _, excluded := range r.Exclude {
		if statusValue == excluded {
			return false
		}
	}
	
	// If no includes specified, and not excluded, then visible
	if len(r.Include) == 0 {
		return true
	}
	
	// Check if value is in include list
	for _, included := range r.Include {
		if statusValue == included {
			return true
		}
	}
	
	return false
}

// TransitionRule defines valid status transitions within a dimension.
type TransitionRule struct {
	Dimension string   // Which dimension this rule applies to
	From      string   // Source status value
	To        []string // Valid destination status values
	Validator func(uid string, adapter WorkflowStoreAdapter) error // Optional custom validation
}

// CanTransition checks if a transition from one value to another is allowed.
func (r TransitionRule) CanTransition(from, to string) bool {
	if r.From != from {
		return false
	}
	for _, validTo := range r.To {
		if validTo == to {
			return true
		}
	}
	return false
}

// AutoTransitionRule defines automatic status changes triggered by certain events.
// For example, when all children of an item are marked as "done", 
// the parent item could automatically be marked as "done" as well.
type AutoTransitionRule struct {
	Trigger        string // Event that triggers this rule (e.g., "child_status_change")
	Condition      string // Condition to check (e.g., "all_children_status_equals")
	ConditionValue string // Value for the condition (e.g., "done")
	TargetDimension string // Which dimension to modify (e.g., "completion")
	Action         string // What action to take (e.g., "set_status")
	ActionValue    string // Value for the action (e.g., "done")
}

// WorkflowConfig contains the complete configuration for a workflow system.
type WorkflowConfig struct {
	Dimensions      []StatusDimension             // Available status dimensions
	Visibility      map[string][]VisibilityRule   // Context -> rules mapping
	Transitions     map[string][]TransitionRule   // Dimension -> rules mapping
	AutoTransitions []AutoTransitionRule          // Rules for automatic transitions
}

// Validate checks if the workflow configuration is valid.
func (c WorkflowConfig) Validate() error {
	// Validate dimensions
	dimensionNames := make(map[string]bool)
	for _, dim := range c.Dimensions {
		if err := dim.Validate(); err != nil {
			return fmt.Errorf("invalid dimension '%s': %w", dim.Name, err)
		}
		if dimensionNames[dim.Name] {
			return fmt.Errorf("duplicate dimension name: %s", dim.Name)
		}
		dimensionNames[dim.Name] = true
	}
	
	// Validate that visibility rules reference valid dimensions
	for context, rules := range c.Visibility {
		for _, rule := range rules {
			if !dimensionNames[rule.Dimension] {
				return fmt.Errorf("visibility rule in context '%s' references unknown dimension '%s'", context, rule.Dimension)
			}
		}
	}
	
	// Validate that transition rules reference valid dimensions and values
	for dimName, rules := range c.Transitions {
		if !dimensionNames[dimName] {
			return fmt.Errorf("transition rules reference unknown dimension '%s'", dimName)
		}
		
		// Find the dimension to validate values
		var dimension StatusDimension
		for _, d := range c.Dimensions {
			if d.Name == dimName {
				dimension = d
				break
			}
		}
		
		for _, rule := range rules {
			if !dimension.HasValue(rule.From) {
				return fmt.Errorf("transition rule has invalid 'from' value '%s' for dimension '%s'", rule.From, dimName)
			}
			for _, to := range rule.To {
				if !dimension.HasValue(to) {
					return fmt.Errorf("transition rule has invalid 'to' value '%s' for dimension '%s'", to, dimName)
				}
			}
		}
	}
	
	// Validate auto-transition rules
	for _, rule := range c.AutoTransitions {
		if rule.TargetDimension != "" && !dimensionNames[rule.TargetDimension] {
			return fmt.Errorf("auto-transition rule references unknown dimension '%s'", rule.TargetDimension)
		}
	}
	
	return nil
}

// GetDimension returns the dimension with the given name, or nil if not found.
func (c WorkflowConfig) GetDimension(name string) *StatusDimension {
	for i, dim := range c.Dimensions {
		if dim.Name == name {
			return &c.Dimensions[i]
		}
	}
	return nil
}

// StatusMetrics provides aggregated information about items and their statuses.
type StatusMetrics struct {
	Total    int                        // Total number of items
	ByStatus map[string]map[string]int  // dimension -> value -> count
	contexts map[string]int             // context -> count (internal)
}

// NewStatusMetrics creates a new StatusMetrics instance.
func NewStatusMetrics() *StatusMetrics {
	return &StatusMetrics{
		ByStatus: make(map[string]map[string]int),
		contexts: make(map[string]int),
	}
}

// AddItem adds an item with the given statuses to the metrics.
func (m *StatusMetrics) AddItem(statuses map[string]string) {
	m.Total++
	for dimension, value := range statuses {
		if m.ByStatus[dimension] == nil {
			m.ByStatus[dimension] = make(map[string]int)
		}
		m.ByStatus[dimension][value]++
	}
}

// GetCount returns the count of items with a specific status value in a dimension.
func (m *StatusMetrics) GetCount(dimension, value string) int {
	if dimMap, exists := m.ByStatus[dimension]; exists {
		return dimMap[value]
	}
	return 0
}

// GetContextCount returns the count of items visible in a specific context.
func (m *StatusMetrics) GetContextCount(context string) int {
	return m.contexts[context]
}

// SetContextCount sets the count of items visible in a specific context.
func (m *StatusMetrics) SetContextCount(context string, count int) {
	m.contexts[context] = count
}