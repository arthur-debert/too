package display

import (
	"fmt"
	"io"
	"os"
	"text/template"

	"github.com/arthur-debert/tdh/pkg/tdh"
	"github.com/arthur-debert/tdh/pkg/tdh/models"
)

// Renderer handles output formatting for tdh commands
type Renderer struct {
	writer io.Writer
}

// NewRenderer creates a new renderer
func NewRenderer(w io.Writer) *Renderer {
	if w == nil {
		w = os.Stdout
	}
	return &Renderer{writer: w}
}

// RenderInit renders the init command result
func (r *Renderer) RenderInit(result *tdh.InitResult) error {
	_, err := fmt.Fprintln(r.writer, result.Message)
	return err
}

// RenderAdd renders the add command result
func (r *Renderer) RenderAdd(result *tdh.AddResult) error {
	_, err := fmt.Fprintf(r.writer, "Added todo #%d: %s\n", result.Todo.ID, result.Todo.Text)
	return err
}

// RenderModify renders the modify command result
func (r *Renderer) RenderModify(result *tdh.ModifyResult) error {
	if _, err := fmt.Fprintf(r.writer, "Modified todo #%d\n", result.Todo.ID); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(r.writer, "  Old: %s\n", result.OldText); err != nil {
		return err
	}
	_, err := fmt.Fprintf(r.writer, "  New: %s\n", result.NewText)
	return err
}

// RenderToggle renders the toggle command result
func (r *Renderer) RenderToggle(result *tdh.ToggleResult) error {
	_, err := fmt.Fprintf(r.writer, "Toggled todo #%d from %s to %s\n",
		result.Todo.ID, result.OldStatus, result.NewStatus)
	return err
}

// RenderClean renders the clean command result
func (r *Renderer) RenderClean(result *tdh.CleanResult) error {
	if result.RemovedCount == 0 {
		if _, err := fmt.Fprintln(r.writer, "No finished todos to clean"); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintf(r.writer, "Removed %d finished todo(s)\n", result.RemovedCount); err != nil {
			return err
		}
		for _, todo := range result.RemovedTodos {
			if _, err := fmt.Fprintf(r.writer, "  - #%d: %s\n", todo.ID, todo.Text); err != nil {
				return err
			}
		}
	}
	_, err := fmt.Fprintf(r.writer, "%d active todo(s) remaining\n", result.ActiveCount)
	return err
}

// RenderReorder renders the reorder command result
func (r *Renderer) RenderReorder(result *tdh.ReorderResult) error {
	_, err := fmt.Fprintf(r.writer, "Swapped todos #%d and #%d\n",
		result.TodoA.ID, result.TodoB.ID)
	return err
}

// RenderSearch renders the search command result
func (r *Renderer) RenderSearch(result *tdh.SearchResult) error {
	if len(result.MatchedTodos) == 0 {
		_, err := fmt.Fprintf(r.writer, "No todos found matching '%s'\n", result.Query)
		return err
	}

	if _, err := fmt.Fprintf(r.writer, "Found %d todo(s) matching '%s':\n",
		len(result.MatchedTodos), result.Query); err != nil {
		return err
	}

	for _, todo := range result.MatchedTodos {
		r.renderTodo(todo)
	}
	return nil
}

// RenderList renders the list command result
func (r *Renderer) RenderList(result *tdh.ListResult) error {
	if len(result.Todos) == 0 {
		_, err := fmt.Fprintln(r.writer, "No todos found")
		return err
	}

	// Optional: render with a template if needed
	// For now, use simple output
	for _, todo := range result.Todos {
		r.renderTodo(todo)
	}

	_, err := fmt.Fprintf(r.writer, "\n%d todo(s), %d done\n",
		result.TotalCount, result.DoneCount)
	return err
}

// renderTodo renders a single todo (helper method)
func (r *Renderer) renderTodo(todo *models.Todo) {
	tdh.MakeOutput(todo, true)
}

// RenderError renders an error message
func (r *Renderer) RenderError(err error) error {
	_, writeErr := fmt.Fprintf(r.writer, "Error: %s\n", err.Error())
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
