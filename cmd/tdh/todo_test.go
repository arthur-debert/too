package main

import "github.com/arthur-debert/tdh/pkg/tdh"

func ExampleTodo() {
	todo := tdh.Todo{
		ID:       0,
		Text:     "Test td",
		Status:   "pending",
		Modified: "",
	}
	todo.MakeOutput(false)
	// Output: 0 | ✕ Test td
}
