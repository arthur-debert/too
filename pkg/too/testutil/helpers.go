package testutil

import (
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/google/uuid"
)

// CreateTestTodo creates a todo for testing
func CreateTestTodo(text string, status models.TodoStatus) *models.Todo {
	return &models.Todo{
		ID:       uuid.New().String(),
		Text:     text,
		Statuses: map[string]string{"completion": string(status)},
		Items:    []*models.Todo{},
	}
}
