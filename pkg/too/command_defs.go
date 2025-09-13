package too

import (
	"fmt"
	"strings"

	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/store"
)

// CommandType represents the type of command for categorization
type CommandType string

const (
	CommandTypeCore  CommandType = "core"  // Mutating todo commands
	CommandTypeExtra CommandType = "extra" // Display/organization commands
	CommandTypeMisc  CommandType = "misc"  // Utility commands
)

// CommandDef defines a unified command structure
type CommandDef struct {
	Name        string
	Aliases     []string
	Type        CommandType
	Description string

	// MutateFunc performs the mutation (if any) and returns affected UIDs
	// If nil, command is display-only
	MutateFunc func(args []string, opts map[string]interface{}) ([]string, error)

	// FilterFunc filters todos for display (optional)
	// If nil, all todos are shown
	FilterFunc func(todos []*models.IDMTodo, opts map[string]interface{}) []*models.IDMTodo

	// MessageFunc generates the command result message
	// Returns empty string for no message (like list)
	MessageFunc func(affectedUIDs []string, todos []*models.IDMTodo) string

	// UsesPositionalArgs indicates if command uses positional arguments
	UsesPositionalArgs bool

	// AcceptsMultiple indicates if command accepts multiple position arguments
	AcceptsMultiple bool
}

// Execute runs the command with unified logic
func (c *CommandDef) Execute(args []string, opts map[string]interface{}) (*ChangeResult, error) {
	var affectedUIDs []string
	var err error

	// Execute mutation if defined
	if c.MutateFunc != nil {
		affectedUIDs, err = c.MutateFunc(args, opts)
		if err != nil {
			return nil, err
		}
	}

	// Get all todos from the collection
	// We need to get the collection path and load todos
	collectionPath, _ := opts["collectionPath"].(string)
	idmStore := store.NewIDMStore(collectionPath)
	manager, err := store.NewPureIDMManager(idmStore, collectionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create manager: %w", err)
	}

	// Get all todos and attach position paths
	todos := manager.ListActive()
	manager.AttachPositionPaths(todos)

	// Apply filter if defined
	displayTodos := todos
	if c.FilterFunc != nil {
		displayTodos = c.FilterFunc(todos, opts)
	}

	// Get affected todos for highlighting
	var affectedTodos []*models.IDMTodo
	if len(affectedUIDs) > 0 {
		affectedMap := make(map[string]bool)
		for _, uid := range affectedUIDs {
			affectedMap[uid] = true
		}
		for _, todo := range todos {
			if affectedMap[todo.UID] {
				affectedTodos = append(affectedTodos, todo)
			}
		}
	}

	// Count stats
	totalCount := len(displayTodos)
	doneCount := 0
	for _, todo := range displayTodos {
		if todo.IsComplete() {
			doneCount++
		}
	}

	// Generate message if defined
	message := ""
	if c.MessageFunc != nil {
		message = c.MessageFunc(affectedUIDs, affectedTodos)
	}

	return NewChangeResult(
		c.Name,
		message,
		affectedTodos,
		displayTodos,
		totalCount,
		doneCount,
	), nil
}

// CommandRegistry holds all command definitions
var CommandRegistry = map[string]*CommandDef{
	"add": {
		Name:               "add",
		Aliases:            []string{"a", "new", "create"},
		Type:               CommandTypeCore,
		Description:        "Add a new todo",
		UsesPositionalArgs: true,
		AcceptsMultiple:    false,
		MutateFunc: func(args []string, opts map[string]interface{}) ([]string, error) {
			text := args[0]
			parentPath := ""
			if v, ok := opts["parent"].(string); ok {
				parentPath = v
			}
			
			collectionPath := ""
			if v, ok := opts["collectionPath"].(string); ok {
				collectionPath = v
			}

			result, err := Add(text, AddOptions{
				CollectionPath: collectionPath,
				ParentPath:     parentPath,
			})
			if err != nil {
				return nil, err
			}
			return []string{result.Todo.UID}, nil
		},
		MessageFunc: func(affectedUIDs []string, todos []*models.IDMTodo) string {
			if len(todos) == 0 {
				return ""
			}
			return fmt.Sprintf("Added todo: %s", todos[0].PositionPath)
		},
	},

	"complete": {
		Name:               "complete",
		Aliases:            []string{"c"},
		Type:               CommandTypeCore,
		Description:        "Mark todos as complete",
		UsesPositionalArgs: true,
		AcceptsMultiple:    true,
		MutateFunc: func(args []string, opts map[string]interface{}) ([]string, error) {
			collectionPath := ""
			if v, ok := opts["collectionPath"].(string); ok {
				collectionPath = v
			}
			
			// Complete each position
			var allUIDs []string
			for _, pos := range args {
				result, err := Complete(pos, CompleteOptions{
					CollectionPath: collectionPath,
				})
				if err != nil {
					return nil, err
				}
				allUIDs = append(allUIDs, result.Todo.UID)
			}
			return allUIDs, nil
		},
		MessageFunc: func(affectedUIDs []string, todos []*models.IDMTodo) string {
			if len(todos) == 0 {
				return "No todos completed"
			}

			positions := make([]string, len(todos))
			for i, todo := range todos {
				positions[i] = todo.PositionPath
			}

			word := "todo"
			if len(todos) > 1 {
				word = "todos"
			}
			return fmt.Sprintf("Completed %s: %s", word, strings.Join(positions, ", "))
		},
	},

	"edit": {
		Name:               "edit",
		Aliases:            []string{"e", "modify"},
		Type:               CommandTypeCore,
		Description:        "Edit the text of an existing todo",
		UsesPositionalArgs: true,
		AcceptsMultiple:    false,
		MutateFunc: func(args []string, opts map[string]interface{}) ([]string, error) {
			collectionPath := ""
			if v, ok := opts["collectionPath"].(string); ok {
				collectionPath = v
			}
			
			position := args[0]
			text := ""
			if len(args) > 1 {
				text = args[1]
			}

			result, err := Modify(position, text, ModifyOptions{
				CollectionPath: collectionPath,
			})
			if err != nil {
				return nil, err
			}
			return []string{result.Todo.UID}, nil
		},
		MessageFunc: func(affectedUIDs []string, todos []*models.IDMTodo) string {
			if len(todos) == 0 {
				return ""
			}
			return fmt.Sprintf("Modified todo: %s", todos[0].PositionPath)
		},
	},

	"reopen": {
		Name:               "reopen",
		Aliases:            []string{"o"},
		Type:               CommandTypeCore,
		Description:        "Mark todos as pending",
		UsesPositionalArgs: true,
		AcceptsMultiple:    true,
		MutateFunc: func(args []string, opts map[string]interface{}) ([]string, error) {
			collectionPath := ""
			if v, ok := opts["collectionPath"].(string); ok {
				collectionPath = v
			}
			
			// Reopen each position
			var allUIDs []string
			for _, pos := range args {
				result, err := Reopen(pos, ReopenOptions{
					CollectionPath: collectionPath,
				})
				if err != nil {
					return nil, err
				}
				allUIDs = append(allUIDs, result.Todo.UID)
			}
			return allUIDs, nil
		},
		MessageFunc: func(affectedUIDs []string, todos []*models.IDMTodo) string {
			if len(todos) == 0 {
				return "No todos reopened"
			}

			positions := make([]string, len(todos))
			for i, todo := range todos {
				positions[i] = todo.PositionPath
			}

			word := "todo"
			if len(todos) > 1 {
				word = "todos"
			}
			return fmt.Sprintf("Reopened %s: %s", word, strings.Join(positions, ", "))
		},
	},

	"move": {
		Name:               "move",
		Aliases:            []string{"m"},
		Type:               CommandTypeExtra,
		Description:        "Move a todo to a different parent",
		UsesPositionalArgs: true,
		AcceptsMultiple:    false,
		MutateFunc: func(args []string, opts map[string]interface{}) ([]string, error) {
			collectionPath := ""
			if v, ok := opts["collectionPath"].(string); ok {
				collectionPath = v
			}
			
			fromPosition := args[0]
			toPosition := args[1]

			result, err := Move(fromPosition, toPosition, MoveOptions{
				CollectionPath: collectionPath,
			})
			if err != nil {
				return nil, err
			}
			return []string{result.Todo.UID}, nil
		},
		MessageFunc: func(affectedUIDs []string, todos []*models.IDMTodo) string {
			if len(todos) == 0 {
				return ""
			}
			return fmt.Sprintf("Moved todo: %s", todos[0].PositionPath)
		},
	},

	"clean": {
		Name:               "clean",
		Aliases:            []string{},
		Type:               CommandTypeMisc,
		Description:        "Remove finished todos",
		UsesPositionalArgs: false,
		AcceptsMultiple:    false,
		MutateFunc: func(args []string, opts map[string]interface{}) ([]string, error) {
			collectionPath := ""
			if v, ok := opts["collectionPath"].(string); ok {
				collectionPath = v
			}
			
			result, err := Clean(CleanOptions{
				CollectionPath: collectionPath,
			})
			if err != nil {
				return nil, err
			}

			uids := make([]string, len(result.RemovedTodos))
			for i, todo := range result.RemovedTodos {
				uids[i] = todo.UID
			}
			return uids, nil
		},
		MessageFunc: func(affectedUIDs []string, todos []*models.IDMTodo) string {
			if len(affectedUIDs) == 0 {
				return "No finished todos to clean"
			}

			// Show first 7 chars of UIDs for cleaned todos
			shortUIDs := make([]string, len(affectedUIDs))
			for i, uid := range affectedUIDs {
				if len(uid) > 7 {
					shortUIDs[i] = uid[:7]
				} else {
					shortUIDs[i] = uid
				}
			}

			word := "todo"
			if len(affectedUIDs) > 1 {
				word = "todos"
			}
			return fmt.Sprintf("Cleaned %s: %s", word, strings.Join(shortUIDs, ", "))
		},
	},

	"list": {
		Name:               "list",
		Aliases:            []string{"ls"},
		Type:               CommandTypeExtra,
		Description:        "List all todos",
		UsesPositionalArgs: false,
		AcceptsMultiple:    false,
		MutateFunc:         nil, // Display only
		FilterFunc: func(todos []*models.IDMTodo, opts map[string]interface{}) []*models.IDMTodo {
			showDone, _ := opts["showDone"].(bool)
			showAll, _ := opts["showAll"].(bool)

			if showAll {
				return todos
			}

			var filtered []*models.IDMTodo
			for _, todo := range todos {
				if showDone && todo.IsComplete() {
					filtered = append(filtered, todo)
				} else if !showDone && !todo.IsComplete() {
					filtered = append(filtered, todo)
				}
			}
			return filtered
		},
		MessageFunc: nil, // No message for list
	},

	"search": {
		Name:               "search",
		Aliases:            []string{"s"},
		Type:               CommandTypeExtra,
		Description:        "Search for todos",
		UsesPositionalArgs: true,
		AcceptsMultiple:    false,
		MutateFunc:         nil, // Display only
		FilterFunc: func(todos []*models.IDMTodo, opts map[string]interface{}) []*models.IDMTodo {
			query, _ := opts["query"].(string)
			if query == "" {
				return todos
			}

			query = strings.ToLower(query)
			var filtered []*models.IDMTodo
			for _, todo := range todos {
				if strings.Contains(strings.ToLower(todo.Text), query) {
					filtered = append(filtered, todo)
				}
			}
			return filtered
		},
		MessageFunc: func(affectedUIDs []string, todos []*models.IDMTodo) string {
			matchCount := len(todos)
			if matchCount == 0 {
				return "No matches found"
			}
			word := "match"
			if matchCount > 1 {
				word = "matches"
			}
			return fmt.Sprintf("Found %d %s", matchCount, word)
		},
	},

	// Utility commands (these don't use the unified execution model)
	"init": {
		Name:        "init",
		Aliases:     []string{"i"},
		Type:        CommandTypeMisc,
		Description: "Initialize a new todo collection",
	},

	"datapath": {
		Name:        "datapath",
		Aliases:     []string{"path"},
		Type:        CommandTypeMisc,
		Description: "Show the path to the todo data file",
	},

	"formats": {
		Name:        "formats",
		Aliases:     []string{},
		Type:        CommandTypeMisc,
		Description: "List available output formats",
	},
}

// GetCommand retrieves a command by name or alias
func GetCommand(name string) (*CommandDef, bool) {
	// Check direct name match
	if cmd, ok := CommandRegistry[name]; ok {
		return cmd, true
	}

	// Check aliases
	for _, cmd := range CommandRegistry {
		for _, alias := range cmd.Aliases {
			if alias == name {
				return cmd, true
			}
		}
	}

	return nil, false
}