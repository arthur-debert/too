package output

import (
	"fmt"
	"strings"

	"github.com/arthur-debert/too/pkg/too"
	"github.com/arthur-debert/too/pkg/too/commands/formats"
	"github.com/arthur-debert/too/pkg/too/models"
)

// renderChangeAsMarkdown converts a ChangeResult to markdown format
func renderChangeAsMarkdown(result *too.ChangeResult) string {
	var sb strings.Builder

	// Show affected todos summary
	if len(result.AffectedTodos) > 0 {
		verb := result.Command
		if !strings.HasSuffix(verb, "ed") {
			verb = verb + "ed"
		}
		sb.WriteString(fmt.Sprintf("%s %d todo(s)\n\n", strings.Title(verb), len(result.AffectedTodos)))
	}

	// Render all todos
	if len(result.AllTodos) > 0 {
		sb.WriteString(renderTodosAsMarkdown(result.AllTodos))
		sb.WriteString(fmt.Sprintf("\n---\n%d todo(s), %d done\n", result.TotalCount, result.DoneCount))
	}

	return sb.String()
}

// renderFormatsAsMarkdown converts a formats result to markdown
func renderFormatsAsMarkdown(result *formats.Result) string {
	var sb strings.Builder
	sb.WriteString("Available output formats:\n")
	
	for _, format := range result.Formats {
		sb.WriteString(fmt.Sprintf("- **%s**: %s\n", format.Name, format.Description))
	}

	return sb.String()
}

// renderTodosAsMarkdown converts a list of todos to markdown format
func renderTodosAsMarkdown(todos []*models.Todo) string {
	// Build hierarchical structure
	hierarchical := models.BuildHierarchy(todos)
	return renderHierarchicalTodosAsMarkdown(hierarchical, 0)
}

// renderHierarchicalTodosAsMarkdown recursively renders hierarchical todos as markdown
func renderHierarchicalTodosAsMarkdown(todos []*models.HierarchicalTodo, indent int) string {
	var sb strings.Builder
	indentStr := strings.Repeat("   ", indent)

	for i, todo := range todos {
		checkbox := "[ ]"
		if todo.GetStatus() == models.StatusDone {
			checkbox = "[x]"
		}

		// Format multi-line text properly
		text := formatMultilineMarkdown(todo.Text, indentStr)
		sb.WriteString(fmt.Sprintf("%s%d. %s %s\n", indentStr, i+1, checkbox, text))

		// Render nested todos
		if len(todo.Children) > 0 {
			sb.WriteString(renderHierarchicalTodosAsMarkdown(todo.Children, indent+1))
		}
	}

	return sb.String()
}

// formatMultilineMarkdown formats multi-line text for markdown with proper indentation
func formatMultilineMarkdown(text string, indentStr string) string {
	lines := strings.Split(text, "\n")
	if len(lines) <= 1 {
		return text
	}

	continuationIndent := indentStr + "      "
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