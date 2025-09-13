package output

import (
	"fmt"
	"io"
	"os"

	"github.com/arthur-debert/too/pkg/too"
)

// Renderer is the main output renderer for too
// Now it wraps the Engine to provide backward compatibility
type Renderer struct {
	engine *Engine
	format string
	writer io.Writer
}

// NewRenderer creates a new renderer with default settings
// This maintains backward compatibility with existing code
func NewRenderer(w io.Writer) *Renderer {
	if w == nil {
		w = os.Stdout
	}

	engine, _ := GetGlobalEngine()
	
	return &Renderer{
		engine: engine,
		format: "term",
		writer: w,
	}
}

// NewRendererWithFormat creates a new renderer with the specified format
func NewRendererWithFormat(format string, w io.Writer) (*Renderer, error) {
	if w == nil {
		w = os.Stdout
	}

	engine, err := GetGlobalEngine()
	if err != nil {
		return nil, fmt.Errorf("failed to get engine: %w", err)
	}

	// Validate format
	formats := engine.ListFormats()
	found := false
	for _, f := range formats {
		if f == format {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("invalid format %q. Available formats: %v", format, formats)
	}

	return &Renderer{
		engine: engine,
		format: format,
		writer: w,
	}, nil
}

// RenderChange renders any command that changes todos
func (r *Renderer) RenderChange(result *too.ChangeResult) error {
	return r.engine.Render(r.writer, r.format, result)
}

// RenderMessage renders a message result
func (r *Renderer) RenderMessage(result *too.MessageResult) error {
	return r.engine.Render(r.writer, r.format, result)
}

// RenderInit renders the init command result
func (r *Renderer) RenderInit(result *too.InitResult) error {
	msg := too.NewInfoMessage(result.Message)
	return r.RenderMessage(msg)
}

// RenderSearch renders the search command result
func (r *Renderer) RenderSearch(result *too.SearchResult) error {
	return r.engine.Render(r.writer, r.format, result)
}

// RenderList renders the list command result
func (r *Renderer) RenderList(result *too.ListResult) error {
	return r.engine.Render(r.writer, r.format, result)
}

// RenderDataPath renders the datapath command result
func (r *Renderer) RenderDataPath(result *too.ShowDataPathResult) error {
	msg := too.NewInfoMessage(result.Path)
	return r.RenderMessage(msg)
}

// RenderFormats renders the formats command result
func (r *Renderer) RenderFormats(result *too.ListFormatsResult) error {
	return r.engine.Render(r.writer, r.format, result)
}

// RenderError renders an error message
func (r *Renderer) RenderError(err error) error {
	return r.engine.RenderError(r.writer, r.format, err)
}

// HasFormatter checks if a formatter is available
func HasFormatter(name string) bool {
	engine, err := GetGlobalEngine()
	if err != nil {
		return false
	}
	
	for _, format := range engine.ListFormats() {
		if format == name {
			return true
		}
	}
	return false
}

// List returns all available format names
func List() []string {
	engine, err := GetGlobalEngine()
	if err != nil {
		return []string{}
	}
	
	return engine.ListFormats()
}