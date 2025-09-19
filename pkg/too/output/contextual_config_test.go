package output

import (
	"bytes"
	"testing"

	"github.com/arthur-debert/too/pkg/too"
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContextualViewConfig(t *testing.T) {
	// Save original config
	originalConfig := too.GetConfig()
	defer too.SetConfig(originalConfig)

	// Create test data
	todos := []*models.Todo{
		{UID: "1", Text: "Item 1", PositionPath: "1"},
		{UID: "2", Text: "Item 2", PositionPath: "2"},
		{UID: "3", Text: "Item 3", PositionPath: "3"},
		{UID: "4", Text: "Item 4", PositionPath: "4"},
		{UID: "5", Text: "Item 5", PositionPath: "5"},
	}

	changeResult := &too.ChangeResult{
		Command:       "edit",
		AffectedTodos: []*models.Todo{todos[2]}, // Item 3 is affected
		AllTodos:      todos,
		TotalCount:    5,
		DoneCount:     0,
	}

	// Create engine
	engine, err := NewEngine()
	require.NoError(t, err)

	t.Run("contextual view disabled", func(t *testing.T) {
		// Disable contextual view
		config := too.DefaultConfig()
		config.Display.UseContextualChangeView = false
		too.SetConfig(config)

		// Render
		var buf bytes.Buffer
		err := engine.GetLipbalmEngine().Render(&buf, "term", changeResult)
		require.NoError(t, err)

		output := buf.String()
		
		// Should show all items
		assert.Contains(t, output, "Item 1")
		assert.Contains(t, output, "Item 2")
		assert.Contains(t, output, "Item 3")
		assert.Contains(t, output, "Item 4") 
		assert.Contains(t, output, "Item 5")
		
		// Should NOT have ellipsis
		assert.NotContains(t, output, "…")
	})

	t.Run("contextual view enabled", func(t *testing.T) {
		// Enable contextual view
		config := too.DefaultConfig()
		config.Display.UseContextualChangeView = true
		too.SetConfig(config)

		// Render
		var buf bytes.Buffer
		err := engine.GetLipbalmEngine().Render(&buf, "term", changeResult)
		require.NoError(t, err)

		output := buf.String()
		
		// Should show context items
		assert.Contains(t, output, "Item 1")
		assert.Contains(t, output, "Item 2")
		assert.Contains(t, output, "Item 3")
		assert.Contains(t, output, "Item 4")
		assert.Contains(t, output, "Item 5")
		
		// Should NOT have ellipsis for this case (all 5 items fit in context)
		assert.NotContains(t, output, "…")
	})

	t.Run("contextual view with ellipsis", func(t *testing.T) {
		// Add more items to trigger ellipsis
		moreTodos := make([]*models.Todo, 10)
		for i := 0; i < 10; i++ {
			moreTodos[i] = &models.Todo{
				UID:          string(rune('a' + i)),
				Text:         string(rune('A' + i)),
				PositionPath: string(rune('1' + i)),
			}
		}

		changeResult := &too.ChangeResult{
			Command:       "edit",
			AffectedTodos: []*models.Todo{moreTodos[8]}, // Second to last item
			AllTodos:      moreTodos,
			TotalCount:    10,
			DoneCount:     0,
		}

		// Enable contextual view
		config := too.DefaultConfig()
		config.Display.UseContextualChangeView = true
		too.SetConfig(config)

		// Render
		var buf bytes.Buffer
		err := engine.GetLipbalmEngine().Render(&buf, "term", changeResult)
		require.NoError(t, err)

		output := buf.String()
		
		// Should have ellipsis before
		assert.Contains(t, output, "…")
		
		// Should show the highlighted item
		assert.Contains(t, output, moreTodos[8].Text)
		
		// Should show 2 items before
		assert.Contains(t, output, moreTodos[6].Text)
		assert.Contains(t, output, moreTodos[7].Text)
		
		// Should show 1 item after (last one)
		assert.Contains(t, output, moreTodos[9].Text)
		
		// Should NOT show items too far away
		assert.NotContains(t, output, moreTodos[0].Text)
		assert.NotContains(t, output, moreTodos[1].Text)
	})
}