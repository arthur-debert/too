#!/bin/zsh

# Create some todos
too add "Buy groceries" --format "${TOO_FORMAT}"
too add "Call dentist" --format "${TOO_FORMAT}"
too add "Plan weekend trip" --format "${TOO_FORMAT}"

# Create hierarchical structure
too add --to 1 "Apples" --format "${TOO_FORMAT}"
too add --to 1 "Bread" --format "${TOO_FORMAT}"
too add --to 3 "Book hotel" --format "${TOO_FORMAT}"
too add --to 3 "Check weather" --format "${TOO_FORMAT}"

# Complete some tasks
too complete 1.1 --format "${TOO_FORMAT}"  # Complete "Apples"
too complete 2 --format "${TOO_FORMAT}"    # Complete "Call dentist"

# List final state
too list --format "${TOO_FORMAT}"
