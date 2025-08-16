package markdown

import (
	"fmt"
	"io"
	"strings"

	"github.com/arthur-debert/tdh/pkg/tdh"
	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/arthur-debert/tdh/pkg/tdh/output"
)

// formatter implements the Formatter interface for Markdown output
type formatter struct{}


// New creates a new Markdown formatter
func New() output.Formatter {
	return &formatter{}
}

// Name returns the formatter name
func (f *formatter) Name() string {
	return "markdown"
}

// Description returns the formatter description
func (f *formatter) Description() string {
	return "Markdown output for documentation and notes"
}

// renderTodos renders a list of todos as nested numbered lists
func (f *formatter) renderTodos(todos []*models.Todo, indent int) string {
	var result strings.Builder
	indentStr := strings.Repeat("   ", indent) // 3 spaces per indent level

	for i, todo := range todos {
		// Format: "1. [x] Todo text" or "1. [ ] Todo text"
		checkbox := "[ ]"
		if todo.Status == models.StatusDone {
			checkbox = "[x]"
		}

		result.WriteString(fmt.Sprintf("%s%d. %s %s\n", indentStr, i+1, checkbox, todo.Text))

		// Render nested todos
		if len(todo.Items) > 0 {
			result.WriteString(f.renderTodos(todo.Items, indent+1))
		}
	}

	return result.String()
}

// RenderAdd renders the add command result as Markdown
func (f *formatter) RenderAdd(w io.Writer, result *tdh.AddResult) error {
	checkbox := "[ ]"
	if result.Todo.Status == models.StatusDone {
		checkbox = "[x]"
	}
	_, err := fmt.Fprintf(w, "Added todo #%d: %s %s\n", result.Todo.Position, checkbox, result.Todo.Text)
	return err
}

// RenderModify renders the modify command result as Markdown
func (f *formatter) RenderModify(w io.Writer, result *tdh.ModifyResult) error {
	checkbox := "[ ]"
	if result.Todo.Status == models.StatusDone {
		checkbox = "[x]"
	}
	_, err := fmt.Fprintf(w, "Modified todo #%d: %s %s\n", result.Todo.Position, checkbox, result.Todo.Text)
	return err
}

// RenderInit renders the init command result as Markdown
func (f *formatter) RenderInit(w io.Writer, result *tdh.InitResult) error {
	_, err := fmt.Fprintf(w, "Initialized todo collection at: `%s`\n", result.DBPath)
	return err
}

// RenderClean renders the clean command result as Markdown
func (f *formatter) RenderClean(w io.Writer, result *tdh.CleanResult) error {
	_, err := fmt.Fprintf(w, "Removed %d completed todo(s)\n", result.RemovedCount)
	return err
}

// RenderSearch renders the search command result as Markdown
func (f *formatter) RenderSearch(w io.Writer, result *tdh.SearchResult) error {
	if len(result.MatchedTodos) == 0 {
		_, err := fmt.Fprintln(w, "No matching todos found")
		return err
	}

	_, err := fmt.Fprintf(w, "Found %d matching todo(s):\n\n", len(result.MatchedTodos))
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(w, f.renderTodos(result.MatchedTodos, 0))
	return err
}

// RenderList renders the list command result as Markdown
func (f *formatter) RenderList(w io.Writer, result *tdh.ListResult) error {
	if len(result.Todos) == 0 {
		_, err := fmt.Fprintln(w, "No todos")
		return err
	}

	// Render the todo list
	_, err := fmt.Fprint(w, f.renderTodos(result.Todos, 0))
	if err != nil {
		return err
	}

	// Add summary
	if result.TotalCount > 0 {
		_, err = fmt.Fprintf(w, "\n---\n%d todo(s), %d done\n", result.TotalCount, result.DoneCount)
	}

	return err
}

// RenderComplete renders the complete command results as Markdown
func (f *formatter) RenderComplete(w io.Writer, results []*tdh.CompleteResult) error {
	for _, result := range results {
		_, err := fmt.Fprintf(w, "Completed todo #%d: [x] %s\n", result.Todo.Position, result.Todo.Text)
		if err != nil {
			return err
		}
	}
	return nil
}

// RenderReopen renders the reopen command results as Markdown
func (f *formatter) RenderReopen(w io.Writer, results []*tdh.ReopenResult) error {
	for _, result := range results {
		_, err := fmt.Fprintf(w, "Reopened todo #%d: [ ] %s\n", result.Todo.Position, result.Todo.Text)
		if err != nil {
			return err
		}
	}
	return nil
}

// RenderMove renders the move command result as Markdown
func (f *formatter) RenderMove(w io.Writer, result *tdh.MoveResult) error {
	_, err := fmt.Fprintf(w, "Moved todo from %s to %s\n", result.OldPath, result.NewPath)
	return err
}

// RenderSwap renders the swap command result as Markdown
func (f *formatter) RenderSwap(w io.Writer, result *tdh.SwapResult) error {
	_, err := fmt.Fprintf(w, "Swapped todo from %s to %s\n", result.OldPath, result.NewPath)
	return err
}

// RenderDataPath renders the datapath command result as Markdown
func (f *formatter) RenderDataPath(w io.Writer, result *tdh.ShowDataPathResult) error {
	_, err := fmt.Fprintf(w, "Data file path: `%s`\n", result.Path)
	return err
}

// RenderFormats renders the formats command result as Markdown
func (f *formatter) RenderFormats(w io.Writer, result *tdh.ListFormatsResult) error {
	_, err := fmt.Fprintln(w, "Available output formats:")
	if err != nil {
		return err
	}

	for _, format := range result.Formats {
		_, err = fmt.Fprintf(w, "- **%s**: %s\n", format.Name, format.Description)
		if err != nil {
			return err
		}
	}

	return nil
}

// RenderError renders an error message as Markdown
func (f *formatter) RenderError(w io.Writer, err error) error {
	_, writeErr := fmt.Fprintf(w, "**Error:** %s\n", err.Error())
	return writeErr
}
