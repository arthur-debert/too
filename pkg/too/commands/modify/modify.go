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

// Execute modifies the text of an existing todo using the pure IDM data model.
// This function now uses IDM internally but maintains backward compatibility 
// by returning the traditional Result format.
func Execute(positionStr string, newText string, opts Options) (*Result, error) {
	// Use IDM implementation and convert result for backward compatibility
	idmResult, err := ExecuteIDM(positionStr, newText, opts)
	if err != nil {
		return nil, err
	}

	return ConvertIDMResultToResult(idmResult), nil
}
