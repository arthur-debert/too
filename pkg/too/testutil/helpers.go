package testutil

import (
	"github.com/arthur-debert/too/pkg/too/models"
)

// CreateTestTodo creates a todo for testing
func CreateTestTodo(text string, status models.TodoStatus) *models.IDMTodo {
	todo := models.NewIDMTodo(text, "")
	todo.EnsureStatuses()
	todo.Statuses["completion"] = string(status)
	return todo
}
