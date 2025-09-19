#!/bin/zsh

too add "Groceries"
too add --to 1 "Milk"
too add --to 1 "Bread"
too add --to 1 "Eggs"
too add --to 1 "Butter"
too add --to 1 "Cheese"
too add --to 1 "Cereal"
too add --to 1 "Yogurt"
too add --to 1 "Fruits"
too add "Pack for Trip"
too add --to 2 "Clothes"
too add --to 2 "Camera Gear"
too add --to 2 "Passport"
too add --to 2 "Toiletries"
too add --to 2 "Snacks"
too add --to 2 "Books"
too add --to 2 "Travel Pillow"
too add --to 2 "Headphones" --contextual

too list --contextual

# lets see how the last item in a long list looks
too complete 2.8 --contextual
# now the first item
too complete 2.1 --contextual
# and one in the middle
too complete 2.4 --contextual
# bread is gone
echo "And now search"
too search "Books"
echo "and this is --all with done items"
too list  --all

# this shell exports the function export_history, which will expoert, sans
# line numbers, the history of commands run in this shell, useful if your
# want to save the commands you ran to a file as a script to replay later
