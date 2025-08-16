package json

import (
	"encoding/json"
	"io"

	"github.com/arthur-debert/tdh/pkg/tdh"
	"github.com/arthur-debert/tdh/pkg/tdh/output"
)

// formatter implements the Formatter interface for JSON output
type formatter struct{}


// New creates a new JSON formatter
func New() output.Formatter {
	return &formatter{}
}

// Name returns the formatter name
func (f *formatter) Name() string {
	return "json"
}

// Description returns the formatter description
func (f *formatter) Description() string {
	return "JSON output for programmatic consumption"
}

// encode is a helper that JSON encodes any value to the writer
func (f *formatter) encode(w io.Writer, v interface{}) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

// RenderAdd renders the add command result as JSON
func (f *formatter) RenderAdd(w io.Writer, result *tdh.AddResult) error {
	return f.encode(w, result)
}

// RenderModify renders the modify command result as JSON
func (f *formatter) RenderModify(w io.Writer, result *tdh.ModifyResult) error {
	return f.encode(w, result)
}

// RenderInit renders the init command result as JSON
func (f *formatter) RenderInit(w io.Writer, result *tdh.InitResult) error {
	return f.encode(w, result)
}

// RenderClean renders the clean command result as JSON
func (f *formatter) RenderClean(w io.Writer, result *tdh.CleanResult) error {
	return f.encode(w, result)
}

// RenderSearch renders the search command result as JSON
func (f *formatter) RenderSearch(w io.Writer, result *tdh.SearchResult) error {
	return f.encode(w, result)
}

// RenderList renders the list command result as JSON
func (f *formatter) RenderList(w io.Writer, result *tdh.ListResult) error {
	return f.encode(w, result)
}

// RenderComplete renders the complete command results as JSON
func (f *formatter) RenderComplete(w io.Writer, results []*tdh.CompleteResult) error {
	return f.encode(w, results)
}

// RenderReopen renders the reopen command results as JSON
func (f *formatter) RenderReopen(w io.Writer, results []*tdh.ReopenResult) error {
	return f.encode(w, results)
}

// RenderMove renders the move command result as JSON
func (f *formatter) RenderMove(w io.Writer, result *tdh.MoveResult) error {
	return f.encode(w, result)
}

// RenderSwap renders the swap command result as JSON
func (f *formatter) RenderSwap(w io.Writer, result *tdh.SwapResult) error {
	return f.encode(w, result)
}

// RenderDataPath renders the datapath command result as JSON
func (f *formatter) RenderDataPath(w io.Writer, result *tdh.ShowDataPathResult) error {
	return f.encode(w, result)
}

// RenderFormats renders the formats command result as JSON
func (f *formatter) RenderFormats(w io.Writer, result *tdh.ListFormatsResult) error {
	return f.encode(w, result)
}

// RenderError renders an error message as JSON
func (f *formatter) RenderError(w io.Writer, err error) error {
	errorResponse := struct {
		Error string `json:"error"`
	}{
		Error: err.Error(),
	}
	return f.encode(w, errorResponse)
}
