package output

import (
	"io"
	"os"
)

// Renderer is the main output renderer for tdh
type Renderer = LipbamlRenderer

// NewRenderer creates a new renderer with default settings
func NewRenderer(w io.Writer) *Renderer {
	if w == nil {
		w = os.Stdout
	}

	// Create lipbaml renderer with colors enabled
	renderer, _ := NewLipbamlRenderer(w, true)
	return renderer
}
