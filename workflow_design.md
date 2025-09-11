# IDM Workflow Layer Design

## Core Concepts

### 1. Status Dimensions
Instead of a single "status" field, support multiple independent status axes:

```go
type StatusDimension struct {
    Name   string   // e.g., "completion", "priority", "visibility"
    Values []string // e.g., ["pending", "done"], ["low", "high"], ["public", "private"]
}

// Examples:
var CompletionDimension = StatusDimension{
    Name:   "completion",
    Values: []string{"pending", "in-progress", "done", "archived"},
}

var PriorityDimension = StatusDimension{
    Name:   "priority", 
    Values: []string{"low", "normal", "high", "urgent"},
}
```

### 2. Visibility Rules
Define which status values are visible in which contexts:

```go
type VisibilityRule struct {
    Context   string            // e.g., "active_scope", "all_scope", "archive_scope"
    Dimension string            // e.g., "completion"
    Include   []string          // Status values to include
    Exclude   []string          // Status values to exclude
}

// Examples for too:
var TooVisibilityRules = []VisibilityRule{
    {
        Context:   "position_paths",  // Items that get HID positions
        Dimension: "completion", 
        Include:   []string{"pending", "in-progress"},
    },
    {
        Context:   "hierarchy_display", // Items shown in list --all
        Dimension: "completion",
        Include:   []string{"pending", "in-progress", "done"},
    },
}

// Examples for CMS:
var CMSVisibilityRules = []VisibilityRule{
    {
        Context:   "public_site",
        Dimension: "publication",
        Include:   []string{"published"},
    },
    {
        Context:   "admin_panel", 
        Dimension: "publication",
        Include:   []string{"draft", "published", "archived"},
    },
}
```

### 3. Transition Rules
Define valid status changes:

```go
type TransitionRule struct {
    Dimension string
    From      string
    To        []string  // Valid next states
}

// Examples:
var TooTransitionRules = []TransitionRule{
    {Dimension: "completion", From: "pending", To: []string{"in-progress", "done"}},
    {Dimension: "completion", From: "in-progress", To: []string{"pending", "done"}},
    {Dimension: "completion", From: "done", To: []string{"pending", "archived"}},
}
```

## API Design

### 1. StatusManager
```go
type StatusManager struct {
    registry    *Registry
    adapter     WorkflowStoreAdapter
    dimensions  map[string]StatusDimension
    visibility  map[string][]VisibilityRule  // context -> rules
    transitions map[string][]TransitionRule  // dimension -> rules
}

func NewStatusManager(
    registry *Registry,
    adapter WorkflowStoreAdapter,
    config WorkflowConfig,
) *StatusManager

// Status operations
func (sm *StatusManager) SetStatus(uid, dimension, value string) error
func (sm *StatusManager) GetStatus(uid, dimension string) (string, error)
func (sm *StatusManager) GetAllStatuses(uid string) (map[string]string, error)

// Transition validation
func (sm *StatusManager) CanTransition(uid, dimension, newValue string) error
func (sm *StatusManager) Transition(uid, dimension, newValue string) error

// Context-aware queries
func (sm *StatusManager) GetChildrenInContext(parentUID, context string) ([]string, error)
func (sm *StatusManager) IsVisibleInContext(uid, context string) (bool, error)
```

### 2. Enhanced Adapter Interface
```go
type WorkflowStoreAdapter interface {
    StoreAdapter  // Inherit basic methods
    
    // Status management
    SetItemStatus(uid, dimension, value string) error
    GetItemStatus(uid, dimension string) (string, error)
    GetItemStatuses(uid string) (map[string]string, error)
    
    // Context-aware queries
    GetChildrenInContext(parentUID, context string) ([]string, error)
}
```

### 3. Workflow Configuration
```go
type WorkflowConfig struct {
    Dimensions  []StatusDimension
    Visibility  map[string][]VisibilityRule  // context -> rules
    Transitions map[string][]TransitionRule  // dimension -> rules
}

// Example too configuration:
var TooWorkflowConfig = WorkflowConfig{
    Dimensions: []StatusDimension{
        {Name: "completion", Values: []string{"pending", "done"}},
    },
    Visibility: map[string][]VisibilityRule{
        "position_paths": {{
            Context: "position_paths", Dimension: "completion", 
            Include: []string{"pending"},
        }},
        "hierarchy_display": {{
            Context: "hierarchy_display", Dimension: "completion",
            Include: []string{"pending", "done"},
        }},
    },
    Transitions: map[string][]TransitionRule{
        "completion": {
            {Dimension: "completion", From: "pending", To: []string{"done"}},
            {Dimension: "completion", From: "done", To: []string{"pending"}},
        },
    },
}
```

## Usage Examples

### In too commands:
```go
// Complete command:
func Execute(positionPath string, opts Options) (*Result, error) {
    // Resolve using position_paths context (only pending items)
    uid, err := registry.ResolvePositionPathInContext("root", positionPath, "position_paths")
    
    // Transition status
    err = statusManager.Transition(uid, "completion", "done")
    
    // The item is now:
    // - Excluded from position_paths context (no HID)
    // - Still included in hierarchy_display context (visible in list --all)
}

// List command:
func Execute(opts Options) (*Result, error) {
    var context string
    if opts.ShowAll {
        context = "hierarchy_display"  // Shows pending + done
    } else {
        context = "position_paths"     // Shows only pending
    }
    
    children, err := statusManager.GetChildrenInContext("root", context)
}
```

### For other applications:
```go
// CMS Article Manager:
var CMSConfig = WorkflowConfig{
    Dimensions: []StatusDimension{
        {Name: "publication", Values: []string{"draft", "review", "published", "archived"}},
        {Name: "priority", Values: []string{"low", "normal", "high"}},
    },
    Visibility: map[string][]VisibilityRule{
        "public_site": {{
            Context: "public_site", Dimension: "publication",
            Include: []string{"published"},
        }},
        "editor_view": {{
            Context: "editor_view", Dimension: "publication", 
            Include: []string{"draft", "review", "published"},
        }},
    },
}

// Issue Tracker:
var IssueTrackerConfig = WorkflowConfig{
    Dimensions: []StatusDimension{
        {Name: "state", Values: []string{"open", "in-progress", "resolved", "closed"}},
        {Name: "priority", Values: []string{"low", "medium", "high", "critical"}},
    },
    Visibility: map[string][]VisibilityRule{
        "active_board": {{
            Context: "active_board", Dimension: "state",
            Include: []string{"open", "in-progress"},
        }},
        "all_issues": {{
            Context: "all_issues", Dimension: "state",
            Include: []string{"open", "in-progress", "resolved", "closed"},
        }},
    },
}
```

## Benefits for too

1. **Complete offloading**: too no longer manages status logic
2. **Context-aware HIDs**: Position paths only include active items
3. **Flexible visibility**: Different views (active vs all) handled by IDM
4. **Validation**: Status transitions validated by IDM
5. **Future extensibility**: Easy to add priority, labels, etc.

## Implementation Strategy

1. **Phase 1**: Add workflow layer to IDM package
2. **Phase 2**: Update too's adapter to support workflow interface  
3. **Phase 3**: Replace too's status management with StatusManager calls
4. **Phase 4**: Remove todo.MarkComplete/MarkPending methods

This would truly complete the IDM adoption while making IDM genuinely reusable for other workflow scenarios.