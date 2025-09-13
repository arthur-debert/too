package lipbalm_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/arthur-debert/too/pkg/lipbalm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderEngine_Basic(t *testing.T) {
	engine := lipbalm.Quick()

	t.Run("JSON formatting", func(t *testing.T) {
		data := map[string]string{"name": "test", "value": "123"}
		var buf bytes.Buffer

		err := engine.Render(&buf, "json", data)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, `"name": "test"`)
		assert.Contains(t, output, `"value": "123"`)
	})

	t.Run("YAML formatting", func(t *testing.T) {
		data := map[string]string{"name": "test", "value": "123"}
		var buf bytes.Buffer

		err := engine.Render(&buf, "yaml", data)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "name: test")
		assert.Contains(t, output, "value: \"123\"")
	})

	t.Run("Plain formatting", func(t *testing.T) {
		data := map[string]string{"name": "test"}
		var buf bytes.Buffer

		err := engine.Render(&buf, "plain", data)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, `{"name":"test"}`)
	})
}

func TestRenderEngine_Callbacks(t *testing.T) {
	t.Run("PreProcess callback", func(t *testing.T) {
		engine := lipbalm.WithCallbacks(lipbalm.RenderCallbacks{
			PreProcess: func(format string, data interface{}) interface{} {
				// Add extra field
				if m, ok := data.(map[string]string); ok {
					m["processed"] = "true"
					return m
				}
				return data
			},
		})

		data := map[string]string{"name": "test"}
		var buf bytes.Buffer

		err := engine.Render(&buf, "json", data)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, `"processed": "true"`)
	})

	t.Run("PostProcess callback", func(t *testing.T) {
		engine := lipbalm.WithCallbacks(lipbalm.RenderCallbacks{
			PostProcess: func(format string, output string) string {
				if format == "json" {
					return "// JSON output\n" + output
				}
				return output
			},
		})

		data := map[string]string{"name": "test"}
		var buf bytes.Buffer

		err := engine.Render(&buf, "json", data)
		require.NoError(t, err)

		output := buf.String()
		assert.True(t, strings.HasPrefix(output, "// JSON output\n"))
	})
}

func TestRenderEngine_ErrorHandling(t *testing.T) {
	engine := lipbalm.Quick()

	t.Run("Unknown format", func(t *testing.T) {
		var buf bytes.Buffer
		err := engine.Render(&buf, "unknown", nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown format")
	})

	t.Run("RenderError method", func(t *testing.T) {
		var buf bytes.Buffer
		testErr := assert.AnError

		err := engine.RenderError(&buf, "json", testErr)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, `"error":`)
		assert.Contains(t, output, testErr.Error())
	})
}

func TestRenderEngine_Formats(t *testing.T) {
	engine := lipbalm.Quick()

	t.Run("List formats", func(t *testing.T) {
		formats := engine.ListFormats()
		assert.Contains(t, formats, "json")
		assert.Contains(t, formats, "yaml")
		assert.Contains(t, formats, "csv")
		assert.Contains(t, formats, "markdown")
		assert.Contains(t, formats, "plain")
		assert.Contains(t, formats, "term")
	})

	t.Run("Set default format", func(t *testing.T) {
		err := engine.SetFormat("yaml")
		require.NoError(t, err)
		assert.Equal(t, "yaml", engine.GetFormat())

		// Test with invalid format
		err = engine.SetFormat("invalid")
		assert.Error(t, err)
	})
}

func TestRenderEngine_CSV(t *testing.T) {
	engine := lipbalm.Quick()

	t.Run("Struct slice to CSV", func(t *testing.T) {
		type Person struct {
			Name  string `json:"name"`
			Age   int    `json:"age"`
			Email string `json:"email"`
		}

		data := []Person{
			{Name: "Alice", Age: 30, Email: "alice@example.com"},
			{Name: "Bob", Age: 25, Email: "bob@example.com"},
		}

		var buf bytes.Buffer
		err := engine.Render(&buf, "csv", data)
		require.NoError(t, err)

		output := buf.String()
		lines := strings.Split(strings.TrimSpace(output), "\n")
		
		// Check headers
		assert.Equal(t, "name,age,email", lines[0])
		
		// Check data
		assert.Equal(t, "Alice,30,alice@example.com", lines[1])
		assert.Equal(t, "Bob,25,bob@example.com", lines[2])
	})

	t.Run("Map to CSV", func(t *testing.T) {
		data := map[string]int{"foo": 1, "bar": 2}

		var buf bytes.Buffer
		err := engine.Render(&buf, "csv", data)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "key,value")
		// Maps are unordered, so just check both entries exist
		assert.Contains(t, output, "foo,1")
		assert.Contains(t, output, "bar,2")
	})
}

func TestRenderEngine_Markdown(t *testing.T) {
	engine := lipbalm.Quick()

	t.Run("Struct to Markdown", func(t *testing.T) {
		type Task struct {
			Title    string `json:"title"`
			Status   string `json:"status"`
			Priority int    `json:"priority"`
		}

		data := Task{
			Title:    "Implement feature",
			Status:   "in_progress",
			Priority: 1,
		}

		var buf bytes.Buffer
		err := engine.Render(&buf, "markdown", data)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "**title:** Implement feature")
		assert.Contains(t, output, "**status:** in_progress")
		assert.Contains(t, output, "**priority:** 1")
	})

	t.Run("Slice to Markdown table", func(t *testing.T) {
		type Item struct {
			Name  string `json:"name"`
			Count int    `json:"count"`
		}

		data := []Item{
			{Name: "Apple", Count: 5},
			{Name: "Banana", Count: 3},
		}

		var buf bytes.Buffer
		err := engine.Render(&buf, "markdown", data)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "| name | count |")
		assert.Contains(t, output, "| --- | --- |")
		assert.Contains(t, output, "| Apple | 5 |")
		assert.Contains(t, output, "| Banana | 3 |")
	})
}