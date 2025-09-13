package lipbalm

import (
	"embed"
)

// WithTemplates creates a RenderEngine with template support
func WithTemplates(embedFS embed.FS, templateDir string, styles StyleMap) (*RenderEngine, error) {
	tm := NewTemplateManager(styles, nil)
	if err := tm.AddTemplatesFromEmbed(embedFS, templateDir); err != nil {
		return nil, err
	}

	return New(&Config{
		AutoDetectTerminal: true,
		Styles:             styles,
		TemplateManager:    tm,
	}), nil
}

// WithCallbacks creates a RenderEngine with custom callbacks
func WithCallbacks(callbacks RenderCallbacks) *RenderEngine {
	return New(&Config{
		AutoDetectTerminal: true,
		Styles:             DefaultStyles(),
		Callbacks:          callbacks,
	})
}


// Config returns the engine's configuration for modification
func (e *RenderEngine) Config() *Config {
	return e.config
}