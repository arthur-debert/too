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

// Execute marks a todo as pending using the pure IDM data model.
// This function now uses IDM internally but maintains backward compatibility 
// by returning the traditional Result format.
func Execute(ref string, opts Options) (*Result, error) {
	// Use IDM implementation and convert result for backward compatibility
	idmResult, err := ExecuteIDM(ref, opts)
	if err != nil {
		return nil, err
	}

	return ConvertIDMResultToResult(idmResult), nil
}
