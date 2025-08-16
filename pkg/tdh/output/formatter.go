package output

import (
	"io"

	"github.com/arthur-debert/tdh/pkg/tdh"
)

// Formatter defines the interface that all output formatters must implement.
// Each formatter is responsible for rendering all command results in its specific format.
type Formatter interface {
	// Format identification
	Name() string        // The format identifier used in CLI (e.g., "json", "term")
	Description() string // Human-readable description for help text

	// Render methods for each command result type
	RenderAdd(w io.Writer, result *tdh.AddResult) error
	RenderModify(w io.Writer, result *tdh.ModifyResult) error
	RenderInit(w io.Writer, result *tdh.InitResult) error
	RenderClean(w io.Writer, result *tdh.CleanResult) error
	RenderSearch(w io.Writer, result *tdh.SearchResult) error
	RenderList(w io.Writer, result *tdh.ListResult) error
	RenderComplete(w io.Writer, results []*tdh.CompleteResult) error
	RenderReopen(w io.Writer, results []*tdh.ReopenResult) error
	RenderMove(w io.Writer, result *tdh.MoveResult) error
	RenderSwap(w io.Writer, result *tdh.SwapResult) error
	RenderDataPath(w io.Writer, result *tdh.ShowDataPathResult) error
	RenderFormats(w io.Writer, result *tdh.ListFormatsResult) error
	RenderError(w io.Writer, err error) error
}
