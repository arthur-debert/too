#!/bin/zsh
# Test for bug #158: Nested completed item path resolution

echo "=== Bug #158: Testing nested completed item paths ==="
too add "Parent task"
too add "Child task" to 1
echo "\nBefore completion:"
too list --all
echo "\nCompleting child task:"
too complete 1.1
echo "\nAfter completion:"
too list --all
echo "\nTrying to reopen with c path (should be c1.1 or similar):"
too reopen c1.1
echo "\nAfter reopen attempt:"
too list --all
echo "Expected: Nested completed items should have proper c-prefixed paths"
