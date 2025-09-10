# Porting `too` to the `idm` Package

This document outlines the process of refactoring `too` to use the new `idm` package for all human-friendly ID (HID) management.

## 1. Porting Process

The migration from the old, model-coupled position logic to the new `idm` registry was performed command by command. The general process for each command was as follows:

1.  **Create the `IDMStoreAdapter`:** A new struct, `store.IDMStoreAdapter`, was created. This struct implements the `idm.StoreAdapter` interface and acts as the bridge between the `idm` registry and `too`'s existing `JSONFileStore`. It knows how to query a `models.Collection` to get the UIDs of children for a given parent UID (a "scope").

2.  **Integrate the Registry:** In each command's `Execute` function, the `idm.Registry` is now initialized and populated. This involves:
    -   Creating a new `store.IDMStoreAdapter`.
    -   Creating a new `idm.Registry`.
    -   Getting all scopes (parent UIDs) from the adapter.
    -   Calling `reg.RebuildScope()` for each scope to populate the registry with the current state of the todo collection.

3.  **Replace Position-Finding Logic:** The key part of the refactoring was replacing all calls to the old `collection.FindItemByPositionPath(path)` with `reg.ResolvePositionPath(store.RootScope, path)`. This new method returns the stable UID of the target todo.

4.  **Update Data Operations:** The command's core logic was updated to use the resolved UID to find and manipulate the todo item via `collection.FindItemByID(uid)`.

5.  **Remove Redundant Code:** The old helper functions for finding items by position (`FindByPosition`, `FindItemByPositionPath`, etc.) and their corresponding tests were removed from the `pkg/too/models` and `pkg/too/internal/helpers` packages.

### Ported Commands

The following commands, which rely on user-provided position paths, have been ported to use the `idm` package:
-   `add` (for the `--to` flag)
-   `edit`
-   `complete`
-   `reopen`
-   `move`
-   `swap`

## 2. Current Difficulties & State

The porting process is functionally complete, but the test suite is currently failing. The primary difficulties are:

1.  **Brittle Test Setups:** Many of the existing tests relied on the implicit, and sometimes incorrect, `Position` values of todos created by test helpers. The new `idm.Registry` is stricter and requires that the HIDs within any given scope are sequential and 1-based (e.g., 1, 2, 3...). The test data does not always conform to this, causing `reg.ResolvePositionPath` to fail with "todo not found" errors where the old logic might have succeeded.

2.  **Reordering Logic:** The fix for the brittle tests is to manually call `collection.Reorder()` within the test setup code to ensure the `Position` fields are correct before the command is executed. This work is tedious and is the source of the remaining test failures.

3.  **Panics in Tests:** Several tests are still panicking due to `nil` pointers. This happens when a test fails to find a todo but continues its execution as if it had succeeded, leading to a crash on the next line. This is a result of the test failures mentioned above and is being progressively fixed by adding checks and `t.FailNow()` calls.

4.  **Short ID Resolution:** The `reopen` command can accept either a position path or a short UID. The `idm` registry does not currently have a mechanism for resolving short UIDs. The refactored `reopen` command contains a temporary workaround that checks if the input *looks like* a position path; if not, it falls back to the old `FindItemByShortID` method. This is a temporary solution, and the `idm` package should eventually be extended to handle this case more gracefully.

The work was paused at the stage of fixing these failing tests. Once all tests are passing, the `feature/port-to-idm` branch will be ready for review.
