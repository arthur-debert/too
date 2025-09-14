#!/bin/zsh
# Test for bug #156: Completed todos not showing in --all view

echo "=== Bug #156: Testing completed todo visibility ==="
too add "Task to complete"
echo "\nBefore completion:"
too list --all
echo "\nCompleting task:"
too complete 1
echo "\nAfter completion with --all:"
too list --all
echo "\nAfter completion with --done:"
too list --done
echo "Expected: Completed task should show with both --all and --done flags"
