package internal

import "github.com/arthur-debert/tdh/pkg/tdh/models"

// FindResult holds the results of a Find operation.
type FindResult struct {
	Todos      []*models.Todo
	TotalCount int
	DoneCount  int
}
