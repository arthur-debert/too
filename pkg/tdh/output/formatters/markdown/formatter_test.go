package markdown_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/arthur-debert/tdh/pkg/tdh"
	"github.com/arthur-debert/tdh/pkg/tdh/commands/formats"
	"github.com/arthur-debert/tdh/pkg/tdh/models"
	markdownformatter "github.com/arthur-debert/tdh/pkg/tdh/output/formatters/markdown"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarkdownFormatter(t *testing.T) {
	formatter := markdownformatter.New()

	t.Run("Name and Description", func(t *testing.T) {
		assert.Equal(t, "markdown", formatter.Name())
		assert.Equal(t, "Markdown output for documentation and notes", formatter.Description())
	})

	t.Run("RenderAdd", func(t *testing.T) {
		var buf bytes.Buffer
		result := &tdh.AddResult{
			Todo: &models.Todo{
				Position: 1,
				Text:     "Test todo",
				Status:   models.StatusPending,
			},
		}

		err := formatter.RenderAdd(&buf, result)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "Added todo #1: [ ] Test todo")
	})

	t.Run("RenderList with nested todos", func(t *testing.T) {
		var buf bytes.Buffer
		result := &tdh.ListResult{
			Todos: []*models.Todo{
				{
					Position: 1,
					Text:     "Parent todo",
					Status:   models.StatusPending,
					Items: []*models.Todo{
						{
							Position: 1,
							Text:     "Child todo 1",
							Status:   models.StatusDone,
						},
						{
							Position: 2,
							Text:     "Child todo 2",
							Status:   models.StatusPending,
							Items: []*models.Todo{
								{
									Position: 1,
									Text:     "Grandchild todo",
									Status:   models.StatusPending,
								},
							},
						},
					},
				},
				{
					Position: 2,
					Text:     "Second parent",
					Status:   models.StatusDone,
				},
			},
			TotalCount: 5,
			DoneCount:  2,
		}

		err := formatter.RenderList(&buf, result)
		require.NoError(t, err)

		output := buf.String()
		lines := strings.Split(output, "\n")

		// Check structure
		assert.Contains(t, lines[0], "1. [ ] Parent todo")
		assert.Contains(t, lines[1], "   1. [x] Child todo 1")
		assert.Contains(t, lines[2], "   2. [ ] Child todo 2")
		assert.Contains(t, lines[3], "      1. [ ] Grandchild todo")
		assert.Contains(t, lines[4], "2. [x] Second parent")

		// Check summary
		assert.Contains(t, output, "5 todo(s), 2 done")
	})

	t.Run("RenderList empty", func(t *testing.T) {
		var buf bytes.Buffer
		result := &tdh.ListResult{
			Todos:      []*models.Todo{},
			TotalCount: 0,
			DoneCount:  0,
		}

		err := formatter.RenderList(&buf, result)
		require.NoError(t, err)
		assert.Equal(t, "No todos\n", buf.String())
	})

	t.Run("RenderComplete", func(t *testing.T) {
		var buf bytes.Buffer
		results := []*tdh.CompleteResult{
			{
				Todo: &models.Todo{
					Position: 1,
					Text:     "Completed todo",
					Status:   models.StatusDone,
				},
			},
		}

		err := formatter.RenderComplete(&buf, results)
		require.NoError(t, err)
		assert.Contains(t, buf.String(), "Completed todo #1: [x] Completed todo")
	})

	t.Run("RenderSearch", func(t *testing.T) {
		var buf bytes.Buffer
		result := &tdh.SearchResult{
			MatchedTodos: []*models.Todo{
				{
					Position: 3,
					Text:     "Matching todo",
					Status:   models.StatusPending,
				},
			},
		}

		err := formatter.RenderSearch(&buf, result)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "Found 1 matching todo(s):")
		assert.Contains(t, output, "1. [ ] Matching todo")
	})

	t.Run("RenderError", func(t *testing.T) {
		var buf bytes.Buffer
		testErr := assert.AnError

		err := formatter.RenderError(&buf, testErr)
		require.NoError(t, err)
		assert.Equal(t, "**Error:** assert.AnError general error for testing\n", buf.String())
	})

	t.Run("RenderFormats", func(t *testing.T) {
		var buf bytes.Buffer
		result := &tdh.ListFormatsResult{
			Formats: []formats.Format{
				{
					Name:        "json",
					Description: "JSON output",
				},
				{
					Name:        "markdown",
					Description: "Markdown output",
				},
			},
		}

		err := formatter.RenderFormats(&buf, result)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "Available output formats:")
		assert.Contains(t, output, "- **json**: JSON output")
		assert.Contains(t, output, "- **markdown**: Markdown output")
	})

	t.Run("RenderInit", func(t *testing.T) {
		var buf bytes.Buffer
		result := &tdh.InitResult{
			DBPath: "/home/user/.todos.json",
		}

		err := formatter.RenderInit(&buf, result)
		require.NoError(t, err)
		assert.Equal(t, "Initialized todo collection at: `/home/user/.todos.json`\n", buf.String())
	})

	t.Run("RenderDataPath", func(t *testing.T) {
		var buf bytes.Buffer
		result := &tdh.ShowDataPathResult{
			Path: "/home/user/.todos.json",
		}

		err := formatter.RenderDataPath(&buf, result)
		require.NoError(t, err)
		assert.Equal(t, "Data file path: `/home/user/.todos.json`\n", buf.String())
	})

	t.Run("RenderList with multiline todos", func(t *testing.T) {
		var buf bytes.Buffer
		result := &tdh.ListResult{
			Todos: []*models.Todo{
				{
					Position: 1,
					Text:     "Single line todo",
					Status:   models.StatusPending,
				},
				{
					Position: 2,
					Text:     "Multi-line todo\nSecond line\nThird line",
					Status:   models.StatusPending,
				},
				{
					Position: 3,
					Text:     "Nested parent",
					Status:   models.StatusPending,
					Items: []*models.Todo{
						{
							Position: 1,
							Text:     "Nested child with\nmultiple lines",
							Status:   models.StatusPending,
						},
					},
				},
			},
			TotalCount: 4,
			DoneCount:  0,
		}

		err := formatter.RenderList(&buf, result)
		require.NoError(t, err)

		output := buf.String()
		lines := strings.Split(output, "\n")

		// Check single line todo
		assert.Contains(t, lines[0], "1. [ ] Single line todo")

		// Check multi-line todo with proper indentation
		assert.Contains(t, lines[1], "2. [ ] Multi-line todo")
		assert.Equal(t, "      Second line", lines[2])
		assert.Equal(t, "      Third line", lines[3])

		// Check nested todos with multi-line content
		assert.Contains(t, lines[4], "3. [ ] Nested parent")
		assert.Contains(t, lines[5], "   1. [ ] Nested child with")
		assert.Equal(t, "         multiple lines", lines[6])
	})

	t.Run("RenderAdd with multiline todo", func(t *testing.T) {
		var buf bytes.Buffer
		result := &tdh.AddResult{
			Todo: &models.Todo{
				Position: 1,
				Text:     "New todo with\nmultiple lines",
				Status:   models.StatusPending,
			},
		}

		err := formatter.RenderAdd(&buf, result)
		require.NoError(t, err)

		output := buf.String()
		// For single-line output commands, we expect the newlines to be preserved
		// but the helper function should format them correctly
		assert.Contains(t, output, "Added todo #1: [ ] New todo with\n      multiple lines")
	})
}

func TestMarkdownFormatterBehavior(t *testing.T) {
	t.Run("Markdown formatter has correct metadata", func(t *testing.T) {
		formatter := markdownformatter.New()
		assert.Equal(t, "markdown", formatter.Name())
		assert.Equal(t, "Markdown output for documentation and notes", formatter.Description())
	})

	t.Run("Markdown formatter renders list correctly", func(t *testing.T) {
		buf := &bytes.Buffer{}
		formatter := markdownformatter.New()

		// Test rendering
		result := &tdh.ListResult{
			Todos: []*models.Todo{
				{
					Position: 1,
					Text:     "Test todo",
					Status:   models.StatusPending,
				},
			},
			TotalCount: 1,
			DoneCount:  0,
		}

		err := formatter.RenderList(buf, result)
		require.NoError(t, err)
		assert.Contains(t, buf.String(), "1. [ ] Test todo")
	})
}
