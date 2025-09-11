package clean

import (
	"github.com/arthur-debert/too/pkg/too/models"
)

// Options contains options for the clean command
type Options struct {
	CollectionPath string
}

// Result contains the result of the clean command
type Result struct {
	RemovedCount int
	RemovedTodos []*models.Todo
	ActiveCount  int
}

// Execute removes finished todos from the collection using the pure IDM data model.
// This function now uses IDM internally but maintains backward compatibility 
// by returning the traditional Result format.
func Execute(opts Options) (*Result, error) {
	// Use IDM implementation and convert result for backward compatibility
	idmResult, err := ExecuteIDM(opts)
	if err != nil {
		return nil, err
	}

	return ConvertIDMResultToResult(idmResult), nil
}

