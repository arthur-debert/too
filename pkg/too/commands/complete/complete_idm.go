package complete

import (
	"fmt"

	"github.com/arthur-debert/too/pkg/logging"
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/store"
)

// IDMResult contains the result of the complete command using pure IDM data structures.
type IDMResult struct {
	Todo       *models.IDMTodo
	OldStatus  string
	NewStatus  string
	Mode       string             // Output mode passed from options
	AllTodos   []*models.IDMTodo  // All todos for long mode
	TotalCount int                // Total count for long mode
	DoneCount  int                // Done count for long mode
}

// ExecuteIDM marks a todo as complete using the pure IDM manager.
func ExecuteIDM(positionPath string, opts Options) (*IDMResult, error) {
	logger := logging.GetLogger("too.commands.complete")
	logger.Debug().
		Str("positionPath", positionPath).
		Str("collectionPath", opts.CollectionPath).
		Msg("executing complete command with pure IDM manager")

	idmStore := store.NewIDMStore(opts.CollectionPath)
	
	// Create pure IDM workflow manager
	manager, err := store.NewPureIDMManager(idmStore, opts.CollectionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create pure IDM manager: %w", err)
	}

	// Resolve position path to UID
	uid, err := manager.ResolvePositionPath(store.RootScope, positionPath)
	if err != nil {
		return nil, fmt.Errorf("todo not found: %w", err)
	}

	// Get the todo using PureIDMManager method
	todo := manager.GetTodoByUID(uid)
	if todo == nil {
		return nil, fmt.Errorf("todo with UID '%s' not found", uid)
	}

	// Capture old status
	oldStatus, err := manager.GetStatus(uid, "completion")
	if err != nil {
		oldStatus = string(models.StatusPending)
	}

	// Set status to done
	if err := manager.SetStatus(uid, "completion", "done"); err != nil {
		return nil, fmt.Errorf("failed to set completion status: %w", err)
	}

	// Save changes
	if err := manager.Save(); err != nil {
		return nil, fmt.Errorf("failed to save collection: %w", err)
	}

	logger.Debug().
		Str("todoUID", uid).
		Str("oldStatus", oldStatus).
		Str("newStatus", "done").
		Msg("marked todo as complete using pure IDM manager")

	// Build result
	result := &IDMResult{
		Todo:      todo,
		OldStatus: oldStatus,
		NewStatus: "done",
		Mode:      opts.Mode,
	}

	// Add long mode data if requested using manager's IDM-aware methods
	if opts.Mode == "long" {
		result.AllTodos = manager.ListActive().([]*models.IDMTodo)
		if result.AllTodos == nil {
			result.AllTodos = []*models.IDMTodo{}
		}
		result.TotalCount, result.DoneCount = manager.CountTodos()
	}

	return result, nil
}

// ConvertIDMResultToResult converts an IDMResult to the traditional Result format.
// This enables backward compatibility while migrating to pure IDM structures.
func ConvertIDMResultToResult(idmResult *IDMResult) *Result {
	result := &Result{
		Todo:       convertIDMTodoToTodo(idmResult.Todo),
		OldStatus:  idmResult.OldStatus,
		NewStatus:  idmResult.NewStatus,
		Mode:       idmResult.Mode,
		TotalCount: idmResult.TotalCount,
		DoneCount:  idmResult.DoneCount,
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