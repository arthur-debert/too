package lipbalm

import (
	"fmt"
	"reflect"
)

// TerminalFormatter handles rich terminal output using templates
type TerminalFormatter struct{}

func (f *TerminalFormatter) Name() string        { return "term" }
func (f *TerminalFormatter) Description() string { return "Rich terminal output with colors and formatting" }

func (f *TerminalFormatter) Render(data interface{}, config *Config) (string, error) {
	if config.TemplateManager == nil {
		// Fallback to simple rendering if no templates
		return fmt.Sprintf("%+v", data), nil
	}

	// Determine template name based on data type
	templateName := getTemplateNameForData(data)
	
	// Check if template exists
	tmpl, ok := config.TemplateManager.GetTemplate(templateName)
	if !ok {
		// Try a generic template
		tmpl, ok = config.TemplateManager.GetTemplate("default")
		if !ok {
			// Fallback to simple rendering
			return fmt.Sprintf("%+v", data), nil
		}
	}

	// Render using the template
	output, err := Render(tmpl, data, config.Styles, config.TemplateManager.GetFuncs())
	if err != nil {
		return "", fmt.Errorf("template rendering failed: %w", err)
	}

	return output, nil
}

// getTemplateNameForData attempts to determine a template name from the data type
func getTemplateNameForData(data interface{}) string {
	t := reflect.TypeOf(data)
	if t == nil {
		return "default"
	}

	// Handle pointers
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Use the type name as template name (lowercase)
	name := t.Name()
	if name == "" {
		return "default"
	}

	// Convert CamelCase to snake_case for template names
	return camelToSnake(name)
}

// camelToSnake converts CamelCase to snake_case
func camelToSnake(s string) string {
	var result []byte
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		if r >= 'A' && r <= 'Z' {
			result = append(result, byte(r-'A'+'a'))
		} else {
			result = append(result, byte(r))
		}
	}
	return string(result)
}