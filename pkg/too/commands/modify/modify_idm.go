package modify

import (
	"fmt"

	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/store"
)

// IDMResult contains the result of the modify command using pure IDM data structures.
type IDMResult struct {
	Todo    *models.IDMTodo
	OldText string
	NewText string
}

// ExecuteIDM modifies the text of an existing todo using the pure IDM manager.
func ExecuteIDM(positionStr string, newText string, opts Options) (*IDMResult, error) {
	if newText == "" {
		return nil, fmt.Errorf("new todo text cannot be empty")
	}

	idmStore := store.NewIDMStore(opts.CollectionPath)
	
	// Create pure IDM workflow manager
	manager, err := store.NewPureIDMManager(idmStore, opts.CollectionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create pure IDM manager: %w", err)
	}

	// Resolve position to UID
	uid, err := manager.ResolvePositionPath(store.RootScope, positionStr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve todo position '%s': %w", positionStr, err)
	}

	// Find and modify the todo using PureIDMManager method
	todo := manager.GetTodoByUID(uid)
	if todo == nil {
		return nil, fmt.Errorf("todo with UID '%s' not found", uid)
	}

	oldText := todo.Text
	todo.Text = newText
	todo.SetModified()

	// Save the collection
	if err := manager.Save(); err != nil {
		return nil, err
	}

	return &IDMResult{
		Todo:    todo,
		OldText: oldText,
		NewText: newText,
	}, nil
}

// ConvertIDMResultToResult converts an IDMResult to the traditional Result format.
// This enables backward compatibility while migrating to pure IDM structures.
func ConvertIDMResultToResult(idmResult *IDMResult) *Result {
	return &Result{
		Todo:    convertIDMTodoToTodo(idmResult.Todo),
		OldText: idmResult.OldText,
		NewText: idmResult.NewText,
	}
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