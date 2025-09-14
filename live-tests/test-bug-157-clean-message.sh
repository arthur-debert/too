#!/bin/zsh
# Test for bug #157: Clean command message issues

echo "=== Bug #157: Testing clean command messages ==="
too add "Task one"
too add "Task two"
too complete 1
echo "\nBefore clean:"
too list --all
echo "\nRunning clean command:"
too clean
echo "\nAfter clean:"
too list --all
echo "Expected: Clean should show what was removed and work properly"
