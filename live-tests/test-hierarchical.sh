#!/bin/zsh

echo "=== Testing Hierarchical Operations ==="

# Test 1: Create nested structure
echo "\n--- Test 1: Create nested structure ---"
too add "Project Alpha"
too add "Task 1" to 1
too add "Task 2" to 1
too add "Subtask 1.1" to 1.1
too add "Subtask 1.2" to 1.1
too list

# Test 2: Complete parent with children
echo "\n--- Test 2: Complete parent with active children ---"
too complete 1.1
too list
echo "Expected: Parent 1.1 completed, children should be hidden in active view"
too list --all
echo "Expected: Shows completed parent and its children"

# Test 3: Move operation
echo "\n--- Test 3: Move todo to different parent ---"
too add "Project Beta"
too move 1.2 to 2
too list

# Test 4: Complete all children to trigger auto-completion
echo "\n--- Test 4: Auto-completion when all children done ---"
too complete 1.1.1
too complete 1.1.2
too list --all
echo "Expected: Parent 1.1 should be auto-completed"

# Test 5: Multiple completion with IDs
echo "\n--- Test 5: Complete multiple items by ID ---"
too add "Item A"
too add "Item B" 
too add "Item C"
too complete 3 4 5
too list
echo "Expected: All three items should be completed"