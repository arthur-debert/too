package main

// Command descriptions
const (
	// Root command
	msgRootShort   = "A simple command-line todo list manager"
	msgRootLong    = "tdh is a simple command-line todo list manager that helps you track tasks.\nIt stores todos in a JSON file and provides commands to add, modify, and search todos."
	msgRootVersion = "tdh version {{.Version}}\n"

	// Add command
	msgAddUse   = "add <text>"
	msgAddShort = "Add a new todo (aliases: a, new, create)"
	msgAddLong  = "Add a new todo with the specified text."

	// Clean command
	msgCleanUse   = "clean"
	msgCleanShort = "Remove finished todos"
	msgCleanLong  = "Remove all todos marked as done from the collection."

	// Edit command
	msgEditUse   = "edit <position> <text>"
	msgEditShort = "Edit the text of an existing todo (aliases: modify, m, e)"
	msgEditLong  = "Edit the text of an existing todo by its position."

	// Init command
	msgInitUse   = "init"
	msgInitShort = "Initialize a new todo collection (alias: i)"
	msgInitLong  = "Initialize a new todo collection in the specified location or the default location (~/.todos.json)."

	// List command
	msgListUse   = "list"
	msgListShort = "List all todos (alias: ls)"
	msgListLong  = "List all todos in the collection."

	// Reorder command
	msgReorderUse   = "reorder"
	msgReorderShort = "Reorder todos by sorting and reassigning sequential positions (alias: r)"
	msgReorderLong  = "Reorder todos by sorting them by their current position and reassigning sequential positions starting from 1."

	// Search command
	msgSearchUse   = "search <query>"
	msgSearchShort = "Search for todos (alias: s)"
	msgSearchLong  = "Search for todos containing the specified text."

	// Complete command
	msgCompleteUse   = "complete <positions...>"
	msgCompleteShort = "Mark todos as complete (alias: c)"
	msgCompleteLong  = "Mark one or more todos as complete. Use dot notation for nested items (e.g., 1.2)."

	// Reopen command
	msgReopenUse   = "reopen <positions...>"
	msgReopenShort = "Mark todos as pending (alias: o)"
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
	aliasesEdit     = []string{"modify", "m", "e"}
	aliasesInit     = []string{"i"}
	aliasesList     = []string{"ls"}
	aliasesReorder  = []string{"r"}
	aliasesSearch   = []string{"s"}
	aliasesComplete = []string{"c"}
	aliasesReopen   = []string{"o"}
)
