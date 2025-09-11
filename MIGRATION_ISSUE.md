# Complete IDM Migration: Remove Adapters and Legacy Patterns

## Overview

The too codebase is currently in a transitional state with ~40% of the IDM migration complete. While status management has been successfully migrated to the workflow system, significant technical debt remains from maintaining both legacy and IDM patterns in parallel. This has resulted in increased code complexity (+600 lines) instead of the expected reduction.

## Problem

The current architecture maintains multiple adapter layers and legacy patterns:
- Data flows through: `models.Todo` → `IDMStoreAdapter` → `WorkflowTodoAdapter` → `WorkflowManager`
- Dual position management (models + IDM)
- Legacy collection-based operations
- Store interface impedance mismatch with IDM

## Goal

Achieve a pure IDM architecture with:
- Direct IDM operations without adapters
- Single source of truth for all data (positions, status, hierarchy)
- Reduced code complexity and line count
- No backwards compatibility or legacy patterns

## Implementation Phases

### Phase 1: Eliminate Adapter Layers ⚡ **[HIGH IMPACT - ~400 lines reduction]**

**Goal**: Remove intermediate adapter layers and make commands work directly with WorkflowManager

- [ ] Remove `IDMStoreAdapter` (`pkg/too/store/idm_adapter.go`)
- [ ] Remove `WorkflowTodoAdapter` (`pkg/too/store/workflow_adapter.go`)
- [ ] Update `WorkflowManager` to work directly with collections
- [ ] Update all commands to use `WorkflowManager` directly without adapter creation
- [ ] Remove adapter interfaces from `pkg/idm/adapter.go`
- [ ] Update tests to work without adapters

**Success Criteria**: Commands operate directly on WorkflowManager without any adapter layers

### Phase 2: Remove Position Field 🎯 **[HIGH IMPACT - ~200 lines reduction]**

**Goal**: Delegate all position management to IDM

- [ ] Remove `Position` field from `models.Todo`
- [ ] Remove position management methods from models:
  - [ ] `ResetActivePositions()`
  - [ ] `ResetRootPositions()` 
  - [ ] `ResetSiblingPositions()`
  - [ ] `ReorderTodos()`
  - [ ] `findHighestPosition()`
- [ ] Update commands to use IDM for position queries
- [ ] Remove manual position resets from commands
- [ ] Update tests for IDM-based positioning

**Success Criteria**: All position management flows through IDM Registry

### Phase 3: Replace Collection Operations 📦 **[MEDIUM IMPACT - ~150 lines reduction]**

**Goal**: Use IDM queries instead of collection filtering

- [ ] Remove filtering methods from models:
  - [ ] `ListActive()`
  - [ ] `ListArchived()`
  - [ ] `ListAll()`
- [ ] Replace with IDM visibility context queries
- [ ] Update commands to use IDM queries directly
- [ ] Remove collection-based filtering helpers
- [ ] Update display logic to work with IDM results

**Success Criteria**: All filtering operations use IDM visibility contexts

### Phase 4: Pure IDM Data Model 🏗️ **[HIGH IMPACT - ~300 lines reduction]**

**Goal**: Replace models.Todo with IDM-native structures

- [ ] Create IDM-native todo representation (UID, text, dimensions only)
- [ ] Remove `models.Todo` struct
- [ ] Remove `models.Collection` struct
- [ ] Update all commands to work with IDM data directly
- [ ] Replace `Todo.Items` with IDM parent-child relationships
- [ ] Update persistence to serialize IDM state

**Success Criteria**: No legacy data models remain

### Phase 5: Direct IDM Store 💾 **[MEDIUM IMPACT - ~100 lines reduction]**

**Goal**: Replace Store interface with IDM persistence

- [ ] Remove Store interface
- [ ] Create IDM-native persistence layer
- [ ] Remove transaction pattern `Store.Update()`
- [ ] Commands receive Manager instance instead of Store
- [ ] Direct IDM state serialization

**Success Criteria**: Persistence layer is IDM-native

## Expected Outcome

- **Code reduction**: ~1,150 lines removed
- **Architecture**: Pure IDM without adapters or legacy patterns
- **Performance**: Direct operations without multiple abstraction layers
- **Maintainability**: Single source of truth for all operations

## Validation

After each phase:
1. All tests must pass
2. Line count should decrease
3. No new adapters or backwards compatibility code

## Priority

Start with Phase 1 (Eliminate Adapters) as it:
- Has the highest immediate impact (~400 lines)
- Proves the migration approach works
- Simplifies subsequent phases