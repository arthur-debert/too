package models

import "sort"

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

// ResetActivePositions resets positions for only active (pending) todos in the slice.
// Done items are left with position 0, pending items get sequential positions starting from 1.
func ResetActivePositions(todos []*Todo) {
	if len(todos) == 0 {
		return
	}

	// First, collect active todos and ensure done todos have position 0
	var activeTodos []*Todo
	for _, todo := range todos {
		switch todo.Status {
		case StatusPending:
			activeTodos = append(activeTodos, todo)
		case StatusDone:
			// Ensure done items have position 0
			todo.Position = 0
		}
	}

	// Sort active todos by their current position
	// Items with position 0 (newly reopened) go to the end
	sort.SliceStable(activeTodos, func(i, j int) bool {
		// If one has position 0 and the other doesn't, the one with 0 goes after
		if activeTodos[i].Position == 0 && activeTodos[j].Position != 0 {
			return false
		}
		if activeTodos[i].Position != 0 && activeTodos[j].Position == 0 {
			return true
		}
		// Otherwise sort by position
		return activeTodos[i].Position < activeTodos[j].Position
	})

	// Assign sequential positions to active todos only
	for i, todo := range activeTodos {
		todo.Position = i + 1
	}

	// Note: We don't recursively process children here because
	// position reset should only happen at the requested level
}
