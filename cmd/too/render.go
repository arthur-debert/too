package main

import (
	"fmt"
	"os"

	"github.com/arthur-debert/too/pkg/too/output"
)

// getRenderer returns a renderer configured with the format flag value
func getRenderer() (*output.Renderer, error) {
	// Validate the format
	if !output.HasFormatter(formatFlag) {
		availableFormats := output.List()
		return nil, fmt.Errorf("invalid format %q. Available formats: %v", formatFlag, availableFormats)
	}

	// Create renderer with specified format
	renderer, err := output.NewRendererWithFormat(formatFlag, os.Stdout)
	if err != nil {
		return nil, fmt.Errorf("failed to create renderer: %w", err)
	}

	return renderer, nil
}
