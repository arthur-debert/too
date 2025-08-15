package output

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/arthur-debert/tdh/pkg/lipbalm"
	"github.com/arthur-debert/tdh/pkg/tdh"
	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/arthur-debert/tdh/pkg/tdh/output/styles"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

// LipbamlRenderer is a renderer that uses lipbalm for styled output
type LipbamlRenderer struct {
	writer    io.Writer
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
			Foreground(styles.MUTED_TEXT),
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
		writer:    w,
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

// templateFuncs returns custom functions for templates
func (r *LipbamlRenderer) templateFuncs() map[string]interface{} {
	return map[string]interface{}{
		"isDone": func(todo *models.Todo) bool {
			return todo.Status == models.StatusDone
		},
		"padPosition": func(pos int) string {
			return fmt.Sprintf("%6d", pos)
		},
		"getIndent": func(level int) string {
			config := tdh.GetConfig()
			indent := strings.Repeat(config.Display.IndentString, config.Display.IndentSize)
			return strings.Repeat(indent, level)
		},
		"renderNestedTodos": func(todos []*models.Todo, parentPath string, level int) string {
			var result strings.Builder
			for _, todo := range todos {
				path := parentPath
				if path == "" {
					path = fmt.Sprintf("%d", todo.Position)
				} else {
					path = fmt.Sprintf("%s.%d", parentPath, todo.Position)
				}

				// Render this todo with its path and indentation
				indent := r.templateFuncs()["getIndent"].(func(int) string)(level)
				isDone := r.templateFuncs()["isDone"].(func(*models.Todo) bool)(todo)
				statusSymbol := "✕"
				statusStyle := "todo-pending"
				if isDone {
					statusSymbol = "✓"
					statusStyle = "todo-done"
				}

				result.WriteString(fmt.Sprintf("%s<subdued>%6s</subdued> | <%s>%s</%s> %s\n",
					indent, path, statusStyle, statusSymbol, statusStyle, todo.Text))

				// Recursively render children
				if len(todo.Items) > 0 {
					childrenOutput := r.templateFuncs()["renderNestedTodos"].(func([]*models.Todo, string, int) string)(
						todo.Items, path, level+1)
					result.WriteString(childrenOutput)
				}
			}
			return result.String()
		},
	}
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
func (r *LipbamlRenderer) RenderAdd(result *tdh.AddResult) error {
	output, err := r.renderTemplate("add_result", result)
	if err != nil {
		return fmt.Errorf("failed to render add result: %w", err)
	}
	_, err = fmt.Fprintln(r.writer, output)
	return err
}

// RenderModify renders the modify command result using lipbalm
func (r *LipbamlRenderer) RenderModify(result *tdh.ModifyResult) error {
	output, err := r.renderTemplate("modify_result", result)
	if err != nil {
		return fmt.Errorf("failed to render modify result: %w", err)
	}
	_, err = fmt.Fprintln(r.writer, output)
	return err
}

// RenderInit renders the init command result using lipbalm
func (r *LipbamlRenderer) RenderInit(result *tdh.InitResult) error {
	output, err := r.renderTemplate("init_result", result)
	if err != nil {
		return fmt.Errorf("failed to render init result: %w", err)
	}
	_, err = fmt.Fprintln(r.writer, output)
	return err
}

// RenderClean renders the clean command result using lipbalm
func (r *LipbamlRenderer) RenderClean(result *tdh.CleanResult) error {
	output, err := r.renderTemplate("clean_result", result)
	if err != nil {
		return fmt.Errorf("failed to render clean result: %w", err)
	}
	_, err = fmt.Fprintln(r.writer, output)
	return err
}

// RenderReorder renders the reorder command result using lipbalm
func (r *LipbamlRenderer) RenderReorder(result *tdh.ReorderResult) error {
	output, err := r.renderTemplate("reorder_result", result)
	if err != nil {
		return fmt.Errorf("failed to render reorder result: %w", err)
	}
	_, err = fmt.Fprintln(r.writer, output)
	return err
}

// RenderSearch renders the search command result using lipbalm
func (r *LipbamlRenderer) RenderSearch(result *tdh.SearchResult) error {
	output, err := r.renderTemplate("search_result", result)
	if err != nil {
		return fmt.Errorf("failed to render search result: %w", err)
	}
	_, err = fmt.Fprintln(r.writer, output)
	return err
}

// RenderList renders the list command result using lipbalm
func (r *LipbamlRenderer) RenderList(result *tdh.ListResult) error {
	output, err := r.renderTemplate("todo_list", result)
	if err != nil {
		return fmt.Errorf("failed to render list: %w", err)
	}
	_, err = fmt.Fprint(r.writer, output)
	return err
}

// RenderComplete renders the complete command results using lipbalm
func (r *LipbamlRenderer) RenderComplete(results []*tdh.CompleteResult) error {
	for _, result := range results {
		output, err := r.renderTemplate("complete_result", result)
		if err != nil {
			return fmt.Errorf("failed to render complete result: %w", err)
		}
		_, err = fmt.Fprintln(r.writer, output)
		if err != nil {
			return err
		}
	}
	return nil
}

// RenderReopen renders the reopen command results using lipbalm
func (r *LipbamlRenderer) RenderReopen(results []*tdh.ReopenResult) error {
	for _, result := range results {
		output, err := r.renderTemplate("reopen_result", result)
		if err != nil {
			return fmt.Errorf("failed to render reopen result: %w", err)
		}
		_, err = fmt.Fprintln(r.writer, output)
		if err != nil {
			return err
		}
	}
	return nil
}

// RenderMove renders the move command result using lipbalm
func (r *LipbamlRenderer) RenderMove(result *tdh.MoveResult) error {
	output, err := r.renderTemplate("move_result", result)
	if err != nil {
		return fmt.Errorf("failed to render move result: %w", err)
	}
	_, err = fmt.Fprintln(r.writer, output)
	return err
}

// RenderError renders an error message
func (r *LipbamlRenderer) RenderError(err error) error {
	output, renderErr := r.renderTemplate("error", err.Error())
	if renderErr != nil {
		return renderErr
	}
	_, writeErr := fmt.Fprintln(r.writer, output)
	return writeErr
}
