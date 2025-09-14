#!/bin/zsh

echo "=== Testing 'to' Keyword Bug ==="

# Test the add command parsing
echo "\n--- Test: Add child todo ---"
too add "Parent task"
echo "Added parent, now adding child:"
too add "Child task" to 1
echo "Current list:"
too list

# Test different variations
echo "\n--- Test: Different variations ---"
too add "Another child" --to 1
echo "Added with --to flag:"
too list

# JSON output to see the actual text
echo "\n--- JSON output to see exact text ---"
too list --format json