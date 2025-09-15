package formats

import (
	"github.com/arthur-debert/too/pkg/lipbalm"
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

// Execute returns the list of available output formats
func Execute(opts Options) (*Result, error) {
	// Get format information directly from lipbalm
	infos := lipbalm.GetFormatInfo()

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
