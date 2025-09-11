package markdown_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/arthur-debert/too/pkg/too"
	"github.com/arthur-debert/too/pkg/too/commands/formats"
	"github.com/arthur-debert/too/pkg/too/models"
	markdownformatter "github.com/arthur-debert/too/pkg/too/output/formatters/markdown"
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
		result := &too.AddResult{
			Todo: &models.Todo{
				Text:     "Test todo",
				Statuses: map[string]string{"completion": string(models.StatusPending)},
				Items:    []*models.Todo{},
			},
		}

		err := formatter.RenderAdd(&buf, result)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "Added todo: [ ] Test todo")
	})

	t.Run("RenderList with nested todos", func(t *testing.T) {
		var buf bytes.Buffer
		result := &too.ListResult{
			Todos: []*models.Todo{
				{
					Text:     "Parent todo",
					Statuses: map[string]string{"completion": string(models.StatusPending)},
					Items: []*models.Todo{
						{
							Text:     "Child todo 1",
							Statuses: map[string]string{"completion": string(models.StatusDone)},
							Items:    []*models.Todo{},
						},
						{
							Text:     "Child todo 2",
							Statuses: map[string]string{"completion": string(models.StatusPending)},
							Items: []*models.Todo{
								{
									Text:     "Grandchild todo",
									Statuses: map[string]string{"completion": string(models.StatusPending)},
									Items:    []*models.Todo{},
								},
							},
						},
					},
				},
				{
					Text:     "Second parent",
					Statuses: map[string]string{"completion": string(models.StatusDone)},
					Items:    []*models.Todo{},
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
		result := &too.ListResult{
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
		results := []*too.CompleteResult{
			{
				Todo: &models.Todo{
					Text:     "Completed todo",
					Statuses: map[string]string{"completion": string(models.StatusDone)},
					Items:    []*models.Todo{},
				},
			},
		}

		err := formatter.RenderComplete(&buf, results)
		require.NoError(t, err)
		assert.Contains(t, buf.String(), "Completed todo: [x] Completed todo")
	})

	t.Run("RenderSearch", func(t *testing.T) {
		var buf bytes.Buffer
		result := &too.SearchResult{
			MatchedTodos: []*models.Todo{
				{
					Text:     "Matching todo",
					Statuses: map[string]string{"completion": string(models.StatusPending)},
					Items:    []*models.Todo{},
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
		result := &too.ListFormatsResult{
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
		result := &too.InitResult{
			DBPath: "/home/user/.todos.json",
		}

		err := formatter.RenderInit(&buf, result)
		require.NoError(t, err)
		assert.Equal(t, "Initialized todo collection at: `/home/user/.todos.json`\n", buf.String())
	})

	t.Run("RenderDataPath", func(t *testing.T) {
		var buf bytes.Buffer
		result := &too.ShowDataPathResult{
			Path: "/home/user/.todos.json",
		}

		err := formatter.RenderDataPath(&buf, result)
		require.NoError(t, err)
		assert.Equal(t, "Data file path: `/home/user/.todos.json`\n", buf.String())
	})

	t.Run("RenderList with multiline todos", func(t *testing.T) {
		var buf bytes.Buffer
		result := &too.ListResult{
			Todos: []*models.Todo{
				{
					Text:     "Single line todo",
					Statuses: map[string]string{"completion": string(models.StatusPending)},
					Items:    []*models.Todo{},
				},
				{
					Text:     "Multi-line todo\nSecond line\nThird line",
					Statuses: map[string]string{"completion": string(models.StatusPending)},
					Items:    []*models.Todo{},
				},
				{
					Text:     "Nested parent",
					Statuses: map[string]string{"completion": string(models.StatusPending)},
					Items: []*models.Todo{
						{
							Text:     "Nested child with\nmultiple lines",
							Statuses: map[string]string{"completion": string(models.StatusPending)},
							Items:    []*models.Todo{},
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
		result := &too.AddResult{
			Todo: &models.Todo{
				Text:     "New todo with\nmultiple lines",
				Statuses: map[string]string{"completion": string(models.StatusPending)},
				Items:    []*models.Todo{},
			},
		}

		err := formatter.RenderAdd(&buf, result)
		require.NoError(t, err)

		output := buf.String()
		// For single-line output commands, we expect the newlines to be preserved
		// but the helper function should format them correctly
		assert.Contains(t, output, "Added todo: [ ] New todo with\n      multiple lines")
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
		result := &too.ListResult{
			Todos: []*models.Todo{
				{
					Text:     "Test todo",
					Statuses: map[string]string{"completion": string(models.StatusPending)},
					Items:    []*models.Todo{},
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
