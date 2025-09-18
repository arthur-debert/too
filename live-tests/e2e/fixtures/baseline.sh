#!/bin/zsh
# Standard test database setup
# Creates the baseline test data structure as specified

too add "Groceries"
too add --to 1 "Bread"
too add --to 1 "Milk"
too add --to 1 "Pancakes"
too add --to 1 "Eggs"
too complete 1.4  # Complete Eggs

too add "Pack"
too add --to 2 "Camera"
too add --to 2 "Clothes"
too add --to 2 "Passport"
too add --to 2 "Bag"
too complete 2.4  # Complete Bag