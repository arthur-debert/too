package add

import (
	"fmt"

	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/store"
)

// UnifiedResult contains the result of the add command that works with both manager types.
type UnifiedResult struct {
	UID          string
	Text         string
	ParentID     string
	PositionPath string
	Mode         string
	AllTodos     interface{} // []*models.Todo or []*models.IDMTodo
	TotalCount   int
	DoneCount    int
	IsPureIDM    bool
}

// ExecuteUnified adds a new todo using the factory pattern.
// It automatically selects the appropriate manager based on storage format.
func ExecuteUnified(text string, opts Options) (*UnifiedResult, error) {
	if text == "" {
		return nil, fmt.Errorf("todo text cannot be empty")
	}

	// Create workflow manager using factory
	manager, err := store.CreateWorkflowManager(opts.CollectionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create workflow manager: %w", err)
	}

	// Resolve parent if specified
	var parentUID string = store.RootScope
	if opts.ParentPath != "" {
		uid, err := manager.ResolvePositionPath(store.RootScope, opts.ParentPath)
		if err != nil {
			return nil, fmt.Errorf("parent todo not found: %w", err)
		}
		parentUID = uid
	}

	// Add the todo
	newUID, err := manager.Add(parentUID, text)
	if err != nil {
		return nil, fmt.Errorf("failed to add todo: %w", err)
	}

	// Save the changes
	if err := manager.Save(); err != nil {
		return nil, fmt.Errorf("failed to save collection: %w", err)
	}

	// Get the created todo
	todoInterface := manager.GetTodoByID(newUID)
	if todoInterface == nil {
		return nil, fmt.Errorf("todo with UID %s not found after creation", newUID)
	}

	// Get the position path of the newly created todo
	positionPath, err := manager.GetPositionPath(store.RootScope, newUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get position path: %w", err)
	}

	// Build unified result
	result := &UnifiedResult{
		UID:          newUID,
		PositionPath: positionPath,
		Mode:         opts.Mode,
		IsPureIDM:    manager.IsPureIDM(),
	}

	// Extract text and parent ID based on type
	if manager.IsPureIDM() {
		todo := todoInterface.(*models.IDMTodo)
		result.Text = todo.Text
		result.ParentID = todo.ParentID
	} else {
		todo := todoInterface.(*models.Todo)
		result.Text = todo.Text
		result.ParentID = todo.ParentID
	}

	// If in long mode, get additional data
	if opts.Mode == "long" {
		result.AllTodos = manager.ListActive()
		result.TotalCount, result.DoneCount = manager.CountTodos()
	}

	return result, nil
}

// ConvertUnifiedToResult converts a UnifiedResult to the traditional Result format.
func ConvertUnifiedToResult(unified *UnifiedResult) *Result {
	result := &Result{
		PositionPath: unified.PositionPath,
		Mode:         unified.Mode,
		TotalCount:   unified.TotalCount,
		DoneCount:    unified.DoneCount,
	}

	// Create a Todo from the unified result
	result.Todo = &models.Todo{
		ID:       unified.UID,
		ParentID: unified.ParentID,
		Text:     unified.Text,
		Items:    []*models.Todo{}, // Empty - hierarchy managed by IDM
	}

	// Convert AllTodos if present
	if unified.AllTodos != nil {
		if unified.IsPureIDM {
			// Convert from IDMTodo slice
			idmTodos := unified.AllTodos.([]*models.IDMTodo)
			result.AllTodos = make([]*models.Todo, len(idmTodos))
			for i, idmTodo := range idmTodos {
				result.AllTodos[i] = convertIDMTodoToTodo(idmTodo)
			}
		} else {
			// Already in Todo format
			result.AllTodos = unified.AllTodos.([]*models.Todo)
		}
	}

	return result
}