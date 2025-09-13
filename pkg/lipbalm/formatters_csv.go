package lipbalm

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// CSVFormatter handles CSV output
type CSVFormatter struct{}

func (f *CSVFormatter) Name() string        { return "csv" }
func (f *CSVFormatter) Description() string { return "CSV output for spreadsheet applications" }

func (f *CSVFormatter) Render(data interface{}, config *Config) (string, error) {
	// Apply custom field renderers if any
	if config.Callbacks.CustomFields != nil {
		data = applyCustomFields(data, "csv", config.Callbacks.CustomFields)
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	
	// Convert data to CSV rows
	headers, rows, err := dataToCSV(data)
	if err != nil {
		return "", err
	}

	// Write headers
	if len(headers) > 0 {
		if err := writer.Write(headers); err != nil {
			return "", err
		}
	}

	// Write rows
	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return "", err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// dataToCSV converts various data types to CSV format
func dataToCSV(data interface{}) (headers []string, rows [][]string, err error) {
	v := reflect.ValueOf(data)
	if !v.IsValid() {
		return nil, nil, fmt.Errorf("invalid data")
	}

	// Handle pointers
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		return sliceToCSV(v)
	case reflect.Struct:
		return structToCSV(v)
	case reflect.Map:
		return mapToCSV(v)
	default:
		// For simple types, create a single cell
		return nil, [][]string{{fmt.Sprint(v.Interface())}}, nil
	}
}

// sliceToCSV converts a slice to CSV rows
func sliceToCSV(v reflect.Value) ([]string, [][]string, error) {
	if v.Len() == 0 {
		return nil, nil, nil
	}

	// Check first element to determine structure
	first := v.Index(0)
	if first.Kind() == reflect.Ptr {
		first = first.Elem()
	}

	// If it's a slice of structs, extract headers from struct fields
	if first.Kind() == reflect.Struct {
		headers := extractHeaders(first.Type())
		rows := make([][]string, 0, v.Len())

		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)
			if elem.Kind() == reflect.Ptr {
				elem = elem.Elem()
			}
			row := extractValues(elem)
			rows = append(rows, row)
		}

		return headers, rows, nil
	}

	// For slice of simple types, no headers
	rows := make([][]string, 0, v.Len())
	for i := 0; i < v.Len(); i++ {
		rows = append(rows, []string{formatValue(v.Index(i).Interface())})
	}

	return nil, rows, nil
}

// structToCSV converts a struct to CSV format
func structToCSV(v reflect.Value) ([]string, [][]string, error) {
	headers := extractHeaders(v.Type())
	row := extractValues(v)
	return headers, [][]string{row}, nil
}

// mapToCSV converts a map to CSV format
func mapToCSV(v reflect.Value) ([]string, [][]string, error) {
	headers := []string{"key", "value"}
	rows := make([][]string, 0, v.Len())

	for _, key := range v.MapKeys() {
		row := []string{
			formatValue(key.Interface()),
			formatValue(v.MapIndex(key).Interface()),
		}
		rows = append(rows, row)
	}

	return headers, rows, nil
}

// extractHeaders extracts CSV headers from struct type
func extractHeaders(t reflect.Type) []string {
	var headers []string

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		
		// Skip unexported fields
		if field.PkgPath != "" {
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
	}

	return headers
}

// extractValues extracts values from a struct
func extractValues(v reflect.Value) []string {
	var values []string

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		
		// Skip unexported fields
		if field.PkgPath != "" {
			continue
		}

		// Skip fields with json:"-"
		if tag := field.Tag.Get("json"); tag == "-" {
			continue
		}

		value := v.Field(i).Interface()
		values = append(values, formatValue(value))
	}

	return values
}

// formatValue formats a value for CSV output
func formatValue(v interface{}) string {
	if v == nil {
		return ""
	}

	// Handle special types
	switch val := v.(type) {
	case time.Time:
		return val.Format(time.RFC3339)
	case fmt.Stringer:
		return val.String()
	default:
		return fmt.Sprint(v)
	}
}