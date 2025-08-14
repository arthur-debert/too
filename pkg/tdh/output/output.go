package output

import (
	"fmt"
	"io"
	"os"
)

// Renderer is the main output renderer for tdh
// This is an alias to TemplateRenderer for backward compatibility
type Renderer = TemplateRenderer

// NewRenderer creates a new renderer with default settings
func NewRenderer(w io.Writer) *Renderer {
	if w == nil {
		w = os.Stdout
	}

	// Create template renderer with colors enabled
	renderer, err := NewTemplateRenderer(w, true)
	if err != nil {
		panic(fmt.Sprintf("Failed to create template renderer: %v", err))
	}

	return renderer
}
