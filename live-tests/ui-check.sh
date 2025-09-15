#!/bin/zsh

too add "Groceries" --format "${TOO_FORMAT}"
too add --to 1 "Milk" --format "${TOO_FORMAT}"
too add --to 1 "Bread" --format "${TOO_FORMAT}"
too add --to 1 "Eggs" --format "${TOO_FORMAT}"
too add "Pack for Trip" --format "${TOO_FORMAT}"
too add --to 2 "Clothes" --format "${TOO_FORMAT}"
too add --to 2 "Camera Gear" --format "${TOO_FORMAT}"
too add --to 2 "Passport"
too list --format "${TOO_FORMAT}"
echo "well look at how various lists look"
echo "now we delete the second subtask of the first task (Bread)"
echo "command prints the --all list version so you can see the effect"
too complete 1.2 --format "${TOO_FORMAT}" # bread is gone
echo "And now search"
too search "Pack" --format "${TOO_FORMAT}"
echo "and this is --all with done items"
too list --all --format "${TOO_FORMAT}"

# this shell exports the function export_history, which will expoert, sans
# line numbers, the history of commands run in this shell, useful if your
# want to save the commands you ran to a file as a script to replay later
