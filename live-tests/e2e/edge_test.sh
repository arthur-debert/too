#!/bin/zsh

# Test empty list
too list --format "${TOO_FORMAT}"

# Add todo with special characters
too add "Todo with \"quotes\" and 'apostrophes'" --format "${TOO_FORMAT}"
too add "Todo with unicode: 🚀 🎯 ✅" --format "${TOO_FORMAT}"

# Test very long text
too add "This is a very long todo item that has a lot of text to test how the system handles longer descriptions and whether everything works correctly with extended content" --format "${TOO_FORMAT}"

# Test hierarchy limits
too add "Level 1" --format "${TOO_FORMAT}"
too add --to 1 "Level 2" --format "${TOO_FORMAT}"
too add --to 1.1 "Level 3" --format "${TOO_FORMAT}"

# List final state
too list --format "${TOO_FORMAT}"
