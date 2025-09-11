package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/arthur-debert/too/pkg/too"
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/output"
)

// formatter implements the Formatter interface for CSV output
type formatter struct{}

// New creates a new CSV formatter
func New() output.Formatter {
	return &formatter{}
}

// Name returns the formatter name
func (f *formatter) Name() string {
	return "csv"
}

// Description returns the formatter description
func (f *formatter) Description() string {
	return "CSV output for spreadsheet applications"
}

// writeCSV is a helper that writes CSV data
func (f *formatter) writeCSV(w io.Writer, headers []string, rows [][]string) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write headers
	if err := writer.Write(headers); err != nil {
		return err
	}

	// Write rows
	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return writer.Error()
}

// flattenTodos converts a hierarchical todo structure to flat CSV rows
func (f *formatter) flattenTodos(todos []*models.Todo, parentPath string) [][]string {
	var rows [][]string

	for _, todo := range todos {
		// Build the current path
		currentPath := todo.Text
		if parentPath != "" {
			currentPath = parentPath + " > " + todo.Text
		}

		// Add the current todo as a row
		row := []string{
			todo.ID,
			parentPath, // parent column for hierarchy
			fmt.Sprintf("%d", todo.Position),
			todo.Text,
			string(todo.Status),
			todo.Modified.Format("2006-01-02T15:04:05"),
		}
		rows = append(rows, row)

		// Recursively add child todos
		if len(todo.Items) > 0 {
			childRows := f.flattenTodos(todo.Items, currentPath)
			rows = append(rows, childRows...)
		}
	}

	return rows
}

// RenderAdd renders the add command result as CSV
func (f *formatter) RenderAdd(w io.Writer, result *too.AddResult) error {
	headers := []string{"id", "parent", "position", "text", "status", "modified"}
	rows := f.flattenTodos([]*models.Todo{result.Todo}, "")
	return f.writeCSV(w, headers, rows)
}

// RenderModify renders the modify command result as CSV
func (f *formatter) RenderModify(w io.Writer, result *too.ModifyResult) error {
	headers := []string{"id", "parent", "position", "text", "status", "modified", "old_text"}
	rows := [][]string{{
		result.Todo.ID,
		"",
		fmt.Sprintf("%d", result.Todo.Position),
		result.Todo.Text,
		string(result.Todo.Status),
		result.Todo.Modified.Format("2006-01-02T15:04:05"),
		result.OldText,
	}}
	return f.writeCSV(w, headers, rows)
}

// RenderInit renders the init command result as CSV
func (f *formatter) RenderInit(w io.Writer, result *too.InitResult) error {
	headers := []string{"message", "path", "created"}
	rows := [][]string{{result.Message, result.DBPath, fmt.Sprintf("%t", result.Created)}}
	return f.writeCSV(w, headers, rows)
}

// RenderClean renders the clean command result as CSV
func (f *formatter) RenderClean(w io.Writer, result *too.CleanResult) error {
	headers := []string{"operation", "removed_count", "active_count"}
	rows := [][]string{{
		"clean",
		fmt.Sprintf("%d", result.RemovedCount),
		fmt.Sprintf("%d", result.ActiveCount),
	}}

	// If there are removed todos, add them
	if len(result.RemovedTodos) > 0 {
		if err := f.writeCSV(w, headers, rows); err != nil {
			return err
		}
		// Write a blank line
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
		// Write removed todos
		todoHeaders := []string{"id", "parent", "position", "text", "status", "modified"}
		todoRows := f.flattenTodos(result.RemovedTodos, "")
		return f.writeCSV(w, todoHeaders, todoRows)
	}

	return f.writeCSV(w, headers, rows)
}

// RenderSearch renders the search command result as CSV
func (f *formatter) RenderSearch(w io.Writer, result *too.SearchResult) error {
	headers := []string{"id", "parent", "position", "text", "status", "modified", "matched"}
	var rows [][]string

	for _, todo := range result.MatchedTodos {
		// Determine if this todo matched or if it's included because a child matched
		matched := "direct"
		if !strings.Contains(strings.ToLower(todo.Text), strings.ToLower(result.Query)) {
			matched = "child"
		}

		row := []string{
			todo.ID,
			"",
			fmt.Sprintf("%d", todo.Position),
			todo.Text,
			string(todo.Status),
			todo.Modified.Format("2006-01-02T15:04:05"),
			matched,
		}
		rows = append(rows, row)

		// Add child todos
		if len(todo.Items) > 0 {
			childRows := f.flattenTodos(todo.Items, todo.Text)
			// Update matched status for children
			for i := range childRows {
				childRows[i] = append(childRows[i], "direct")
			}
			rows = append(rows, childRows...)
		}
	}

	return f.writeCSV(w, headers, rows)
}

// RenderList renders the list command result as CSV
func (f *formatter) RenderList(w io.Writer, result *too.ListResult) error {
	// First write summary
	summaryHeaders := []string{"total_count", "done_count", "pending_count"}
	summaryRows := [][]string{{
		fmt.Sprintf("%d", result.TotalCount),
		fmt.Sprintf("%d", result.DoneCount),
		fmt.Sprintf("%d", result.TotalCount-result.DoneCount),
	}}

	if err := f.writeCSV(w, summaryHeaders, summaryRows); err != nil {
		return err
	}

	// Write a blank line
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}

	// Then write todos
	headers := []string{"id", "parent", "position", "text", "status", "modified"}
	rows := f.flattenTodos(result.Todos, "")
	return f.writeCSV(w, headers, rows)
}

// RenderComplete renders the complete command results as CSV
func (f *formatter) RenderComplete(w io.Writer, results []*too.CompleteResult) error {
	headers := []string{"id", "parent", "position", "text", "status", "modified"}
	var rows [][]string

	for _, result := range results {
		todoRows := f.flattenTodos([]*models.Todo{result.Todo}, "")
		rows = append(rows, todoRows...)
	}

	return f.writeCSV(w, headers, rows)
}

// RenderReopen renders the reopen command results as CSV
func (f *formatter) RenderReopen(w io.Writer, results []*too.ReopenResult) error {
	headers := []string{"id", "parent", "position", "text", "status", "modified"}
	var rows [][]string

	for _, result := range results {
		todoRows := f.flattenTodos([]*models.Todo{result.Todo}, "")
		rows = append(rows, todoRows...)
	}

	return f.writeCSV(w, headers, rows)
}

// RenderMove renders the move command result as CSV
func (f *formatter) RenderMove(w io.Writer, result *too.MoveResult) error {
	headers := []string{"id", "text", "old_path", "new_path", "status"}
	rows := [][]string{{
		result.Todo.ID,
		result.Todo.Text,
		result.OldPath,
		result.NewPath,
		string(result.Todo.Status),
	}}
	return f.writeCSV(w, headers, rows)
}

// RenderDataPath renders the datapath command result as CSV
func (f *formatter) RenderDataPath(w io.Writer, result *too.ShowDataPathResult) error {
	headers := []string{"data_path"}
	rows := [][]string{{result.Path}}
	return f.writeCSV(w, headers, rows)
}

// RenderFormats renders the formats command result as CSV
func (f *formatter) RenderFormats(w io.Writer, result *too.ListFormatsResult) error {
	headers := []string{"name", "description"}
	var rows [][]string

	for _, format := range result.Formats {
		rows = append(rows, []string{format.Name, format.Description})
	}

	return f.writeCSV(w, headers, rows)
}

// RenderError renders an error message as CSV
func (f *formatter) RenderError(w io.Writer, err error) error {
	headers := []string{"error"}
	rows := [][]string{{err.Error()}}
	return f.writeCSV(w, headers, rows)
}
