package internal

import "github.com/arthur-debert/too/pkg/too/models"

// CountTodos recursively calculates the total count and done count from a collection of todos.
// This helper function provides consistent counting logic across all store implementations.
func CountTodos(todos []*models.Todo) (totalCount, doneCount int) {
	for _, todo := range todos {
		totalCount++
		if todo.GetStatus() == models.StatusDone {
			doneCount++
		}
		// Recursively count children
		childTotal, childDone := CountTodos(todo.Items)
		totalCount += childTotal
		doneCount += childDone
	}
	return totalCount, doneCount
}
