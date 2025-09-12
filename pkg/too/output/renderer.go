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

	// Get the style map from the styles package
	styleMap := styles.GetLipbalmStyleMap()

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

// renderTodoCommand is a unified method for rendering todo commands with the todo_list template
func (r *LipbamlRenderer) renderTodoCommand(message, messageType string, todos []*models.IDMTodo, totalCount, doneCount int, highlightID string) error {
	wrapped := &TodoListWithMessage{
		Message:     message,
		MessageType: messageType,
		Todos:       todos,
		TotalCount:  totalCount,
		DoneCount:   doneCount,
		HighlightID: highlightID,
	}
	output, err := r.renderTemplate("todo_list", wrapped)
	if err != nil {
		return fmt.Errorf("failed to render todo command: %w", err)
	}
	_, err = fmt.Fprintln(r.Writer, output)
	return err
}


// templateFuncs returns custom functions for templates
func (r *LipbamlRenderer) templateFuncs() map[string]interface{} {
	return map[string]interface{}{
		"isDone": func(todo interface{}) bool {
			switch t := todo.(type) {
			case *models.IDMTodo:
				return t.GetStatus() == models.StatusDone
			case *HierarchicalTodo:
				return t.IDMTodo.GetStatus() == models.StatusDone
			default:
				return false
			}
		},
		"indent": func(level int) string {
			// Use 2 spaces per level for indentation
			return strings.Repeat("  ", level)
		},
		"lines": func(text string) []string {
			return strings.Split(text, "\n")
		},
		"getSymbol": func(status string) string {
			return styles.GetStatusSymbol(status)
		},
		"add": func(a, b int) int {
			return a + b
		},
		"len": func(s string) int {
			return len(s)
		},
		"repeat": func(s string, n int) string {
			return strings.Repeat(s, n)
		},
		"buildHierarchy": func(todos []*models.IDMTodo) []*HierarchicalTodo {
			return BuildHierarchy(todos)
		},
		"dict": func(values ...interface{}) map[string]interface{} {
			if len(values)%2 != 0 {
				panic("dict requires even number of arguments")
			}
			dict := make(map[string]interface{})
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					panic(fmt.Sprintf("dict keys must be strings, got %T", values[i]))
				}
				dict[key] = values[i+1]
			}
			return dict
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
func (r *LipbamlRenderer) RenderAdd(result *too.AddResult) error {
	message := fmt.Sprintf("Added todo #%s: %s", result.PositionPath, result.Todo.Text)
	
	// For short mode, just show the message without todos
	if result.Mode == "short" {
		return r.renderTodoCommand(message, "success", nil, 0, 0, "")
	}
	
	// For long mode, show message + todo list with highlight
	return r.renderTodoCommand(
		message,
		"success",
		result.AllTodos,
		result.TotalCount,
		result.DoneCount,
		result.Todo.UID,
	)
}

// RenderModify renders the modify command result using lipbalm
func (r *LipbamlRenderer) RenderModify(result *too.ModifyResult) error {
	return r.renderTodoCommand(
		fmt.Sprintf("Modified todo: %s", result.Todo.Text),
		"info",
		result.AllTodos,
		result.TotalCount,
		result.DoneCount,
		result.Todo.UID,
	)
}

// RenderInit renders the init command result using lipbalm
func (r *LipbamlRenderer) RenderInit(result *too.InitResult) error {
	message := NewPlainMessage(result.Message)
	output, err := r.renderTemplate("message", message)
	if err != nil {
		return fmt.Errorf("failed to render init result: %w", err)
	}
	_, err = fmt.Fprintln(r.Writer, output)
	return err
}

// RenderClean renders the clean command result using lipbalm
func (r *LipbamlRenderer) RenderClean(result *too.CleanResult) error {
	message := fmt.Sprintf("Removed %d finished todo(s)", result.RemovedCount)
	messageType := "success"
	if result.RemovedCount == 0 {
		message = "No finished todos to clean"
		messageType = "warning"
	}
	
	return r.renderTodoCommand(
		message,
		messageType,
		result.ActiveTodos,
		result.ActiveCount,
		0, // After clean, no done todos remain
		"", // No highlight
	)
}

// RenderSearch renders the search command result using lipbalm
func (r *LipbamlRenderer) RenderSearch(result *too.SearchResult) error {
	matchCount := len(result.MatchedTodos)
	message := fmt.Sprintf("Found %d match", matchCount)
	if matchCount != 1 {
		message = fmt.Sprintf("Found %d matches", matchCount)
	}
	if result.Query != "" {
		message += fmt.Sprintf(" for \"%s\"", result.Query)
	}
	
	messageType := "info"
	if matchCount == 0 {
		messageType = "warning"
	}
	
	return r.renderTodoCommand(
		message,
		messageType,
		result.MatchedTodos,
		result.TotalCount,
		0, // Search doesn't track done count separately
		"", // No highlight
	)
}

// RenderList renders the list command result using lipbalm
func (r *LipbamlRenderer) RenderList(result *too.ListResult) error {
	return r.renderTodoCommand(
		"", // No message for basic list
		"",
		result.Todos,
		result.TotalCount,
		result.DoneCount,
		"", // No highlight
	)
}

// RenderComplete renders the complete command results using lipbalm
func (r *LipbamlRenderer) RenderComplete(results []*too.CompleteResult) error {
	for _, result := range results {
		message := fmt.Sprintf("✓ Completed: %s", result.Todo.Text)
		
		// For short mode, just show the message
		if result.Mode == "short" {
			if err := r.renderTodoCommand(message, "success", nil, 0, 0, ""); err != nil {
				return err
			}
			continue
		}
		
		// For long mode, show message + todo list with highlight
		if err := r.renderTodoCommand(
			message,
			"success",
			result.AllTodos,
			result.TotalCount,
			result.DoneCount,
			result.Todo.UID,
		); err != nil {
			return err
		}
	}
	return nil
}

// RenderReopen renders the reopen command results using lipbalm
func (r *LipbamlRenderer) RenderReopen(results []*too.ReopenResult) error {
	for _, result := range results {
		if err := r.renderTodoCommand(
			fmt.Sprintf("○ Reopened: %s", result.Todo.Text),
			"warning",
			result.AllTodos,
			result.TotalCount,
			result.DoneCount,
			result.Todo.UID,
		); err != nil {
			return err
		}
	}
	return nil
}

// RenderMove renders the move command result using lipbalm
func (r *LipbamlRenderer) RenderMove(result *too.MoveResult) error {
	return r.renderTodoCommand(
		fmt.Sprintf("Moved todo from %s to %s: %s", result.OldPath, result.NewPath, result.Todo.Text),
		"success",
		result.AllTodos,
		result.TotalCount,
		result.DoneCount,
		result.Todo.UID,
	)
}

// RenderDataPath renders the datapath command result using lipbalm
func (r *LipbamlRenderer) RenderDataPath(result *too.ShowDataPathResult) error {
	message := NewPlainMessage(result.Path)
	output, err := r.renderTemplate("message", message)
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
	message := &Message{
		Text:  "Error: " + err.Error(),
		Level: "error",
	}
	output, renderErr := r.renderTemplate("message", message)
	if renderErr != nil {
		return renderErr
	}
	_, writeErr := fmt.Fprintln(r.Writer, output)
	return writeErr
}
