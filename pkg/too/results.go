package too

import "github.com/arthur-debert/too/pkg/too/models"

// ChangeResult represents the result of any command that modifies todos
type ChangeResult struct {
	Command        string              // The command that was executed (add, modify, complete, etc.)
	AffectedTodos  []*models.IDMTodo   // The todos that were affected by the command
	AllTodos       []*models.IDMTodo   // All todos in the collection after the change
	TotalCount     int                 // Total number of todos
	DoneCount      int                 // Number of completed todos
}

// NewChangeResult creates a new ChangeResult
func NewChangeResult(command string, affected []*models.IDMTodo, all []*models.IDMTodo, total, done int) *ChangeResult {
	return &ChangeResult{
		Command:       command,
		AffectedTodos: affected,
		AllTodos:      all,
		TotalCount:    total,
		DoneCount:     done,
	}
}