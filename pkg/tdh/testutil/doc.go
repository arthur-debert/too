// Package testutil provides domain-specific test helpers and assertions for tdh tests.
//
// # Testing Strategy
//
// The core testing strategy is to separate business logic from I/O.
//
//   - For command and business logic tests (the common case), use the setup helpers
//     like CreatePopulatedStore() to get a fast, in-memory store.
//
//   - For tests of the storage layer itself, use t.TempDir() to create safe,
//     temporary file-based stores.
//
// Always prefer the custom assertions in this package (e.g., AssertTodoInList)
// over manual checks to make tests clearer and more robust.
//
// This package reduces boilerplate in tests and makes them more expressive by providing
// focused helper functions for common test scenarios.
//
// # Setup Helpers
//
// The package provides helpers to quickly set up test stores with pre-populated data:
//
//	// Create an in-memory store with todos
//	store := testutil.CreatePopulatedStore(t, "Buy milk", "Walk dog", "Write tests")
//
//	// Create a store with specific todo states
//	store := testutil.CreateStoreWithSpecs(t, []testutil.TodoSpec{
//	    {Text: "Buy milk", Status: models.StatusDone},
//	    {Text: "Walk dog", Status: models.StatusPending},
//	})
//
// # Assertions
//
// Domain-specific assertions make tests more readable and provide better error messages:
//
//	// Check if a todo exists in a list
//	testutil.AssertTodoInList(t, todos, "Buy milk")
//
//	// Verify todo counts
//	testutil.AssertTodoCount(t, result, 10, 3) // 10 total, 3 done
//
//	// Check todo status
//	testutil.AssertTodoHasStatus(t, todo, models.StatusDone)
//
// # Test Helpers
//
// Additional helpers for common test operations:
//
//	// Get a temporary directory that's automatically cleaned up
//	dir := testutil.TempDir(t)
//
//	// Create a test collection
//	collection := testutil.NewTestCollection(todos...)
package testutil
