package csv_test

import (
	"bytes"
	"encoding/csv"
	"strings"
	"testing"

	"github.com/arthur-debert/too/pkg/too"
	"github.com/arthur-debert/too/pkg/too/commands/formats"
	"github.com/arthur-debert/too/pkg/too/models"
	csvformatter "github.com/arthur-debert/too/pkg/too/output/formatters/csv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCSVFormatter(t *testing.T) {
	formatter := csvformatter.New()

	t.Run("Name and Description", func(t *testing.T) {
		assert.Equal(t, "csv", formatter.Name())
		assert.Equal(t, "CSV output for spreadsheet applications", formatter.Description())
	})

	t.Run("RenderAdd", func(t *testing.T) {
		var buf bytes.Buffer
		result := &too.AddResult{
			Todo: &models.Todo{
				ID:       "123",
				Position: 1,
				Text:     "Test todo",
				Statuses: map[string]string{"completion": string(models.StatusPending)},
				Items:    []*models.Todo{},
			},
		}

		err := formatter.RenderAdd(&buf, result)
		require.NoError(t, err)

		// Parse CSV
		reader := csv.NewReader(&buf)
		records, err := reader.ReadAll()
		require.NoError(t, err)

		// Check headers
		assert.Equal(t, []string{"id", "parent", "position", "text", "status", "modified"}, records[0])

		// Check data
		assert.Len(t, records, 2) // header + 1 row
		assert.Equal(t, "123", records[1][0])
		assert.Equal(t, "", records[1][1]) // no parent
		assert.Equal(t, "1", records[1][2])
		assert.Equal(t, "Test todo", records[1][3])
		assert.Equal(t, "pending", records[1][4])
	})

	t.Run("RenderList", func(t *testing.T) {
		var buf bytes.Buffer
		result := &too.ListResult{
			Todos: []*models.Todo{
				{
					ID:       "1",
					Position: 1,
					Text:     "First todo",
					Statuses: map[string]string{"completion": string(models.StatusPending)},
					Items:    []*models.Todo{},
				},
				{
					ID:       "2",
					Position: 2,
					Text:     "Second todo, with comma",
					Statuses: map[string]string{"completion": string(models.StatusDone)},
					Items:    []*models.Todo{},
				},
			},
			TotalCount: 2,
			DoneCount:  1,
		}

		err := formatter.RenderList(&buf, result)
		require.NoError(t, err)

		// Split by blank line
		parts := strings.Split(buf.String(), "\n\n")
		require.Len(t, parts, 2)

		// Parse summary CSV
		summaryReader := csv.NewReader(strings.NewReader(parts[0]))
		summaryRecords, err := summaryReader.ReadAll()
		require.NoError(t, err)
		assert.Equal(t, []string{"total_count", "done_count", "pending_count"}, summaryRecords[0])
		assert.Equal(t, []string{"2", "1", "1"}, summaryRecords[1])

		// Parse todos CSV
		todosReader := csv.NewReader(strings.NewReader(parts[1]))
		todosRecords, err := todosReader.ReadAll()
		require.NoError(t, err)
		assert.Len(t, todosRecords, 3) // header + 2 rows

		// Check that comma in text is properly escaped
		assert.Equal(t, "Second todo, with comma", todosRecords[2][3])
	})

	t.Run("RenderError", func(t *testing.T) {
		var buf bytes.Buffer
		testErr := assert.AnError

		err := formatter.RenderError(&buf, testErr)
		require.NoError(t, err)

		// Parse CSV
		reader := csv.NewReader(&buf)
		records, err := reader.ReadAll()
		require.NoError(t, err)

		assert.Equal(t, []string{"error"}, records[0])
		assert.Equal(t, testErr.Error(), records[1][0])
	})

	t.Run("RenderFormats", func(t *testing.T) {
		var buf bytes.Buffer
		result := &too.ListFormatsResult{
			Formats: []formats.Format{
				{
					Name:        "csv",
					Description: "CSV output",
				},
				{
					Name:        "json",
					Description: "JSON output",
				},
			},
		}

		err := formatter.RenderFormats(&buf, result)
		require.NoError(t, err)

		// Parse CSV
		reader := csv.NewReader(&buf)
		records, err := reader.ReadAll()
		require.NoError(t, err)

		assert.Len(t, records, 3) // header + 2 rows
		assert.Equal(t, []string{"name", "description"}, records[0])
		assert.Equal(t, "csv", records[1][0])
		assert.Equal(t, "json", records[2][0])
	})

	t.Run("Nested todos with hierarchy", func(t *testing.T) {
		var buf bytes.Buffer
		result := &too.ListResult{
			Todos: []*models.Todo{
				{
					ID:       "1",
					Position: 1,
					Text:     "Parent todo",
					Statuses: map[string]string{"completion": string(models.StatusPending)},
					Items: []*models.Todo{
						{
							ID:       "1.1",
							Position: 1,
							Text:     "Child todo",
							Statuses: map[string]string{"completion": string(models.StatusPending)},
							Items: []*models.Todo{
								{
									ID:       "1.1.1",
									Position: 1,
									Text:     "Grandchild todo",
									Statuses: map[string]string{"completion": string(models.StatusPending)},
									Items:    []*models.Todo{},
								},
							},
						},
					},
				},
			},
			TotalCount: 3,
			DoneCount:  0,
		}

		err := formatter.RenderList(&buf, result)
		require.NoError(t, err)

		// Split by blank line and parse todos
		parts := strings.Split(buf.String(), "\n\n")
		todosReader := csv.NewReader(strings.NewReader(parts[1]))
		todosRecords, err := todosReader.ReadAll()
		require.NoError(t, err)

		assert.Len(t, todosRecords, 4) // header + 3 rows

		// Check parent hierarchy
		assert.Equal(t, "", todosRecords[1][1])                         // Parent has no parent
		assert.Equal(t, "Parent todo", todosRecords[2][1])              // Child's parent
		assert.Equal(t, "Parent todo > Child todo", todosRecords[3][1]) // Grandchild's parent path
	})

	t.Run("Text with newlines", func(t *testing.T) {
		var buf bytes.Buffer
		result := &too.AddResult{
			Todo: &models.Todo{
				ID:       "123",
				Position: 1,
				Text:     "Todo with\nnewline",
				Statuses: map[string]string{"completion": string(models.StatusPending)},
				Items:    []*models.Todo{},
			},
		}

		err := formatter.RenderAdd(&buf, result)
		require.NoError(t, err)

		// Parse CSV - the CSV package should handle newlines properly
		reader := csv.NewReader(&buf)
		records, err := reader.ReadAll()
		require.NoError(t, err)

		assert.Len(t, records, 2)
		assert.Equal(t, "Todo with\nnewline", records[1][3])
	})

	t.Run("Special characters", func(t *testing.T) {
		var buf bytes.Buffer
		result := &too.AddResult{
			Todo: &models.Todo{
				ID:       "123",
				Position: 1,
				Text:     `Todo with "quotes", commas, and 'apostrophes'`,
				Statuses: map[string]string{"completion": string(models.StatusPending)},
				Items:    []*models.Todo{},
			},
		}

		err := formatter.RenderAdd(&buf, result)
		require.NoError(t, err)

		// Parse CSV
		reader := csv.NewReader(&buf)
		records, err := reader.ReadAll()
		require.NoError(t, err)

		assert.Len(t, records, 2)
		assert.Equal(t, `Todo with "quotes", commas, and 'apostrophes'`, records[1][3])
	})

	t.Run("RenderMove", func(t *testing.T) {
		var buf bytes.Buffer
		result := &too.MoveResult{
			Todo: &models.Todo{
				ID:       "123",
				Position: 3,
				Text:     "Moved todo",
				Statuses: map[string]string{"completion": string(models.StatusPending)},
				Items:    []*models.Todo{},
			},
			OldPath: "1",
			NewPath: "3",
		}

		err := formatter.RenderMove(&buf, result)
		require.NoError(t, err)

		// Parse CSV
		reader := csv.NewReader(&buf)
		records, err := reader.ReadAll()
		require.NoError(t, err)

		assert.Equal(t, []string{"id", "text", "old_path", "new_path", "status"}, records[0])
		assert.Equal(t, "1", records[1][2]) // old path
		assert.Equal(t, "3", records[1][3]) // new path
	})
}
