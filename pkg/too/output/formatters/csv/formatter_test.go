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

	t.Run("RenderChange", func(t *testing.T) {
		var buf bytes.Buffer
		todo := &models.IDMTodo{
			UID:      "123",
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

		// CSV has two sections separated by blank line
		content := buf.String()
		sections := strings.Split(content, "\n\n")
		require.Len(t, sections, 2)

		// Parse summary section
		summaryReader := csv.NewReader(strings.NewReader(sections[0]))
		summaryRecords, err := summaryReader.ReadAll()
		require.NoError(t, err)
		assert.Equal(t, []string{"command", "affected_count", "total_count", "done_count"}, summaryRecords[0])
		assert.Equal(t, "add", summaryRecords[1][0])
		assert.Equal(t, "1", summaryRecords[1][1])
		assert.Equal(t, "1", summaryRecords[1][2])
		assert.Equal(t, "0", summaryRecords[1][3])

		// Parse data section
		dataReader := csv.NewReader(strings.NewReader(sections[1]))
		dataRecords, err := dataReader.ReadAll()
		require.NoError(t, err)
		assert.Equal(t, []string{"id", "parent", "position", "text", "status", "modified"}, dataRecords[0])
		assert.Equal(t, "123", dataRecords[1][0])
		assert.Equal(t, "Test todo", dataRecords[1][3])
		assert.Equal(t, "pending", dataRecords[1][4])
	})

	t.Run("RenderList", func(t *testing.T) {
		var buf bytes.Buffer
		result := &too.ListResult{
			Todos: []*models.IDMTodo{
				{
					UID:      "1",
					Text:     "First todo",
					Statuses: map[string]string{"completion": string(models.StatusPending)},
				},
				{
					UID:      "2",
					Text:     "Second todo, with comma",
					Statuses: map[string]string{"completion": string(models.StatusDone)},
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
			Todos: []*models.IDMTodo{
				{
					UID:      "parent-uid",
					Text:     "Parent todo",
					Statuses: map[string]string{"completion": string(models.StatusPending)},
					ParentID: "",
				},
				{
					UID:      "child-uid",
					Text:     "Child todo",
					Statuses: map[string]string{"completion": string(models.StatusPending)},
					ParentID: "parent-uid",
				},
				{
					UID:      "grandchild-uid",
					Text:     "Grandchild todo",
					Statuses: map[string]string{"completion": string(models.StatusPending)},
					ParentID: "child-uid",
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

		// With flat structure, verify the todos are present
		var parentFound, childFound, grandchildFound bool
		for i := 1; i < len(todosRecords); i++ {
			text := todosRecords[i][3] // text field
			if text == "Parent todo" {
				parentFound = true
			} else if text == "Child todo" {
				childFound = true
			} else if text == "Grandchild todo" {
				grandchildFound = true
			}
		}
		assert.True(t, parentFound, "Should have parent todo")
		assert.True(t, childFound, "Should have child todo")
		assert.True(t, grandchildFound, "Should have grandchild todo")
	})

	t.Run("Text with newlines", func(t *testing.T) {
		var buf bytes.Buffer
		todo := &models.IDMTodo{
			UID:      "123",
			Text:     "Todo with\nnewline",
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

		// Parse CSV - should have summary + blank line + data
		content := buf.String()
		sections := strings.Split(content, "\n\n")
		require.Len(t, sections, 2)
		
		// Parse data section
		reader := csv.NewReader(strings.NewReader(sections[1]))
		records, err := reader.ReadAll()
		require.NoError(t, err)

		assert.Len(t, records, 2)
		assert.Equal(t, "Todo with\nnewline", records[1][3])
	})

	t.Run("Special characters", func(t *testing.T) {
		var buf bytes.Buffer
		todo := &models.IDMTodo{
			UID:      "123",
			Text:     `Todo with "quotes", commas, and 'apostrophes'`,
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

		// Parse data section
		content := buf.String()
		sections := strings.Split(content, "\n\n")
		require.Len(t, sections, 2)
		
		reader := csv.NewReader(strings.NewReader(sections[1]))
		records, err := reader.ReadAll()
		require.NoError(t, err)

		assert.Len(t, records, 2)
		assert.Equal(t, `Todo with "quotes", commas, and 'apostrophes'`, records[1][3])
	})

	t.Run("RenderChange - Move", func(t *testing.T) {
		var buf bytes.Buffer
		todo := &models.IDMTodo{
			UID:      "123",
			Text:     "Moved todo",
			Statuses: map[string]string{"completion": string(models.StatusPending)},
		}
		result := too.NewChangeResult(
			"moved",
			[]*models.IDMTodo{todo},
			[]*models.IDMTodo{todo},
			1,
			0,
		)

		err := formatter.RenderChange(&buf, result)
		require.NoError(t, err)

		// CSV has two sections separated by blank line
		content := buf.String()
		sections := strings.Split(content, "\n\n")
		require.Len(t, sections, 2)

		// Parse summary section
		summaryReader := csv.NewReader(strings.NewReader(sections[0]))
		summaryRecords, err := summaryReader.ReadAll()
		require.NoError(t, err)
		assert.Equal(t, "moved", summaryRecords[1][0])

		// Parse data section
		dataReader := csv.NewReader(strings.NewReader(sections[1]))
		dataRecords, err := dataReader.ReadAll()
		require.NoError(t, err)
		assert.Equal(t, []string{"id", "parent", "position", "text", "status", "modified"}, dataRecords[0])
		assert.Equal(t, "123", dataRecords[1][0])
		assert.Equal(t, "Moved todo", dataRecords[1][3])
	})
}
