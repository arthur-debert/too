package lipbalm

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/beevik/etree"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// StyleMap defines a map of tag names to lipgloss styles.
type StyleMap map[string]lipgloss.Style

// TemplateManager manages template loading and caching
type TemplateManager struct {
	templates map[string]string
	styles    StyleMap
	funcs     template.FuncMap
}

// NewTemplateManager creates a new template manager
func NewTemplateManager(styles StyleMap, funcs template.FuncMap) *TemplateManager {
	return &TemplateManager{
		templates: make(map[string]string),
		styles:    styles,
		funcs:     funcs,
	}
}

// AddTemplatesFromEmbed loads templates from an embedded filesystem
func (tm *TemplateManager) AddTemplatesFromEmbed(embedFS embed.FS, dir string) error {
	return tm.addTemplatesFromFS(embedFS, dir)
}

// AddTemplatesFromDir loads templates from a filesystem directory
func (tm *TemplateManager) AddTemplatesFromDir(dir string) error {
	return tm.addTemplatesFromFS(os.DirFS(dir), ".")
}

// addTemplatesFromFS loads templates from any filesystem
func (tm *TemplateManager) addTemplatesFromFS(fsys fs.FS, dir string) error {
	return fs.WalkDir(fsys, dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".tmpl") {
			return nil
		}

		content, err := fs.ReadFile(fsys, path)
		if err != nil {
			return fmt.Errorf("failed to read template %s: %w", path, err)
		}

		// Use filename without extension as template name
		name := strings.TrimSuffix(filepath.Base(path), ".tmpl")
		tm.templates[name] = string(content)
		return nil
	})
}

// AddTemplate adds a template by name and content
func (tm *TemplateManager) AddTemplate(name, content string) {
	tm.templates[name] = content
}

// RenderTemplate renders a template by name
func (tm *TemplateManager) RenderTemplate(name string, data interface{}) (string, error) {
	templateString, ok := tm.templates[name]
	if !ok {
		return "", fmt.Errorf("template '%s' not found", name)
	}

	return Render(templateString, data, tm.styles, tm.funcs)
}

// ListTemplates returns all loaded template names
func (tm *TemplateManager) ListTemplates() []string {
	names := make([]string, 0, len(tm.templates))
	for name := range tm.templates {
		names = append(names, name)
	}
	return names
}

// GetTemplate returns a template by name and whether it exists
func (tm *TemplateManager) GetTemplate(name string) (string, bool) {
	template, ok := tm.templates[name]
	return template, ok
}

// GetStyles returns the style map
func (tm *TemplateManager) GetStyles() StyleMap {
	return tm.styles
}

var defaultRenderer = lipgloss.NewRenderer(os.Stdout)

// SetDefaultRenderer sets the default renderer for the package.
// This is useful for testing.
func SetDefaultRenderer(renderer *lipgloss.Renderer) {
	defaultRenderer = renderer
}

// Render processes a template string, first with Go's text/template engine,
// and then applies lipgloss styles based on the XML-like tags.
// Optional template functions can be provided as the last parameter.
// Automatically includes Sprig functions for enhanced template capabilities.
func Render(templateString string, data interface{}, styles StyleMap, funcs ...template.FuncMap) (string, error) {
	// Phase 1: Go template expansion
	tmpl := template.New("lipbalm")
	
	// Add Sprig functions first
	tmpl = tmpl.Funcs(sprig.FuncMap())
	
	// Add custom functions if provided (these override Sprig functions with same names)
	if len(funcs) > 0 {
		tmpl = tmpl.Funcs(funcs[0])
	}
	
	tmpl, err := tmpl.Parse(templateString)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	expandedTemplate := buf.String()

	// Phase 2: Lipbaml expansion
	return ExpandTags(expandedTemplate, styles)
}

// ExpandTags applies lipgloss styles based on the XML-like tags.
func ExpandTags(s string, styles StyleMap) (string, error) {
	hasColor := defaultRenderer.ColorProfile() != termenv.Ascii

	doc := etree.NewDocument()
	// We need a single root element for the XML parser.
	if err := doc.ReadFromString("<root>" + s + "</root>"); err != nil {
		// If parsing fails, it might be because the template is not valid XML.
		// In this case, we'll return the raw expanded template.
		// This is a fallback for templates that don't use lipbalm tags.
		return s, nil
	}

	var result bytes.Buffer
	root := doc.SelectElement("root")
	for _, token := range root.Child {
		processToken(token, &result, styles, hasColor)
	}

	return result.String(), nil
}

func processToken(token etree.Token, w io.Writer, styles StyleMap, hasColor bool) {
	switch t := token.(type) {
	case *etree.Element:
		if t.Tag == "no-format" {
			if !hasColor {
				for _, child := range t.Child {
					processToken(child, w, styles, hasColor)
				}
			}
			return
		}

		style, styleExists := styles[t.Tag]

		if hasColor && styleExists {
			var innerContent bytes.Buffer
			for _, child := range t.Child {
				processToken(child, &innerContent, styles, hasColor)
			}
			_, _ = w.Write([]byte(style.Render(innerContent.String())))
		} else {
			for _, child := range t.Child {
				processToken(child, w, styles, hasColor)
			}
		}
	case *etree.CharData:
		_, _ = w.Write([]byte(t.Data))
	}
}
