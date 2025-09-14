package too

import (
	"fmt"
	"strings"

	"github.com/arthur-debert/too/pkg/logging"
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/store"
	"github.com/rs/zerolog"
)

// NanoEngine is a simplified engine that uses nanostore
type NanoEngine struct {
	adapter *store.NanoStoreAdapter
	logger  zerolog.Logger
}

// NewNanoEngine creates a new engine instance
func NewNanoEngine(dataPath string) (*NanoEngine, error) {
	// Convert .json to .db
	if strings.HasSuffix(dataPath, ".json") {
		dataPath = strings.TrimSuffix(dataPath, ".json") + ".db"
	}

	adapter, err := store.NewNanoStoreAdapter(dataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create nanostore adapter: %w", err)
	}

	return &NanoEngine{
		adapter: adapter,
		logger:  logging.GetLogger("too.engine"),
	}, nil
}

// Close releases resources
func (e *NanoEngine) Close() error {
	if e.adapter != nil {
		return e.adapter.Close()
	}
	return nil
}

// Add creates a new todo
func (e *NanoEngine) Add(text string, parentID *string) (*models.Todo, error) {
	todo, err := e.adapter.Add(text, parentID)
	if err != nil {
		return nil, err
	}

	e.logger.Debug().
		Str("uid", todo.UID).
		Str("position", todo.PositionPath).
		Str("text", todo.Text).
		Msg("added todo")

	return todo, nil
}

// Complete marks a todo as done
func (e *NanoEngine) Complete(positionPath string) (*models.Todo, error) {
	// Get the todo before completing to return it
	uuid, err := e.adapter.ResolvePositionPath(positionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve position path: %w", err)
	}

	todo, err := e.adapter.GetByUUID(uuid)
	if err != nil {
		return nil, fmt.Errorf("failed to get todo: %w", err)
	}

	// Complete it
	if err := e.adapter.Complete(positionPath); err != nil {
		return nil, err
	}

	// Update status in returned todo
	todo.Statuses["completion"] = "done"

	e.logger.Debug().
		Str("uid", todo.UID).
		Str("position", positionPath).
		Msg("completed todo")

	return todo, nil
}

// Reopen marks a todo as pending
func (e *NanoEngine) Reopen(positionPath string) (*models.Todo, error) {
	// Get the todo before reopening to return it
	uuid, err := e.adapter.ResolvePositionPath(positionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve position path: %w", err)
	}

	todo, err := e.adapter.GetByUUID(uuid)
	if err != nil {
		return nil, fmt.Errorf("failed to get todo: %w", err)
	}

	// Reopen it
	if err := e.adapter.Reopen(positionPath); err != nil {
		return nil, err
	}

	// Update status in returned todo
	todo.Statuses["completion"] = "pending"

	e.logger.Debug().
		Str("uid", todo.UID).
		Str("position", positionPath).
		Msg("reopened todo")

	return todo, nil
}

// Update modifies a todo's text
func (e *NanoEngine) Update(positionPath string, text string) (*models.Todo, error) {
	if err := e.adapter.Update(positionPath, text); err != nil {
		return nil, err
	}

	// Get updated todo
	uuid, err := e.adapter.ResolvePositionPath(positionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve position path: %w", err)
	}

	todo, err := e.adapter.GetByUUID(uuid)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated todo: %w", err)
	}

	e.logger.Debug().
		Str("uid", todo.UID).
		Str("position", positionPath).
		Str("newText", text).
		Msg("updated todo")

	return todo, nil
}

// Move changes a todo's parent
func (e *NanoEngine) Move(positionPath string, newParentPath *string) (*models.Todo, string, error) {
	if err := e.adapter.Move(positionPath, newParentPath); err != nil {
		return nil, "", err
	}

	// Get updated todo with new position
	uuid, err := e.adapter.ResolvePositionPath(positionPath)
	if err != nil {
		// Todo might have a new position after move, try to get by UUID
		// This is a limitation - we need to list all to find the new position
		todos, err := e.adapter.List(true)
		if err != nil {
			return nil, "", fmt.Errorf("failed to find moved todo: %w", err)
		}
		
		// Find by matching text (not ideal but works for now)
		for _, t := range todos {
			if t.UID == uuid {
				return t, t.PositionPath, nil
			}
		}
		return nil, "", fmt.Errorf("moved todo not found")
	}

	todo, err := e.adapter.GetByUUID(uuid)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get moved todo: %w", err)
	}

	e.logger.Debug().
		Str("uid", todo.UID).
		Str("oldPosition", positionPath).
		Str("newPosition", todo.PositionPath).
		Msg("moved todo")

	return todo, todo.PositionPath, nil
}

// Delete removes a todo
func (e *NanoEngine) Delete(positionPath string, cascade bool) (int, error) {
	// Count before delete
	beforeCount := 0
	if cascade {
		todos, _ := e.adapter.List(true)
		uuid, _ := e.adapter.ResolvePositionPath(positionPath)
		beforeCount = e.countTodoAndChildren(todos, uuid)
	} else {
		beforeCount = 1
	}

	if err := e.adapter.Delete(positionPath, cascade); err != nil {
		return 0, err
	}

	e.logger.Debug().
		Str("position", positionPath).
		Bool("cascade", cascade).
		Int("deleted", beforeCount).
		Msg("deleted todo")

	return beforeCount, nil
}

// List returns todos
func (e *NanoEngine) List(showAll bool) ([]*models.Todo, error) {
	return e.adapter.List(showAll)
}

// Search finds todos matching query
func (e *NanoEngine) Search(query string, showAll bool) ([]*models.Todo, error) {
	return e.adapter.Search(query, showAll)
}

// ResolveReference converts a user-facing ID to UUID
func (e *NanoEngine) ResolveReference(ref string) (string, error) {
	return e.adapter.ResolvePositionPath(ref)
}

// MutateAttributeByUUID changes a single attribute on a todo by its UUID
func (e *NanoEngine) MutateAttributeByUUID(uuid string, attr models.AttributeType, value interface{}) (string, error) {
	// Apply mutation based on attribute type
	var err error
	switch attr {
	case models.AttributeCompletion:
		status := value.(string)
		// For UUID-based operations, we need to use the adapter directly
		// since Complete/Reopen methods expect user-facing IDs
		if status == string(models.StatusDone) {
			err = e.adapter.CompleteByUUID(uuid)
		} else {
			err = e.adapter.ReopenByUUID(uuid)
		}
	case models.AttributeText:
		text := value.(string)
		err = e.adapter.UpdateByUUID(uuid, text)
	case models.AttributeParent:
		newParent := value.(string)
		var parentPtr *string
		if newParent != "" {
			parentPtr = &newParent
		}
		err = e.adapter.MoveByUUID(uuid, parentPtr)
	default:
		return "", fmt.Errorf("unknown attribute: %s", attr)
	}

	if err != nil {
		return "", err
	}

	return uuid, nil
}

// MutateAttribute changes a single attribute on a todo
func (e *NanoEngine) MutateAttribute(ref string, attr models.AttributeType, value interface{}) (string, error) {
	// Resolve reference to UID
	uuid, err := e.adapter.ResolvePositionPath(ref)
	if err != nil {
		return "", fmt.Errorf("failed to resolve reference '%s': %w", ref, err)
	}

	// Apply mutation based on attribute type
	switch attr {
	case models.AttributeCompletion:
		status := value.(string)
		if status == string(models.StatusDone) {
			err = e.adapter.Complete(ref)
		} else {
			err = e.adapter.Reopen(ref)
		}
	case models.AttributeText:
		text := value.(string)
		err = e.adapter.Update(ref, text)
	case models.AttributeParent:
		newParent := value.(string)
		var parentPtr *string
		if newParent != "" {
			parentPtr = &newParent
		}
		err = e.adapter.Move(ref, parentPtr)
	default:
		return "", fmt.Errorf("unknown attribute: %s", attr)
	}

	if err != nil {
		return "", err
	}

	return uuid, nil
}

// GetStats returns todo counts
func (e *NanoEngine) GetStats() (total, done int) {
	todos, err := e.adapter.List(true)
	if err != nil {
		return 0, 0
	}
	
	total = len(todos)
	done = 0
	for _, todo := range todos {
		if todo.GetStatus() == models.StatusDone {
			done++
		}
	}
	
	return total, done
}

// Clean removes all completed todos
func (e *NanoEngine) Clean() ([]*models.Todo, error) {
	// Get all completed todos first (before deletion) to return them
	allTodos, err := e.adapter.List(true)
	if err != nil {
		return nil, err
	}

	var removedTodos []*models.Todo
	for _, todo := range allTodos {
		if todo.GetStatus() == models.StatusDone {
			removedTodos = append(removedTodos, todo)
		}
	}

	// Now delete all completed todos in one operation
	_, err = e.adapter.DeleteCompleted()
	if err != nil {
		return nil, fmt.Errorf("failed to delete completed todos: %w", err)
	}

	return removedTodos, nil
}

// GetTodos returns todos for display
func (e *NanoEngine) GetTodos(filter FilterFunc) ([]*models.Todo, error) {
	// Get all todos
	allTodos, err := e.adapter.List(true)
	if err != nil {
		return nil, err
	}

	// Apply filter if provided
	if filter != nil {
		var filtered []*models.Todo
		for _, todo := range allTodos {
			if filter([]*models.Todo{todo}) != nil && len(filter([]*models.Todo{todo})) > 0 {
				filtered = append(filtered, todo)
			}
		}
		return filtered, nil
	}

	return allTodos, nil
}

// Save is a no-op for nanostore (auto-saves)
func (e *NanoEngine) Save() error {
	// Nanostore automatically persists changes
	return nil
}

// GetTodoByUID retrieves a todo by its UID
func (e *NanoEngine) GetTodoByUID(uid string) (*models.Todo, error) {
	return e.adapter.GetByUUID(uid)
}

// countTodoAndChildren counts a todo and all its descendants
func (e *NanoEngine) countTodoAndChildren(todos []*models.Todo, uuid string) int {
	count := 1 // Count the todo itself
	
	// Find all children
	for _, todo := range todos {
		if todo.ParentID == uuid {
			count += e.countTodoAndChildren(todos, todo.UID)
		}
	}
	
	return count
}