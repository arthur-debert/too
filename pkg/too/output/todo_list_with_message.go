package output

import "github.com/arthur-debert/too/pkg/too/models"

// TodoListWithMessage wraps any result that contains todos to add a message
type TodoListWithMessage struct {
	Message     string            // Optional message to display
	MessageType string            // Type: success, error, warning, info
	Todos       []*models.Todo // The todos to display
	TotalCount  int              // Total count of todos
	DoneCount   int              // Count of done todos
	HighlightID string           // Optional UID to highlight
}