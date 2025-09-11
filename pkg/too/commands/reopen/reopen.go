package reopen

import (
	"github.com/arthur-debert/too/pkg/too/models"
)

// Options contains options for the reopen command
type Options struct {
	CollectionPath string
}

// Result contains the result of the reopen command
type Result struct {
	Todo      *models.Todo
	OldStatus string
	NewStatus string
}

// Execute marks a todo as pending by finding it via a user-provided reference,
// which can be either a position path (e.g., "1.2") or a short ID.
// Uses WorkflowManager for status management.
func Execute(ref string, opts Options) (*Result, error) {
	return ExecuteDirect(ref, opts)
}
