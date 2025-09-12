package main

import (
	_ "embed"
)

// Command descriptions
const (
	// Root command
	msgRootShort   = "A simple command-line todo list manager"
	msgRootLong    = "Fast project-aware, nested, command-line todo list."
	msgRootVersion = "too version {{.Version}}\n"

	// Add command
	msgAddUse   = "add <text> [parent-position]"
	msgAddShort = "Add a new todo"
	msgAddLong  = `Add a new todo with the specified text.

You can specify a parent in three ways:
  too add "Buy milk" 1        # Add as child of todo #1
  too add "Buy milk" 1.2      # Add as child of todo #1.2
  too add "Buy milk" --to 1   # Using the --to flag`

	// Clean command
	msgCleanUse   = "clean"
	msgCleanShort = "Remove finished todos"
	msgCleanLong  = "Remove all todos marked as done from the collection."

	// Edit command
	msgEditUse   = "edit <position> <text>"
	msgEditShort = "Edit the text of an existing todo"
	msgEditLong  = "Edit the text of an existing todo by its position."

	// Init command
	msgInitUse   = "init"
	msgInitShort = "Initialize a new todo collection"
	msgInitLong  = "Initialize a new todo collection in the specified location or the default location (~/.todos.json)."

	// List command
	msgListUse   = "list"
	msgListShort = "List all todos"
	msgListLong  = "List all todos in the collection."

	// Search command
	msgSearchUse   = "search <query>"
	msgSearchShort = "Search for todos"
	msgSearchLong  = "Search for todos containing the specified text."

	// Complete command
	msgCompleteUse   = "complete <positions...>"
	msgCompleteShort = "Mark todos as complete"
	msgCompleteLong  = "Mark one or more todos as complete. Use dot notation for nested items (e.g., 1.2)."

	// Reopen command
	msgReopenUse   = "reopen <positions...>"
	msgReopenShort = "Mark todos as pending"
	msgReopenLong  = "Mark one or more todos as pending. Use dot notation for nested items (e.g., 1.2)."

	// Move command
	msgMoveUse   = "move <source_path> <destination_parent_path>"
	msgMoveShort = "Move a todo to a different parent"
	msgMoveLong  = "Move a todo from one location to another in the hierarchy. Use dot notation for paths (e.g., 1.2). Use empty string \"\" for root level."
)

// Flag descriptions
const (
	// Global flags
	msgFlagVerbose  = "Increase verbosity (-v, -vv, -vvv)"
	msgFlagDataPath = "path to todo collection (default: $HOME/.todos.json)"
	msgFlagFormat   = "output format (term, json, markdown)"
	msgFlagMode     = "output mode (short, long)"

	// List command flags
	msgFlagDone = "print done todos"
	msgFlagAll  = "print all todos"

	// Search command flags
	msgFlagCaseSensitive = "Perform case-sensitive search"
)

// Error messages
const (
	msgCommandFailed = "Command failed"
)

// Command aliases
var (
	aliasesAdd      = []string{"a", "new", "create"}
	aliasesEdit     = []string{"modify", "e"}
	aliasesInit     = []string{"i"}
	aliasesList     = []string{"ls"}
	aliasesSearch   = []string{"s"}
	aliasesComplete = []string{"c"}
	aliasesReopen   = []string{"o"}
	aliasesMove     = []string{"m"}
)

//go:embed templates/help.txt
var helpTemplate string
