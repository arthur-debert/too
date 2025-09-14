package too

import (
	"github.com/arthur-debert/too/pkg/lipbalm"
	"github.com/arthur-debert/too/pkg/too/models"
)

// ChangeResult represents the result of any command that modifies todos
type ChangeResult struct {
	Command        string              // The command that was executed (add, modify, complete, etc.)
	Message        string              // Optional message to display
	AffectedTodos  []*models.Todo   // The todos that were affected by the command
	AllTodos       []*models.Todo   // All todos in the collection after the change
	TotalCount     int                 // Total number of todos
	DoneCount      int                 // Number of completed todos
}

// MessageType returns the appropriate message type for this result
func (r *ChangeResult) MessageType() string {
	switch r.Command {
	case "edit", "modify":
		return "info"
	case "reopen":
		return "warning"
	case "clean":
		if len(r.AffectedTodos) == 0 {
			return "warning"
		}
		return "success"
	default:
		return "success"
	}
}

// NewChangeResult creates a new ChangeResult
func NewChangeResult(command string, message string, affected []*models.Todo, all []*models.Todo, total, done int) *ChangeResult {
	return &ChangeResult{
		Command:       command,
		Message:       message,
		AffectedTodos: affected,
		AllTodos:      all,
		TotalCount:    total,
		DoneCount:     done,
	}
}

// MessageResult is an alias for lipbalm's Message type for backward compatibility
type MessageResult = lipbalm.Message

// NewMessageResult creates a new MessageResult with the specified level
func NewMessageResult(text, level string) *MessageResult {
	return lipbalm.NewMessage(text, level)
}

// NewInfoMessage creates a new info message result
func NewInfoMessage(text string) *MessageResult {
	return lipbalm.NewInfoMessage(text)
}