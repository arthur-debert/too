package output

import (
	"github.com/arthur-debert/too/pkg/too/models"
)

// HierarchicalTodo is an alias for models.HierarchicalTodo for backward compatibility
type HierarchicalTodo = models.HierarchicalTodo

// BuildHierarchy is a convenience wrapper for models.BuildHierarchy
func BuildHierarchy(todos []*models.Todo) []*HierarchicalTodo {
	return models.BuildHierarchy(todos)
}

// FlattenHierarchy is a convenience wrapper for models.FlattenHierarchy
func FlattenHierarchy(todos []*HierarchicalTodo) []*models.Todo {
	return models.FlattenHierarchy(todos)
}