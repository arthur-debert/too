package json_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/arthur-debert/too/pkg/too"
	"github.com/arthur-debert/too/pkg/too/commands/formats"
	"github.com/arthur-debert/too/pkg/too/models"
	jsonformatter "github.com/arthur-debert/too/pkg/too/output/formatters/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONFormatter(t *testing.T) {
	formatter := jsonformatter.New()

	t.Run("Name and Description", func(t *testing.T) {
		assert.Equal(t, "json", formatter.Name())
		assert.Equal(t, "JSON output for programmatic consumption", formatter.Description())
	})

	t.Run("RenderChange", func(t *testing.T) {
		var buf bytes.Buffer
		todo := &models.IDMTodo{
			Text:     "Test todo",
			Statuses: map[string]string{"completion": string(models.StatusPending)},
		}
		result := too.NewChangeResult(
			"add",
			[]*models.IDMTodo{todo},
			[]*models.IDMTodo{todo},
			1,
			0,
		)

		err := formatter.RenderChange(&buf, result)
		require.NoError(t, err)

		// Verify it's valid JSON
		var decoded too.ChangeResult
		err = json.Unmarshal(buf.Bytes(), &decoded)
		require.NoError(t, err)
		assert.Equal(t, "add", decoded.Command)
		assert.Len(t, decoded.AffectedTodos, 1)
		assert.Equal(t, todo.Text, decoded.AffectedTodos[0].Text)
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

		// Verify it's valid JSON
		var decoded too.ListResult
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

	t.Run("RenderChange - Complete", func(t *testing.T) {
		var buf bytes.Buffer
		todo := &models.IDMTodo{
			Text:     "Completed todo",
			Statuses: map[string]string{"completion": string(models.StatusDone)},
		}
		result := too.NewChangeResult(
			"completed",
			[]*models.IDMTodo{todo},
			[]*models.IDMTodo{todo},
			1,
			1,
		)

		err := formatter.RenderChange(&buf, result)
		require.NoError(t, err)

		// Verify it's valid JSON
		var decoded too.ChangeResult
		err = json.Unmarshal(buf.Bytes(), &decoded)
		require.NoError(t, err)
		assert.Equal(t, "completed", decoded.Command)
		assert.Len(t, decoded.AffectedTodos, 1)
		assert.Equal(t, todo.Text, decoded.AffectedTodos[0].Text)
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
					Name:        "term",
					Description: "Terminal output",
				},
			},
		}

		err := formatter.RenderFormats(&buf, result)
		require.NoError(t, err)

		// Verify it's valid JSON
		var decoded too.ListFormatsResult
		err = json.Unmarshal(buf.Bytes(), &decoded)
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
		err = json.Unmarshal(buf.Bytes(), &decoded)
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
