package testutil

import (
	"github.com/arthur-debert/too/pkg/too/models"
)

// CreateTestTodo creates a todo for testing
func CreateTestTodo(text string, status models.TodoStatus) *models.Todo {
	todo := &models.Todo{
		UID:      "test-" + text,
		Text:     text,
		ParentID: "",
		Statuses: map[string]string{
			"completion": string(status),
		},
	}
	return todo
}
