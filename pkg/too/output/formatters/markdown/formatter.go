package markdown

import (
	"fmt"
	"io"
	"strings"

	"github.com/arthur-debert/too/pkg/too"
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/output"
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

// formatMultilineMarkdown formats text with newlines for markdown, indenting subsequent lines
func formatMultilineMarkdown(text string, indentStr string) string {
	lines := strings.Split(text, "\n")
	if len(lines) <= 1 {
		return text
	}

	// For markdown, we need to indent continuation lines with the same indent + extra spaces
	// to align with the text after the checkbox
	continuationIndent := indentStr + "      " // indent + "1. [ ] " equivalent spacing

	var result strings.Builder
	for i, line := range lines {
		if i == 0 {
			result.WriteString(line)
		} else {
			result.WriteString("\n" + continuationIndent + line)
		}
	}
	return result.String()
}

// renderTodos renders a list of todos as nested numbered lists
func (f *formatter) renderTodos(todos []*models.Todo, indent int) string {
	var result strings.Builder
	indentStr := strings.Repeat("   ", indent) // 3 spaces per indent level

	for i, todo := range todos {
		// Format: "1. [x] Todo text" or "1. [ ] Todo text"
		checkbox := "[ ]"
		if todo.GetStatus() == models.StatusDone {
			checkbox = "[x]"
		}

		// Format multi-line text properly
		formattedText := formatMultilineMarkdown(todo.Text, indentStr)
		result.WriteString(fmt.Sprintf("%s%d. %s %s\n", indentStr, i+1, checkbox, formattedText))

		// Render nested todos
		if len(todo.Items) > 0 {
			result.WriteString(f.renderTodos(todo.Items, indent+1))
		}
	}

	return result.String()
}

// RenderAdd renders the add command result as Markdown
func (f *formatter) RenderAdd(w io.Writer, result *too.AddResult) error {
	checkbox := "[ ]"
	if result.Todo.GetStatus() == models.StatusDone {
		checkbox = "[x]"
	}
	_, err := fmt.Fprintf(w, "Added todo: %s %s\n", checkbox, formatMultilineMarkdown(result.Todo.Text, ""))
	return err
}

// RenderModify renders the modify command result as Markdown
func (f *formatter) RenderModify(w io.Writer, result *too.ModifyResult) error {
	checkbox := "[ ]"
	if result.Todo.GetStatus() == models.StatusDone {
		checkbox = "[x]"
	}
	_, err := fmt.Fprintf(w, "Modified todo: %s %s\n", checkbox, formatMultilineMarkdown(result.Todo.Text, ""))
	return err
}

// RenderInit renders the init command result as Markdown
func (f *formatter) RenderInit(w io.Writer, result *too.InitResult) error {
	_, err := fmt.Fprintf(w, "Initialized todo collection at: `%s`\n", result.DBPath)
	return err
}

// RenderClean renders the clean command result as Markdown
func (f *formatter) RenderClean(w io.Writer, result *too.CleanResult) error {
	_, err := fmt.Fprintf(w, "Removed %d completed todo(s)\n", result.RemovedCount)
	return err
}

// RenderSearch renders the search command result as Markdown
func (f *formatter) RenderSearch(w io.Writer, result *too.SearchResult) error {
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
func (f *formatter) RenderList(w io.Writer, result *too.ListResult) error {
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
func (f *formatter) RenderComplete(w io.Writer, results []*too.CompleteResult) error {
	for _, result := range results {
		_, err := fmt.Fprintf(w, "Completed todo: [x] %s\n", formatMultilineMarkdown(result.Todo.Text, ""))
		if err != nil {
			return err
		}
	}
	return nil
}

// RenderReopen renders the reopen command results as Markdown
func (f *formatter) RenderReopen(w io.Writer, results []*too.ReopenResult) error {
	for _, result := range results {
		_, err := fmt.Fprintf(w, "Reopened todo: [ ] %s\n", formatMultilineMarkdown(result.Todo.Text, ""))
		if err != nil {
			return err
		}
	}
	return nil
}

// RenderMove renders the move command result as Markdown
func (f *formatter) RenderMove(w io.Writer, result *too.MoveResult) error {
	_, err := fmt.Fprintf(w, "Moved todo from %s to %s\n", result.OldPath, result.NewPath)
	return err
}

// RenderSwap renders the swap command result as Markdown

// RenderDataPath renders the datapath command result as Markdown
func (f *formatter) RenderDataPath(w io.Writer, result *too.ShowDataPathResult) error {
	_, err := fmt.Fprintf(w, "Data file path: `%s`\n", result.Path)
	return err
}

// RenderFormats renders the formats command result as Markdown
func (f *formatter) RenderFormats(w io.Writer, result *too.ListFormatsResult) error {
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
