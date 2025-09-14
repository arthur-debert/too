#!/bin/zsh
# Test for bug #155: Multiple completion doesn't work

echo "=== Bug #155: Testing multiple completion ==="
too add "Task one"
too add "Task two"
too add "Task three"
echo "\nBefore completion:"
too list
echo "\nTrying to complete multiple tasks (1 2):"
too complete 1 2
echo "\nAfter completion attempt:"
too list --all
echo "Expected: Both tasks 1 and 2 should be completed"
