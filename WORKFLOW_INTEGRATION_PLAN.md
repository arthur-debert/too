# Workflow Integration Plan for Too

This document outlines the plan for integrating the new IDM workflow layer into the `too` application, enabling multi-dimensional status management and advanced workflow features.

## Overview

The integration will extend `too` from a simple pending/done task manager to a comprehensive workflow-enabled system supporting:
- Multiple status dimensions (completion, priority, urgency, etc.)
- Context-aware filtering and display
- Automatic transitions and workflow rules
- Preset workflow configurations for different use cases

## Current Architecture Analysis

### Existing Components
- **Models**: `pkg/too/models/models.go` - Simple Todo with Status (pending/done)
- **Store**: `pkg/too/store/` - IDM adapter for persistence and hierarchy
- **Commands**: `pkg/too/commands/` - Business logic for operations
- **CLI**: `cmd/too/` - Command-line interface

### Current Status Management
- Binary status: `StatusPending` and `StatusDone`
- Status transitions in `complete.go` and `reopen.go`
- Bottom-up completion logic for hierarchical completion

## Integration Strategy

### Phase 1: Foundation Layer
1. **Workflow Store Adapter**: Create adapter between workflow system and existing store
2. **Configuration Management**: Add workflow config loading and validation
3. **Backward Compatibility**: Ensure existing functionality continues to work

### Phase 2: Enhanced Commands
1. **Extended Complete/Reopen**: Support multiple status dimensions
2. **New Status Commands**: Add commands for managing different dimensions
3. **Context-Aware Listing**: Filter todos based on workflow contexts

### Phase 3: Advanced Features
1. **Priority Management**: Add priority dimension support
2. **Workflow Presets**: Enable switching between different workflow configurations
3. **Auto-transitions**: Implement workflow automation

## Implementation Plan

### 1. Workflow Store Integration

#### Create WorkflowTodoAdapter
```go
// pkg/too/store/workflow_adapter.go
type WorkflowTodoAdapter struct {
    store Store
    collection *models.Collection
}

// Implement workflow.WorkflowStoreAdapter interface
func (a *WorkflowTodoAdapter) SetItemStatus(uid, dimension, value string) error
func (a *WorkflowTodoAdapter) GetItemStatus(uid, dimension string) (string, error)
// ... other workflow interface methods
```

#### Extend Todo Model
```go
// pkg/too/models/models.go
type Todo struct {
    ID       string                 `json:"id"`
    ParentID string                 `json:"parentId"`
    Position int                    `json:"position"`
    Text     string                 `json:"text"`
    Status   TodoStatus             `json:"status"`           // Legacy field
    Statuses map[string]string      `json:"statuses"`         // New multi-dimensional status
    Modified time.Time             `json:"modified"`
    Items    []*Todo               `json:"items"`
}
```

### 2. Configuration System

#### Add Workflow Configuration
```go
// pkg/too/config/workflow.go
type WorkflowConfig struct {
    Enabled      bool                  `json:"enabled"`
    PresetName   string               `json:"preset"`      // "todo", "priority", "kanban", etc.
    CustomConfig *workflow.WorkflowConfig `json:"custom"`  // Custom workflow if not using preset
}

func LoadWorkflowConfig(configPath string) (*WorkflowConfig, error)
func (c *WorkflowConfig) GetWorkflowConfig() workflow.WorkflowConfig
```

### 3. Enhanced Commands

#### Extend Complete Command
```go
// pkg/too/commands/complete/complete.go
type Options struct {
    CollectionPath string
    Mode           string
    Dimension      string  // New: specify which dimension to complete
    Value          string  // New: specify target value (default: "done")
    WorkflowMode   bool    // New: enable workflow features
}

func Execute(positionPath string, opts Options) (*Result, error) {
    // Use workflow manager if enabled, otherwise fall back to legacy behavior
}
```

#### Add New Status Command
```go
// pkg/too/commands/status/status.go
type Options struct {
    CollectionPath string
    Dimension      string  // Which dimension to modify
    Value          string  // Target value
    WorkflowMode   bool
}

type Result struct {
    Todo        *models.Todo
    Dimension   string
    OldValue    string
    NewValue    string
    Transitions []AutoTransitionResult // Any triggered auto-transitions
}

func Execute(positionPath string, opts Options) (*Result, error)
```

#### Add Priority Command
```go
// pkg/too/commands/priority/priority.go
type Options struct {
    CollectionPath string
    Priority       string  // "low", "medium", "high"
}

func Execute(positionPath string, opts Options) (*Result, error)
```

### 4. Context-Aware Listing

#### Extend List Command
```go
// pkg/too/commands/list/list.go
type Options struct {
    CollectionPath string
    Context        string  // New: "active", "all", "high_priority", etc.
    ShowDone       bool    // Legacy compatibility
    WorkflowMode   bool    // Enable workflow features
}

func Execute(opts Options) (*Result, error) {
    // Use workflow context filtering if enabled
}
```

### 5. CLI Integration

#### Add Workflow Commands
```bash
# New commands
too status 1.2 --dimension priority --value high
too priority 1.2 high
too workflow-config --preset todo-priority
too contexts list
too contexts show active

# Enhanced existing commands
too complete 1.2 --dimension completion --value done
too list --context active
too list --context high_priority
```

#### Configuration Commands
```bash
too init --workflow todo          # Initialize with basic todo workflow
too init --workflow priority      # Initialize with priority workflow
too init --workflow kanban        # Initialize with kanban workflow
```

## Migration Strategy

### Backward Compatibility
1. **Legacy Status Field**: Maintain existing `Status` field for backward compatibility
2. **Default Behavior**: Non-workflow mode behaves exactly as before
3. **Gradual Migration**: Users can opt-in to workflow features

### Data Migration
```go
// pkg/too/store/migration.go
func MigrateToWorkflow(collection *models.Collection) error {
    for _, todo := range collection.AllTodos() {
        if todo.Statuses == nil {
            todo.Statuses = make(map[string]string)
            // Migrate legacy status to completion dimension
            todo.Statuses["completion"] = string(todo.Status)
        }
    }
    return nil
}
```

## Testing Strategy

### Unit Tests
- Test workflow adapter implementation
- Test configuration loading and validation
- Test command execution with workflow features
- Test backward compatibility scenarios

### Integration Tests
- Test complete workflow scenarios (todo creation → status changes → completion)
- Test auto-transitions and workflow automation
- Test context filtering and visibility rules
- Test migration from legacy to workflow format

### CLI Tests
- Test new commands and flags
- Test workflow configuration commands
- Test preset workflow switching

## Benefits

### For Users
1. **Enhanced Productivity**: Priority management, context filtering
2. **Workflow Automation**: Auto-transitions, bottom-up completion
3. **Flexibility**: Multiple preset workflows or custom configurations
4. **Backward Compatibility**: Existing workflows continue unchanged

### For Developers
1. **Extensibility**: Easy to add new dimensions and workflows
2. **Type Safety**: Validated configurations and transitions
3. **Clean Architecture**: Separation between workflow logic and business logic
4. **Testing**: Comprehensive test coverage for workflow features

## Rollout Plan

### Phase 1: Foundation (Week 1-2)
- Implement workflow store adapter
- Add configuration system
- Ensure backward compatibility
- Basic unit tests

### Phase 2: Core Features (Week 3-4)
- Extend complete/reopen commands
- Add status and priority commands
- Implement context-aware listing
- Integration tests

### Phase 3: Advanced Features (Week 5-6)
- Add workflow preset switching
- Implement auto-transitions
- Add workflow configuration CLI
- Full CLI integration

### Phase 4: Polish (Week 7)
- Documentation updates
- Performance optimization
- User experience improvements
- Release preparation

## Success Criteria

1. **Backward Compatibility**: All existing functionality works unchanged
2. **Feature Completeness**: All planned workflow features implemented
3. **Performance**: No significant performance regression
4. **Usability**: Intuitive CLI for workflow features
5. **Documentation**: Complete user and developer documentation
6. **Testing**: >95% test coverage for new functionality

This integration will transform `too` from a simple task manager into a powerful, workflow-enabled productivity tool while maintaining its simplicity and ease of use.