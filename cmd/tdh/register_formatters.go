package main

import (
	"fmt"

	"github.com/arthur-debert/tdh/pkg/tdh/formatter"
	"github.com/arthur-debert/tdh/pkg/tdh/output"
	"github.com/arthur-debert/tdh/pkg/tdh/output/formatters/json"
	"github.com/arthur-debert/tdh/pkg/tdh/output/formatters/markdown"
	"github.com/arthur-debert/tdh/pkg/tdh/output/formatters/term"
)

func init() {
	// Register all built-in formatters
	registerBuiltinFormatters()
}

// registerBuiltinFormatters registers all built-in formatters.
func registerBuiltinFormatters() {
	// Register JSON formatter
	if err := output.Register(&output.FormatterInfo{
		Info: formatter.Info{
			Name:        "json",
			Description: "JSON output for programmatic consumption",
		},
		Factory: func() (output.Formatter, error) {
			return json.New(), nil
		},
	}); err != nil {
		panic(fmt.Sprintf("failed to register json formatter: %v", err))
	}
	
	// Register Markdown formatter
	if err := output.Register(&output.FormatterInfo{
		Info: formatter.Info{
			Name:        "markdown",
			Description: "Markdown output for documentation and notes",
		},
		Factory: func() (output.Formatter, error) {
			return markdown.New(), nil
		},
	}); err != nil {
		panic(fmt.Sprintf("failed to register markdown formatter: %v", err))
	}
	
	// Register Terminal formatter
	if err := output.Register(&output.FormatterInfo{
		Info: formatter.Info{
			Name:        "term",
			Description: "Rich terminal output with colors and formatting (default)",
		},
		Factory: func() (output.Formatter, error) {
			return term.New()
		},
	}); err != nil {
		panic(fmt.Sprintf("failed to register term formatter: %v", err))
	}
}