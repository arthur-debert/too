package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/arthur-debert/tdh/pkg/tdh"
	"github.com/arthur-debert/tdh/pkg/tdh/models"
)

func TestRenderer_RenderList_WithTemplates(t *testing.T) {
	tests := []struct {
		name     string
		result   *tdh.ListResult
		contains []string
	}{
		{
			name: "render list with multiple todos",
			result: &tdh.ListResult{
				Todos: []*models.Todo{
					{
						Position: 1,
						Text:     "First todo #urgent",
						Status:   models.StatusPending,
					},
					{
						Position: 2,
						Text:     "Second todo #done",
						Status:   models.StatusDone,
					},
				},
				TotalCount: 2,
				DoneCount:  1,
			},
			contains: []string{
				"1", "✕", "First todo #urgent",
				"2", "✓", "Second todo #done",
				"2 todo(s), 1 done",
			},
		},
		{
			name: "render empty list",
			result: &tdh.ListResult{
				Todos:      []*models.Todo{},
				TotalCount: 0,
				DoneCount:  0,
			},
			contains: []string{"No todos found"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			renderer := NewRenderer(&buf)

			err := renderer.RenderList(tt.result)
			if err != nil {
				t.Fatalf("RenderList failed: %v", err)
			}

			output := buf.String()
			for _, expected := range tt.contains {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain %q, got:\n%s", expected, output)
				}
			}
		})
	}
}

func TestRenderer_TemplateIntegration(t *testing.T) {
	// Test that template renderer is properly integrated
	var buf bytes.Buffer
	renderer := NewRenderer(&buf)

	// Check that template renderer was created
	if renderer.templateRenderer == nil {
		t.Error("Template renderer should be initialized")
	}

	// Test rendering a list
	result := &tdh.ListResult{
		Todos: []*models.Todo{
			{
				Position: 1,
				Text:     "Test integration #works",
				Status:   models.StatusPending,
			},
		},
		TotalCount: 1,
		DoneCount:  0,
	}

	err := renderer.RenderList(result)
	if err != nil {
		t.Fatalf("RenderList failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Test integration") {
		t.Errorf("Expected rendered output, got: %s", output)
	}
}
