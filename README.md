# tdh

> A minimal terminal based todo list manager that's feature rich. Design to be
> fast on a command line first and foremost.

# No long commands, no need to quote 
tdh a This is a todo # long forms are avaialable too tdh add "This is a todo"
tdh c 1 # complete todo 1, tdh complete 1
>
Let's expand on minimal and feature rich :)
tdh is not a task manager doesn't support multi user and other interfaces
but the cli.

Withing that scope, tdh has plenty of useful features such as: 
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
  tdh init                    # creates a new todo list here  
  tdh add "Buy Groceries"
      Added todo #1: Buy Groceries
  tdh add --to 1 "Milk"
  tdh complete 1.1            # completes todo item 1 (Groceries)'s first item (Milk)
  tdh reopen 1.1              # My bad, we still need milk
  tdh search bread
  tdh  list --format=markdown # prints all todos in markdown format
  tdh clean # remove completed todos


### Installation

- From *homebrew*: `brew install tdh`
- From *binary*: go to the [release page](https://github.com/arthur-debert/tdh/releases)
- From *source*: `go get github.com/arthur-debert/tdh`
- From .deb: get the .deb in the github releases page.

### Data Files

*tdh* will look at a `.todos` files to store your todos (like Git does: it will try recursively in each parent folder). This permit to have different list of todos per folder.

If it doesn't find a `.todos`, *tdh* will store in $HOME/.todos, unless you override this with the  `TODO_DB_PATH` environment variable.

