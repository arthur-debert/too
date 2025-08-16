package term

import (
	"fmt"
	"io"
	"os"

	"github.com/arthur-debert/tdh/pkg/tdh"
	"github.com/arthur-debert/tdh/pkg/tdh/output"
)

// formatter implements the Formatter interface for terminal output
type formatter struct {
	renderer *output.LipbamlRenderer
}

// init registers the terminal formatter
func init() {
	output.Register(&output.FormatterInfo{
		Name:        "term",
		Description: "Rich terminal output with colors and formatting (default)",
		Factory: func() (output.Formatter, error) {
			return New()
		},
	})
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

// RenderAdd renders the add command result
func (f *formatter) RenderAdd(w io.Writer, result *tdh.AddResult) error {
	// Update renderer's writer for this render
	f.renderer.Writer = w
	return f.renderer.RenderAdd(result)
}

// RenderModify renders the modify command result
func (f *formatter) RenderModify(w io.Writer, result *tdh.ModifyResult) error {
	f.renderer.Writer = w
	return f.renderer.RenderModify(result)
}

// RenderInit renders the init command result
func (f *formatter) RenderInit(w io.Writer, result *tdh.InitResult) error {
	f.renderer.Writer = w
	return f.renderer.RenderInit(result)
}

// RenderClean renders the clean command result
func (f *formatter) RenderClean(w io.Writer, result *tdh.CleanResult) error {
	f.renderer.Writer = w
	return f.renderer.RenderClean(result)
}

// RenderSearch renders the search command result
func (f *formatter) RenderSearch(w io.Writer, result *tdh.SearchResult) error {
	f.renderer.Writer = w
	return f.renderer.RenderSearch(result)
}

// RenderList renders the list command result
func (f *formatter) RenderList(w io.Writer, result *tdh.ListResult) error {
	f.renderer.Writer = w
	return f.renderer.RenderList(result)
}

// RenderComplete renders the complete command results
func (f *formatter) RenderComplete(w io.Writer, results []*tdh.CompleteResult) error {
	f.renderer.Writer = w
	return f.renderer.RenderComplete(results)
}

// RenderReopen renders the reopen command results
func (f *formatter) RenderReopen(w io.Writer, results []*tdh.ReopenResult) error {
	f.renderer.Writer = w
	return f.renderer.RenderReopen(results)
}

// RenderMove renders the move command result
func (f *formatter) RenderMove(w io.Writer, result *tdh.MoveResult) error {
	f.renderer.Writer = w
	return f.renderer.RenderMove(result)
}

// RenderSwap renders the swap command result
func (f *formatter) RenderSwap(w io.Writer, result *tdh.SwapResult) error {
	f.renderer.Writer = w
	return f.renderer.RenderSwap(result)
}

// RenderDataPath renders the datapath command result
func (f *formatter) RenderDataPath(w io.Writer, result *tdh.ShowDataPathResult) error {
	f.renderer.Writer = w
	return f.renderer.RenderDataPath(result)
}

// RenderFormats renders the formats command result
func (f *formatter) RenderFormats(w io.Writer, result *tdh.ListFormatsResult) error {
	f.renderer.Writer = w
	return f.renderer.RenderFormats(result)
}

// RenderError renders an error message
func (f *formatter) RenderError(w io.Writer, err error) error {
	f.renderer.Writer = w
	return f.renderer.RenderError(err)
}
