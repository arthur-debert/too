package tdh_test

import (
	"testing"

	"github.com/arthur-debert/tdh/pkg/tdh"
	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/arthur-debert/tdh/pkg/tdh/testutil"
)

// Example: TestAddCommand using testutil
func TestAddCommand_WithTestutil(t *testing.T) {
	// Much simpler setup with testutil
	store := testutil.CreatePopulatedStore(t) // Start with empty store

	// Add a todo
	opts := tdh.AddOptions{CollectionPath: store.Path()}
	result, err := tdh.Add("My first todo", opts)

	// Use testutil assertions
	testutil.AssertNoError(t, err)

	// Verify the todo was created correctly
	if result.Todo.Text != "My first todo" {
		t.Errorf("expected todo text to be 'My first todo', got %q", result.Todo.Text)
	}

	// Load and verify it was saved
	collection, err := store.Load()
	testutil.AssertNoError(t, err)

	testutil.AssertCollectionSize(t, collection, 1)
	testutil.AssertTodoInList(t, collection.Todos, "My first todo")
	testutil.AssertTodoHasStatus(t, collection.Todos[0], models.StatusPending)
}

// Example: TestSearchCommand using testutil
func TestSearchCommand_WithTestutil(t *testing.T) {
	// Create a store with predefined todos
	store := testutil.CreateStoreWithSpecs(t, []testutil.TodoSpec{
		{Text: "Buy milk", Status: models.StatusPending},
		{Text: "Buy bread", Status: models.StatusDone},
		{Text: "Walk the dog", Status: models.StatusPending},
		{Text: "Write tests", Status: models.StatusDone},
	})

	// Search for "Buy"
	opts := tdh.SearchOptions{
		CollectionPath: store.Path(),
		CaseSensitive:  false,
	}
	result, err := tdh.Search("Buy", opts)

	testutil.AssertNoError(t, err)

	// Should find 2 todos
	if len(result.MatchedTodos) != 2 {
		t.Errorf("expected 2 matches, got %d", len(result.MatchedTodos))
	}

	// Verify the matched todos
	testutil.AssertTodoInList(t, result.MatchedTodos, "Buy milk")
	testutil.AssertTodoInList(t, result.MatchedTodos, "Buy bread")

	// Should not include the others
	testutil.AssertTodoNotInList(t, result.MatchedTodos, "Walk the dog")
	testutil.AssertTodoNotInList(t, result.MatchedTodos, "Write tests")
}

// Example: TestCleanCommand using testutil
func TestCleanCommand_WithTestutil(t *testing.T) {
	// Create a store with mixed pending/done todos
	store := testutil.CreateStoreWithSpecs(t, []testutil.TodoSpec{
		{Text: "Pending task 1", Status: models.StatusPending},
		{Text: "Done task 1", Status: models.StatusDone},
		{Text: "Pending task 2", Status: models.StatusPending},
		{Text: "Done task 2", Status: models.StatusDone},
		{Text: "Done task 3", Status: models.StatusDone},
	})

	// Run clean command
	opts := tdh.CleanOptions{CollectionPath: store.Path()}
	result, err := tdh.Clean(opts)

	testutil.AssertNoError(t, err)

	// Verify the result
	if result.RemovedCount != 3 {
		t.Errorf("expected 3 removed, got %d", result.RemovedCount)
	}
	if result.ActiveCount != 2 {
		t.Errorf("expected 2 active, got %d", result.ActiveCount)
	}

	// Verify removed todos
	testutil.AssertTodoInList(t, result.RemovedTodos, "Done task 1")
	testutil.AssertTodoInList(t, result.RemovedTodos, "Done task 2")
	testutil.AssertTodoInList(t, result.RemovedTodos, "Done task 3")

	// Load collection and verify only pending tasks remain
	collection, _ := store.Load()
	testutil.AssertCollectionSize(t, collection, 2)
	testutil.AssertTodoInList(t, collection.Todos, "Pending task 1")
	testutil.AssertTodoInList(t, collection.Todos, "Pending task 2")
}
