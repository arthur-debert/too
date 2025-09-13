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

func TestEngine_Formats(t *testing.T) {
	engine, err := output.NewEngine()
	require.NoError(t, err)

	// Test JSON format
	t.Run("JSON format", func(t *testing.T) {
		todo := &models.IDMTodo{
			UID:      "test-123",
			Text:     "Test todo",
			Statuses: map[string]string{"completion": string(models.StatusPending)},
			Modified: time.Now(),
		}

		result := &too.ListResult{
			Todos:      []*models.IDMTodo{todo},
			TotalCount: 1,
			DoneCount:  0,
		}

		var buf bytes.Buffer
		err := engine.GetLipbalmEngine().Render(&buf, "json", result)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, `"Todos"`)
		assert.Contains(t, output, `"test-123"`)
		assert.Contains(t, output, `"Test todo"`)
	})

	// Test YAML format
	t.Run("YAML format", func(t *testing.T) {
		result := &too.MessageResult{
			Text:  "Test message",
			Level: "info",
		}

		var buf bytes.Buffer
		err := engine.GetLipbalmEngine().Render(&buf, "yaml", result)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "text: Test message")
		assert.Contains(t, output, "level: info")
	})

	// Test Markdown format
	t.Run("Markdown format", func(t *testing.T) {
		todo1 := &models.IDMTodo{
			UID:      "parent-1",
			Text:     "Parent todo",
			Statuses: map[string]string{"completion": string(models.StatusPending)},
			Modified: time.Now(),
		}

		todo2 := &models.IDMTodo{
			UID:      "child-1",
			Text:     "Child todo",
			Statuses: map[string]string{"completion": string(models.StatusDone)},
			ParentID: "parent-1",
			Modified: time.Now(),
		}

		result := &too.ListResult{
			Todos:      []*models.IDMTodo{todo1, todo2},
			TotalCount: 2,
			DoneCount:  1,
		}

		var buf bytes.Buffer
		err := engine.GetLipbalmEngine().Render(&buf, "markdown", result)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "1. [ ] Parent todo")
		assert.Contains(t, output, "   1. [x] Child todo")
		assert.Contains(t, output, "2 todo(s), 1 done")
	})

	// Test CSV format
	t.Run("CSV format", func(t *testing.T) {
		todo := &models.IDMTodo{
			UID:      "test-123",
			Text:     "Test todo",
			Statuses: map[string]string{"completion": string(models.StatusPending)},
			Modified: time.Now(),
		}

		result := &too.ListResult{
			Todos:      []*models.IDMTodo{todo},
			TotalCount: 1,
			DoneCount:  0,
		}

		var buf bytes.Buffer
		err := engine.GetLipbalmEngine().Render(&buf, "csv", result)
		require.NoError(t, err)

		output := buf.String()
		// For now, just check the CSV contains the data
		// The actual format depends on how lipbalm renders structs to CSV
		assert.NotEmpty(t, output)
		// TODO: Implement proper CSV rendering for ListResult
	})
}

func TestEngine_ErrorHandling(t *testing.T) {
	engine, err := output.NewEngine()
	require.NoError(t, err)

	t.Run("RenderError", func(t *testing.T) {
		testErr := assert.AnError
		var buf bytes.Buffer

		err := engine.GetLipbalmEngine().RenderError(&buf, "json", testErr)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, `"error"`)
		assert.Contains(t, output, testErr.Error())
	})
}

func TestEngine_ChangeResult(t *testing.T) {
	engine, err := output.NewEngine()
	require.NoError(t, err)

	todo := &models.IDMTodo{
		UID:      "test-123",
		Text:     "Test todo",
		Statuses: map[string]string{"completion": string(models.StatusPending)},
		Modified: time.Now(),
	}

	result := &too.ChangeResult{
		Command:       "add",
		Message:       "Added todo",
		AffectedTodos: []*models.IDMTodo{todo},
		AllTodos:      []*models.IDMTodo{todo},
		TotalCount:    1,
		DoneCount:     0,
	}

	t.Run("Markdown shows affected count", func(t *testing.T) {
		var buf bytes.Buffer
		err := engine.GetLipbalmEngine().Render(&buf, "markdown", result)
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "Added 1 todo(s)")
		assert.Contains(t, output, "1. [ ] Test todo")
	})
}