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
func (f *formatter) renderTodos(todos []*models.IDMTodo, indent int) string {
	// Build hierarchical structure
	hierarchical := output.BuildHierarchy(todos)
	return f.renderHierarchicalTodos(hierarchical, indent)
}

// renderHierarchicalTodos renders hierarchical todos
func (f *formatter) renderHierarchicalTodos(todos []*output.HierarchicalTodo, indent int) string {
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
		if len(todo.Children) > 0 {
			result.WriteString(f.renderHierarchicalTodos(todo.Children, indent+1))
		}
	}

	return result.String()
}

// RenderChange renders any command that changes todos as Markdown
func (f *formatter) RenderChange(w io.Writer, result *too.ChangeResult) error {
	// Show affected todos summary
	if len(result.AffectedTodos) > 0 {
		verb := result.Command
		if !strings.HasSuffix(verb, "ed") {
			verb = verb + "ed"
		}
		_, err := fmt.Fprintf(w, "%s %d todo(s)\n\n", strings.Title(verb), len(result.AffectedTodos))
		if err != nil {
			return err
		}
	}
	
	// Render all todos
	if len(result.AllTodos) > 0 {
		_, err := fmt.Fprint(w, f.renderTodos(result.AllTodos, 0))
		if err != nil {
			return err
		}
		
		// Add summary
		_, err = fmt.Fprintf(w, "\n---\n%d todo(s), %d done\n", result.TotalCount, result.DoneCount)
		return err
	}
	
	return nil
}

// RenderInit renders the init command result as Markdown
func (f *formatter) RenderInit(w io.Writer, result *too.InitResult) error {
	_, err := fmt.Fprintf(w, "Initialized todo collection at: `%s`\n", result.DBPath)
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
