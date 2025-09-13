package output

import (
	"bytes"
	"testing"

	"github.com/arthur-debert/too/pkg/too"
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRenderer(t *testing.T) {
	t.Run("Default renderer uses term formatter", func(t *testing.T) {
		renderer := NewRenderer(nil)
		require.NotNil(t, renderer)
		assert.NotNil(t, renderer.engine)
		assert.Equal(t, "term", renderer.format)
	})

	t.Run("Custom writer is used", func(t *testing.T) {
		buf := &bytes.Buffer{}
		renderer := NewRenderer(buf)
		require.NotNil(t, renderer)
		assert.Equal(t, buf, renderer.writer)
	})
}

func TestNewRendererWithFormat(t *testing.T) {
	t.Run("Valid format", func(t *testing.T) {
		renderer, err := NewRendererWithFormat("json", nil)
		require.NoError(t, err)
		require.NotNil(t, renderer)
		assert.Equal(t, "json", renderer.format)
	})

	t.Run("Invalid format returns error", func(t *testing.T) {
		_, err := NewRendererWithFormat("invalid", nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid format")
	})
}

func TestRenderer_RenderMethods(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewRenderer(&buf)

	t.Run("RenderChange", func(t *testing.T) {
		buf.Reset()
		result := &too.ChangeResult{
			Command: "test",
			Message: "Test message",
			AllTodos: []*models.IDMTodo{},
			TotalCount: 0,
			DoneCount: 0,
		}
		err := renderer.RenderChange(result)
		require.NoError(t, err)
		assert.NotEmpty(t, buf.String())
	})

	t.Run("RenderMessage", func(t *testing.T) {
		buf.Reset()
		result := &too.MessageResult{
			Text: "Test message",
			Level: "info",
		}
		err := renderer.RenderMessage(result)
		require.NoError(t, err)
		assert.NotEmpty(t, buf.String())
	})

	t.Run("RenderInit", func(t *testing.T) {
		buf.Reset()
		result := &too.InitResult{
			Message: "Initialized",
		}
		err := renderer.RenderInit(result)
		require.NoError(t, err)
		assert.NotEmpty(t, buf.String())
	})

	t.Run("RenderSearch", func(t *testing.T) {
		buf.Reset()
		result := &too.SearchResult{
			Query: "test",
			MatchedTodos: []*models.IDMTodo{},
			TotalCount: 0,
		}
		err := renderer.RenderSearch(result)
		require.NoError(t, err)
		assert.NotEmpty(t, buf.String())
	})

	t.Run("RenderList", func(t *testing.T) {
		buf.Reset()
		result := &too.ListResult{
			Todos: []*models.IDMTodo{},
			TotalCount: 0,
			DoneCount: 0,
		}
		err := renderer.RenderList(result)
		require.NoError(t, err)
		assert.NotEmpty(t, buf.String())
	})

	t.Run("RenderDataPath", func(t *testing.T) {
		buf.Reset()
		result := &too.ShowDataPathResult{
			Path: "/test/path",
		}
		err := renderer.RenderDataPath(result)
		require.NoError(t, err)
		assert.NotEmpty(t, buf.String())
	})

	t.Run("RenderFormats", func(t *testing.T) {
		buf.Reset()
		result := &too.ListFormatsResult{
			Formats: []too.FormatInfo{
				{Name: "json", Description: "JSON format"},
			},
		}
		err := renderer.RenderFormats(result)
		require.NoError(t, err)
		assert.NotEmpty(t, buf.String())
	})

	t.Run("RenderError", func(t *testing.T) {
		buf.Reset()
		err := renderer.RenderError(assert.AnError)
		require.NoError(t, err)
		assert.NotEmpty(t, buf.String())
	})
}

func TestHasFormatter(t *testing.T) {
	// Test for known formatters
	assert.True(t, HasFormatter("json"))
	assert.True(t, HasFormatter("yaml"))
	assert.True(t, HasFormatter("term"))
	
	// Test for unknown formatter
	assert.False(t, HasFormatter("unknown"))
}

func TestList(t *testing.T) {
	formats := List()
	assert.NotEmpty(t, formats)
	
	// Check that standard formats are present
	assert.Contains(t, formats, "json")
	assert.Contains(t, formats, "yaml")
	assert.Contains(t, formats, "csv")
	assert.Contains(t, formats, "markdown")
	assert.Contains(t, formats, "term")
}