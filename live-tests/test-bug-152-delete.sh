#!/bin/zsh
# Test for bug #152: Delete command is missing

echo "=== Bug #152: Testing delete command ==="
too add "Todo to delete"
too list
echo "Attempting to delete todo 1:"
too delete 1
echo "Expected: Error 'unknown command delete'"