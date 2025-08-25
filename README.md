# too

> A minimal terminal based todo list manager that's feature rich. Design to be
> fast on a command line first and foremost.

# No long commands, no need to quote 
too a This is a todo # long forms are avaialable too too add "This is a todo"
too c 1 # complete todo 1, too complete 1
>
Let's expand on minimal and feature rich :)
too is not a task manager doesn't support multi user and other interfaces
but the cli.

Withing that scope, too has plenty of useful features such as: 
- Nested todos
- Automatic scoping per git repos.
- Multi-line todos
- Search
- Rich terminal output
- Various outputs formats, including mardown and json.
- Can use $EDITOR to edit more complex todos.
- Move and reorder of todo items.
- Archiving and cleaning of todos.

## Usage
  too init                    # creates a new todo list here  
  too add "Buy Groceries"
      Added todo #1: Buy Groceries
  too add --to 1 "Milk"
  too complete 1.1            # completes todo item 1 (Groceries)'s first item (Milk)
  too reopen 1.1              # My bad, we still need milk
  too search bread
  too  list --format=markdown # prints all todos in markdown format
  too clean # remove completed todos


### Installation

- From *homebrew*: `brew install too`
- From *binary*: go to the [release page](https://github.com/arthur-debert/too/releases)
- From *source*: `go get github.com/arthur-debert/too`
- From .deb: get the .deb in the github releases page.

### Data Files

*too* will look at a `.todos` files to store your todos (like Git does: it will try recursively in each parent folder). This permit to have different list of todos per folder.

If it doesn't find a `.todos`, *too* will store in $HOME/.todos, unless you override this with the  `TODO_DB_PATH` environment variable.

