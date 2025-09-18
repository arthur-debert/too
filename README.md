# too

> A minimal terminal based todo list manager that's feature rich. Design to be
> fast on a command line first and foremost.

# No long commands, no need to quote 
too a This is a todo # long forms are available too: too add "This is a todo"
too c 1 # complete todo 1, too complete 1

# Even simpler - naked execution
too Buy milk         # adds "Buy milk" as a new todo
too                  # lists all todos
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

### Quick Start - Naked Execution
  too                         # lists all todos
  too Buy groceries           # adds "Buy groceries" as a new todo
  too --to 1 Milk            # adds "Milk" as a sub-todo of #1
  
### Standard Commands  
  too init                    # creates a new todo list here  
  too add "Buy Groceries"
      Added todo #1: Buy Groceries
  too add --to 1 "Milk"
  too complete 1.1            # completes todo item 1 (Groceries)'s first item (Milk)
  too reopen 1.1              # My bad, we still need milk
  too search bread
  too list --format=markdown  # prints all todos in markdown format
  too clean                   # remove completed todos


### Installation

- From *homebrew*: `brew install too`
- From *binary*: go to the [release page](https://github.com/arthur-debert/too/releases)
- From *source*: `go get github.com/arthur-debert/too`
- From .deb: get the .deb in the github releases page.

### Data Files

*too* determines where to store your todos using this precedence:

1. **Explicit path** - If you use `--data-path` flag, that path is used
2. **Environment variable** - If `TODO_DB_PATH` is set, that path is used  
3. **Project-local** - Searches current directory and parents for `.todos.json` (like Git)
4. **Home directory** - Falls back to `~/.todos.json` if it exists
5. **Default** - Creates `.todos.json` in the current directory

This allows you to have different todo lists per project while maintaining a global list.

