package formats

import (
	"github.com/arthur-debert/too/pkg/too/formatter"
)

// Options contains options for the formats command
type Options struct {
	// No options needed for formats command currently
}

// Format describes an available output format
type Format struct {
	Name        string
	Description string
}

// Result contains the result of the formats command
type Result struct {
	Formats []Format
}

// GetFormatterInfoFunc is a function that returns formatter information.
// This is set by the output package to avoid import cycles.
var GetFormatterInfoFunc func() []*formatter.Info

// Execute returns the list of available output formats
func Execute(opts Options) (*Result, error) {
	if GetFormatterInfoFunc == nil {
		// Return empty result if the function is not set
		return &Result{
			Formats: []Format{},
		}, nil
	}

	// Get formatter information
	infos := GetFormatterInfoFunc()

	// Convert to our result format
	formats := make([]Format, len(infos))
	for i, info := range infos {
		formats[i] = Format{
			Name:        info.Name,
			Description: info.Description,
		}
	}

	return &Result{
		Formats: formats,
	}, nil
}
