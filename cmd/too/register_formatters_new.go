// +build ignore

package main

import (
	"fmt"
	"io"

	"github.com/arthur-debert/too/pkg/too"
	"github.com/arthur-debert/too/pkg/too/formatter"
	"github.com/arthur-debert/too/pkg/too/output"
)

// This file shows how to migrate to the new lipbalm-based engine

func registerFormattersNew() {
	// Create the global engine
	engine, err := output.GetGlobalEngine()
	if err != nil {
		panic(fmt.Sprintf("failed to create output engine: %v", err))
	}

	// Get available formats from lipbalm
	formats := engine.ListFormats()

	// Register adapters for each format
	for _, format := range formats {
		formatName := format // Capture for closure
		
		// Create adapter that implements the old Formatter interface
		adapter := &formatAdapter{
			engine: engine,
			format: formatName,
		}

		// Register with the old registry
		if err := output.Register(&output.FormatterInfo{
			Info: formatter.Info{
				Name:        formatName,
				Description: getFormatDescription(formatName),
			},
			Factory: func() (output.Formatter, error) {
				return adapter, nil
			},
		}); err != nil {
			panic(fmt.Sprintf("failed to register %s formatter: %v", formatName, err))
		}
	}
}

// formatAdapter implements the old Formatter interface using the new engine
type formatAdapter struct {
	engine *output.Engine
	format string
}

func (a *formatAdapter) Name() string {
	return a.format
}

func (a *formatAdapter) Description() string {
	return getFormatDescription(a.format)
}

func (a *formatAdapter) RenderChange(w io.Writer, result *too.ChangeResult) error {
	return a.engine.Render(w, a.format, result)
}

func (a *formatAdapter) RenderMessage(w io.Writer, result *too.MessageResult) error {
	return a.engine.Render(w, a.format, result)
}

func (a *formatAdapter) RenderSearch(w io.Writer, result *too.SearchResult) error {
	return a.engine.Render(w, a.format, result)
}

func (a *formatAdapter) RenderList(w io.Writer, result *too.ListResult) error {
	return a.engine.Render(w, a.format, result)
}

func (a *formatAdapter) RenderFormats(w io.Writer, result *too.ListFormatsResult) error {
	return a.engine.Render(w, a.format, result)
}

func (a *formatAdapter) RenderError(w io.Writer, err error) error {
	return a.engine.RenderError(w, a.format, err)
}

func getFormatDescription(format string) string {
	switch format {
	case "json":
		return "JSON output for programmatic consumption"
	case "yaml":
		return "YAML output for programmatic consumption"
	case "csv":
		return "CSV output for spreadsheet applications"
	case "markdown":
		return "Markdown output for documentation and notes"
	case "term":
		return "Rich terminal output with colors and formatting (default)"
	case "plain":
		return "Plain text output without formatting"
	default:
		return fmt.Sprintf("%s output format", format)
	}
}