package too

import (
	"fmt"
	"strings"

	"github.com/arthur-debert/too/pkg/too/models"
)

// UnifiedCommand represents a simplified command definition
type UnifiedCommand struct {
	Name        string
	Aliases     []string
	Type        CommandType
	Description string
	
	// Execution configuration
	Attribute       AttributeType      // What attribute to change (if any)
	AttributeValue  interface{}        // Fixed value (e.g., StatusDone) or nil if from args
	RequiresRef     bool              // Does it need a todo reference?
	AcceptsMultiple bool              // Can operate on multiple todos?
	RequiresText    bool              // Does it need text input (edit, add)?
	
	// Optional functions
	ValidateFunc    func(args []string, opts map[string]interface{}) error
	GetFilterFunc   func(opts map[string]interface{}) FilterFunc
	GetMessageFunc  func(affectedCount int, affectedTodos []*models.IDMTodo) string
}

// UnifiedCommands defines all commands in a declarative way
var UnifiedCommands = map[string]*UnifiedCommand{
	// Status changers
	"complete": {
		Name:            "complete",
		Aliases:         []string{"c"},
		Type:            CommandTypeCore,
		Description:     "Mark todos as complete",
		Attribute:       AttributeCompletion,
		AttributeValue:  string(models.StatusDone),
		RequiresRef:     true,
		AcceptsMultiple: true,
		GetMessageFunc: func(count int, todos []*models.IDMTodo) string {
			if count == 0 {
				return "No todos completed"
			}
			return formatMessage("Completed", todos)
		},
	},
	
	"reopen": {
		Name:            "reopen",
		Aliases:         []string{"o"},
		Type:            CommandTypeCore,
		Description:     "Mark todos as pending",
		Attribute:       AttributeCompletion,
		AttributeValue:  string(models.StatusPending),
		RequiresRef:     true,
		AcceptsMultiple: true,
		GetMessageFunc: func(count int, todos []*models.IDMTodo) string {
			if count == 0 {
				return "No todos reopened"
			}
			return formatMessage("Reopened", todos)
		},
	},
	
	// Attribute editors
	"edit": {
		Name:         "edit",
		Aliases:      []string{"e", "modify"},
		Type:         CommandTypeCore,
		Description:  "Edit the text of an existing todo",
		Attribute:    AttributeText,
		RequiresRef:  true,
		RequiresText: true,
		ValidateFunc: func(args []string, opts map[string]interface{}) error {
			if len(args) < 2 {
				return fmt.Errorf("edit requires position and new text")
			}
			return nil
		},
		GetMessageFunc: func(count int, todos []*models.IDMTodo) string {
			if count == 0 {
				return ""
			}
			return fmt.Sprintf("Modified todo: %s", todos[0].PositionPath)
		},
	},
	
	"move": {
		Name:        "move",
		Aliases:     []string{"m"},
		Type:        CommandTypeExtra,
		Description: "Move a todo to a different parent",
		Attribute:   AttributeParent,
		RequiresRef: true,
		ValidateFunc: func(args []string, opts map[string]interface{}) error {
			if len(args) < 2 {
				return fmt.Errorf("move requires source and destination")
			}
			return nil
		},
		GetMessageFunc: func(count int, todos []*models.IDMTodo) string {
			if count == 0 {
				return ""
			}
			return fmt.Sprintf("Moved todo: %s", todos[0].PositionPath)
		},
	},
	
	// Create/Delete
	"add": {
		Name:         "add",
		Aliases:      []string{"a", "new", "create"},
		Type:         CommandTypeCore,
		Description:  "Add a new todo",
		RequiresText: true,
		ValidateFunc: func(args []string, opts map[string]interface{}) error {
			if len(args) < 1 || args[0] == "" {
				return fmt.Errorf("add requires todo text")
			}
			return nil
		},
		GetMessageFunc: func(count int, todos []*models.IDMTodo) string {
			if count == 0 {
				return ""
			}
			return fmt.Sprintf("Added todo: %s", todos[0].PositionPath)
		},
	},
	
	"clean": {
		Name:        "clean",
		Aliases:     []string{},
		Type:        CommandTypeMisc,
		Description: "Remove finished todos",
		GetMessageFunc: func(count int, todos []*models.IDMTodo) string {
			if count == 0 {
				return "No finished todos to clean"
			}
			word := "todo"
			if count > 1 {
				word = "todos"
			}
			return fmt.Sprintf("Cleaned %d %s", count, word)
		},
	},
	
	// Listers
	"list": {
		Name:        "list",
		Aliases:     []string{"ls"},
		Type:        CommandTypeExtra,
		Description: "List all todos",
		GetFilterFunc: func(opts map[string]interface{}) FilterFunc {
			if done, _ := opts["done"].(bool); done {
				return FilterDone()
			}
			if all, _ := opts["all"].(bool); all {
				return FilterAll()
			}
			return FilterPending()
		},
	},
	
	"search": {
		Name:        "search",
		Aliases:     []string{"s"},
		Type:        CommandTypeExtra,
		Description: "Search for todos",
		ValidateFunc: func(args []string, opts map[string]interface{}) error {
			if len(args) < 1 {
				return fmt.Errorf("search requires a query")
			}
			return nil
		},
		GetFilterFunc: func(opts map[string]interface{}) FilterFunc {
			query, _ := opts["query"].(string)
			if query == "" {
				return FilterAll()
			}
			return FilterByQuery(query)
		},
		GetMessageFunc: func(count int, todos []*models.IDMTodo) string {
			word := "match"
			if count != 1 {
				word = "matches"
			}
			return fmt.Sprintf("Found %d %s", count, word)
		},
	},
}

// ExecuteUnifiedCommand executes a command using the unified engine
func ExecuteUnifiedCommand(cmdName string, args []string, opts map[string]interface{}) (*ChangeResult, error) {
	// Find command
	cmd, ok := UnifiedCommands[cmdName]
	if !ok {
		// Check aliases
		for name, c := range UnifiedCommands {
			for _, alias := range c.Aliases {
				if alias == cmdName {
					cmd = c
					cmdName = name
					break
				}
			}
			if cmd != nil {
				break
			}
		}
		if cmd == nil {
			return nil, fmt.Errorf("unknown command: %s", cmdName)
		}
	}
	
	// Validate args
	if cmd.ValidateFunc != nil {
		if err := cmd.ValidateFunc(args, opts); err != nil {
			return nil, err
		}
	}
	
	// Create engine
	collectionPath, _ := opts["collectionPath"].(string)
	engine, err := NewCommandEngine(collectionPath)
	if err != nil {
		return nil, err
	}
	
	// Execute command
	var affectedUIDs []string
	
	switch cmdName {
	case "add":
		// Special case: create new todo
		text := args[0]
		parentRef := ""
		if p, ok := opts["parent"].(string); ok {
			parentRef = p
		}
		
		uid, err := engine.Add(text, parentRef)
		if err != nil {
			return nil, err
		}
		affectedUIDs = []string{uid}
		
	case "clean":
		// Special case: remove completed todos
		affectedUIDs, err = engine.Clean()
		if err != nil {
			return nil, err
		}
		
	default:
		// Standard attribute mutation
		if cmd.Attribute != "" {
			if cmd.AcceptsMultiple && len(args) > 1 {
				// Multiple refs (complete, reopen)
				for _, ref := range args {
					uid, err := engine.MutateAttribute(ref, cmd.Attribute, cmd.AttributeValue)
					if err != nil {
						return nil, err
					}
					affectedUIDs = append(affectedUIDs, uid)
				}
			} else if cmd.RequiresRef {
				// Single ref with possible value
				ref := args[0]
				value := cmd.AttributeValue
				
				// Get value from args if needed
				if value == nil && len(args) > 1 {
					if cmd.Attribute == AttributeText {
						value = args[1]
					} else if cmd.Attribute == AttributeParent {
						value = args[1]
					}
				}
				
				uid, err := engine.MutateAttribute(ref, cmd.Attribute, value)
				if err != nil {
					return nil, err
				}
				affectedUIDs = []string{uid}
			}
		}
	}
	
	// Save changes
	if len(affectedUIDs) > 0 {
		if err := engine.Save(); err != nil {
			return nil, fmt.Errorf("failed to save: %w", err)
		}
	}
	
	// Get todos for display
	var filter FilterFunc
	if cmd.GetFilterFunc != nil {
		filter = cmd.GetFilterFunc(opts)
	}
	todos := engine.GetTodos(filter)
	
	// Get affected todos from all todos (not filtered)
	var affectedTodos []*models.IDMTodo
	if len(affectedUIDs) > 0 {
		allTodos := engine.GetTodos(nil)  // Get all todos
		affectedMap := make(map[string]bool)
		for _, uid := range affectedUIDs {
			affectedMap[uid] = true
		}
		for _, todo := range allTodos {
			if affectedMap[todo.UID] {
				affectedTodos = append(affectedTodos, todo)
			}
		}
	}
	
	// Get stats
	totalCount, doneCount := engine.GetStats()
	
	// Generate message
	message := ""
	if cmd.GetMessageFunc != nil {
		// For list/search commands, use the filtered todos count
		messageCount := len(affectedUIDs)
		if cmdName == "search" || cmdName == "list" {
			messageCount = len(todos)
		}
		message = cmd.GetMessageFunc(messageCount, affectedTodos)
	}
	
	return NewChangeResult(
		cmdName,
		message,
		affectedTodos,
		todos,
		totalCount,
		doneCount,
	), nil
}

// formatMessage formats a standard action message
func formatMessage(action string, todos []*models.IDMTodo) string {
	if len(todos) == 0 {
		return ""
	}
	
	positions := make([]string, len(todos))
	for i, todo := range todos {
		positions[i] = todo.PositionPath
	}
	
	word := "todo"
	if len(todos) > 1 {
		word = "todos"
	}
	
	return fmt.Sprintf("%s %s: %s", action, word, strings.Join(positions, ", "))
}