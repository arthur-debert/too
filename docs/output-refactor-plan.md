# Output Package Refactoring Plan

## Current State
- ~1100 LOC in output package (excluding tests)
- Multiple layers of abstraction (Renderer -> Engine -> lipbalm)
- Duplicate message/style definitions
- Too much generic code that should be in lipbalm

## What Should Move to lipbalm

### 1. Generic Message Type
- The `Message` struct with Text and Level fields
- Standard message levels: success, error, warning, info
- This is a common pattern across CLI apps

### 2. Generic Styles
Most styles in lipbalm_styles.go are generic:
- success, error, warning, info (standard semantic colors)
- muted, highlighted, subdued, accent (common UI states)
- label, value, count (common data display patterns)

Only keep todo-specific styles in too:
- todo-done, todo-pending
- Custom symbols for todo states

### 3. Standard Template Functions
Move to lipbalm's default template functions:
- indent(level) - common for hierarchical display
- lines(text) - splitting multi-line text
- Basic string manipulation

Keep in too:
- isDone() - checks todo completion status
- buildHierarchy() - builds todo tree structure
- getSymbol() - gets todo-specific symbols

## Simplified Architecture

### 1. Direct lipbalm Usage
Instead of:
```
Renderer -> Engine -> lipbalm.RenderEngine
```

Use:
```
Renderer -> lipbalm.RenderEngine (configured with too-specific callbacks)
```

### 2. Configuration-Based Approach
```go
// In init.go or a setup function
lipbalm.Configure(&lipbalm.Config{
    Styles: mergeStyles(lipbalm.DefaultStyles(), tooSpecificStyles()),
    Templates: templateFS,
    Callbacks: tooCallbacks(),
})
```

### 3. Result Types with Self-Describing Behavior
Add methods to result types:
- MessageType() string - already added to ChangeResult
- ShouldHighlight() []string - returns UIDs to highlight
- GetTemplate() string - returns template name for this result type

## Files to Keep (Simplified)

1. **output.go** - Thin compatibility layer for existing code
2. **hierarchy.go** - Todo-specific hierarchical display logic
3. **styles/symbols.go** - Todo-specific symbols
4. **templates/*.tmpl** - Todo-specific templates
5. **init.go** - Configure lipbalm with too-specific settings

## Files to Remove/Merge

1. **engine.go** - Merge useful parts into init.go, remove wrapper
2. **renderer.go** - Merge with output.go
3. **message.go** - Move to lipbalm
4. **todo_list_with_message.go** - Inline into preprocessing callback
5. **styles/colors.go** - Use lipbalm's color system
6. **styles/styles.go** - Merge todo-specific parts into init.go
7. **styles/lipbalm_styles.go** - Split between lipbalm defaults and too-specific

## Implementation Steps

1. Add generic Message type to lipbalm
2. Add default styles to lipbalm (success, error, warning, etc.)
3. Update too to use lipbalm's Message type
4. Simplify Engine to just configuration and callbacks
5. Merge renderer files and remove duplication
6. Move generic template functions to lipbalm
7. Update all imports and tests

## Expected Outcome
- Reduce output package from ~1100 LOC to ~300-400 LOC
- Only todo-specific logic remains in too
- Cleaner separation of concerns
- lipbalm becomes more useful for other CLI apps