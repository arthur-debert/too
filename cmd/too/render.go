package main

import (
	"fmt"
	"io"
	"os"

	"github.com/arthur-debert/too/pkg/lipbalm"
	"github.com/arthur-debert/too/pkg/too/output"
)

// renderEngine is a cached lipbalm engine for rendering
var renderEngine *lipbalm.RenderEngine

// getRenderEngine returns a configured lipbalm render engine
func getRenderEngine() (*lipbalm.RenderEngine, error) {
	if renderEngine != nil {
		return renderEngine, nil
	}

	// Create output engine with too-specific functionality
	engine, err := output.NewEngine()
	if err != nil {
		return nil, fmt.Errorf("failed to create render engine: %w", err)
	}

	renderEngine = engine.GetLipbalmEngine()
	return renderEngine, nil
}

// render renders data in the specified format to the writer
func render(writer io.Writer, format string, data interface{}) error {
	// Validate the format
	if !lipbalm.HasFormat(format) {
		availableFormats := lipbalm.ListFormats()
		return fmt.Errorf("invalid format %q. Available formats: %v", format, availableFormats)
	}

	// Get render engine
	engine, err := getRenderEngine()
	if err != nil {
		return err
	}

	// Render the data
	return engine.Render(writer, format, data)
}

// renderToStdout renders data using the global format flag to stdout
func renderToStdout(data interface{}) error {
	return render(os.Stdout, formatFlag, data)
}

// renderError renders an error message
func renderError(writer io.Writer, format string, err error) error {
	// Get render engine
	engine, engineErr := getRenderEngine()
	if engineErr != nil {
		return engineErr
	}

	// Render the error
	return engine.RenderError(writer, format, err)
}
