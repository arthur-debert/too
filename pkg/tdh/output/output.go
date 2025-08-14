package output

import (
	"fmt"
	"io"
	"os"
	"text/template"

	"github.com/arthur-debert/tdh/pkg/tdh"
)

// Renderer handles output formatting for tdh commands
type Renderer struct {
	writer           io.Writer
	templateRenderer *TemplateRenderer
}

// NewRenderer creates a new renderer
func NewRenderer(w io.Writer) *Renderer {
	if w == nil {
		w = os.Stdout
	}

	// Create template renderer - panic if it fails
	templateRenderer, err := NewTemplateRenderer(w, true)
	if err != nil {
		panic(fmt.Sprintf("Failed to create template renderer: %v", err))
	}

	return &Renderer{
		writer:           w,
		templateRenderer: templateRenderer,
	}
}

// RenderInit renders the init command result
func (r *Renderer) RenderInit(result *tdh.InitResult) error {
	if err := r.templateRenderer.Render("init_result", result); err != nil {
		return err
	}
	_, err := fmt.Fprintln(r.writer)
	return err
}

// RenderAdd renders the add command result
func (r *Renderer) RenderAdd(result *tdh.AddResult) error {
	if err := r.templateRenderer.Render("add_result", result); err != nil {
		return err
	}
	_, err := fmt.Fprintln(r.writer)
	return err
}

// RenderModify renders the modify command result
func (r *Renderer) RenderModify(result *tdh.ModifyResult) error {
	if err := r.templateRenderer.Render("modify_result", result); err != nil {
		return err
	}
	_, err := fmt.Fprintln(r.writer)
	return err
}

// RenderToggle renders the toggle command result
func (r *Renderer) RenderToggle(result *tdh.ToggleResult) error {
	if err := r.templateRenderer.Render("toggle_result", result); err != nil {
		return err
	}
	_, err := fmt.Fprintln(r.writer)
	return err
}

// RenderClean renders the clean command result
func (r *Renderer) RenderClean(result *tdh.CleanResult) error {
	if err := r.templateRenderer.Render("clean_result", result); err != nil {
		return err
	}
	_, err := fmt.Fprintln(r.writer)
	return err
}

// RenderReorder renders the reorder command result
func (r *Renderer) RenderReorder(result *tdh.ReorderResult) error {
	if err := r.templateRenderer.Render("reorder_result", result); err != nil {
		return err
	}
	_, err := fmt.Fprintln(r.writer)
	return err
}

// RenderSearch renders the search command result
func (r *Renderer) RenderSearch(result *tdh.SearchResult) error {
	if err := r.templateRenderer.Render("search_result", result); err != nil {
		return err
	}
	_, err := fmt.Fprintln(r.writer)
	return err
}

// RenderList renders the list command result
func (r *Renderer) RenderList(result *tdh.ListResult) error {
	return r.templateRenderer.Render("todo_list", result)
}

// RenderError renders an error message
func (r *Renderer) RenderError(err error) error {
	if renderErr := r.templateRenderer.Render("error", err.Error()); renderErr != nil {
		return renderErr
	}
	_, writeErr := fmt.Fprintln(r.writer)
	return writeErr
}

// RenderTemplate renders a template with data
func (r *Renderer) RenderTemplate(tmpl string, data interface{}) error {
	t, err := template.New("output").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	err = t.Execute(r.writer, data)
	if err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}
