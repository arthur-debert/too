package output

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/arthur-debert/too/pkg/lipbalm"
	"github.com/arthur-debert/too/pkg/too"
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/output/styles"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

// LipbamlRenderer is a renderer that uses lipbalm for styled output
type LipbamlRenderer struct {
	Writer    io.Writer // Exported to allow formatter to update it
	useColor  bool
	styles    lipbalm.StyleMap
	templates map[string]string
}

// NewLipbamlRenderer creates a new lipbalm-based renderer
func NewLipbamlRenderer(w io.Writer, useColor bool) (*LipbamlRenderer, error) {
	if w == nil {
		w = os.Stdout
	}

	// Set up lipgloss renderer with proper color detection
	lipglossRenderer := lipgloss.NewRenderer(w)
	if useColor {
		// Force color output for testing
		lipglossRenderer.SetColorProfile(termenv.TrueColor)
	} else {
		lipglossRenderer.SetColorProfile(termenv.Ascii)
	}
	lipbalm.SetDefaultRenderer(lipglossRenderer)

	// Define the style map with semantic names
	styleMap := lipbalm.StyleMap{
		// Status and result styles
		"success": lipgloss.NewStyle().
			Foreground(styles.SUCCESS_COLOR),
		"error": lipgloss.NewStyle().
			Foreground(styles.ERROR_COLOR).
			Bold(true),
		"warning": lipgloss.NewStyle().
			Foreground(styles.WARNING_COLOR),
		"info": lipgloss.NewStyle().
			Foreground(styles.INFO_COLOR),

		// Todo state styles
		"todo-done": lipgloss.NewStyle().
			Foreground(styles.SUCCESS_COLOR).
			Bold(true),
		"todo-pending": lipgloss.NewStyle().
			Foreground(styles.ERROR_COLOR).
			Bold(true),

		// UI element styles
		"position": lipgloss.NewStyle().
			Foreground(styles.SUBDUED_TEXT),
		"muted": lipgloss.NewStyle().
			Foreground(styles.VERY_FAINT_TEXT).
			Faint(true),
		"highlighted-todo": lipgloss.NewStyle().
			Bold(true),
		"subdued": lipgloss.NewStyle().
			Foreground(styles.SUBDUED_TEXT),
		"accent": lipgloss.NewStyle().
			Foreground(styles.ACCENT_COLOR),
		"count": lipgloss.NewStyle().
			Foreground(styles.INFO_COLOR),
		"label": lipgloss.NewStyle().
			Foreground(styles.SUBDUED_TEXT),
		"value": lipgloss.NewStyle().
			Foreground(styles.PRIMARY_TEXT),
	}

	r := &LipbamlRenderer{
		Writer:    w,
		useColor:  useColor,
		styles:    styleMap,
		templates: make(map[string]string),
	}

	// Load all templates
	entries, err := templateFS.ReadDir("templates")
	if err != nil {
		return nil, fmt.Errorf("failed to read templates directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".tmpl") {
			content, err := templateFS.ReadFile("templates/" + entry.Name())
			if err != nil {
				return nil, fmt.Errorf("failed to read template %s: %w", entry.Name(), err)
			}
			// Store without .tmpl extension
			templateName := strings.TrimSuffix(entry.Name(), ".tmpl")
			r.templates[templateName] = string(content)
		}
	}

	return r, nil
}

// getStatusSymbol returns the appropriate unicode symbol based on todo status
// ○ scheduled -> start status (pending)
// ◐ in progress -> parent with mixed completion states
// ● done -> completed
// ⊘ deleted -> deleted (if supported)
func getStatusSymbol(todo *models.IDMTodo, children []*HierarchicalTodo) string {
	// Check if deleted status exists
	if status, exists := todo.GetWorkflowStatus("status"); exists && status == "deleted" {
		return "⊘"
	}
	
	// Check completion status
	if todo.GetStatus() == models.StatusDone {
		return "●"
	}
	
	// Check if this todo has children with mixed completion states
	if len(children) > 0 {
		hasComplete := false
		hasPending := false
		
		for _, child := range children {
			if child.IDMTodo.GetStatus() == models.StatusDone {
				hasComplete = true
			} else {
				hasPending = true
			}
			
			if hasComplete && hasPending {
				return "◐" // In progress - mixed states
			}
		}
	}
	
	// Default to scheduled/pending
	return "○"
}

// formatMultilineText formats text with newlines, indenting subsequent lines
func formatMultilineText(text string, baseIndent string, columnWidth int) string {
	lines := strings.Split(text, "\n")
	if len(lines) <= 1 {
		return text
	}

	// Calculate the indentation for continuation lines
	// baseIndent + position column (6) + " | " (3) + status symbol (1) + " " (1) = baseIndent + 11
	continuationIndent := baseIndent + strings.Repeat(" ", columnWidth+11)

	// Build the result with proper indentation
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

// templateFuncs returns custom functions for templates
func (r *LipbamlRenderer) templateFuncs() map[string]interface{} {
	return map[string]interface{}{
		"isDone": func(todo *models.IDMTodo) bool {
			return todo.GetStatus() == models.StatusDone
		},
		"padPosition": func(pos int) string {
			return fmt.Sprintf("%6d", pos)
		},
		"getIndent": func(level int) string {
			config := too.GetConfig()
			indent := strings.Repeat(config.Display.IndentString, config.Display.IndentSize)
			return strings.Repeat(indent, level)
		},
		"formatMultiline": func(text string, indent int) string {
			// For use in templates where we don't have the full context
			baseIndent := strings.Repeat(" ", indent)
			return formatMultilineText(text, baseIndent, 6)
		},
		"renderNestedTodosWithHighlight": func(todos []*models.IDMTodo, parentPath string, level int, highlightID string) string {
			// Build hierarchical structure from flat list
			hierarchical := BuildHierarchy(todos)
			return r.renderHierarchicalTodosWithHighlight(hierarchical, parentPath, level, highlightID)
		},
		"renderNestedTodos": func(todos []*models.IDMTodo, parentPath string, level int) string {
			// Build hierarchical structure from flat list
			hierarchical := BuildHierarchy(todos)
			return r.renderHierarchicalTodos(hierarchical, parentPath, level)
		},
	}
}

// renderHierarchicalTodosWithHighlight renders hierarchical todos with optional highlighting
func (r *LipbamlRenderer) renderHierarchicalTodosWithHighlight(todos []*HierarchicalTodo, parentPath string, level int, highlightID string) string {
	var result strings.Builder
	for i, todo := range todos {
		// Use array index + 1 as position
		position := i + 1
		path := parentPath
		if path == "" {
			path = fmt.Sprintf("%d", position)
		} else {
			path = fmt.Sprintf("%s.%d", parentPath, position)
		}

		// Render this todo with its path and indentation
		indent := r.templateFuncs()["getIndent"].(func(int) string)(level)
		statusSymbol := getStatusSymbol(todo.IDMTodo, todo.Children)

		// Format the todo text with proper indentation for multi-line content
		formattedText := formatMultilineText(todo.Text, indent, 6)

		// Apply muted style if this is not the highlighted todo
		if highlightID != "" && todo.UID != highlightID {
			// For non-highlighted todos, wrap entire line in muted tag
			result.WriteString(fmt.Sprintf("%s<muted>%6s | %s %s</muted>\n",
				indent, path, statusSymbol, formattedText))
		} else {
			// For highlighted todo, wrap everything in highlighted-todo tag for bold
			result.WriteString(fmt.Sprintf("%s<highlighted-todo>%6s | %s %s</highlighted-todo>\n",
				indent, path, statusSymbol, formattedText))
		}

		// Recursively render children
		if len(todo.Children) > 0 {
			childrenOutput := r.renderHierarchicalTodosWithHighlight(todo.Children, path, level+1, highlightID)
			result.WriteString(childrenOutput)
		}
	}
	return result.String()
}

// renderHierarchicalTodos renders hierarchical todos without highlighting
func (r *LipbamlRenderer) renderHierarchicalTodos(todos []*HierarchicalTodo, parentPath string, level int) string {
	var result strings.Builder
	for i, todo := range todos {
		// Use array index + 1 as position
		position := i + 1
		path := parentPath
		if path == "" {
			path = fmt.Sprintf("%d", position)
		} else {
			path = fmt.Sprintf("%s.%d", parentPath, position)
		}

		// Render this todo with its path and indentation
		indent := r.templateFuncs()["getIndent"].(func(int) string)(level)
		statusSymbol := getStatusSymbol(todo.IDMTodo, todo.Children)
		
		// Determine style based on status
		statusStyle := "todo-pending"
		if todo.IDMTodo.GetStatus() == models.StatusDone {
			statusStyle = "todo-done"
		}

		// Format the todo text with proper indentation for multi-line content
		formattedText := formatMultilineText(todo.Text, indent, 6)
		result.WriteString(fmt.Sprintf("%s<subdued>%6s</subdued> | <%s>%s</%s> %s\n",
			indent, path, statusStyle, statusSymbol, statusStyle, formattedText))

		// Recursively render children
		if len(todo.Children) > 0 {
			childrenOutput := r.renderHierarchicalTodos(todo.Children, path, level+1)
			result.WriteString(childrenOutput)
		}
	}
	return result.String()
}

// renderTemplate renders a template with the given data
func (r *LipbamlRenderer) renderTemplate(templateName string, data interface{}) (string, error) {
	tmplContent, ok := r.templates[templateName]
	if !ok {
		return "", fmt.Errorf("template '%s' not found", templateName)
	}

	// For templates that use other templates, we need to process all of them together
	if strings.Contains(tmplContent, "{{template") {
		// Combine all templates into one for parsing
		var allTemplates strings.Builder
		for name, content := range r.templates {
			allTemplates.WriteString(fmt.Sprintf(`{{define "%s.tmpl"}}%s{{end}}`, name, content))
		}

		// Parse and execute with custom functions using standard template package
		tmpl, err := template.New("combined").Funcs(template.FuncMap(r.templateFuncs())).Parse(allTemplates.String())
		if err != nil {
			return "", fmt.Errorf("failed to parse templates: %w", err)
		}

		var buf bytes.Buffer
		if err := tmpl.ExecuteTemplate(&buf, templateName+".tmpl", data); err != nil {
			return "", fmt.Errorf("failed to execute template: %w", err)
		}

		// Now expand the lipbalm tags
		return lipbalm.ExpandTags(buf.String(), r.styles)
	}

	// For simple templates, we need to parse with functions first, then use lipbalm
	tmpl, err := template.New(templateName).Funcs(template.FuncMap(r.templateFuncs())).Parse(tmplContent)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	// Now expand the lipbalm tags
	return lipbalm.ExpandTags(buf.String(), r.styles)
}

// RenderAdd renders the add command result using lipbalm
func (r *LipbamlRenderer) RenderAdd(result *too.AddResult) error {
	output, err := r.renderTemplate("add_result", result)
	if err != nil {
		return fmt.Errorf("failed to render add result: %w", err)
	}
	_, err = fmt.Fprintln(r.Writer, output)
	return err
}

// RenderModify renders the modify command result using lipbalm
func (r *LipbamlRenderer) RenderModify(result *too.ModifyResult) error {
	output, err := r.renderTemplate("modify_result", result)
	if err != nil {
		return fmt.Errorf("failed to render modify result: %w", err)
	}
	_, err = fmt.Fprintln(r.Writer, output)
	return err
}

// RenderInit renders the init command result using lipbalm
func (r *LipbamlRenderer) RenderInit(result *too.InitResult) error {
	output, err := r.renderTemplate("init_result", result)
	if err != nil {
		return fmt.Errorf("failed to render init result: %w", err)
	}
	_, err = fmt.Fprintln(r.Writer, output)
	return err
}

// RenderClean renders the clean command result using lipbalm
func (r *LipbamlRenderer) RenderClean(result *too.CleanResult) error {
	output, err := r.renderTemplate("clean_result", result)
	if err != nil {
		return fmt.Errorf("failed to render clean result: %w", err)
	}
	_, err = fmt.Fprintln(r.Writer, output)
	return err
}

// RenderSearch renders the search command result using lipbalm
func (r *LipbamlRenderer) RenderSearch(result *too.SearchResult) error {
	output, err := r.renderTemplate("search_result", result)
	if err != nil {
		return fmt.Errorf("failed to render search result: %w", err)
	}
	_, err = fmt.Fprintln(r.Writer, output)
	return err
}

// RenderList renders the list command result using lipbalm
func (r *LipbamlRenderer) RenderList(result *too.ListResult) error {
	output, err := r.renderTemplate("todo_list", result)
	if err != nil {
		return fmt.Errorf("failed to render list: %w", err)
	}
	_, err = fmt.Fprint(r.Writer, output)
	return err
}

// RenderComplete renders the complete command results using lipbalm
func (r *LipbamlRenderer) RenderComplete(results []*too.CompleteResult) error {
	for _, result := range results {
		output, err := r.renderTemplate("complete_result", result)
		if err != nil {
			return fmt.Errorf("failed to render complete result: %w", err)
		}
		_, err = fmt.Fprintln(r.Writer, output)
		if err != nil {
			return err
		}
	}
	return nil
}

// RenderReopen renders the reopen command results using lipbalm
func (r *LipbamlRenderer) RenderReopen(results []*too.ReopenResult) error {
	for _, result := range results {
		output, err := r.renderTemplate("reopen_result", result)
		if err != nil {
			return fmt.Errorf("failed to render reopen result: %w", err)
		}
		_, err = fmt.Fprintln(r.Writer, output)
		if err != nil {
			return err
		}
	}
	return nil
}

// RenderMove renders the move command result using lipbalm
func (r *LipbamlRenderer) RenderMove(result *too.MoveResult) error {
	output, err := r.renderTemplate("move_result", result)
	if err != nil {
		return fmt.Errorf("failed to render move result: %w", err)
	}
	_, err = fmt.Fprintln(r.Writer, output)
	return err
}

// RenderDataPath renders the datapath command result using lipbalm
func (r *LipbamlRenderer) RenderDataPath(result *too.ShowDataPathResult) error {
	output, err := r.renderTemplate("datapath_result", result)
	if err != nil {
		return fmt.Errorf("failed to render datapath result: %w", err)
	}
	_, err = fmt.Fprintln(r.Writer, output)
	return err
}

// RenderFormats renders the formats command result using lipbalm
func (r *LipbamlRenderer) RenderFormats(result *too.ListFormatsResult) error {
	output, err := r.renderTemplate("formats_result", result)
	if err != nil {
		return fmt.Errorf("failed to render formats result: %w", err)
	}
	_, err = fmt.Fprintln(r.Writer, output)
	return err
}

// RenderError renders an error message
func (r *LipbamlRenderer) RenderError(err error) error {
	output, renderErr := r.renderTemplate("error", err.Error())
	if renderErr != nil {
		return renderErr
	}
	_, writeErr := fmt.Fprintln(r.Writer, output)
	return writeErr
}
