package lipbalm

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"sync"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"golang.org/x/term"
)

// RenderEngine is the main API for formatting output
type RenderEngine struct {
	registry      *FormatRegistry
	config        *Config
	defaultFormat string
	mu            sync.RWMutex
}

// Config holds engine configuration
type Config struct {
	// Terminal detection settings
	AutoDetectTerminal bool
	ForceColor         bool

	// Default styles for terminal output
	Styles StyleMap

	// Template configuration
	TemplateManager *TemplateManager

	// Custom render callbacks
	Callbacks RenderCallbacks

	// Default output writer
	DefaultWriter io.Writer
}

// RenderCallbacks allows domain-specific customization
type RenderCallbacks struct {
	// Called before rendering any object to transform/enrich data
	PreProcess func(format string, data interface{}) interface{}

	// Custom field renderers for specific formats
	CustomFields map[string]FieldRenderer

	// Post-process the rendered output
	PostProcess func(format string, output string) string
}

// FieldRenderer customizes how a specific field is rendered
// Returns (rendered string, handled bool). If handled is false, default rendering is used.
type FieldRenderer func(format string, fieldName string, value interface{}) (string, bool)

// New creates a new RenderEngine with the given configuration
func New(config *Config) *RenderEngine {
	if config.DefaultWriter == nil {
		config.DefaultWriter = os.Stdout
	}

	engine := &RenderEngine{
		registry: newDefaultRegistry(),
		config:   config,
	}

	if config.AutoDetectTerminal {
		engine.defaultFormat = engine.detectDefaultFormat()
	} else {
		engine.defaultFormat = "term"
	}

	// Set up lipgloss renderer based on config
	if config.ForceColor || (config.AutoDetectTerminal && isTerminal(config.DefaultWriter)) {
		renderer := lipgloss.NewRenderer(config.DefaultWriter)
		renderer.SetColorProfile(termenv.TrueColor)
		SetDefaultRenderer(renderer)
	} else {
		renderer := lipgloss.NewRenderer(config.DefaultWriter)
		renderer.SetColorProfile(termenv.Ascii)
		SetDefaultRenderer(renderer)
	}

	return engine
}

// Render renders data in the specified format to the writer
func (e *RenderEngine) Render(w io.Writer, format string, data interface{}) error {
	if format == "" {
		format = e.defaultFormat
	}

	formatter, err := e.registry.Get(format)
	if err != nil {
		return err
	}

	// Apply pre-processing callback if defined
	processedData := data
	if e.config.Callbacks.PreProcess != nil {
		processedData = e.config.Callbacks.PreProcess(format, data)
	}

	// Render using the formatter
	output, err := formatter.Render(processedData, e.config)
	if err != nil {
		return err
	}

	// Apply post-processing callback if defined
	if e.config.Callbacks.PostProcess != nil {
		output = e.config.Callbacks.PostProcess(format, output)
	}

	_, err = w.Write([]byte(output))
	return err
}

// RenderError renders an error in the specified format
func (e *RenderEngine) RenderError(w io.Writer, format string, err error) error {
	if format == "" {
		format = e.defaultFormat
	}

	// Use a simple map for error representation
	errorData := map[string]string{"error": err.Error()}
	
	// For terminal format, use template if available
	if format == "term" && e.config.TemplateManager != nil {
		tmpl, ok := e.config.TemplateManager.GetTemplate("error")
		if ok {
			output, renderErr := Render(tmpl, errorData, e.config.Styles)
			if renderErr == nil {
				_, writeErr := w.Write([]byte(output + "\n"))
				return writeErr
			}
		}
	}

	return e.Render(w, format, errorData)
}

// GetFormat returns the current default format
func (e *RenderEngine) GetFormat() string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.defaultFormat
}

// SetFormat sets the default format
func (e *RenderEngine) SetFormat(format string) error {
	if !e.registry.Has(format) {
		return fmt.Errorf("unknown format: %s", format)
	}
	
	e.mu.Lock()
	defer e.mu.Unlock()
	e.defaultFormat = format
	return nil
}

// ListFormats returns all available format names
func (e *RenderEngine) ListFormats() []string {
	return e.registry.List()
}

// GetRegistry returns the format registry for advanced use cases
func (e *RenderEngine) GetRegistry() *FormatRegistry {
	return e.registry
}

// detectDefaultFormat detects the appropriate default format
func (e *RenderEngine) detectDefaultFormat() string {
	if !isTerminal(e.config.DefaultWriter) || isPiped() {
		return "plain"
	}
	return "term"
}

// isTerminal checks if the writer is a terminal
func isTerminal(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		return term.IsTerminal(int(f.Fd()))
	}
	return false
}

// isPiped checks if output is being piped
func isPiped() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) == 0
}

// Quick creates a RenderEngine with sensible defaults
func Quick() *RenderEngine {
	return New(&Config{
		AutoDetectTerminal: true,
		Styles:             DefaultStyles(),
		DefaultWriter:      os.Stdout,
	})
}

// DefaultStyles returns a basic set of styles
func DefaultStyles() StyleMap {
	return StyleMap{
		"title":   lipgloss.NewStyle().Bold(true),
		"error":   lipgloss.NewStyle().Foreground(lipgloss.Color("196")),
		"warning": lipgloss.NewStyle().Foreground(lipgloss.Color("214")),
		"success": lipgloss.NewStyle().Foreground(lipgloss.Color("46")),
		"info":    lipgloss.NewStyle().Foreground(lipgloss.Color("33")),
		"muted":   lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
	}
}

// FormatRegistry manages available formats
type FormatRegistry struct {
	mu      sync.RWMutex
	formats map[string]Formatter
}

// Formatter interface for all output formats
type Formatter interface {
	Name() string
	Description() string
	Render(data interface{}, config *Config) (string, error)
}

// newDefaultRegistry creates a registry with built-in formatters
func newDefaultRegistry() *FormatRegistry {
	r := &FormatRegistry{
		formats: make(map[string]Formatter),
	}

	// Register built-in formatters
	r.Register(&JSONFormatter{})
	r.Register(&YAMLFormatter{})
	r.Register(&CSVFormatter{})
	r.Register(&MarkdownFormatter{})
	r.Register(&PlainFormatter{})
	r.Register(&TerminalFormatter{})

	return r
}

// Register adds a formatter
func (r *FormatRegistry) Register(f Formatter) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.formats[f.Name()] = f
}

// Get retrieves a formatter by name
func (r *FormatRegistry) Get(name string) (Formatter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	f, ok := r.formats[name]
	if !ok {
		return nil, fmt.Errorf("unknown format: %s", name)
	}
	return f, nil
}

// Has checks if a format is registered
func (r *FormatRegistry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.formats[name]
	return ok
}

// List returns all available format names
func (r *FormatRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	names := make([]string, 0, len(r.formats))
	for name := range r.formats {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// JSONFormatter handles JSON output
type JSONFormatter struct{}

func (f *JSONFormatter) Name() string        { return "json" }
func (f *JSONFormatter) Description() string { return "JSON output for programmatic consumption" }

func (f *JSONFormatter) Render(data interface{}, config *Config) (string, error) {
	// Apply custom field renderers if any
	if config.Callbacks.CustomFields != nil {
		data = applyCustomFields(data, "json", config.Callbacks.CustomFields)
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// PlainFormatter handles plain text output (no formatting)
type PlainFormatter struct{}

func (f *PlainFormatter) Name() string        { return "plain" }
func (f *PlainFormatter) Description() string { return "Plain text output without formatting" }

func (f *PlainFormatter) Render(data interface{}, config *Config) (string, error) {
	// For now, just use JSON without indentation
	// TODO: Implement better plain text formatting
	b, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// applyCustomFields applies custom field renderers to data
func applyCustomFields(data interface{}, format string, customFields map[string]FieldRenderer) interface{} {
	if len(customFields) == 0 {
		return data
	}

	// Check if there's a renderer for the entire data type
	v := reflect.ValueOf(data)
	if !v.IsValid() {
		return data
	}

	typeName := v.Type().String()
	if renderer, ok := customFields[typeName]; ok {
		if result, handled := renderer(format, typeName, data); handled {
			return result
		}
	}

	// TODO: Implement full reflection-based field walking
	// For now, return data as-is
	return data
}