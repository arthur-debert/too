package add

import (
	"fmt"

	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/store"
)

// IDMResult contains the result of the add command using pure IDM data structures.
type IDMResult struct {
	Todo         *models.IDMTodo
	PositionPath string             // Position path of the newly created todo (e.g., "1", "1.2")
	Mode         string             // Output mode passed from options
	AllTodos     []*models.IDMTodo  // All todos for long mode
	TotalCount   int                // Total count for long mode
	DoneCount    int                // Done count for long mode
}

// RunIDM executes the add command using the pure IDM manager.
func RunIDM(idmStore store.IDMStore, text string, opts Options) (*IDMResult, error) {
	// Create pure IDM workflow manager
	manager, err := store.NewPureIDMManager(idmStore, idmStore.Path())
	if err != nil {
		return nil, fmt.Errorf("failed to create pure IDM manager: %w", err)
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

	// Add the todo directly
	newUID, err := manager.Add(parentUID, text)
	if err != nil {
		return nil, fmt.Errorf("failed to add todo: %w", err)
	}

	// Save the changes
	if err := manager.Save(); err != nil {
		return nil, fmt.Errorf("failed to save collection: %w", err)
	}

	// Get the created todo using PureIDMManager method
	todo := manager.GetTodoByUID(newUID)
	if todo == nil {
		return nil, fmt.Errorf("todo with UID %s not found after creation", newUID)
	}

	// Get the position path of the newly created todo
	positionPath, err := manager.GetPositionPath(store.RootScope, newUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get position path: %w", err)
	}

	result := &IDMResult{
		Todo:         todo,
		PositionPath: positionPath,
		Mode:         opts.Mode,
	}

	// If in long mode, get additional data using manager's IDM-aware methods
	if opts.Mode == "long" {
		result.AllTodos = manager.ListActive().([]*models.IDMTodo)
		result.TotalCount, result.DoneCount = manager.CountTodos()
	}

	return result, nil
}

// ExecuteIDM adds a new todo to the collection using pure IDM data structures.
func ExecuteIDM(text string, opts Options) (*IDMResult, error) {
	if text == "" {
		return nil, fmt.Errorf("todo text cannot be empty")
	}

	idmStore := store.NewIDMStore(opts.CollectionPath)
	return RunIDM(idmStore, text, opts)
}

// ConvertIDMResultToResult converts an IDMResult to the traditional Result format.
// This enables backward compatibility while migrating to pure IDM structures.
func ConvertIDMResultToResult(idmResult *IDMResult) *Result {
	result := &Result{
		Todo:         convertIDMTodoToTodo(idmResult.Todo),
		PositionPath: idmResult.PositionPath,
		Mode:         idmResult.Mode,
		TotalCount:   idmResult.TotalCount,
		DoneCount:    idmResult.DoneCount,
	}

	// Convert all todos if present
	if idmResult.AllTodos != nil {
		result.AllTodos = make([]*models.Todo, len(idmResult.AllTodos))
		for i, idmTodo := range idmResult.AllTodos {
			result.AllTodos[i] = convertIDMTodoToTodo(idmTodo)
		}
	}

	return result
}

// convertIDMTodoToTodo converts a single IDMTodo to Todo for backward compatibility.
func convertIDMTodoToTodo(idmTodo *models.IDMTodo) *models.Todo {
	todo := &models.Todo{
		ID:       idmTodo.UID,
		ParentID: idmTodo.ParentID,
		Text:     idmTodo.Text,
		Statuses: make(map[string]string),
		Modified: idmTodo.Modified,
		Items:    []*models.Todo{}, // Empty - hierarchy managed by IDM
	}

	// Copy statuses
	for k, v := range idmTodo.Statuses {
		todo.Statuses[k] = v
	}

	return todo
}