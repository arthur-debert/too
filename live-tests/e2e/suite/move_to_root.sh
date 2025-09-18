#!/bin/zsh
# Create nested structure
too add "Parent"
too add --to 1 "Nested child"

# Move nested child to root level
too move 1.1 .

too list
