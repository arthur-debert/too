# Fix Notes for ResetActivePositions

## What Was Fixed
The core issue was that `ResetActivePositions` was only updating Position fields but not reordering the actual slice. This caused the integration test to fail because after reopening a todo, it would get position 3 but still appear in the middle of the list instead of at the end.

The fix changes `ResetActivePositions` to:
1. Take a pointer to the slice (`*[]*Todo` instead of `[]*Todo`)
2. Actually reorder the slice to put active items first (in order), then done items
3. This ensures the slice order matches the position numbers

## What Still Needs Fixing
Many tests are failing because they expect items to be in specific positions in the slice. With the new behavior:
- Active items come first in the slice (sorted by position)
- Done items come after all active items
- This affects any test that accesses items by index (e.g., `todos[0]`, `parent.Items[1]`)

## Tests That Need Updates
1. `pkg/tdh/models/status_test.go` - Several tests checking slice positions
2. `pkg/tdh/commands/complete/complete_bottom_up_test.go` - Tests accessing items by index
3. `pkg/tdh/internal/helpers` tests - May have similar issues

## The Fix Pattern
When updating tests, instead of expecting items at specific indices, either:
1. Update the index expectations to match the new order
2. Find items by ID or text instead of index
3. Use helper functions that search for items

Example:
```go
// Old way (assumes order):
assert.Equal(t, "Item 1", todos[0].Text)

// New way (finds by property):
var item1 *models.Todo
for _, todo := range todos {
    if todo.Text == "Item 1" {
        item1 = todo
        break
    }
}
assert.NotNil(t, item1)
```