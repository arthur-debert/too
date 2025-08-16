package output

import (
	"bytes"
	"testing"

	"github.com/arthur-debert/tdh/pkg/tdh"
	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRenderer(t *testing.T) {
	t.Run("Default renderer uses term formatter", func(t *testing.T) {
		renderer := NewRenderer(nil)
		require.NotNil(t, renderer)
		assert.NotNil(t, renderer.formatter)
		assert.Equal(t, "term", renderer.formatter.Name())
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
		renderer, err := NewRendererWithFormat("term", nil)
		require.NoError(t, err)
		require.NotNil(t, renderer)
		assert.Equal(t, "term", renderer.formatter.Name())
	})

	t.Run("Invalid format returns error", func(t *testing.T) {
		_, err := NewRendererWithFormat("invalid", nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get formatter")
	})
}

func TestRendererMethods(t *testing.T) {
	// Create a renderer with buffer to capture output
	buf := &bytes.Buffer{}
	renderer := NewRenderer(buf)

	t.Run("RenderAdd", func(t *testing.T) {
		result := &tdh.AddResult{
			Todo: &models.Todo{
				Position: 1,
				Text:     "Test todo",
				Status:   models.StatusPending,
			},
		}
		err := renderer.RenderAdd(result)
		require.NoError(t, err)
		// Check that something was written
		assert.NotEmpty(t, buf.String())
	})

	t.Run("RenderList", func(t *testing.T) {
		buf.Reset()
		result := &tdh.ListResult{
			Todos: []*models.Todo{
				{
					Position: 1,
					Text:     "First todo",
					Status:   models.StatusPending,
				},
				{
					Position: 2,
					Text:     "Second todo",
					Status:   models.StatusDone,
				},
			},
			TotalCount: 2,
			DoneCount:  1,
		}
		err := renderer.RenderList(result)
		require.NoError(t, err)
		assert.NotEmpty(t, buf.String())
	})
}
