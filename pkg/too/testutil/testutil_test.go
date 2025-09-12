package testutil_test

import (
	"fmt"
	"testing"

	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/testutil"
)

func TestCreatePopulatedStore(t *testing.T) {
	// Test creating a store with some todos
	s := testutil.CreatePopulatedStore(t, "Buy milk", "Walk dog", "Write tests")

	collection, err := s.LoadIDM()
	if err != nil {
		t.Fatalf("failed to load collection: %v", err)
	}

	// Verify we have 3 todos
	testutil.AssertCollectionSize(t, collection, 3)

	// Verify the todos exist
	testutil.AssertTodoInList(t, collection.Items, "Buy milk")
	testutil.AssertTodoInList(t, collection.Items, "Walk dog")
	testutil.AssertTodoInList(t, collection.Items, "Write tests")

	// All should be pending by default
	for _, todo := range collection.Items {
		testutil.AssertTodoHasStatus(t, todo, models.StatusPending)
	}
}

func TestCreateStoreWithSpecs(t *testing.T) {
	// Test creating a store with specific todo states
	specs := []testutil.TodoSpec{
		{Text: "Completed task", Status: models.StatusDone},
		{Text: "Pending task", Status: models.StatusPending},
		{Text: "Another done task", Status: models.StatusDone},
	}

	s := testutil.CreateStoreWithSpecs(t, specs)

	// Load collection and count manually
	collection, err := s.LoadIDM()
	testutil.AssertNoError(t, err)

	// Verify counts
	totalCount := len(collection.Items)
	doneCount := 0
	for _, todo := range collection.Items {
		if todo.GetStatus() == models.StatusDone {
			doneCount++
		}
	}
	
	if totalCount != 3 {
		t.Errorf("expected 3 total todos, got %d", totalCount)
	}
	if doneCount != 2 {
		t.Errorf("expected 2 done todos, got %d", doneCount)
	}

	// Verify individual todos
	for _, todo := range collection.Items {
		switch todo.Text {
		case "Completed task", "Another done task":
			testutil.AssertTodoHasStatus(t, todo, models.StatusDone)
		case "Pending task":
			testutil.AssertTodoHasStatus(t, todo, models.StatusPending)
		}
	}
}

func TestAssertTodoByID(t *testing.T) {
	s := testutil.CreatePopulatedStore(t, "Test todo")
	collection, _ := s.LoadIDM()

	// Should find the todo
	todo := testutil.AssertTodoByID(t, collection.Items, collection.Items[0].UID)
	if todo.Text != "Test todo" {
		t.Errorf("expected todo text to be 'Test todo', got %q", todo.Text)
	}
}

func TestAssertError(t *testing.T) {
	// Test with an actual error
	err := fmt.Errorf("file not found: test.txt")

	// This should not panic
	testutil.AssertError(t, err, "not found")

	// Test AssertNoError with nil
	testutil.AssertNoError(t, nil)
}

func TestAssertTodoNotInList(t *testing.T) {
	s := testutil.CreatePopulatedStore(t, "Buy milk", "Walk dog")
	collection, _ := s.LoadIDM()

	// This should pass - "Write tests" is not in the list
	testutil.AssertTodoNotInList(t, collection.Items, "Write tests")
}
