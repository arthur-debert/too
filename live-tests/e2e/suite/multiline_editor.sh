#!/bin/zsh
# Create a file with multiple todos
cat > todos.txt << 'TODOS'
First todo
Second todo
Third todo
TODOS

# Add todos from the file using editor mode
EDITOR="cat todos.txt >" too add -e --format "${TOO_FORMAT}"

too list --format "${TOO_FORMAT}"
rm -f todos.txt
