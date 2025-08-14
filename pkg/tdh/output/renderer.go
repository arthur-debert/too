package output

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"os"
	"regexp"
	"text/template"

	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/arthur-debert/tdh/pkg/tdh/output/styles"
	"github.com/charmbracelet/lipgloss"
	"github.com/pterm/pterm"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

// TemplateRenderer is the new template-based renderer
type TemplateRenderer struct {
	writer    io.Writer
	templates map[string]*template.Template
	useColor  bool
}

// NewTemplateRenderer creates a new template-based renderer
func NewTemplateRenderer(w io.Writer, useColor bool) (*TemplateRenderer, error) {
	if w == nil {
		w = os.Stdout
	}

	r := &TemplateRenderer{
		writer:    w,
		templates: make(map[string]*template.Template),
		useColor:  useColor,
	}

	// Load all templates
	entries, err := templateFS.ReadDir("templates")
	if err != nil {
		return nil, fmt.Errorf("failed to read templates directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		content, err := templateFS.ReadFile("templates/" + name)
		if err != nil {
			return nil, fmt.Errorf("failed to read template %s: %w", name, err)
		}

		// Create template with custom functions
		tmpl, err := template.New(name).Funcs(r.templateFuncs()).Parse(string(content))
		if err != nil {
			return nil, fmt.Errorf("failed to parse template %s: %w", name, err)
		}

		// Store without .tmpl extension
		templateName := name[:len(name)-5] // Remove .tmpl
		r.templates[templateName] = tmpl
	}

	return r, nil
}

// templateFuncs returns the custom template functions
func (r *TemplateRenderer) templateFuncs() template.FuncMap {
	return template.FuncMap{
		"style":    r.applyStyle,
		"color":    r.applyColor,
		"hashtags": r.highlightHashtags,
	}
}

// applyStyle applies a named style to text
func (r *TemplateRenderer) applyStyle(styleName string, text string) string {
	if !r.useColor {
		return text
	}

	var style lipgloss.Style
	switch styleName {
	case "done":
		style = styles.StatusDone
	case "pending":
		style = styles.StatusPending
	case "position":
		style = styles.Position
	case "hashtag":
		style = styles.Hashtag
	case "error":
		style = styles.Error
	case "success":
		style = styles.Success
	case "info":
		style = styles.Info
	default:
		style = styles.Base
	}

	return style.Render(text)
}

// applyColor applies a color to text using pterm
func (r *TemplateRenderer) applyColor(color string, text string) string {
	if !r.useColor {
		return text
	}

	switch color {
	case "green":
		return pterm.FgGreen.Sprint(text)
	case "red":
		return pterm.FgRed.Sprint(text)
	case "yellow":
		return pterm.FgYellow.Sprint(text)
	case "blue":
		return pterm.FgBlue.Sprint(text)
	case "gray":
		return pterm.FgGray.Sprint(text)
	default:
		return text
	}
}

// highlightHashtags highlights hashtags in text
func (r *TemplateRenderer) highlightHashtags(text string) string {
	if !r.useColor {
		return text
	}

	hashtagRegex := regexp.MustCompile(`#[^\s]+`)
	return hashtagRegex.ReplaceAllStringFunc(text, func(match string) string {
		return styles.Hashtag.Render(match)
	})
}

// Render renders a template with the given data
func (r *TemplateRenderer) Render(templateName string, data interface{}) error {
	tmpl, ok := r.templates[templateName]
	if !ok {
		return fmt.Errorf("template '%s' not found", templateName)
	}

	// Prepare data for rendering
	renderData := r.prepareData(data)

	// Execute template
	var buf bytes.Buffer
	err := tmpl.Execute(&buf, renderData)
	if err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Write to output
	_, err = r.writer.Write(buf.Bytes())
	return err
}

// prepareData prepares data for template rendering
func (r *TemplateRenderer) prepareData(data interface{}) interface{} {
	// If it's a Todo, enrich it with rendered elements
	if todo, ok := data.(*models.Todo); ok {
		// Format position with padding
		positionStr := fmt.Sprintf("%6d", todo.Position)

		return map[string]interface{}{
			"Todo":        todo,
			"Position":    positionStr,
			"PositionRaw": todo.Position,
			"Symbol":      r.getStatusSymbol(todo.Status),
			"Text":        r.highlightHashtags(todo.Text),
			"TextRaw":     todo.Text,
			"Status":      todo.Status,
			"IsDone":      todo.Status == models.StatusDone,
			"IsPending":   todo.Status == models.StatusPending,
		}
	}

	return data
}

// getStatusSymbol returns the styled status symbol
func (r *TemplateRenderer) getStatusSymbol(status models.TodoStatus) string {
	if status == models.StatusDone {
		return r.applyStyle("done", "✓")
	}
	return r.applyStyle("pending", "✕")
}
