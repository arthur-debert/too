package output

import (
	"io"

	"github.com/arthur-debert/too/pkg/too"
)

// Formatter defines the interface that all output formatters must implement.
// Each formatter is responsible for rendering all command results in its specific format.
type Formatter interface {
	// Format identification
	Name() string        // The format identifier used in CLI (e.g., "json", "term")
	Description() string // Human-readable description for help text

	// Render methods for different result types
	RenderChange(w io.Writer, result *too.ChangeResult) error    // For all todo-modifying commands
	RenderMessage(w io.Writer, result *too.MessageResult) error  // For simple messages
	RenderSearch(w io.Writer, result *too.SearchResult) error    // Special case: filtered view
	RenderList(w io.Writer, result *too.ListResult) error        // Special case: no message
	RenderFormats(w io.Writer, result *too.ListFormatsResult) error
	RenderError(w io.Writer, err error) error
}
