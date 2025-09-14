#!/bin/zsh

echo "=== Testing Reopen Functionality ==="

# Test 1: Simple reopen
echo "\n--- Test 1: Simple reopen ---"
too add "Test todo 1"
too add "Test todo 2"
too list
echo "Before completion:"
too list --all

# Complete one todo
too complete 1
echo "After completing todo 1:"
too list --all
echo "Position paths for completed items:"

# Try different ways to reopen
echo "Trying reopen with different position references:"
too reopen c1
echo "Tried reopen c1"
too list --all

echo "Trying reopen with position 1:"
too reopen 1  # This should reopen "Test todo 2" since it's now position 1
too list --all

# Test 2: Show business logic directly via list
echo "\n--- Test 2: Show all todos to understand positions ---"
too list --format json