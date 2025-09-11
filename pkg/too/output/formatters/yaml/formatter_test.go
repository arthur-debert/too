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
			Todo: &models.IDMTodo{
				Text:     "Test todo",
				Statuses: map[string]string{"completion": string(models.StatusPending)},
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
			Todos: []*models.IDMTodo{
				{
					Text:     "First todo",
					Statuses: map[string]string{"completion": string(models.StatusPending)},
				},
				{
					Text:     "Second todo",
					Statuses: map[string]string{"completion": string(models.StatusDone)},
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
				Todo: &models.IDMTodo{
					Text:     "Completed todo",
					Statuses: map[string]string{"completion": string(models.StatusDone)},
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
			Todos: []*models.IDMTodo{
				{
					Text:     "Parent todo",
					Statuses: map[string]string{"completion": string(models.StatusPending)},
					ParentID: "",
				},
				{
					Text:     "Child todo",
					Statuses: map[string]string{"completion": string(models.StatusPending)},
					ParentID: "parent-uid",
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
		// With flat structure, we have both parent and child
		assert.Len(t, decoded.Todos, 2)
		// Verify we have both parent and child todos
		var parentFound, childFound bool
		for _, todo := range decoded.Todos {
			if todo.Text == "Parent todo" {
				parentFound = true
			} else if todo.Text == "Child todo" {
				childFound = true
			}
		}
		assert.True(t, parentFound, "Should have parent todo")
		assert.True(t, childFound, "Should have child todo")
	})
}
