package complete

import (
	"fmt"

	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/store"
)

// UnifiedResult contains the result of the complete command that works with both manager types.
type UnifiedResult struct {
	UID        string
	Text       string
	OldStatus  string
	NewStatus  string
	Mode       string
	AllTodos   interface{} // []*models.Todo or []*models.IDMTodo
	TotalCount int
	DoneCount  int
	IsPureIDM  bool
}

// ExecuteUnified marks a todo as complete using the factory pattern.
// It automatically selects the appropriate manager based on storage format.
func ExecuteUnified(positionPath string, opts Options) (*UnifiedResult, error) {
	// Create workflow manager using factory
	manager, err := store.CreateWorkflowManager(opts.CollectionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create workflow manager: %w", err)
	}

	// Resolve the position path to get the UID
	uid, err := manager.ResolvePositionPath(store.RootScope, positionPath)
	if err != nil {
		return nil, fmt.Errorf("todo not found at position %s: %w", positionPath, err)
	}

	// Get current status
	oldStatus, err := manager.GetStatus(uid, "completion")
	if err != nil {
		// Default to pending if not found
		oldStatus = string(models.StatusPending)
	}

	// Set to done
	newStatus := string(models.StatusDone)
	if err := manager.SetStatus(uid, "completion", newStatus); err != nil {
		return nil, fmt.Errorf("failed to mark todo as complete: %w", err)
	}

	// Save changes
	if err := manager.Save(); err != nil {
		return nil, fmt.Errorf("failed to save collection: %w", err)
	}

	// Get the todo details
	todoInterface := manager.GetTodoByID(uid)
	if todoInterface == nil {
		return nil, fmt.Errorf("todo with UID %s not found after update", uid)
	}

	// Build unified result
	result := &UnifiedResult{
		UID:       uid,
		OldStatus: oldStatus,
		NewStatus: newStatus,
		Mode:      opts.Mode,
		IsPureIDM: manager.IsPureIDM(),
	}

	// Extract text based on type
	if manager.IsPureIDM() {
		todo := todoInterface.(*models.IDMTodo)
		result.Text = todo.Text
	} else {
		todo := todoInterface.(*models.Todo)
		result.Text = todo.Text
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
		OldStatus:  unified.OldStatus,
		NewStatus:  unified.NewStatus,
		Mode:       unified.Mode,
		TotalCount: unified.TotalCount,
		DoneCount:  unified.DoneCount,
	}

	// Create a Todo from the unified result
	result.Todo = &models.Todo{
		ID:       unified.UID,
		Text:     unified.Text,
		Items:    []*models.Todo{}, // Empty - hierarchy managed by IDM
		Statuses: map[string]string{"completion": unified.NewStatus},
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

