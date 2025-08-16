package models

import (
	"sort"
)

// ReorderTodos sorts todos by their current position and reassigns sequential positions starting from 1.
// This is a pure function that performs an in-memory data transformation.
// With nested lists, this now also recursively reorders child items.
func ReorderTodos(todos []*Todo) {
	if len(todos) == 0 {
		return
	}

	// Sort todos by their current position
	// Using a stable sort to maintain relative order of todos with same position
	sort.SliceStable(todos, func(i, j int) bool {
		return todos[i].Position < todos[j].Position
	})

	// Reassign positions sequentially starting from 1
	for i := range todos {
		todos[i].Position = i + 1

		// Recursively reorder child items
		if len(todos[i].Items) > 0 {
			ReorderTodos(todos[i].Items)
		}
	}
}

// ResetActivePositions reorders the slice to put active items first (in order),
// followed by done items, and assigns proper positions.
func ResetActivePositions(todos *[]*Todo) {
	if todos == nil || len(*todos) == 0 {
		return
	}

	// Separate existing active todos from newly reopened ones (pos 0)
	var activeTodos []*Todo
	var reopenedTodos []*Todo
	var doneTodos []*Todo

	for _, todo := range *todos {
		switch todo.Status {
		case StatusPending, "": // Handle empty status as pending
			if todo.Position == 0 {
				reopenedTodos = append(reopenedTodos, todo)
			} else {
				activeTodos = append(activeTodos, todo)
			}
		case StatusDone:
			// Ensure done items have position 0 and collect them
			todo.Position = 0
			doneTodos = append(doneTodos, todo)
		default:
			// Handle any other unexpected status as pending
			if todo.Position == 0 {
				reopenedTodos = append(reopenedTodos, todo)
			} else {
				activeTodos = append(activeTodos, todo)
			}
		}
	}

	// Sort the existing active todos by their current position
	sort.SliceStable(activeTodos, func(i, j int) bool {
		return activeTodos[i].Position < activeTodos[j].Position
	})

	// Build the new order: existing active, then reopened, then done
	combinedActive := append(activeTodos, reopenedTodos...)

	// Assign sequential positions to the combined active list
	for i, todo := range combinedActive {
		todo.Position = i + 1
	}

	// Reconstruct the slice in the new order
	newOrder := append(combinedActive, doneTodos...)
	*todos = newOrder
}
