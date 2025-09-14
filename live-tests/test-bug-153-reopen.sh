#!/bin/zsh
# Test for bug #153: Reopen command not working

echo "=== Bug #153: Testing reopen command ==="
too add "Todo to complete and reopen"
echo "Before completion:"
too list --all

too complete 1
echo "After completion:"
too list --all

echo "Attempting to reopen with 'too reopen c1':"
too reopen c1
echo "After reopen attempt:"
too list --all
echo "Expected: Todo should be back in pending state"