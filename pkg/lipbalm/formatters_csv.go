package lipbalm

import (
	"bytes"
	"reflect"
	"github.com/gocarina/gocsv"
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
	
	// gocsv requires a slice, not a single struct
	// Check if data is already a slice/array
	rv := reflect.ValueOf(data)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		// Wrap single value in a slice
		sliceType := reflect.SliceOf(rv.Type())
		slice := reflect.MakeSlice(sliceType, 1, 1)
		slice.Index(0).Set(rv)
		data = slice.Interface()
	}
	
	// Use gocsv to handle the marshaling
	// Note: gocsv uses csv tags by default, not json tags
	// It automatically handles pointer dereferencing
	err := gocsv.Marshal(data, &buf)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}