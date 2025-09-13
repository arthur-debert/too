package lipbalm

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
)

// MarkdownFormatter handles Markdown output
type MarkdownFormatter struct{}

func (f *MarkdownFormatter) Name() string        { return "markdown" }
func (f *MarkdownFormatter) Description() string { return "Markdown output for documentation and notes" }

func (f *MarkdownFormatter) Render(data interface{}, config *Config) (string, error) {
	// Check if there's a custom markdown renderer in callbacks
	if config.Callbacks.CustomFields != nil {
		if renderer, ok := config.Callbacks.CustomFields["__markdown__"]; ok {
			if result, handled := renderer("markdown", "__markdown__", data); handled {
				return result, nil
			}
		}
	}

	// Default markdown rendering
	return renderAsMarkdown(data, config), nil
}

// renderAsMarkdown converts data to markdown format
func renderAsMarkdown(data interface{}, config *Config) string {
	var buf bytes.Buffer
	
	v := reflect.ValueOf(data)
	if !v.IsValid() {
		return ""
	}

	// Handle pointers
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		renderSliceAsMarkdown(&buf, v, config)
	case reflect.Struct:
		renderStructAsMarkdown(&buf, v, config)
	case reflect.Map:
		renderMapAsMarkdown(&buf, v, config)
	default:
		buf.WriteString(fmt.Sprint(v.Interface()))
	}

	return buf.String()
}

// renderSliceAsMarkdown renders a slice as a markdown list or table
func renderSliceAsMarkdown(buf *bytes.Buffer, v reflect.Value, config *Config) {
	if v.Len() == 0 {
		buf.WriteString("*No items*\n")
		return
	}

	// Check first element to determine structure
	first := v.Index(0)
	if first.Kind() == reflect.Ptr {
		first = first.Elem()
	}

	// If it's a slice of structs, render as table
	if first.Kind() == reflect.Struct {
		renderStructSliceAsTable(buf, v, config)
		return
	}

	// Otherwise render as list
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		buf.WriteString(fmt.Sprintf("- %s\n", formatMarkdownValue(elem.Interface(), config)))
	}
}

// renderStructSliceAsTable renders a slice of structs as a markdown table
func renderStructSliceAsTable(buf *bytes.Buffer, v reflect.Value, config *Config) {
	if v.Len() == 0 {
		return
	}

	// Get type info from first element
	first := v.Index(0)
	if first.Kind() == reflect.Ptr {
		first = first.Elem()
	}
	t := first.Type()

	// Extract headers
	var headers []string
	var fieldIndices []int
	
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		
		// Skip unexported fields
		if field.PkgPath != "" {
			continue
		}

		// Check for json:"-" tag
		if tag := field.Tag.Get("json"); tag == "-" {
			continue
		}

		// Use json tag if available, otherwise use field name
		name := field.Name
		if tag := field.Tag.Get("json"); tag != "" {
			parts := strings.Split(tag, ",")
			if parts[0] != "" && parts[0] != "-" {
				name = parts[0]
			}
		}

		headers = append(headers, name)
		fieldIndices = append(fieldIndices, i)
	}

	// Write headers
	buf.WriteString("| " + strings.Join(headers, " | ") + " |\n")
	buf.WriteString("|" + strings.Repeat(" --- |", len(headers)) + "\n")

	// Write rows
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		if elem.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}

		buf.WriteString("| ")
		for j, idx := range fieldIndices {
			if j > 0 {
				buf.WriteString(" | ")
			}
			value := elem.Field(idx).Interface()
			buf.WriteString(formatMarkdownValue(value, config))
		}
		buf.WriteString(" |\n")
	}
}

// renderStructAsMarkdown renders a struct as markdown
func renderStructAsMarkdown(buf *bytes.Buffer, v reflect.Value, config *Config) {
	t := v.Type()
	
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		
		// Skip unexported fields
		if field.PkgPath != "" {
			continue
		}

		// Check for json:"-" tag
		if tag := field.Tag.Get("json"); tag == "-" {
			continue
		}

		// Use json tag if available, otherwise use field name
		name := field.Name
		if tag := field.Tag.Get("json"); tag != "" {
			parts := strings.Split(tag, ",")
			if parts[0] != "" && parts[0] != "-" {
				name = parts[0]
			}
		}

		value := v.Field(i).Interface()
		formattedValue := formatMarkdownValue(value, config)
		
		// Bold field names
		buf.WriteString(fmt.Sprintf("**%s:** %s\n", name, formattedValue))
	}
}

// renderMapAsMarkdown renders a map as markdown
func renderMapAsMarkdown(buf *bytes.Buffer, v reflect.Value, config *Config) {
	keys := v.MapKeys()
	if len(keys) == 0 {
		buf.WriteString("*Empty map*\n")
		return
	}

	// Create a table for maps
	buf.WriteString("| Key | Value |\n")
	buf.WriteString("| --- | --- |\n")
	
	for _, key := range keys {
		keyStr := formatMarkdownValue(key.Interface(), config)
		valueStr := formatMarkdownValue(v.MapIndex(key).Interface(), config)
		buf.WriteString(fmt.Sprintf("| %s | %s |\n", keyStr, valueStr))
	}
}

// formatMarkdownValue formats a value for markdown, checking for custom renderers
func formatMarkdownValue(v interface{}, config *Config) string {
	if v == nil {
		return "*nil*"
	}

	// Check for custom field renderer
	if config != nil && config.Callbacks.CustomFields != nil {
		// Get the type name for custom rendering
		typeName := reflect.TypeOf(v).Name()
		if renderer, ok := config.Callbacks.CustomFields[typeName]; ok {
			if result, handled := renderer("markdown", typeName, v); handled {
				return result
			}
		}
	}

	// Default formatting
	switch val := v.(type) {
	case string:
		// Escape markdown special characters
		escaped := strings.ReplaceAll(val, "|", "\\|")
		escaped = strings.ReplaceAll(escaped, "\n", " ")
		return escaped
	case fmt.Stringer:
		return val.String()
	default:
		return fmt.Sprint(v)
	}
}