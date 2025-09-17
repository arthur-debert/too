#!/bin/zsh
# Standard test database setup
# Creates the baseline test data structure as specified

too add "Groceries" --format "${TOO_FORMAT}"
too add --to 1 "Bread" --format "${TOO_FORMAT}"
too add --to 1 "Milk" --format "${TOO_FORMAT}"
too add --to 1 "Pancakes" --format "${TOO_FORMAT}"
too add --to 1 "Eggs" --format "${TOO_FORMAT}"
too complete 1.4 --format "${TOO_FORMAT}"  # Complete Eggs

too add "Pack" --format "${TOO_FORMAT}"
too add --to 2 "Camera" --format "${TOO_FORMAT}"
too add --to 2 "Clothes" --format "${TOO_FORMAT}"
too add --to 2 "Passport" --format "${TOO_FORMAT}"
too add --to 2 "Bag" --format "${TOO_FORMAT}"
too complete 2.4 --format "${TOO_FORMAT}"  # Complete Bag