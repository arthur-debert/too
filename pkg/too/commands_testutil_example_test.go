package too_test

import (
	"testing"

	"github.com/arthur-debert/too/pkg/too"
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/store"
	"github.com/arthur-debert/too/pkg/too/testutil"
)

// Example: TestAddCommand using testutil
func TestAddCommand_WithTestutil(t *testing.T) {
	// Create empty test store to get the path
	adapter, dbPath := testutil.CreateTestStore(t)
	adapter.Close() // Close it since ExecuteUnifiedCommand will create its own

	// Add a todo using unified command
	opts := map[string]interface{}{
		"collectionPath": dbPath,
	}
	result, err := too.ExecuteUnifiedCommand("add", []string{"My first todo"}, opts)

	// Use testutil assertions
	testutil.AssertNoError(t, err)

	// Verify the todo was created correctly
	if len(result.AffectedTodos) == 0 || result.AffectedTodos[0].Text != "My first todo" {
		t.Errorf("expected todo text to be 'My first todo', got %v", result.AffectedTodos)
	}

	// Create a new adapter to load and verify it was saved
	adapter, err = store.NewNanoStoreAdapter(dbPath)
	testutil.AssertNoError(t, err)
	defer adapter.Close()
	
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

	// Search for "Buy" using unified command (with --all to include completed)
	opts := map[string]interface{}{
		"collectionPath": dbPath,
		"query":          "Buy",
		"all":            true,
	}
	result, err := too.ExecuteUnifiedCommand("search", []string{"Buy"}, opts)

	testutil.AssertNoError(t, err)

	// Should find 2 todos
	if len(result.AllTodos) != 2 {
		t.Errorf("expected 2 matches, got %d", len(result.AllTodos))
	}

	// Verify the matched todos
	testutil.AssertTodoInList(t, result.AllTodos, "Buy milk")
	testutil.AssertTodoInList(t, result.AllTodos, "Buy bread")

	// Should not include the others
	testutil.AssertTodoNotInList(t, result.AllTodos, "Walk the dog")
	testutil.AssertTodoNotInList(t, result.AllTodos, "Write tests")
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

	// Run clean command using unified command
	opts := map[string]interface{}{
		"collectionPath": dbPath,
	}
	result, err := too.ExecuteUnifiedCommand("clean", []string{}, opts)

	testutil.AssertNoError(t, err)

	// Verify the result
	if len(result.AffectedTodos) != 3 {
		t.Errorf("expected 3 removed, got %d", len(result.AffectedTodos))
	}
	if len(result.AllTodos) != 2 {
		t.Errorf("expected 2 active, got %d", len(result.AllTodos))
	}

	// Verify removed todos
	testutil.AssertTodoInList(t, result.AffectedTodos, "Done task 1")
	testutil.AssertTodoInList(t, result.AffectedTodos, "Done task 2")
	testutil.AssertTodoInList(t, result.AffectedTodos, "Done task 3")

	// Load collection and verify only pending tasks remain
	todos := testutil.LoadTodos(t, adapter, false)
	testutil.AssertTodoCount(t, todos, 2)
	testutil.AssertTodoInList(t, todos, "Pending task 1")
	testutil.AssertTodoInList(t, todos, "Pending task 2")
}
