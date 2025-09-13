package lipbalm

import (
	"bytes"

	"gopkg.in/yaml.v3"
)

// YAMLFormatter handles YAML output
type YAMLFormatter struct{}

func (f *YAMLFormatter) Name() string        { return "yaml" }
func (f *YAMLFormatter) Description() string { return "YAML output for programmatic consumption" }

func (f *YAMLFormatter) Render(data interface{}, config *Config) (string, error) {
	// Apply custom field renderers if any
	if config.Callbacks.CustomFields != nil {
		data = applyCustomFields(data, "yaml", config.Callbacks.CustomFields)
	}

	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(data); err != nil {
		return "", err
	}
	return buf.String(), nil
}