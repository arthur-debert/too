package models

import "sort"

// ReorderTodos sorts todos by their current position and reassigns sequential positions starting from 1.
// This is a pure function that performs an in-memory data transformation.
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
	}
}
