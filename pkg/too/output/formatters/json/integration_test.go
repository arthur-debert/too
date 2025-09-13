package json_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/arthur-debert/too/pkg/too"
	"github.com/arthur-debert/too/pkg/too/models"
	. "github.com/arthur-debert/too/pkg/too/output/formatters/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestJSONFormatterBehavior tests the JSON formatter behavior directly
func TestJSONFormatterBehavior(t *testing.T) {
	t.Run("JSON formatter renders correctly", func(t *testing.T) {
		buf := &bytes.Buffer{}
		formatter := New()
		
		// Test rendering a change result
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

		err := formatter.RenderChange(buf, result)
		require.NoError(t, err)

		// Verify JSON output
		var decoded too.ChangeResult
		err = json.Unmarshal(buf.Bytes(), &decoded)
		require.NoError(t, err)
		assert.Equal(t, "add", decoded.Command)
		assert.Len(t, decoded.AffectedTodos, 1)
		assert.Equal(t, "Test todo", decoded.AffectedTodos[0].Text)
	})
	
	t.Run("JSON formatter has correct metadata", func(t *testing.T) {
		formatter := New()
		assert.Equal(t, "json", formatter.Name())
		assert.Equal(t, "JSON output for programmatic consumption", formatter.Description())
	})
}
