#!/bin/zsh
# Create nested structure
too add "Parent" --format "${TOO_FORMAT}"
too add --to 1 "Nested child" --format "${TOO_FORMAT}"

# Move nested child to root level
too move 1.1 . --format "${TOO_FORMAT}"

too list --format "${TOO_FORMAT}"
