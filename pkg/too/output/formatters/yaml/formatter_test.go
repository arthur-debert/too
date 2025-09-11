package yaml_test

import (
	"bytes"
	"testing"

	"github.com/arthur-debert/too/pkg/too"
	"github.com/arthur-debert/too/pkg/too/commands/formats"
	"github.com/arthur-debert/too/pkg/too/models"
	yamlformatter "github.com/arthur-debert/too/pkg/too/output/formatters/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestYAMLFormatter(t *testing.T) {
	formatter := yamlformatter.New()

	t.Run("Name and Description", func(t *testing.T) {
		assert.Equal(t, "yaml", formatter.Name())
		assert.Equal(t, "YAML output for programmatic consumption", formatter.Description())
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

		// Verify it's valid YAML
		var decoded too.AddResult
		err = yaml.Unmarshal(buf.Bytes(), &decoded)
		require.NoError(t, err)
		assert.Equal(t, result.Todo.Text, decoded.Todo.Text)
	})

	t.Run("RenderList", func(t *testing.T) {
		var buf bytes.Buffer
		result := &too.ListResult{
			Todos: []*models.Todo{
				{
					Text:     "First todo",
					Statuses: map[string]string{"completion": string(models.StatusPending)},
					Items:    []*models.Todo{},
				},
				{
					Text:     "Second todo",
					Statuses: map[string]string{"completion": string(models.StatusDone)},
					Items:    []*models.Todo{},
				},
			},
			TotalCount: 2,
			DoneCount:  1,
		}

		err := formatter.RenderList(&buf, result)
		require.NoError(t, err)

		// Verify it's valid YAML
		var decoded too.ListResult
		err = yaml.Unmarshal(buf.Bytes(), &decoded)
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

		// Verify it's valid YAML with error field
		var decoded map[string]string
		err = yaml.Unmarshal(buf.Bytes(), &decoded)
		require.NoError(t, err)
		assert.Equal(t, testErr.Error(), decoded["error"])
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

		// Verify it's valid YAML array
		var decoded []*too.CompleteResult
		err = yaml.Unmarshal(buf.Bytes(), &decoded)
		require.NoError(t, err)
		assert.Len(t, decoded, 1)
		assert.Equal(t, results[0].Todo.Text, decoded[0].Todo.Text)
	})

	t.Run("RenderFormats", func(t *testing.T) {
		var buf bytes.Buffer
		result := &too.ListFormatsResult{
			Formats: []formats.Format{
				{
					Name:        "yaml",
					Description: "YAML output",
				},
				{
					Name:        "term",
					Description: "Terminal output",
				},
			},
		}

		err := formatter.RenderFormats(&buf, result)
		require.NoError(t, err)

		// Verify it's valid YAML
		var decoded too.ListFormatsResult
		err = yaml.Unmarshal(buf.Bytes(), &decoded)
		require.NoError(t, err)
		assert.Len(t, decoded.Formats, 2)
	})

	t.Run("Nested todos", func(t *testing.T) {
		var buf bytes.Buffer
		result := &too.ListResult{
			Todos: []*models.Todo{
				{
					Text:     "Parent todo",
					Statuses: map[string]string{"completion": string(models.StatusPending)},
					Items: []*models.Todo{
						{
							Text:     "Child todo",
							Statuses: map[string]string{"completion": string(models.StatusPending)},
							Items:    []*models.Todo{},
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
		var decoded too.ListResult
		err = yaml.Unmarshal(buf.Bytes(), &decoded)
		require.NoError(t, err)
		assert.Len(t, decoded.Todos, 1)
		assert.Len(t, decoded.Todos[0].Items, 1)
		assert.Equal(t, "Child todo", decoded.Todos[0].Items[0].Text)
	})
}
