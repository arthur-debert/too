# IDM Workflow Layer Implementation Plan

## Phase 1: Core Workflow Layer in IDM

### New IDM Package Structure
```
pkg/idm/
├── registry.go          # Existing - core UID/HID mapping
├── manager.go           # Existing - hierarchy operations  
├── adapter.go           # Existing - basic interfaces
├── workflow/
│   ├── types.go         # Status dimensions, rules, config
│   ├── manager.go       # StatusManager implementation
│   ├── transitions.go   # Status transition logic
│   ├── visibility.go    # Context-aware filtering
│   ├── adapter.go       # WorkflowStoreAdapter interface
│   └── presets.go       # Common workflow configs (todo, cms, etc.)
```

### 1. Core Types (`workflow/types.go`)
```go
package workflow

type StatusDimension struct {
    Name         string
    Values       []string
    DefaultValue string
}

type VisibilityRule struct {
    Context   string
    Dimension string  
    Include   []string
    Exclude   []string
}

type TransitionRule struct {
    Dimension string
    From      string
    To        []string
    Validator func(uid string, adapter WorkflowStoreAdapter) error
}

type AutoTransitionRule struct {
    Trigger        string // "child_status_change", "time_based", etc.
    Condition      string // "all_children_status_equals", etc.
    ConditionValue string
    Action         string // "set_status", "move_item", etc.
    ActionValue    string
}

type WorkflowConfig struct {
    Dimensions      []StatusDimension
    Visibility      map[string][]VisibilityRule
    Transitions     map[string][]TransitionRule
    AutoTransitions []AutoTransitionRule
}
```

### 2. Workflow Manager (`workflow/manager.go`)
```go
package workflow

type StatusManager struct {
    registry        *idm.Registry
    hierarchyMgr    *idm.Manager
    adapter         WorkflowStoreAdapter
    config          WorkflowConfig
    visibilityCache map[string]map[string]bool // context -> uid -> visible
}

func NewStatusManager(
    registry *idm.Registry,
    hierarchyMgr *idm.Manager, 
    adapter WorkflowStoreAdapter,
    config WorkflowConfig,
) *StatusManager

// Core status operations
func (sm *StatusManager) GetStatus(uid, dimension string) (string, error)
func (sm *StatusManager) SetStatus(uid, dimension, value string) error
func (sm *StatusManager) GetAllStatuses(uid string) (map[string]string, error)

// Validated transitions
func (sm *StatusManager) CanTransition(uid, dimension, newValue string) error
func (sm *StatusManager) Transition(uid, dimension, newValue string) error

// Context-aware queries
func (sm *StatusManager) GetChildrenInContext(parentUID, context string) ([]string, error)
func (sm *StatusManager) ResolvePositionPathInContext(startScope, path, context string) (string, error)
func (sm *StatusManager) GetPositionPathInContext(startScope, uid, context string) (string, error)
func (sm *StatusManager) IsVisibleInContext(uid, context string) (bool, error)

// Auto-transitions (e.g., bottom-up completion)
func (sm *StatusManager) TriggerAutoTransitions(triggerType, targetUID string) error

// Metrics and aggregation
func (sm *StatusManager) GetMetrics(scope string) StatusMetrics

type StatusMetrics struct {
    Total     int
    ByStatus  map[string]map[string]int // dimension -> value -> count
    // Convenience accessors for common patterns
    Active    int // Items visible in "active" context
    Done      int // Items with completion="done"
}
```

### 3. Enhanced Adapter (`workflow/adapter.go`)
```go
package workflow

type WorkflowStoreAdapter interface {
    idm.ManagedStoreAdapter // Inherit hierarchy operations
    
    // Status management
    SetItemStatus(uid, dimension, value string) error
    GetItemStatus(uid, dimension string) (string, error)
    GetItemStatuses(uid string) (map[string]string, error)
    
    // Context-aware queries  
    GetChildrenInContext(parentUID, context string) ([]string, error)
    
    // Bulk operations for performance
    SetMultipleStatuses(updates map[string]map[string]string) error
    GetStatusesBulk(uids []string) (map[string]map[string]string, error)
}
```

### 4. Preset Configurations (`workflow/presets.go`)
```go
package workflow

// Todo app workflow
var TodoWorkflow = WorkflowConfig{
    Dimensions: []StatusDimension{
        {Name: "completion", Values: []string{"pending", "done"}, DefaultValue: "pending"},
    },
    Visibility: map[string][]VisibilityRule{
        "active": {{Context: "active", Dimension: "completion", Include: []string{"pending"}}},
        "all":    {{Context: "all", Dimension: "completion", Include: []string{"pending", "done"}}},
    },
    Transitions: map[string][]TransitionRule{
        "completion": {
            {Dimension: "completion", From: "pending", To: []string{"done"}},
            {Dimension: "completion", From: "done", To: []string{"pending"}},
        },
    },
    AutoTransitions: []AutoTransitionRule{
        {
            Trigger:        "child_status_change",
            Condition:      "all_children_status_equals", 
            ConditionValue: "done",
            Action:         "set_status",
            ActionValue:    "completion:done",
        },
    },
}

// CMS workflow
var CMSWorkflow = WorkflowConfig{
    Dimensions: []StatusDimension{
        {Name: "publication", Values: []string{"draft", "review", "published", "archived"}, DefaultValue: "draft"},
        {Name: "priority", Values: []string{"low", "normal", "high"}, DefaultValue: "normal"},
    },
    Visibility: map[string][]VisibilityRule{
        "public":     {{Context: "public", Dimension: "publication", Include: []string{"published"}}},
        "editorial":  {{Context: "editorial", Dimension: "publication", Include: []string{"draft", "review", "published"}}},
        "admin":      {{Context: "admin", Dimension: "publication", Include: []string{"draft", "review", "published", "archived"}}},
    },
    // ... transitions
}

// Issue tracker workflow  
var IssueTrackerWorkflow = WorkflowConfig{
    Dimensions: []StatusDimension{
        {Name: "state", Values: []string{"open", "in-progress", "resolved", "closed"}, DefaultValue: "open"},
        {Name: "priority", Values: []string{"low", "medium", "high", "critical"}, DefaultValue: "medium"},
    },
    // ... visibility rules and transitions
}
```

## Phase 2: Update too's Adapter

### Enhanced IDM Adapter (`pkg/too/store/workflow_adapter.go`)
```go
package store

type WorkflowIDMAdapter struct {
    *IDMStoreAdapter // Embed existing adapter
}

func NewWorkflowAdapter(collection *models.Collection) *WorkflowIDMAdapter {
    baseAdapter := &IDMStoreAdapter{collection: collection}
    return &WorkflowIDMAdapter{IDMStoreAdapter: baseAdapter}
}

// Implement workflow-specific methods
func (a *WorkflowIDMAdapter) SetItemStatus(uid, dimension, value string) error {
    todo := a.collection.FindItemByID(uid)
    if todo == nil {
        return fmt.Errorf("todo %s not found", uid)
    }
    
    switch dimension {
    case "completion":
        switch value {
        case "pending":
            todo.Status = models.StatusPending
        case "done": 
            todo.Status = models.StatusDone
        default:
            return fmt.Errorf("unknown completion value: %s", value)
        }
    default:
        return fmt.Errorf("unknown dimension: %s", dimension)
    }
    
    todo.Modified = time.Now()
    return nil
}

func (a *WorkflowIDMAdapter) GetItemStatus(uid, dimension string) (string, error) {
    todo := a.collection.FindItemByID(uid)
    if todo == nil {
        return "", fmt.Errorf("todo %s not found", uid)
    }
    
    switch dimension {
    case "completion":
        switch todo.Status {
        case models.StatusPending:
            return "pending", nil
        case models.StatusDone:
            return "done", nil
        default:
            return "", fmt.Errorf("unknown status: %s", todo.Status)
        }
    default:
        return "", fmt.Errorf("unknown dimension: %s", dimension)
    }
}

func (a *WorkflowIDMAdapter) GetChildrenInContext(parentUID, context string) ([]string, error) {
    // Get all children first
    allChildren, err := a.GetChildren(parentUID)
    if err != nil {
        return nil, err
    }
    
    // Filter based on context
    var filtered []string
    for _, uid := range allChildren {
        todo := a.collection.FindItemByID(uid)
        if todo == nil {
            continue
        }
        
        // Apply context-specific filtering
        switch context {
        case "active":
            if todo.Status == models.StatusPending {
                filtered = append(filtered, uid)
            }
        case "all":
            filtered = append(filtered, uid) // Include all
        }
    }
    
    return filtered, nil
}
```

### Convenience Factory (`pkg/too/store/workflow_factory.go`)
```go
package store

func NewWorkflowManager(collection *models.Collection) *workflow.StatusManager {
    adapter := NewWorkflowAdapter(collection)
    registry := idm.NewRegistry()
    
    // Build registry from current collection state
    scopes, _ := adapter.GetScopes()
    for _, scope := range scopes {
        registry.RebuildScope(adapter, scope)
    }
    
    hierarchyMgr := &idm.Manager{} // Initialize with registry and adapter
    
    return workflow.NewStatusManager(
        registry,
        hierarchyMgr,
        adapter,
        workflow.TodoWorkflow, // Use preset config
    )
}
```

## Phase 3: Update too Commands

### Example: Complete Command Refactor
```go
// pkg/too/commands/complete/complete.go
func Execute(positionPath string, opts Options) (*Result, error) {
    s := store.NewStore(opts.CollectionPath)
    var result *Result

    err := s.Update(func(collection *models.Collection) error {
        workflow := store.NewWorkflowManager(collection)
        
        // Resolve in active context (only pending items have positions)
        uid, err := workflow.ResolvePositionPathInContext("root", positionPath, "active")
        if err != nil {
            return fmt.Errorf("todo not found: %w", err)
        }

        // Get current status for result
        oldStatus, err := workflow.GetStatus(uid, "completion")
        if err != nil {
            return err
        }
        
        // Transition status (handles auto-transitions like bottom-up completion)
        err = workflow.Transition(uid, "completion", "done")
        if err != nil {
            return err
        }

        // Build result
        todo := collection.FindItemByID(uid)
        result = &Result{
            Todo:      todo,
            OldStatus: oldStatus,
            NewStatus: "done",
            Mode:      opts.Mode,
        }
        
        // Add metrics for long mode
        if opts.Mode == "long" {
            metrics := workflow.GetMetrics("root")
            result.TotalCount = metrics.Total
            result.DoneCount = metrics.Done
            result.AllTodos, _ = workflow.GetTodosInContext("root", "all")
        }
        
        return nil
    })

    return result, err
}
```

### Example: List Command Refactor  
```go
// pkg/too/commands/list/list.go
func Execute(opts Options) (*Result, error) {
    s := store.NewStore(opts.CollectionPath)
    collection, err := s.Load()
    if err != nil {
        return nil, err
    }

    workflow := store.NewWorkflowManager(collection)
    
    // Determine context from options
    context := "active"
    if opts.ShowDone {
        context = "all"
    }
    
    // Get todos (filtering and positioning handled by workflow)
    todos, err := workflow.GetTodosInContext("root", context)
    if err != nil {
        return nil, err
    }
    
    // Convert UIDs to Todo objects
    todoObjects := make([]*models.Todo, len(todos))
    for i, uid := range todos {
        todoObjects[i] = collection.FindItemByID(uid)
    }
    
    // Get metrics
    metrics := workflow.GetMetrics("root")
    
    return &Result{
        Todos:      todoObjects,
        TotalCount: metrics.Total,
        DoneCount:  metrics.Done,
        Mode:       opts.Mode,
    }, nil
}
```

## Phase 4: Remove Legacy Code

### Files to Remove/Simplify:
- `pkg/too/models/models.go`: Remove `MarkComplete`, `MarkPending`, position reset logic
- `pkg/too/commands/complete/complete.go`: Remove bottom-up completion logic  
- All commands: Remove manual position management
- All commands: Remove manual filtering logic

### Estimated Impact:
- **300+ lines removed** from too codebase
- **Complex logic centralized** in reusable workflow layer
- **Status management** completely offloaded to IDM
- **Position management** automatically handled
- **Context switching** declaratively configured

## Benefits Beyond Line Reduction

### 1. Maintainability
- Status logic in one place (IDM workflow)
- Declarative configuration vs. imperative code
- Consistent behavior across all commands

### 2. Extensibility  
- Add priority dimension: one config change
- Add labels/tags: add dimension + visibility rules
- Add custom transitions: update transition rules

### 3. Reusability
- Workflow layer usable for CMS, issue trackers, etc.
- Preset configurations for common patterns
- Generic enough for any hierarchical + status application

### 4. Performance
- Context-aware caching in workflow layer
- Bulk operations support
- Registry optimizations apply to status queries

### 5. Correctness
- Status transitions validated automatically
- Auto-transitions (bottom-up completion) reliable
- Context isolation prevents logic errors

This workflow layer would transform IDM from a "position path utility" into a **comprehensive hierarchical workflow engine** while massively simplifying too's codebase.