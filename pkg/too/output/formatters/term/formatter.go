package term

import (
	"fmt"
	"io"
	"os"

	"github.com/arthur-debert/too/pkg/too"
	"github.com/arthur-debert/too/pkg/too/output"
)

// formatter implements the Formatter interface for terminal output
type formatter struct {
	renderer *output.LipbamlRenderer
}


// New creates a new terminal formatter
func New() (output.Formatter, error) {
	// Default to stdout with color support
	renderer, err := output.NewLipbamlRenderer(os.Stdout, true)
	if err != nil {
		return nil, fmt.Errorf("failed to create terminal renderer: %w", err)
	}

	return &formatter{
		renderer: renderer,
	}, nil
}

// NewWithWriter creates a new terminal formatter with a custom writer
func NewWithWriter(w io.Writer, useColor bool) (output.Formatter, error) {
	renderer, err := output.NewLipbamlRenderer(w, useColor)
	if err != nil {
		return nil, fmt.Errorf("failed to create terminal renderer: %w", err)
	}

	return &formatter{
		renderer: renderer,
	}, nil
}

// Name returns the formatter name
func (f *formatter) Name() string {
	return "term"
}

// Description returns the formatter description
func (f *formatter) Description() string {
	return "Rich terminal output with colors and formatting (default)"
}

// RenderChange renders any command that changes todos
func (f *formatter) RenderChange(w io.Writer, result *too.ChangeResult) error {
	f.renderer.Writer = w
	return f.renderer.RenderChange(result)
}

// RenderInit renders the init command result
func (f *formatter) RenderInit(w io.Writer, result *too.InitResult) error {
	f.renderer.Writer = w
	return f.renderer.RenderInit(result)
}


// RenderSearch renders the search command result
func (f *formatter) RenderSearch(w io.Writer, result *too.SearchResult) error {
	f.renderer.Writer = w
	return f.renderer.RenderSearch(result)
}

// RenderList renders the list command result
func (f *formatter) RenderList(w io.Writer, result *too.ListResult) error {
	f.renderer.Writer = w
	return f.renderer.RenderList(result)
}


// RenderDataPath renders the datapath command result
func (f *formatter) RenderDataPath(w io.Writer, result *too.ShowDataPathResult) error {
	f.renderer.Writer = w
	return f.renderer.RenderDataPath(result)
}

// RenderFormats renders the formats command result
func (f *formatter) RenderFormats(w io.Writer, result *too.ListFormatsResult) error {
	f.renderer.Writer = w
	return f.renderer.RenderFormats(result)
}

// RenderError renders an error message
func (f *formatter) RenderError(w io.Writer, err error) error {
	f.renderer.Writer = w
	return f.renderer.RenderError(err)
}
