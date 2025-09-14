package output_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/arthur-debert/too/pkg/too"
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEngine_ChangeResult(t *testing.T) {
	engine, err := output.NewEngine()
	require.NoError(t, err)

	// Test JSON format with ChangeResult
	t.Run("JSON format for ChangeResult", func(t *testing.T) {
		todo := &models.Todo{
			UID:      "test-123",
			Text:     "Test todo",
			Statuses: map[string]string{"completion": string(models.StatusPending)},
			Modified: time.Now(),
		}

		result := &too.ChangeResult{
			Command:       "add",
			Message:       "Added todo",
			AffectedTodos: []*models.Todo{todo},
			AllTodos:      []*models.Todo{todo},
			TotalCount:    1,
			DoneCount:     0,
		}

		var buf bytes.Buffer
		err := engine.GetLipbalmEngine().Render(&buf, "json", result)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, `"Command"`)
		assert.Contains(t, output, `"test-123"`)
		assert.Contains(t, output, `"Test todo"`)
	})

	// Test YAML format with ChangeResult  
	t.Run("YAML format for ChangeResult", func(t *testing.T) {
		result := &too.ChangeResult{
			Command: "add",
			Message: "Test message",
		}

		var buf bytes.Buffer
		err := engine.GetLipbalmEngine().Render(&buf, "yaml", result)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "command: add")
		assert.Contains(t, output, "message: Test message")
	})
}

func TestEngine_Formats(t *testing.T) {
	engine, err := output.NewEngine()
	require.NoError(t, err)

	// Test that engine is created successfully
	lipbalmEngine := engine.GetLipbalmEngine()
	require.NotNil(t, lipbalmEngine)

	// Test that basic formats are available
	formats := lipbalmEngine.ListFormats()
	assert.Contains(t, formats, "json")
	assert.Contains(t, formats, "yaml")
	assert.Contains(t, formats, "term")
}

func TestEngine_Hierarchy(t *testing.T) {
	// Test hierarchical todo building
	todos := []*models.Todo{
		{
			UID:          "1",
			Text:         "Parent",
			PositionPath: "1",
		},
		{
			UID:          "2",
			Text:         "Child",
			ParentID:     "1",
			PositionPath: "1.1",
		},
	}

	hierarchical := output.BuildHierarchy(todos)
	require.Len(t, hierarchical, 1)
	assert.Equal(t, "Parent", hierarchical[0].Text)
	assert.Len(t, hierarchical[0].Children, 1)
	assert.Equal(t, "Child", hierarchical[0].Children[0].Text)
}