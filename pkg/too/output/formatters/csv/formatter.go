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
func (f *formatter) flattenTodos(todos []*models.IDMTodo, parentPath string) [][]string {
	// Build hierarchical structure
	hierarchical := output.BuildHierarchy(todos)
	return f.flattenHierarchicalTodos(hierarchical, parentPath)
}

// flattenHierarchicalTodos converts hierarchical todos to flat CSV rows
func (f *formatter) flattenHierarchicalTodos(todos []*output.HierarchicalTodo, parentPath string) [][]string {
	var rows [][]string

	for i, todo := range todos {
		// Build the current path
		currentPath := todo.Text
		if parentPath != "" {
			currentPath = parentPath + " > " + todo.Text
		}

		// Use array index + 1 as position
		position := i + 1

		// Add the current todo as a row
		row := []string{
			todo.UID,
			parentPath, // parent column for hierarchy
			fmt.Sprintf("%d", position),
			todo.Text,
			string(todo.GetStatus()),
			todo.Modified.Format("2006-01-02T15:04:05"),
		}
		rows = append(rows, row)

		// Recursively add child todos
		if len(todo.Children) > 0 {
			childRows := f.flattenHierarchicalTodos(todo.Children, currentPath)
			rows = append(rows, childRows...)
		}
	}

	return rows
}

// RenderChange renders any command that changes todos as CSV
func (f *formatter) RenderChange(w io.Writer, result *too.ChangeResult) error {
	// Write summary
	headers := []string{"command", "affected_count", "total_count", "done_count"}
	rows := [][]string{{
		result.Command,
		fmt.Sprintf("%d", len(result.AffectedTodos)),
		fmt.Sprintf("%d", result.TotalCount),
		fmt.Sprintf("%d", result.DoneCount),
	}}
	
	if err := f.writeCSV(w, headers, rows); err != nil {
		return err
	}
	
	// Write a blank line
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	
	// Write all todos
	todoHeaders := []string{"id", "parent", "position", "text", "status", "modified"}
	todoRows := f.flattenTodos(result.AllTodos, "")
	return f.writeCSV(w, todoHeaders, todoRows)
}

// RenderInit renders the init command result as CSV
func (f *formatter) RenderInit(w io.Writer, result *too.InitResult) error {
	headers := []string{"message", "path", "created"}
	rows := [][]string{{result.Message, result.DBPath, fmt.Sprintf("%t", result.Created)}}
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
			todo.UID,
			"",
			"1", // For single todo output, position is always 1
			todo.Text,
			string(todo.GetStatus()),
			todo.Modified.Format("2006-01-02T15:04:05"),
			matched,
		}
		rows = append(rows, row)

		// Note: Child handling moved to flattenTodos which builds hierarchy
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
