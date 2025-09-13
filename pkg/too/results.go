package too

import "github.com/arthur-debert/too/pkg/too/models"

// ChangeResult represents the result of any command that modifies todos
type ChangeResult struct {
	Command        string              // The command that was executed (add, modify, complete, etc.)
	Message        string              // Optional message to display
	AffectedTodos  []*models.IDMTodo   // The todos that were affected by the command
	AllTodos       []*models.IDMTodo   // All todos in the collection after the change
	TotalCount     int                 // Total number of todos
	DoneCount      int                 // Number of completed todos
}

// NewChangeResult creates a new ChangeResult
func NewChangeResult(command string, message string, affected []*models.IDMTodo, all []*models.IDMTodo, total, done int) *ChangeResult {
	return &ChangeResult{
		Command:       command,
		Message:       message,
		AffectedTodos: affected,
		AllTodos:      all,
		TotalCount:    total,
		DoneCount:     done,
	}
}

// MessageResult represents a simple message output
type MessageResult struct {
	Text  string // The message text
	Level string // Message level: info, success, warning, error
}

// NewMessageResult creates a new MessageResult with the specified level
func NewMessageResult(text, level string) *MessageResult {
	return &MessageResult{
		Text:  text,
		Level: level,
	}
}

// NewInfoMessage creates a new info message result
func NewInfoMessage(text string) *MessageResult {
	return NewMessageResult(text, "info")
}