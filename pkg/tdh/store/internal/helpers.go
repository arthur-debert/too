package internal

import "github.com/arthur-debert/tdh/pkg/tdh/models"

// CountTodos calculates the total count and done count from a collection of todos.
// This helper function provides consistent counting logic across all store implementations.
func CountTodos(todos []*models.Todo) (totalCount, doneCount int) {
	totalCount = len(todos)
	for _, todo := range todos {
		if todo.Status == models.StatusDone {
			doneCount++
		}
	}
	return totalCount, doneCount
}
