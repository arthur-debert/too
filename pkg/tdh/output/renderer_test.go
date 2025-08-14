package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/arthur-debert/tdh/pkg/tdh/models"
)

func TestTemplateRenderer_RenderTodoItem(t *testing.T) {
	tests := []struct {
		name     string
		todo     *models.Todo
		useColor bool
		contains []string
	}{
		{
			name: "render pending todo without color",
			todo: &models.Todo{
				Position: 1,
				Text:     "Test todo with #hashtag",
				Status:   models.StatusPending,
			},
			useColor: false,
			contains: []string{"1", "|", "✕", "Test todo with #hashtag"},
		},
		{
			name: "render done todo without color",
			todo: &models.Todo{
				Position: 42,
				Text:     "Completed task",
				Status:   models.StatusDone,
			},
			useColor: false,
			contains: []string{"42", "|", "✓", "Completed task"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			renderer, err := NewTemplateRenderer(&buf, tt.useColor)
			if err != nil {
				t.Fatalf("Failed to create renderer: %v", err)
			}

			err = renderer.Render("todo_item", tt.todo)
			if err != nil {
				t.Fatalf("Failed to render: %v", err)
			}

			output := buf.String()
			for _, expected := range tt.contains {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain %q, got: %s", expected, output)
				}
			}
		})
	}
}

func TestTemplateRenderer_Integration(t *testing.T) {
	// Create a test todo
	todo := &models.Todo{
		Position: 1,
		Text:     "Integration test #urgent #important",
		Status:   models.StatusPending,
	}

	// Test without colors
	var buf bytes.Buffer
	renderer, err := NewTemplateRenderer(&buf, false)
	if err != nil {
		t.Fatalf("Failed to create renderer: %v", err)
	}

	err = renderer.Render("todo_item", todo)
	if err != nil {
		t.Fatalf("Failed to render: %v", err)
	}

	output := buf.String()
	t.Logf("Rendered output: %s", output)

	// Basic validation
	if !strings.Contains(output, "1") {
		t.Error("Output should contain position")
	}
	if !strings.Contains(output, "✕") {
		t.Error("Output should contain pending symbol")
	}
	if !strings.Contains(output, "#urgent") {
		t.Error("Output should contain hashtags")
	}
}
