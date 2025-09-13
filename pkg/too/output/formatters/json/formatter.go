package json

import (
	"encoding/json"
	"io"

	"github.com/arthur-debert/too/pkg/too"
	"github.com/arthur-debert/too/pkg/too/output"
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

// RenderChange renders any command that changes todos as JSON
func (f *formatter) RenderChange(w io.Writer, result *too.ChangeResult) error {
	return f.encode(w, result)
}

// RenderMessage renders a message result as JSON
func (f *formatter) RenderMessage(w io.Writer, result *too.MessageResult) error {
	return f.encode(w, result)
}


// RenderSearch renders the search command result as JSON
func (f *formatter) RenderSearch(w io.Writer, result *too.SearchResult) error {
	return f.encode(w, result)
}

// RenderList renders the list command result as JSON
func (f *formatter) RenderList(w io.Writer, result *too.ListResult) error {
	return f.encode(w, result)
}


// RenderSwap renders the swap command result as JSON


// RenderFormats renders the formats command result as JSON
func (f *formatter) RenderFormats(w io.Writer, result *too.ListFormatsResult) error {
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
