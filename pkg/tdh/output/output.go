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

	// Create template renderer - panic if it fails
	templateRenderer, err := NewTemplateRenderer(w, true)
	if err != nil {
		panic(fmt.Sprintf("Failed to create template renderer: %v", err))
	}

	return &Renderer{
		writer:           w,
		templateRenderer: templateRenderer,
	}
}

// RenderInit renders the init command result
func (r *Renderer) RenderInit(result *tdh.InitResult) error {
	if err := r.templateRenderer.Render("init_result", result); err != nil {
		return err
	}
	_, err := fmt.Fprintln(r.writer)
	return err
}

// RenderAdd renders the add command result
func (r *Renderer) RenderAdd(result *tdh.AddResult) error {
	if err := r.templateRenderer.Render("add_result", result); err != nil {
		return err
	}
	_, err := fmt.Fprintln(r.writer)
	return err
}

// RenderModify renders the modify command result
func (r *Renderer) RenderModify(result *tdh.ModifyResult) error {
	if err := r.templateRenderer.Render("modify_result", result); err != nil {
		return err
	}
	_, err := fmt.Fprintln(r.writer)
	return err
}

// RenderToggle renders the toggle command result
func (r *Renderer) RenderToggle(result *tdh.ToggleResult) error {
	if err := r.templateRenderer.Render("toggle_result", result); err != nil {
		return err
	}
	_, err := fmt.Fprintln(r.writer)
	return err
}

// RenderClean renders the clean command result
func (r *Renderer) RenderClean(result *tdh.CleanResult) error {
	if err := r.templateRenderer.Render("clean_result", result); err != nil {
		return err
	}
	_, err := fmt.Fprintln(r.writer)
	return err
}

// RenderReorder renders the reorder command result
func (r *Renderer) RenderReorder(result *tdh.ReorderResult) error {
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

	if err := r.templateRenderer.Render("reorder_result", reorderData); err != nil {
		return err
	}
	_, err := fmt.Fprintln(r.writer)
	return err
}

// RenderSearch renders the search command result
func (r *Renderer) RenderSearch(result *tdh.SearchResult) error {
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

	if err := r.templateRenderer.Render("search_result", searchData); err != nil {
		return err
	}
	_, err := fmt.Fprintln(r.writer)
	return err
}

// RenderList renders the list command result
func (r *Renderer) RenderList(result *tdh.ListResult) error {
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

	return r.templateRenderer.Render("todo_list", listData)
}

// RenderError renders an error message
func (r *Renderer) RenderError(err error) error {
	if renderErr := r.templateRenderer.Render("error", err.Error()); renderErr != nil {
		return renderErr
	}
	_, writeErr := fmt.Fprintln(r.writer)
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
