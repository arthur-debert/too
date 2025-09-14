package too_test

import (
	"testing"

	"github.com/arthur-debert/too/pkg/too"
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/testutil"
)

// Example: TestAddCommand using testutil
func TestAddCommand_WithTestutil(t *testing.T) {
	// Create empty test store
	adapter, dbPath := testutil.CreateTestStore(t)
	defer adapter.Close()

	// Add a todo
	opts := too.AddOptions{CollectionPath: dbPath}
	result, err := too.Add("My first todo", opts)

	// Use testutil assertions
	testutil.AssertNoError(t, err)

	// Verify the todo was created correctly
	if result.Todo.Text != "My first todo" {
		t.Errorf("expected todo text to be 'My first todo', got %q", result.Todo.Text)
	}

	// Load and verify it was saved
	todos := testutil.LoadTodos(t, adapter, false)
	testutil.AssertTodoCount(t, todos, 1)
	testutil.AssertTodoInList(t, todos, "My first todo")
	testutil.AssertTodoHasStatus(t, todos[0], models.StatusPending)
}

// Example: TestSearchCommand using testutil
func TestSearchCommand_WithTestutil(t *testing.T) {
	// Create a store with predefined todos
	adapter, dbPath := testutil.CreateStoreWithSpecs(t, 
		testutil.TodoSpec{Text: "Buy milk"},
		testutil.TodoSpec{Text: "Buy bread", Complete: true},
		testutil.TodoSpec{Text: "Walk the dog"},
		testutil.TodoSpec{Text: "Write tests", Complete: true},
	)
	defer adapter.Close()

	// Search for "Buy"
	opts := too.SearchOptions{
		CollectionPath: dbPath,
		CaseSensitive:  false,
	}
	result, err := too.Search("Buy", opts)

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
	adapter, dbPath := testutil.CreateStoreWithSpecs(t,
		testutil.TodoSpec{Text: "Pending task 1"},
		testutil.TodoSpec{Text: "Done task 1", Complete: true},
		testutil.TodoSpec{Text: "Pending task 2"},
		testutil.TodoSpec{Text: "Done task 2", Complete: true},
		testutil.TodoSpec{Text: "Done task 3", Complete: true},
	)
	defer adapter.Close()

	// Run clean command
	opts := too.CleanOptions{CollectionPath: dbPath}
	result, err := too.Clean(opts)

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
	todos := testutil.LoadTodos(t, adapter, false)
	testutil.AssertTodoCount(t, todos, 2)
	testutil.AssertTodoInList(t, todos, "Pending task 1")
	testutil.AssertTodoInList(t, todos, "Pending task 2")
}
