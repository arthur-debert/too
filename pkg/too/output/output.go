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

// RenderChange renders any command that changes todos
func (r *Renderer) RenderChange(result *too.ChangeResult) error {
	return r.formatter.RenderChange(r.writer, result)
}


// RenderMessage renders a message result
func (r *Renderer) RenderMessage(result *too.MessageResult) error {
	return r.formatter.RenderMessage(r.writer, result)
}

// RenderInit renders the init command result
func (r *Renderer) RenderInit(result *too.InitResult) error {
	msg := too.NewInfoMessage(result.Message)
	return r.RenderMessage(msg)
}


// RenderSearch renders the search command result
func (r *Renderer) RenderSearch(result *too.SearchResult) error {
	return r.formatter.RenderSearch(r.writer, result)
}

// RenderList renders the list command result
func (r *Renderer) RenderList(result *too.ListResult) error {
	return r.formatter.RenderList(r.writer, result)
}


// RenderDataPath renders the datapath command result
func (r *Renderer) RenderDataPath(result *too.ShowDataPathResult) error {
	msg := too.NewInfoMessage(result.Path)
	return r.RenderMessage(msg)
}

// RenderFormats renders the formats command result
func (r *Renderer) RenderFormats(result *too.ListFormatsResult) error {
	return r.formatter.RenderFormats(r.writer, result)
}

// RenderError renders an error message
func (r *Renderer) RenderError(err error) error {
	return r.formatter.RenderError(r.writer, err)
}
