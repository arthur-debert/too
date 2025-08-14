package output

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/arthur-debert/tdh/pkg/tdh/models"
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

	// Parse all templates together so they can reference each other
	tmpl := template.New("").Funcs(r.templateFuncs())

	// Parse all template files
	tmpl, err := tmpl.ParseFS(templateFS, "templates/*.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	// Store individual templates by name for easy access
	for _, t := range tmpl.Templates() {
		name := t.Name()
		if name == "" {
			continue
		}
		// Remove .tmpl extension
		if strings.HasSuffix(name, ".tmpl") {
			baseName := name[:len(name)-5]
			r.templates[baseName] = t
		}
	}

	return r, nil
}

// templateFuncs returns the custom template functions
func (r *TemplateRenderer) templateFuncs() template.FuncMap {
	return template.FuncMap{
		"isDone": func(todo *models.Todo) bool {
			return todo.Status == models.StatusDone
		},
		"isPending": func(todo *models.Todo) bool {
			return todo.Status == models.StatusPending
		},
		"padPosition": func(pos int) string {
			return fmt.Sprintf("%6d", pos)
		},
		// Color functions using pterm
		"red": func(s string) string {
			if r.useColor {
				return pterm.FgRed.Sprint(s)
			}
			return s
		},
		"green": func(s string) string {
			if r.useColor {
				return pterm.FgGreen.Sprint(s)
			}
			return s
		},
		"gray": func(s string) string {
			if r.useColor {
				return pterm.FgGray.Sprint(s)
			}
			return s
		},
		"yellow": func(s string) string {
			if r.useColor {
				return pterm.FgYellow.Sprint(s)
			}
			return s
		},
		"cyan": func(s string) string {
			if r.useColor {
				return pterm.FgCyan.Sprint(s)
			}
			return s
		},
	}
}

// Render renders a template with the given data
func (r *TemplateRenderer) Render(templateName string, data interface{}) error {
	tmpl, ok := r.templates[templateName]
	if !ok {
		return fmt.Errorf("template '%s' not found", templateName)
	}

	// Execute template to get markup
	var buf bytes.Buffer
	err := tmpl.Execute(&buf, data)
	if err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Write output directly - colors are already applied by template functions
	_, err = r.writer.Write(buf.Bytes())
	return err
}

// PrepareData prepares data for template rendering
func (r *TemplateRenderer) PrepareData(data interface{}) interface{} {
	// Pass raw data - let templates handle all formatting
	return data
}
