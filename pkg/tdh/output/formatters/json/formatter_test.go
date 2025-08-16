package json_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/arthur-debert/tdh/pkg/tdh"
	"github.com/arthur-debert/tdh/pkg/tdh/commands/formats"
	"github.com/arthur-debert/tdh/pkg/tdh/models"
	jsonformatter "github.com/arthur-debert/tdh/pkg/tdh/output/formatters/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONFormatter(t *testing.T) {
	formatter := jsonformatter.New()

	t.Run("Name and Description", func(t *testing.T) {
		assert.Equal(t, "json", formatter.Name())
		assert.Equal(t, "JSON output for programmatic consumption", formatter.Description())
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

		// Verify it's valid JSON
		var decoded tdh.AddResult
		err = json.Unmarshal(buf.Bytes(), &decoded)
		require.NoError(t, err)
		assert.Equal(t, result.Todo.Text, decoded.Todo.Text)
		assert.Equal(t, result.Todo.Position, decoded.Todo.Position)
	})

	t.Run("RenderList", func(t *testing.T) {
		var buf bytes.Buffer
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

		err := formatter.RenderList(&buf, result)
		require.NoError(t, err)

		// Verify it's valid JSON
		var decoded tdh.ListResult
		err = json.Unmarshal(buf.Bytes(), &decoded)
		require.NoError(t, err)
		assert.Len(t, decoded.Todos, 2)
		assert.Equal(t, result.TotalCount, decoded.TotalCount)
		assert.Equal(t, result.DoneCount, decoded.DoneCount)
	})

	t.Run("RenderError", func(t *testing.T) {
		var buf bytes.Buffer
		testErr := assert.AnError

		err := formatter.RenderError(&buf, testErr)
		require.NoError(t, err)

		// Verify it's valid JSON with error field
		var decoded map[string]string
		err = json.Unmarshal(buf.Bytes(), &decoded)
		require.NoError(t, err)
		assert.Equal(t, testErr.Error(), decoded["error"])
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

		// Verify it's valid JSON array
		var decoded []*tdh.CompleteResult
		err = json.Unmarshal(buf.Bytes(), &decoded)
		require.NoError(t, err)
		assert.Len(t, decoded, 1)
		assert.Equal(t, results[0].Todo.Text, decoded[0].Todo.Text)
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
					Name:        "term",
					Description: "Terminal output",
				},
			},
		}

		err := formatter.RenderFormats(&buf, result)
		require.NoError(t, err)

		// Verify it's valid JSON
		var decoded tdh.ListFormatsResult
		err = json.Unmarshal(buf.Bytes(), &decoded)
		require.NoError(t, err)
		assert.Len(t, decoded.Formats, 2)
	})

	t.Run("Nested todos", func(t *testing.T) {
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
							Text:     "Child todo",
							Status:   models.StatusPending,
						},
					},
				},
			},
			TotalCount: 2,
			DoneCount:  0,
		}

		err := formatter.RenderList(&buf, result)
		require.NoError(t, err)

		// Verify nested structure is preserved
		var decoded tdh.ListResult
		err = json.Unmarshal(buf.Bytes(), &decoded)
		require.NoError(t, err)
		assert.Len(t, decoded.Todos, 1)
		assert.Len(t, decoded.Todos[0].Items, 1)
		assert.Equal(t, "Child todo", decoded.Todos[0].Items[0].Text)
	})
}
