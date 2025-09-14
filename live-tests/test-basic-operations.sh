#!/bin/zsh

echo "=== Testing Basic Operations ==="

# Test 1: Basic add and list
echo "\n--- Test 1: Basic add and list ---"
too add "First todo"
too add "Second todo"
too add "Third todo"
too list

# Test 2: Complete and verify renumbering
echo "\n--- Test 2: Complete middle item and verify renumbering ---"
too complete 2
too list
echo "Expected: First todo=1, Third todo=2"

# Test 3: List all shows completed items
echo "\n--- Test 3: List all shows completed items ---"
too list --all
echo "Expected: Shows Second todo as completed with stable position"

# Test 4: Edit todo text
echo "\n--- Test 4: Edit todo text ---"
too edit 1 "First todo EDITED"
too list

# Test 5: Reopen completed item
echo "\n--- Test 5: Reopen completed item ---"
too reopen 2  # Using stable position from --all view
too list
echo "Expected: All three todos active again"

# Test 6: Delete todo
echo "\n--- Test 6: Delete todo ---"
too delete 2
too list
echo "Expected: Only two todos remain"

# Test 7: Search functionality
echo "\n--- Test 7: Search functionality ---"
too add "Search test item"
too search "test"
echo "Expected: Shows 'Search test item'"

# Test 8: Clean completed todos
echo "\n--- Test 8: Clean completed todos ---"
too complete 1
too complete 2
too clean
too list
echo "Expected: Only 'Search test item' remains"