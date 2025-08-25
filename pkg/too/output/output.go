package output

import (
	"fmt"
	"io"
	"os"

	"github.com/arthur-debert/too/pkg/too"
)

// Renderer is the main output renderer for too
// It wraps a Formatter to provide backward compatibility with existing code
type Renderer struct {
	formatter Formatter
	writer    io.Writer
}

// NewRenderer creates a new renderer with default settings
// This maintains backward compatibility with existing code
func NewRenderer(w io.Writer) *Renderer {
	if w == nil {
		w = os.Stdout
	}

	// Default to terminal formatter for backward compatibility
	formatter, _ := Get("term")

	return &Renderer{
		formatter: formatter,
		writer:    w,
	}
}

// NewRendererWithFormat creates a new renderer with the specified format
func NewRendererWithFormat(format string, w io.Writer) (*Renderer, error) {
	if w == nil {
		w = os.Stdout
	}

	formatter, err := Get(format)
	if err != nil {
		return nil, fmt.Errorf("failed to get formatter %q: %w", format, err)
	}

	return &Renderer{
		formatter: formatter,
		writer:    w,
	}, nil
}

// RenderAdd renders the add command result
func (r *Renderer) RenderAdd(result *too.AddResult) error {
	return r.formatter.RenderAdd(r.writer, result)
}

// RenderModify renders the modify command result
func (r *Renderer) RenderModify(result *too.ModifyResult) error {
	return r.formatter.RenderModify(r.writer, result)
}

// RenderInit renders the init command result
func (r *Renderer) RenderInit(result *too.InitResult) error {
	return r.formatter.RenderInit(r.writer, result)
}

// RenderClean renders the clean command result
func (r *Renderer) RenderClean(result *too.CleanResult) error {
	return r.formatter.RenderClean(r.writer, result)
}

// RenderSearch renders the search command result
func (r *Renderer) RenderSearch(result *too.SearchResult) error {
	return r.formatter.RenderSearch(r.writer, result)
}

// RenderList renders the list command result
func (r *Renderer) RenderList(result *too.ListResult) error {
	return r.formatter.RenderList(r.writer, result)
}

// RenderComplete renders the complete command results
func (r *Renderer) RenderComplete(results []*too.CompleteResult) error {
	return r.formatter.RenderComplete(r.writer, results)
}

// RenderReopen renders the reopen command results
func (r *Renderer) RenderReopen(results []*too.ReopenResult) error {
	return r.formatter.RenderReopen(r.writer, results)
}

// RenderMove renders the move command result
func (r *Renderer) RenderMove(result *too.MoveResult) error {
	return r.formatter.RenderMove(r.writer, result)
}

// RenderSwap renders the swap command result
func (r *Renderer) RenderSwap(result *too.SwapResult) error {
	return r.formatter.RenderSwap(r.writer, result)
}

// RenderDataPath renders the datapath command result
func (r *Renderer) RenderDataPath(result *too.ShowDataPathResult) error {
	return r.formatter.RenderDataPath(r.writer, result)
}

// RenderFormats renders the formats command result
func (r *Renderer) RenderFormats(result *too.ListFormatsResult) error {
	return r.formatter.RenderFormats(r.writer, result)
}

// RenderError renders an error message
func (r *Renderer) RenderError(err error) error {
	return r.formatter.RenderError(r.writer, err)
}
