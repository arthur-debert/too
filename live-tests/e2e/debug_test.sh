#!/bin/zsh

# Create some todos
too add "Buy groceries"
too add "Call dentist"
too add "Plan weekend trip"

# Create hierarchical structure
too add --to 1 "Apples"
too add --to 1 "Bread"
too add --to 3 "Book hotel"
too add --to 3 "Check weather"

# Complete some tasks
too complete 1.1  # Complete "Apples"
too complete 2    # Complete "Call dentist"

# List final state
too list
