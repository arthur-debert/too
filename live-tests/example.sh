#!/bin/zsh

too add "Groceries"
too add --to 1 "Milk"
too add --to 1 "Bread"
too add --to 1 "Eggs"
too add "Pack for Trip"
too add --to 2 "Clothes"
too add --to 2 "Camera Gear"
too add --to 2 "Passport"
too list
echo "now we delete the second subtask of the first task (Bread)"
echo "command prints the --all list version so you can see the effect"
too complete 1.2 # bread is gone
echo "but the normal list command shows the current state"
too list

# this shell exports the function export_history, which will expoert, sans
# line numbers, the history of commands run in this shell, useful if your
# want to save the commands you ran to a file as a script to replay later
