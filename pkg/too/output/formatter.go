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

	// Render methods for each command result type
	RenderAdd(w io.Writer, result *too.AddResult) error
	RenderModify(w io.Writer, result *too.ModifyResult) error
	RenderInit(w io.Writer, result *too.InitResult) error
	RenderClean(w io.Writer, result *too.CleanResult) error
	RenderSearch(w io.Writer, result *too.SearchResult) error
	RenderList(w io.Writer, result *too.ListResult) error
	RenderComplete(w io.Writer, results []*too.CompleteResult) error
	RenderReopen(w io.Writer, results []*too.ReopenResult) error
	RenderMove(w io.Writer, result *too.MoveResult) error
	RenderSwap(w io.Writer, result *too.SwapResult) error
	RenderDataPath(w io.Writer, result *too.ShowDataPathResult) error
	RenderFormats(w io.Writer, result *too.ListFormatsResult) error
	RenderError(w io.Writer, err error) error
}
