package modify

import (
	"github.com/arthur-debert/too/pkg/too/models"
)

// Options contains options for the modify command
type Options struct {
	CollectionPath string
}

// Result contains the result of the modify command
type Result struct {
	Todo    *models.Todo
	OldText string
	NewText string
}

// Execute modifies the text of an existing todo
func Execute(positionStr string, newText string, opts Options) (*Result, error) {
	return ExecuteDirect(positionStr, newText, opts)
}
