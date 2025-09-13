# Formatter Migration Plan

## Current State
- `pkg/too/output/formatters/` contains ~1200 LOC
- Most formatters (JSON, YAML) are just calling encode()
- CSV and Markdown have too-specific logic (hierarchical todos, checkboxes)
- Each formatter implements the too-specific Formatter interface

## Migration Steps

### 1. Update commands to use lipbalm engine directly
Instead of:
```go
formatter := output.Get(formatFlag)
formatter.RenderChange(os.Stdout, result)
```

Use:
```go
engine := output.GetGlobalEngine()
engine.Render(os.Stdout, formatFlag, result)
```

### 2. Move too-specific rendering logic to engine callbacks
- CSV: Hierarchical todo flattening
- Markdown: Todo checkboxes, hierarchical rendering

### 3. Delete generic formatters
Remove these packages entirely:
- `pkg/too/output/formatters/json/`
- `pkg/too/output/formatters/yaml/`
- `pkg/too/output/formatters/term/`

### 4. Keep only domain-specific code
Either:
- Move CSV/Markdown too-specific logic to engine.go
- Or create a minimal formatter for truly custom formats

### 5. Remove the old Formatter interface
- Delete `pkg/too/output/formatter.go`
- Delete `pkg/too/output/registry.go`
- Update `cmd/too/register_formatters.go`

## End Result
- All generic formatting handled by lipbalm
- Too only contains domain-specific rendering logic
- ~1000+ LOC removed from too
- Cleaner separation of concerns