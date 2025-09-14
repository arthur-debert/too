package main

import (
	"fmt"
)

// Example of how completed namespace should work

type Todo struct {
	UID      string
	Text     string
	Status   string
	ParentID string
	Position string // This will be calculated
}

// SimulateIDMNamespace shows how the completed namespace pattern should work
func SimulateIDMNamespace() {
	// Initial state: 3 active todos under root
	todos := []Todo{
		{UID: "uid1", Text: "Groceries", Status: "pending", ParentID: ""},
		{UID: "uid2", Text: "Milk", Status: "pending", ParentID: "uid1"},
		{UID: "uid3", Text: "Bread", Status: "pending", ParentID: "uid1"},
		{UID: "uid4", Text: "Eggs", Status: "pending", ParentID: "uid1"},
		{UID: "uid5", Text: "Pack for Trip", Status: "pending", ParentID: ""},
	}

	fmt.Println("=== Initial State ===")
	assignPositions(todos)
	printTodos(todos)

	// Complete "Bread" (position 1.2)
	fmt.Println("\n=== After completing Bread (1.2) ===")
	todos[2].Status = "done"
	assignPositions(todos)
	printTodos(todos)
	
	// Complete "Milk" (now position 1.1)
	fmt.Println("\n=== After completing Milk (1.1) ===")
	todos[1].Status = "done"
	assignPositions(todos)
	printTodos(todos)

	// Reopen "Bread" (currently 1.c1)
	fmt.Println("\n=== After reopening Bread (1.c1) ===")
	todos[2].Status = "pending"
	assignPositions(todos)
	printTodos(todos)
}

func assignPositions(todos []Todo) {
	// Group by parent
	childrenByParent := make(map[string][]int)
	for i, todo := range todos {
		parent := todo.ParentID
		if parent == "" {
			parent = "root"
		}
		childrenByParent[parent] = append(childrenByParent[parent], i)
	}

	// Assign positions for each parent's children
	for parent, indices := range childrenByParent {
		activeCount := 0
		completedCount := 0
		
		// First pass: count and assign to active items
		for _, i := range indices {
			if todos[i].Status == "pending" {
				activeCount++
				if parent == "root" {
					todos[i].Position = fmt.Sprintf("%d", activeCount)
				} else {
					// Find parent's position
					parentPos := ""
					for _, t := range todos {
						if t.UID == parent {
							parentPos = t.Position
							break
						}
					}
					todos[i].Position = fmt.Sprintf("%s.%d", parentPos, activeCount)
				}
			}
		}
		
		// Second pass: assign to completed items
		for _, i := range indices {
			if todos[i].Status == "done" {
				completedCount++
				if parent == "root" {
					todos[i].Position = fmt.Sprintf("c%d", completedCount)
				} else {
					// Find parent's position
					parentPos := ""
					for _, t := range todos {
						if t.UID == parent {
							parentPos = t.Position
							break
						}
					}
					todos[i].Position = fmt.Sprintf("%s.c%d", parentPos, completedCount)
				}
			}
		}
	}
}

func printTodos(todos []Todo) {
	// Print in a nice tree format
	for _, todo := range todos {
		if todo.ParentID == "" {
			prefix := "○"
			if todo.Status == "done" {
				prefix = "●"
			}
			fmt.Printf("%s %s. %s\n", prefix, todo.Position, todo.Text)
			
			// Print children
			for _, child := range todos {
				if child.ParentID == todo.UID {
					prefix := "○"
					if child.Status == "done" {
						prefix = "●"
					}
					fmt.Printf("  %s %s. %s\n", prefix, child.Position, child.Text)
				}
			}
		}
	}
}

func main() {
	SimulateIDMNamespace()
	
	fmt.Println("\n=== Key Principles ===")
	fmt.Println("1. Active items: consecutive numbering (1, 2, 3...)")
	fmt.Println("2. Completed items: 'c' namespace (c1, c2, c3...)")
	fmt.Println("3. Position paths are stable - '1.2' always refers to 2nd active child")
	fmt.Println("4. Completing moves to 'c' namespace, reopening inserts at position 1")
	fmt.Println("5. No view-dependent IDs - same position path in all contexts")
	
	fmt.Println("\n=== Example Commands ===")
	fmt.Println("too complete 1.2  # Completes 2nd active child of item 1")
	fmt.Println("too reopen 1.c1   # Reopens 1st completed child, makes it 1.1")
	fmt.Println("too list          # Shows only active with clean numbering")
	fmt.Println("too list --all    # Shows all with 'c' prefix for completed")
}