package models

import "sort"

// ReorderTodos sorts todos by their current position and reassigns sequential positions starting from 1.
// This is a pure function that performs an in-memory data transformation.
// Returns the number of todos that had their position changed.
func ReorderTodos(todos []*Todo) int {
	if len(todos) == 0 {
		return 0
	}

	// Sort todos by their current position
	// Using a stable sort to maintain relative order of todos with same position
	sort.SliceStable(todos, func(i, j int) bool {
		return todos[i].Position < todos[j].Position
	})

	// Reassign positions sequentially starting from 1
	changed := 0
	for i := range todos {
		newPosition := i + 1
		if todos[i].Position != newPosition {
			todos[i].Position = newPosition
			changed++
		}
	}

	return changed
}
