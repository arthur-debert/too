package json_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/arthur-debert/tdh/pkg/tdh"
	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/arthur-debert/tdh/pkg/tdh/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONFormatterRegistration(t *testing.T) {
	t.Run("JSON formatter is registered", func(t *testing.T) {
		// Check that JSON formatter is in the list
		names := output.List()
		assert.Contains(t, names, "json")

		// Get the formatter
		formatter, err := output.Get("json")
		require.NoError(t, err)
		assert.Equal(t, "json", formatter.Name())
	})

	t.Run("Can create renderer with JSON format", func(t *testing.T) {
		buf := &bytes.Buffer{}
		renderer, err := output.NewRendererWithFormat("json", buf)
		require.NoError(t, err)
		require.NotNil(t, renderer)

		// Test rendering
		result := &tdh.AddResult{
			Todo: &models.Todo{
				Position: 1,
				Text:     "Test todo",
				Status:   models.StatusPending,
			},
		}

		err = renderer.RenderAdd(result)
		require.NoError(t, err)

		// Verify JSON output
		var decoded tdh.AddResult
		err = json.Unmarshal(buf.Bytes(), &decoded)
		require.NoError(t, err)
		assert.Equal(t, "Test todo", decoded.Todo.Text)
	})
}
