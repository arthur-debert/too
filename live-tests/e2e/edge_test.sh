#!/bin/zsh

# Test empty list
too list

# Add todo with special characters
too add "Todo with \"quotes\" and 'apostrophes'"
too add "Todo with unicode: 🚀 🎯 ✅"

# Test very long text
too add "This is a very long todo item that has a lot of text to test how the system handles longer descriptions and whether everything works correctly with extended content"

# Test hierarchy limits
too add "Level 1"
too add --to 1 "Level 2"
too add --to 1.1 "Level 3"

# List final state
too list
