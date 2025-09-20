package too

import (
	"fmt"
	"strings"

	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/rs/zerolog/log"
)

// UnifiedCommand represents a simplified command definition
type UnifiedCommand struct {
	Name        string
	Aliases     []string
	Type        models.CommandType
	Description string
	
	// Execution configuration
	Attribute       models.AttributeType      // What attribute to change (if any)
	AttributeValue  interface{}        // Fixed value (e.g., StatusDone) or nil if from args
	RequiresRef     bool              // Does it need a todo reference?
	AcceptsMultiple bool              // Can operate on multiple todos?
	RequiresText    bool              // Does it need text input (edit, add)?
	
	// Optional functions
	ValidateFunc    func(args []string, opts map[string]interface{}) error
	GetFilterFunc   func(opts map[string]interface{}) FilterFunc
	GetMessageFunc  func(affectedCount int, affectedTodos []*models.Todo) string
}

// UnifiedCommands defines all commands in a declarative way
var UnifiedCommands = map[string]*UnifiedCommand{
	// Status changers
	"complete": {
		Name:            "complete",
		Aliases:         []string{"c"},
		Type:            models.CommandTypeCore,
		Description:     "Mark todos as complete",
		Attribute:       models.AttributeCompletion,
		AttributeValue:  string(models.StatusDone),
		RequiresRef:     true,
		AcceptsMultiple: true,
		GetMessageFunc: func(count int, todos []*models.Todo) string {
			return "" // Visual highlight is sufficient
		},
	},
	
	"reopen": {
		Name:            "reopen",
		Aliases:         []string{"o"},
		Type:            models.CommandTypeCore,
		Description:     "Mark todos as pending",
		Attribute:       models.AttributeCompletion,
		AttributeValue:  string(models.StatusPending),
		RequiresRef:     true,
		AcceptsMultiple: true,
		GetMessageFunc: func(count int, todos []*models.Todo) string {
			return "" // Visual highlight is sufficient
		},
	},
	
	// Attribute editors
	"edit": {
		Name:         "edit",
		Aliases:      []string{"e", "modify"},
		Type:         models.CommandTypeCore,
		Description:  "Edit the text of an existing todo",
		Attribute:    models.AttributeText,
		RequiresRef:  true,
		RequiresText: true,
		ValidateFunc: func(args []string, opts map[string]interface{}) error {
			if len(args) < 2 {
				return fmt.Errorf("edit requires position and new text")
			}
			return nil
		},
		GetMessageFunc: func(count int, todos []*models.Todo) string {
			return "" // Visual highlight is sufficient
		},
	},
	
	"move": {
		Name:        "move",
		Aliases:     []string{"m"},
		Type:        models.CommandTypeExtra,
		Description: "Move a todo to a different parent",
		Attribute:   models.AttributeParent,
		RequiresRef: true,
		ValidateFunc: func(args []string, opts map[string]interface{}) error {
			if len(args) < 2 {
				return fmt.Errorf("move requires source and destination")
			}
			return nil
		},
		GetMessageFunc: func(count int, todos []*models.Todo) string {
			return "" // Visual highlight is sufficient
		},
	},
	
	// Create/Delete
	"add": {
		Name:         "add",
		Aliases:      []string{"a", "new", "create"},
		Type:         models.CommandTypeCore,
		Description:  "Add a new todo",
		RequiresText: true,
		ValidateFunc: func(args []string, opts map[string]interface{}) error {
			if len(args) < 1 || args[0] == "" {
				return fmt.Errorf("add requires todo text")
			}
			return nil
		},
		GetMessageFunc: func(count int, todos []*models.Todo) string {
			return "" // Visual highlight is sufficient
		},
	},
	
	"clean": {
		Name:        "clean",
		Aliases:     []string{},
		Type:        models.CommandTypeMisc,
		Description: "Remove finished todos",
		GetMessageFunc: func(count int, todos []*models.Todo) string {
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
		Type:        models.CommandTypeExtra,
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
		Type:        models.CommandTypeExtra,
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
		GetMessageFunc: func(count int, todos []*models.Todo) string {
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
	
	// Create engine - use NanoEngine instead
	collectionPath, _ := opts["collectionPath"].(string)
	engine, err := NewNanoEngine(collectionPath)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := engine.Close(); err != nil {
			log.Debug().Err(err).Msg("error closing engine during cleanup")
		}
	}()
	
	// Execute command
	var affectedUIDs []string
	var affectedTodos []*models.Todo
	var todos []*models.Todo
	
	switch cmdName {
	case "add":
		// Special case: create new todo
		text := args[0]
		parentRef := ""
		if p, ok := opts["parent"].(string); ok {
			parentRef = p
		}
		
		todo, err := engine.Add(text, &parentRef)
		if err != nil {
			return nil, err
		}
		affectedUIDs = []string{todo.UID}
		affectedTodos = []*models.Todo{todo}
		
	case "clean":
		// Special case: remove completed todos
		var removedTodos []*models.Todo
		removedTodos, err = engine.Clean()
		if err != nil {
			return nil, err
		}
		// Store removed todos directly as affected todos
		affectedTodos = removedTodos
		// Also collect UIDs for consistency
		for _, todo := range removedTodos {
			affectedUIDs = append(affectedUIDs, todo.UID)
		}
		
	case "search":
		// Special case: use native search
		query := strings.Join(args, " ")
		showAll := false
		if all, ok := opts["all"].(bool); ok {
			showAll = all
		} else if done, ok := opts["done"].(bool); ok && done {
			showAll = true
		}
		
		searchResults, err := engine.Search(query, showAll)
		if err != nil {
			return nil, err
		}
		
		// For search, the results are the todos to display
		todos = searchResults
		// No affected todos for search
		affectedTodos = nil
		
	case "list":
		// Special case: use filter function for list
		filter := cmd.GetFilterFunc(opts)
		listResults, err := engine.GetTodos(filter)
		if err != nil {
			return nil, err
		}
		todos = listResults
		// No affected todos for list
		affectedTodos = nil
		
	default:
		// Standard attribute mutation
		if cmd.Attribute != "" {
			if cmd.AcceptsMultiple {
				// Multiple refs (complete, reopen)
				// First, resolve all IDs to UUIDs before any mutations
				// This is important because IDs can shift after each mutation
				resolvedRefs := make(map[string]string) // ref -> uuid
				for _, ref := range args {
					uuid, err := engine.ResolveReference(ref)
					if err != nil {
						return nil, fmt.Errorf("failed to resolve reference '%s': %w", ref, err)
					}
					resolvedRefs[ref] = uuid
				}
				
				// Now mutate using the resolved UUIDs
				for _, ref := range args {
					uuid := resolvedRefs[ref]
					uid, err := engine.MutateAttributeByUUID(uuid, cmd.Attribute, cmd.AttributeValue)
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
					if cmd.Attribute == models.AttributeText {
						value = args[1]
					} else if cmd.Attribute == models.AttributeParent {
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
	
	// Get todos for display (skip for search/list as they're already populated)
	if cmdName != "search" && cmdName != "list" && todos == nil {
		var filter FilterFunc
		if cmd.GetFilterFunc != nil {
			filter = cmd.GetFilterFunc(opts)
		}
		todos, err = engine.GetTodos(filter)
		if err != nil {
			return nil, err
		}
	}
	
	// Get affected todos from all todos (not filtered)
	// Skip this for clean command as we already have the removed todos
	if cmdName != "clean" && len(affectedUIDs) > 0 {
		allTodos, err := engine.GetTodos(nil)  // Get all todos
		if err != nil {
			return nil, err
		}
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
func formatMessage(action string, todos []*models.Todo) string {
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