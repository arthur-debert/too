package too

import (
	"fmt"
	"strings"

	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/parser"
	"github.com/arthur-debert/too/pkg/too/store"
)

// AttributeType represents the type of attribute to mutate
type AttributeType string

const (
	AttributeCompletion AttributeType = "completion"
	AttributeText       AttributeType = "text"
	AttributeParent     AttributeType = "parent"
)

// CommandEngine provides unified command execution
type CommandEngine struct {
	manager *store.PureIDMManager
}

// NewCommandEngine creates a new command engine
func NewCommandEngine(collectionPath string) (*CommandEngine, error) {
	idmStore := store.NewIDMStore(collectionPath)
	manager, err := store.NewPureIDMManager(idmStore, collectionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create manager: %w", err)
	}
	
	return &CommandEngine{
		manager: manager,
	}, nil
}

// MutateAttribute changes a single attribute on a todo
func (e *CommandEngine) MutateAttribute(ref string, attr AttributeType, value interface{}) (string, error) {
	// Resolve reference to UID
	uid, err := e.resolveReference(ref)
	if err != nil {
		return "", err
	}
	
	// Apply mutation
	switch attr {
	case AttributeCompletion:
		status := value.(string)
		if err := e.manager.SetStatus(uid, "completion", status); err != nil {
			return "", fmt.Errorf("failed to set completion status: %w", err)
		}
		
	case AttributeText:
		text := value.(string)
		todo := e.manager.GetTodoByUID(uid)
		if todo == nil {
			return "", fmt.Errorf("todo not found: %s", uid)
		}
		todo.Text = text
		todo.SetModified()
		
	case AttributeParent:
		parentUID := store.RootScope
		if parentRef := value.(string); parentRef != "" && parentRef != store.RootScope {
			parentUID, err = e.resolveReference(parentRef)
			if err != nil {
				return "", fmt.Errorf("failed to resolve parent: %w", err)
			}
		}
		// Get current parent for Move method
		todo := e.manager.GetTodoByUID(uid)
		if todo == nil {
			return "", fmt.Errorf("todo not found: %s", uid)
		}
		oldParentUID := todo.ParentID
		if oldParentUID == "" {
			oldParentUID = store.RootScope
		}
		
		if err := e.manager.Move(uid, oldParentUID, parentUID); err != nil {
			return "", fmt.Errorf("failed to move todo: %w", err)
		}
		
	default:
		return "", fmt.Errorf("unknown attribute: %s", attr)
	}
	
	return uid, nil
}

// Add creates a new todo
func (e *CommandEngine) Add(text string, parentRef string) (string, error) {
	parentUID := store.RootScope
	if parentRef != "" {
		var err error
		parentUID, err = e.resolveReference(parentRef)
		if err != nil {
			return "", fmt.Errorf("failed to resolve parent: %w", err)
		}
	}
	
	uid, err := e.manager.Add(parentUID, text)
	if err != nil {
		return "", fmt.Errorf("failed to add todo: %w", err)
	}
	
	return uid, nil
}

// Clean removes completed todos
func (e *CommandEngine) Clean() ([]*models.IDMTodo, error) {
	// Use the manager's integrated clean operation
	removedTodos, _, err := e.manager.CleanFinishedTodos()
	if err != nil {
		return nil, fmt.Errorf("failed to clean finished todos: %w", err)
	}
	
	return removedTodos, nil
}

// Save persists changes
func (e *CommandEngine) Save() error {
	return e.manager.Save()
}

// GetTodos returns todos with optional filtering
func (e *CommandEngine) GetTodos(filter FilterFunc) []*models.IDMTodo {
	// Get all todos (including completed) for filtering
	todos := e.manager.ListAll()
	
	// Apply filter first
	if filter != nil {
		todos = filter(todos)
	}
	
	// Attach position paths based on what's being displayed
	// If we're showing only active todos, use active-only paths
	showingOnlyActive := true
	for _, todo := range todos {
		if todo.GetStatus() == models.StatusDone {
			showingOnlyActive = false
			break
		}
	}
	
	if showingOnlyActive {
		e.manager.AttachActiveOnlyPositionPaths(todos)
	} else {
		e.manager.AttachPositionPaths(todos)
	}
	
	return todos
}

// GetStats returns todo counts
func (e *CommandEngine) GetStats() (total, done int) {
	return e.manager.CountTodos()
}

// GetTodoByUID returns a specific todo
func (e *CommandEngine) GetTodoByUID(uid string) *models.IDMTodo {
	return e.manager.GetTodoByUID(uid)
}

// resolveReference converts a position path or short ID to UID
func (e *CommandEngine) resolveReference(ref string) (string, error) {
	// Try as position path first
	if parser.IsPositionPath(ref) {
		uid, err := e.manager.ResolvePositionPath(store.RootScope, ref)
		if err == nil {
			return uid, nil
		}
	}
	
	// Try as short ID
	todo, err := e.manager.GetTodoByShortID(ref)
	if err != nil {
		return "", fmt.Errorf("no todo found matching '%s': %w", ref, err)
	}
	
	return todo.UID, nil
}

// FilterFunc filters a list of todos
type FilterFunc func([]*models.IDMTodo) []*models.IDMTodo

// FilterByStatus returns todos with the given status
func FilterByStatus(status string) FilterFunc {
	return func(todos []*models.IDMTodo) []*models.IDMTodo {
		var filtered []*models.IDMTodo
		for _, todo := range todos {
			if string(todo.GetStatus()) == status {
				filtered = append(filtered, todo)
			}
		}
		return filtered
	}
}

// FilterByQuery returns todos containing the query text
func FilterByQuery(query string) FilterFunc {
	lowerQuery := strings.ToLower(query)
	return func(todos []*models.IDMTodo) []*models.IDMTodo {
		var filtered []*models.IDMTodo
		for _, todo := range todos {
			if strings.Contains(strings.ToLower(todo.Text), lowerQuery) {
				filtered = append(filtered, todo)
			}
		}
		return filtered
	}
}

// FilterAll returns all todos (no filtering)
func FilterAll() FilterFunc {
	return func(todos []*models.IDMTodo) []*models.IDMTodo {
		return todos
	}
}

// FilterPending returns only pending todos
func FilterPending() FilterFunc {
	return FilterByStatus(string(models.StatusPending))
}

// FilterDone returns only done todos  
func FilterDone() FilterFunc {
	return FilterByStatus(string(models.StatusDone))
}


// truncateText truncates text to maxLen with ellipsis
func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen-3] + "..."
}