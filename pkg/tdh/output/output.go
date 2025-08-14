package output

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/arthur-debert/tdh/pkg/tdh"
	"github.com/arthur-debert/tdh/pkg/tdh/models"
	ct "github.com/daviddengcn/go-colortext"
)

var hashtagRegex = regexp.MustCompile(`#[^\s]*`)

// TodoRenderer handles rendering of individual todos
type TodoRenderer struct {
	writer   io.Writer
	useColor bool
}

// NewTodoRenderer creates a new todo renderer
func NewTodoRenderer(w io.Writer, useColor bool) *TodoRenderer {
	if w == nil {
		w = os.Stdout
	}
	return &TodoRenderer{
		writer:   w,
		useColor: useColor,
	}
}

// RenderTodo renders a single todo with formatting
func (r *TodoRenderer) RenderTodo(t *models.Todo) {
	var symbol string
	var color ct.Color

	if t.Status == models.StatusDone {
		color = ct.Green
		symbol = "✓"
	} else {
		color = ct.Red
		symbol = "✕"
	}

	// Right-align the Position with padding
	spaceCount := 6 - len(strconv.Itoa(t.Position))
	_, _ = fmt.Fprint(r.writer, strings.Repeat(" ", spaceCount), t.Position, " | ")

	// Print status symbol with color
	if r.useColor {
		ct.ChangeColor(color, false, ct.None, false)
	}
	_, _ = fmt.Fprint(r.writer, symbol)
	if r.useColor {
		ct.ResetColor()
	}
	_, _ = fmt.Fprint(r.writer, " ")

	// Print text with hashtag highlighting
	r.printWithHashtagHighlight(t.Text)
	_, _ = fmt.Fprintln(r.writer)
}

// printWithHashtagHighlight prints text with hashtags highlighted in yellow.
func (r *TodoRenderer) printWithHashtagHighlight(text string) {
	pos := 0
	for _, match := range hashtagRegex.FindAllStringIndex(text, -1) {
		// Print text before hashtag
		_, _ = fmt.Fprint(r.writer, text[pos:match[0]])

		// Print hashtag with color
		if r.useColor {
			ct.ChangeColor(ct.Yellow, false, ct.None, false)
		}
		_, _ = fmt.Fprint(r.writer, text[match[0]:match[1]])
		if r.useColor {
			ct.ResetColor()
		}

		pos = match[1]
	}
	// Print remaining text
	_, _ = fmt.Fprint(r.writer, text[pos:])
}

// MakeOutput is a compatibility function that renders a todo to stdout
// Deprecated: Use TodoRenderer.RenderTodo instead
func MakeOutput(t *models.Todo, useColor bool) {
	renderer := NewTodoRenderer(os.Stdout, useColor)
	renderer.RenderTodo(t)
}

// Renderer handles output formatting for tdh commands
type Renderer struct {
	writer           io.Writer
	templateRenderer *TemplateRenderer
}

// NewRenderer creates a new renderer
func NewRenderer(w io.Writer) *Renderer {
	if w == nil {
		w = os.Stdout
	}

	// Try to create template renderer
	templateRenderer, _ := NewTemplateRenderer(w, true)

	return &Renderer{
		writer:           w,
		templateRenderer: templateRenderer,
	}
}

// RenderInit renders the init command result
func (r *Renderer) RenderInit(result *tdh.InitResult) error {
	_, err := fmt.Fprintln(r.writer, result.Message)
	return err
}

// RenderAdd renders the add command result
func (r *Renderer) RenderAdd(result *tdh.AddResult) error {
	// Use template renderer if available
	if r.templateRenderer != nil {
		if err := r.templateRenderer.Render("add_result", result); err == nil {
			_, _ = fmt.Fprintln(r.writer)
			return nil
		}
	}

	// Fallback to old rendering
	_, err := fmt.Fprintf(r.writer, "Added todo #%d: %s\n", result.Todo.Position, result.Todo.Text)
	return err
}

// RenderModify renders the modify command result
func (r *Renderer) RenderModify(result *tdh.ModifyResult) error {
	// Use template renderer if available
	if r.templateRenderer != nil {
		if err := r.templateRenderer.Render("modify_result", result); err == nil {
			_, _ = fmt.Fprintln(r.writer)
			return nil
		}
	}

	// Fallback to old rendering
	if _, err := fmt.Fprintf(r.writer, "Modified todo #%d\n", result.Todo.Position); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(r.writer, "  Old: %s\n", result.OldText); err != nil {
		return err
	}
	_, err := fmt.Fprintf(r.writer, "  New: %s\n", result.NewText)
	return err
}

// RenderToggle renders the toggle command result
func (r *Renderer) RenderToggle(result *tdh.ToggleResult) error {
	// Use template renderer if available
	if r.templateRenderer != nil {
		if err := r.templateRenderer.Render("toggle_result", result); err == nil {
			_, _ = fmt.Fprintln(r.writer)
			return nil
		}
	}

	// Fallback to old rendering
	_, err := fmt.Fprintf(r.writer, "Toggled todo #%d from %s to %s\n",
		result.Todo.Position, result.OldStatus, result.NewStatus)
	return err
}

// RenderClean renders the clean command result
func (r *Renderer) RenderClean(result *tdh.CleanResult) error {
	// Use template renderer if available
	if r.templateRenderer != nil {
		if err := r.templateRenderer.Render("clean_result", result); err == nil {
			_, _ = fmt.Fprintln(r.writer)
			return nil
		}
	}

	// Fallback to old rendering
	if result.RemovedCount == 0 {
		if _, err := fmt.Fprintln(r.writer, "No finished todos to clean"); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintf(r.writer, "Removed %d finished todo(s)\n", result.RemovedCount); err != nil {
			return err
		}
		for _, todo := range result.RemovedTodos {
			if _, err := fmt.Fprintf(r.writer, "  - #%d: %s\n", todo.Position, todo.Text); err != nil {
				return err
			}
		}
	}
	_, err := fmt.Fprintf(r.writer, "%d active todo(s) remaining\n", result.ActiveCount)
	return err
}

// RenderReorder renders the reorder command result
func (r *Renderer) RenderReorder(result *tdh.ReorderResult) error {
	// Use template renderer if available
	if r.templateRenderer != nil {
		// Prepare reorder data with pre-rendered todos
		reorderData := map[string]interface{}{
			"ReorderedCount": result.ReorderedCount,
			"Todos":          make([]map[string]interface{}, 0, len(result.Todos)),
		}

		// Pre-render each todo
		for _, todo := range result.Todos {
			todoData := r.templateRenderer.PrepareData(todo)
			if todoMap, ok := todoData.(map[string]interface{}); ok {
				reorderData["Todos"] = append(reorderData["Todos"].([]map[string]interface{}), todoMap)
			}
		}

		if err := r.templateRenderer.Render("reorder_result", reorderData); err == nil {
			_, _ = fmt.Fprintln(r.writer)
			return nil
		}
	}

	// Fallback to old rendering
	if result.ReorderedCount == 0 {
		_, err := fmt.Fprintln(r.writer, "All todos are already in sequential order")
		return err
	}

	_, err := fmt.Fprintf(r.writer, "Reordered %d todo(s) to sequential positions\n", result.ReorderedCount)
	if err != nil {
		return err
	}

	// Optionally show the reordered list
	if len(result.Todos) > 0 {
		_, err = fmt.Fprintln(r.writer, "\nCurrent order:")
		if err != nil {
			return err
		}

		for _, todo := range result.Todos {
			r.renderTodo(todo)
		}
	}

	return nil
}

// RenderSearch renders the search command result
func (r *Renderer) RenderSearch(result *tdh.SearchResult) error {
	// Use template renderer if available
	if r.templateRenderer != nil {
		// Prepare search data with pre-rendered todos
		searchData := map[string]interface{}{
			"Query":        result.Query,
			"MatchedTodos": make([]map[string]interface{}, 0, len(result.MatchedTodos)),
		}

		// Pre-render each todo
		for _, todo := range result.MatchedTodos {
			todoData := r.templateRenderer.PrepareData(todo)
			if todoMap, ok := todoData.(map[string]interface{}); ok {
				searchData["MatchedTodos"] = append(searchData["MatchedTodos"].([]map[string]interface{}), todoMap)
			}
		}

		if err := r.templateRenderer.Render("search_result", searchData); err == nil {
			_, _ = fmt.Fprintln(r.writer)
			return nil
		}
	}

	// Fallback to old rendering
	if len(result.MatchedTodos) == 0 {
		_, err := fmt.Fprintf(r.writer, "No todos found matching '%s'\n", result.Query)
		return err
	}

	if _, err := fmt.Fprintf(r.writer, "Found %d todo(s) matching '%s':\n",
		len(result.MatchedTodos), result.Query); err != nil {
		return err
	}

	for _, todo := range result.MatchedTodos {
		r.renderTodo(todo)
	}
	return nil
}

// RenderList renders the list command result
func (r *Renderer) RenderList(result *tdh.ListResult) error {
	// Use template renderer if available
	if r.templateRenderer != nil {
		// Prepare list data with pre-rendered todos
		listData := map[string]interface{}{
			"Todos":      make([]map[string]interface{}, 0, len(result.Todos)),
			"TotalCount": result.TotalCount,
			"DoneCount":  result.DoneCount,
		}

		// Pre-render each todo
		for _, todo := range result.Todos {
			todoData := r.templateRenderer.PrepareData(todo)
			if todoMap, ok := todoData.(map[string]interface{}); ok {
				listData["Todos"] = append(listData["Todos"].([]map[string]interface{}), todoMap)
			}
		}

		// Try to render with todo_list template
		if err := r.templateRenderer.Render("todo_list", listData); err == nil {
			return nil
		}
		// Fall back to individual rendering if todo_list fails
	}

	// Fallback to old rendering
	if len(result.Todos) == 0 {
		_, err := fmt.Fprintln(r.writer, "No todos found")
		return err
	}

	for _, todo := range result.Todos {
		r.renderTodo(todo)
	}

	_, err := fmt.Fprintf(r.writer, "\n%d todo(s), %d done\n",
		result.TotalCount, result.DoneCount)
	return err
}

// renderTodo renders a single todo (helper method)
func (r *Renderer) renderTodo(todo *models.Todo) {
	renderer := NewTodoRenderer(r.writer, true)
	renderer.RenderTodo(todo)
}

// RenderError renders an error message
func (r *Renderer) RenderError(err error) error {
	_, writeErr := fmt.Fprintf(r.writer, "Error: %s\n", err.Error())
	return writeErr
}

// RenderTemplate renders a template with data
func (r *Renderer) RenderTemplate(tmpl string, data interface{}) error {
	t, err := template.New("output").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	err = t.Execute(r.writer, data)
	if err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}
