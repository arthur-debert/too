package markdown_test

import (
	"bytes"
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

	t.Run("RenderChange", func(t *testing.T) {
		var buf bytes.Buffer
		todo := &models.IDMTodo{
			Text:     "Test todo",
			Statuses: map[string]string{"completion": string(models.StatusPending)},
		}
		result := too.NewChangeResult(
			"add",
			"",
			[]*models.IDMTodo{todo},
			[]*models.IDMTodo{todo},
			1,
			0,
		)

		err := formatter.RenderChange(&buf, result)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "Added 1 todo(s)")
		assert.Contains(t, output, "1. [ ] Test todo")
	})

	t.Run("RenderList with nested todos", func(t *testing.T) {
		var buf bytes.Buffer
		result := &too.ListResult{
			Todos: []*models.IDMTodo{
				{
					UID:      "parent-uid",
					Text:     "Parent todo",
					Statuses: map[string]string{"completion": string(models.StatusPending)},
					ParentID: "",
				},
				{
					UID:      "child1-uid",
					Text:     "Child todo 1",
					Statuses: map[string]string{"completion": string(models.StatusDone)},
					ParentID: "parent-uid",
				},
				{
					UID:      "child2-uid",
					Text:     "Child todo 2",
					Statuses: map[string]string{"completion": string(models.StatusPending)},
					ParentID: "parent-uid",
				},
				{
					UID:      "grandchild-uid",
					Text:     "Grandchild todo",
					Statuses: map[string]string{"completion": string(models.StatusPending)},
					ParentID: "child2-uid",
				},
				{
					UID:      "second-parent-uid",
					Text:     "Second parent",
					Statuses: map[string]string{"completion": string(models.StatusDone)},
					ParentID: "",
				},
			},
			TotalCount: 5,
			DoneCount:  2,
		}

		err := formatter.RenderList(&buf, result)
		require.NoError(t, err)

		output := buf.String()

		// The markdown formatter should use hierarchy building to display nested structure
		// But the input is now flat IDMTodos - the formatter will build hierarchy internally
		assert.Contains(t, output, "Parent todo")
		assert.Contains(t, output, "Child todo 1") 
		assert.Contains(t, output, "Child todo 2")
		assert.Contains(t, output, "Grandchild todo")
		assert.Contains(t, output, "Second parent")

		// Check summary
		assert.Contains(t, output, "5 todo(s), 2 done")
	})

	t.Run("RenderList empty", func(t *testing.T) {
		var buf bytes.Buffer
		result := &too.ListResult{
			Todos:      []*models.IDMTodo{},
			TotalCount: 0,
			DoneCount:  0,
		}

		err := formatter.RenderList(&buf, result)
		require.NoError(t, err)
		assert.Equal(t, "No todos\n", buf.String())
	})

	t.Run("RenderChange - Complete", func(t *testing.T) {
		var buf bytes.Buffer
		todo := &models.IDMTodo{
			Text:     "Completed todo",
			Statuses: map[string]string{"completion": string(models.StatusDone)},
		}
		result := too.NewChangeResult(
			"completed",
			"",
			[]*models.IDMTodo{todo},
			[]*models.IDMTodo{todo},
			1,
			1,
		)

		err := formatter.RenderChange(&buf, result)
		require.NoError(t, err)
		output := buf.String()
		assert.Contains(t, output, "Completed 1 todo(s)")
		assert.Contains(t, output, "[x] Completed todo")
	})

	t.Run("RenderSearch", func(t *testing.T) {
		var buf bytes.Buffer
		result := &too.SearchResult{
			MatchedTodos: []*models.IDMTodo{
				{
					Text:     "Matching todo",
					Statuses: map[string]string{"completion": string(models.StatusPending)},
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

	t.Run("RenderMessage", func(t *testing.T) {
		var buf bytes.Buffer
		result := too.NewInfoMessage("Todo collection initialized")

		err := formatter.RenderMessage(&buf, result)
		require.NoError(t, err)
		assert.Equal(t, "Todo collection initialized\n", buf.String())
	})

	t.Run("RenderMessage with levels", func(t *testing.T) {
		var buf bytes.Buffer
		
		// Test success message
		result := too.NewMessageResult("Operation successful", "success")
		err := formatter.RenderMessage(&buf, result)
		require.NoError(t, err)
		assert.Equal(t, "✓ Operation successful\n", buf.String())
		
		// Test warning message
		buf.Reset()
		result = too.NewMessageResult("Warning message", "warning")
		err = formatter.RenderMessage(&buf, result)
		require.NoError(t, err)
		assert.Equal(t, "**Warning:** Warning message\n", buf.String())
	})

	t.Run("RenderList with multiline todos", func(t *testing.T) {
		var buf bytes.Buffer
		result := &too.ListResult{
			Todos: []*models.IDMTodo{
				{
					UID:      "single-uid",
					Text:     "Single line todo",
					Statuses: map[string]string{"completion": string(models.StatusPending)},
					ParentID: "",
				},
				{
					UID:      "multiline-uid",
					Text:     "Multi-line todo\nSecond line\nThird line",
					Statuses: map[string]string{"completion": string(models.StatusPending)},
					ParentID: "",
				},
				{
					UID:      "nested-parent-uid",
					Text:     "Nested parent",
					Statuses: map[string]string{"completion": string(models.StatusPending)},
					ParentID: "",
				},
				{
					UID:      "nested-child-uid",
					Text:     "Nested child with\nmultiple lines",
					Statuses: map[string]string{"completion": string(models.StatusPending)},
					ParentID: "nested-parent-uid",
				},
			},
			TotalCount: 4,
			DoneCount:  0,
		}

		err := formatter.RenderList(&buf, result)
		require.NoError(t, err)

		output := buf.String()

		// Check that all todos are present (exact line structure may depend on formatter implementation)
		assert.Contains(t, output, "Single line todo")
		assert.Contains(t, output, "Multi-line todo")
		assert.Contains(t, output, "Second line")
		assert.Contains(t, output, "Third line")
		assert.Contains(t, output, "Nested parent") 
		assert.Contains(t, output, "Nested child with")
		assert.Contains(t, output, "multiple lines")
	})

	t.Run("RenderChange with multiline todo", func(t *testing.T) {
		var buf bytes.Buffer
		todo := &models.IDMTodo{
			Text:     "New todo with\nmultiple lines",
			Statuses: map[string]string{"completion": string(models.StatusPending)},
		}
		result := too.NewChangeResult(
			"add",
			"",
			[]*models.IDMTodo{todo},
			[]*models.IDMTodo{todo},
			1,
			0,
		)

		err := formatter.RenderChange(&buf, result)
		require.NoError(t, err)

		output := buf.String()
		// For single-line output commands, we expect the newlines to be preserved
		// but the helper function should format them correctly
		assert.Contains(t, output, "Added 1 todo(s)")
		assert.Contains(t, output, "New todo with")
		assert.Contains(t, output, "multiple lines")
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
			Todos: []*models.IDMTodo{
				{
					Text:     "Test todo",
					Statuses: map[string]string{"completion": string(models.StatusPending)},
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
