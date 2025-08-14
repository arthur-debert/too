package testutil

import (
	"testing"

	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/arthur-debert/tdh/pkg/tdh/store"
)

// AssertTodoInList checks if a todo with the given text exists in the list.
func AssertTodoInList(t *testing.T, todos []*models.Todo, expectedText string) {
	t.Helper()

	for _, todo := range todos {
		if todo.Text == expectedText {
			return
		}
	}

	t.Errorf("expected todo with text %q not found in list of %d todos", expectedText, len(todos))
}

// AssertTodoNotInList checks that a todo with the given text does not exist in the list.
func AssertTodoNotInList(t *testing.T, todos []*models.Todo, unexpectedText string) {
	t.Helper()

	for _, todo := range todos {
		if todo.Text == unexpectedText {
			t.Errorf("unexpected todo with text %q found in list", unexpectedText)
			return
		}
	}
}

// AssertTodoCount checks the total and done counts from a FindResult.
func AssertTodoCount(t *testing.T, result *store.FindResult, expectedTotal, expectedDone int) {
	t.Helper()

	if result.TotalCount != expectedTotal {
		t.Errorf("expected total count %d, got %d", expectedTotal, result.TotalCount)
	}

	if result.DoneCount != expectedDone {
		t.Errorf("expected done count %d, got %d", expectedDone, result.DoneCount)
	}
}

// AssertTodoHasStatus checks that a todo has the expected status.
func AssertTodoHasStatus(t *testing.T, todo *models.Todo, expectedStatus models.TodoStatus) {
	t.Helper()

	if todo.Status != expectedStatus {
		t.Errorf("expected todo %q to have status %q, got %q", todo.Text, expectedStatus, todo.Status)
	}
}

// AssertCollectionSize checks that a collection has the expected number of todos.
func AssertCollectionSize(t *testing.T, collection *models.Collection, expectedSize int) {
	t.Helper()

	actualSize := len(collection.Todos)
	if actualSize != expectedSize {
		t.Errorf("expected collection to have %d todos, got %d", expectedSize, actualSize)
	}
}

// AssertTodoByID finds a todo by ID and verifies it exists.
// Returns the todo if found, allowing further assertions.
func AssertTodoByID(t *testing.T, todos []*models.Todo, id int64) *models.Todo {
	t.Helper()

	for _, todo := range todos {
		if todo.ID == id {
			return todo
		}
	}

	t.Errorf("todo with ID %d not found", id)
	return nil
}

// AssertError checks that an error occurred and optionally contains a substring.
func AssertError(t *testing.T, err error, contains string) {
	t.Helper()

	if err == nil {
		t.Error("expected error but got nil")
		return
	}

	if contains != "" && !containsString(err.Error(), contains) {
		t.Errorf("expected error to contain %q, got %q", contains, err.Error())
	}
}

// AssertNoError checks that no error occurred.
func AssertNoError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// containsString is a helper to check if a string contains a substring.
func containsString(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
