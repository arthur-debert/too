#!/bin/zsh
# Test for bug #154: 'to' keyword not working in add command

echo "=== Bug #154: Testing 'to' keyword in add command ==="
too add "Parent task"
echo "\nAdding child with 'to' keyword:"
too add "Child task" to 1
echo "\nCurrent list:"
too list --all
echo "Expected: Child should appear nested under parent"
