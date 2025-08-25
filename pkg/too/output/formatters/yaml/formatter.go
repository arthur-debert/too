package yaml

import (
	"io"

	"github.com/arthur-debert/too/pkg/too"
	"github.com/arthur-debert/too/pkg/too/output"
	"gopkg.in/yaml.v3"
)

// formatter implements the Formatter interface for YAML output
type formatter struct{}

// New creates a new YAML formatter
func New() output.Formatter {
	return &formatter{}
}

// Name returns the formatter name
func (f *formatter) Name() string {
	return "yaml"
}

// Description returns the formatter description
func (f *formatter) Description() string {
	return "YAML output for programmatic consumption"
}

// encode is a helper that YAML encodes any value to the writer
func (f *formatter) encode(w io.Writer, v interface{}) error {
	encoder := yaml.NewEncoder(w)
	encoder.SetIndent(2)
	return encoder.Encode(v)
}

// RenderAdd renders the add command result as YAML
func (f *formatter) RenderAdd(w io.Writer, result *too.AddResult) error {
	return f.encode(w, result)
}

// RenderModify renders the modify command result as YAML
func (f *formatter) RenderModify(w io.Writer, result *too.ModifyResult) error {
	return f.encode(w, result)
}

// RenderInit renders the init command result as YAML
func (f *formatter) RenderInit(w io.Writer, result *too.InitResult) error {
	return f.encode(w, result)
}

// RenderClean renders the clean command result as YAML
func (f *formatter) RenderClean(w io.Writer, result *too.CleanResult) error {
	return f.encode(w, result)
}

// RenderSearch renders the search command result as YAML
func (f *formatter) RenderSearch(w io.Writer, result *too.SearchResult) error {
	return f.encode(w, result)
}

// RenderList renders the list command result as YAML
func (f *formatter) RenderList(w io.Writer, result *too.ListResult) error {
	return f.encode(w, result)
}

// RenderComplete renders the complete command results as YAML
func (f *formatter) RenderComplete(w io.Writer, results []*too.CompleteResult) error {
	return f.encode(w, results)
}

// RenderReopen renders the reopen command results as YAML
func (f *formatter) RenderReopen(w io.Writer, results []*too.ReopenResult) error {
	return f.encode(w, results)
}

// RenderMove renders the move command result as YAML
func (f *formatter) RenderMove(w io.Writer, result *too.MoveResult) error {
	return f.encode(w, result)
}

// RenderSwap renders the swap command result as YAML
func (f *formatter) RenderSwap(w io.Writer, result *too.SwapResult) error {
	return f.encode(w, result)
}

// RenderDataPath renders the datapath command result as YAML
func (f *formatter) RenderDataPath(w io.Writer, result *too.ShowDataPathResult) error {
	return f.encode(w, result)
}

// RenderFormats renders the formats command result as YAML
func (f *formatter) RenderFormats(w io.Writer, result *too.ListFormatsResult) error {
	return f.encode(w, result)
}

// RenderError renders an error message as YAML
func (f *formatter) RenderError(w io.Writer, err error) error {
	errorResponse := struct {
		Error string `yaml:"error"`
	}{
		Error: err.Error(),
	}
	return f.encode(w, errorResponse)
}
