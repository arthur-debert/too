#!/bin/zsh
# Add multiple todos from a plain text multiline string
too add "First todo
Second todo
Third todo" --format "${TOO_FORMAT}"

too list --format "${TOO_FORMAT}"
