#!/bin/zsh

echo "=== Testing Edge Cases ==="

# Test 1: Empty todo text
echo "\n--- Test 1: Empty todo text ---"
too add ""
echo "Expected: Should error or create todo with empty text"

# Test 2: Very long todo text
echo "\n--- Test 2: Very long todo text ---"
too add "This is a very long todo text that should wrap properly when displayed and not cause any issues with the display formatting or storage system even when it gets extremely long like this"
too list

# Test 3: Special characters
echo "\n--- Test 3: Special characters in todo text ---"
too add "Todo with 'quotes' and \"double quotes\""
too add "Todo with \$pecial ch@racters & symbols!"
too add "Todo with unicode: 你好世界 🚀 ✨"
too list

# Test 4: Invalid position references
echo "\n--- Test 4: Invalid position references ---"
too complete 999
echo "Expected: Error message about invalid position"
too edit xyz "Invalid position"
echo "Expected: Error message about invalid position"

# Test 5: Cascade delete behavior
echo "\n--- Test 5: Cascade delete with children ---"
too add "Parent to delete"
too add "Child 1" to 4
too add "Child 2" to 4
too list
# Note: delete command doesn't exist, but this shows what we'd test

# Test 6: Datapath command
echo "\n--- Test 6: Datapath command ---"
too datapath
echo "Expected: Shows path to .todos.db file"

# Test 7: Version command
echo "\n--- Test 7: Version command ---"
too version