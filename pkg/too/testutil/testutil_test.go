package testutil_test

import (
	"testing"

	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/testutil"
)

func TestCreatePopulatedStore(t *testing.T) {
	// Test creating a populated store
	adapter, _ := testutil.CreatePopulatedStore(t)
	defer adapter.Close()

	// Load all todos
	todos := testutil.LoadTodos(t, adapter, true)

	// Should have 3 todos (2 root + 1 child)
	testutil.AssertTodoCount(t, todos, 3)

	// Verify the todos exist
	testutil.AssertTodoInList(t, todos, "First todo")
	testutil.AssertTodoInList(t, todos, "Second todo")
	testutil.AssertTodoInList(t, todos, "Child of first")

	// Check statuses - second should be done
	foundSecond := false
	for _, todo := range todos {
		if todo.Text == "Second todo" {
			testutil.AssertTodoHasStatus(t, todo, models.StatusDone)
			foundSecond = true
		}
	}
	if !foundSecond {
		t.Error("Second todo not found")
	}
}

func TestCreateStoreWithSpecs(t *testing.T) {
	// Test creating a store with specific todo states
	specs := []testutil.TodoSpec{
		{Text: "Completed task", Complete: true},
		{Text: "Pending task"},
		{Text: "Another done task", Complete: true},
		{Text: "Child task", ParentPos: "1"}, // Child of first todo
	}

	adapter, _ := testutil.CreateStoreWithSpecs(t, specs...)
	defer adapter.Close()

	// Load all todos
	todos := testutil.LoadTodos(t, adapter, true)

	// Verify counts
	testutil.AssertTodoCount(t, todos, 4)

	// Count done todos
	doneCount := 0
	for _, todo := range todos {
		if todo.GetStatus() == models.StatusDone {
			doneCount++
		}
	}

	if doneCount != 2 {
		t.Errorf("expected 2 done todos, got %d", doneCount)
	}

	// Verify specific todos
	testutil.AssertTodoInList(t, todos, "Completed task")
	testutil.AssertTodoInList(t, todos, "Pending task")
	testutil.AssertTodoInList(t, todos, "Child task")
}

func TestCreateTestStore(t *testing.T) {
	// Test creating an empty store
	adapter, dbPath := testutil.CreateTestStore(t)
	defer adapter.Close()

	// Should have created a valid path
	if dbPath == "" {
		t.Error("expected non-empty db path")
	}

	// Should be able to add todos
	todo, err := adapter.Add("Test todo", nil)
	testutil.AssertNoError(t, err)

	if todo.Text != "Test todo" {
		t.Errorf("expected todo text 'Test todo', got %q", todo.Text)
	}

	// Verify it was saved
	todos := testutil.LoadTodos(t, adapter, false)
	testutil.AssertTodoCount(t, todos, 1)
}

func TestLoadTodos(t *testing.T) {
	adapter, _ := testutil.CreateStoreWithSpecs(t,
		testutil.TodoSpec{Text: "Active 1"},
		testutil.TodoSpec{Text: "Done 1", Complete: true},
		testutil.TodoSpec{Text: "Active 2"},
		testutil.TodoSpec{Text: "Done 2", Complete: true},
	)
	defer adapter.Close()

	// Test loading only active todos
	activeTodos := testutil.LoadTodos(t, adapter, false)
	testutil.AssertTodoCount(t, activeTodos, 2)
	testutil.AssertTodoInList(t, activeTodos, "Active 1")
	testutil.AssertTodoInList(t, activeTodos, "Active 2")

	// Test loading all todos
	allTodos := testutil.LoadTodos(t, adapter, true)
	testutil.AssertTodoCount(t, allTodos, 4)
	testutil.AssertTodoInList(t, allTodos, "Done 1")
	testutil.AssertTodoInList(t, allTodos, "Done 2")
}