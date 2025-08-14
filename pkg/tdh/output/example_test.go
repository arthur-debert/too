package output_test

import (
	"fmt"
	"os"

	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/arthur-debert/tdh/pkg/tdh/output"
)

func ExampleTemplateRenderer() {
	// Create a renderer with color support
	renderer, err := output.NewTemplateRenderer(os.Stdout, true)
	if err != nil {
		fmt.Printf("Error creating renderer: %v\n", err)
		return
	}

	// Create a sample todo
	todo := &models.Todo{
		Position: 1,
		Text:     "Implement template-based rendering #milestone1",
		Status:   models.StatusPending,
	}

	// Render the todo using the template
	err = renderer.Render("todo_item", todo)
	if err != nil {
		fmt.Printf("Error rendering: %v\n", err)
		return
	}
}
