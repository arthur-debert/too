package output

import (
	"embed"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/arthur-debert/too/pkg/lipbalm"
	"github.com/arthur-debert/too/pkg/too"
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/output/styles"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

// LipbamlRenderer is a renderer that uses lipbalm for styled output
type LipbamlRenderer struct {
	Writer         io.Writer // Exported to allow formatter to update it
	useColor       bool
	templateManager *lipbalm.TemplateManager
}

// NewLipbamlRenderer creates a new lipbalm-based renderer
func NewLipbamlRenderer(w io.Writer, useColor bool) (*LipbamlRenderer, error) {
	if w == nil {
		w = os.Stdout
	}

	// Set up lipgloss renderer with proper color detection
	lipglossRenderer := lipgloss.NewRenderer(w)
	if useColor {
		// Force color output for testing
		lipglossRenderer.SetColorProfile(termenv.TrueColor)
	} else {
		lipglossRenderer.SetColorProfile(termenv.Ascii)
	}
	lipbalm.SetDefaultRenderer(lipglossRenderer)

	// Get the style map and create template manager
	styleMap := styles.GetLipbalmStyleMap()
	tm := lipbalm.NewTemplateManager(styleMap, nil)

	// Load templates from embedded filesystem
	if err := tm.AddTemplatesFromEmbed(templateFS, "templates"); err != nil {
		return nil, fmt.Errorf("failed to load templates: %w", err)
	}

	return &LipbamlRenderer{
		Writer:          w,
		useColor:        useColor,
		templateManager: tm,
	}, nil
}

// renderTodoCommand is a unified method for rendering todo commands with the todo_list template
func (r *LipbamlRenderer) renderTodoCommand(message, messageType string, todos []*models.IDMTodo, totalCount, doneCount int, highlightID string) error {
	wrapped := &TodoListWithMessage{
		Message:     message,
		MessageType: messageType,
		Todos:       todos,
		TotalCount:  totalCount,
		DoneCount:   doneCount,
		HighlightID: highlightID,
	}
	
	// Get template content and render with domain-specific functions
	template, ok := r.templateManager.GetTemplate("todo_list_with_message")
	if !ok {
		return fmt.Errorf("template 'todo_list_with_message' not found")
	}
	
	output, err := lipbalm.Render(template, wrapped, r.templateManager.GetStyles(), templateFuncs())
	if err != nil {
		return fmt.Errorf("failed to render todo command: %w", err)
	}
	_, err = fmt.Fprintln(r.Writer, output)
	return err
}

// RenderChange renders any command that changes todos
func (r *LipbamlRenderer) RenderChange(result *too.ChangeResult) error {
	// Determine message type based on command
	messageType := "success"
	switch result.Command {
	case "edit", "modify":
		messageType = "info"
	case "reopen":
		messageType = "warning"
	case "clean":
		if len(result.AffectedTodos) == 0 {
			messageType = "warning"
		}
	}
	
	// Get first affected todo's UID for highlighting (if any)
	highlightID := ""
	if len(result.AffectedTodos) > 0 {
		highlightID = result.AffectedTodos[0].UID
	}
	
	return r.renderTodoCommand(
		result.Message,
		messageType,
		result.AllTodos,
		result.TotalCount,
		result.DoneCount,
		highlightID,
	)
}


// templateFuncs returns too-specific template functions
func templateFuncs() template.FuncMap {
	// Start with lipbalm's default functions
	funcs := lipbalm.DefaultTemplateFuncs()
	
	// Add too-specific functions
	funcs["isDone"] = func(todo interface{}) bool {
		switch t := todo.(type) {
		case *models.IDMTodo:
			return t.GetStatus() == models.StatusDone
		case *HierarchicalTodo:
			return t.IDMTodo.GetStatus() == models.StatusDone
		default:
			return false
		}
	}
	funcs["getSymbol"] = func(status string) string {
		return styles.GetStatusSymbol(status)
	}
	funcs["buildHierarchy"] = func(todos []*models.IDMTodo) []*HierarchicalTodo {
		return BuildHierarchy(todos)
	}
	
	return funcs
}




// RenderMessage renders a simple message result
func (r *LipbamlRenderer) RenderMessage(result *too.MessageResult) error {
	template, ok := r.templateManager.GetTemplate("message")
	if !ok {
		return fmt.Errorf("template 'message' not found")
	}
	
	output, err := lipbalm.Render(template, result, r.templateManager.GetStyles(), templateFuncs())
	if err != nil {
		return fmt.Errorf("failed to render message: %w", err)
	}
	_, err = fmt.Fprintln(r.Writer, output)
	return err
}


// RenderSearch renders the search command result using lipbalm
func (r *LipbamlRenderer) RenderSearch(result *too.SearchResult) error {
	matchCount := len(result.MatchedTodos)
	message := fmt.Sprintf("Found %d match", matchCount)
	if matchCount != 1 {
		message = fmt.Sprintf("Found %d matches", matchCount)
	}
	if result.Query != "" {
		message += fmt.Sprintf(" for \"%s\"", result.Query)
	}
	
	messageType := "info"
	if matchCount == 0 {
		messageType = "warning"
	}
	
	return r.renderTodoCommand(
		message,
		messageType,
		result.MatchedTodos,
		result.TotalCount,
		0, // Search doesn't track done count separately
		"", // No highlight
	)
}

// RenderList renders the list command result using lipbalm
func (r *LipbamlRenderer) RenderList(result *too.ListResult) error {
	return r.renderTodoCommand(
		"", // No message for basic list
		"",
		result.Todos,
		result.TotalCount,
		result.DoneCount,
		"", // No highlight
	)
}





// RenderFormats renders the formats command result using lipbalm
func (r *LipbamlRenderer) RenderFormats(result *too.ListFormatsResult) error {
	template, ok := r.templateManager.GetTemplate("formats_result")
	if !ok {
		return fmt.Errorf("template 'formats_result' not found")
	}
	
	output, err := lipbalm.Render(template, result, r.templateManager.GetStyles(), templateFuncs())
	if err != nil {
		return fmt.Errorf("failed to render formats result: %w", err)
	}
	_, err = fmt.Fprintln(r.Writer, output)
	return err
}

// RenderError renders an error message
func (r *LipbamlRenderer) RenderError(err error) error {
	message := lipbalm.NewErrorMessage("Error: " + err.Error())
	
	template, ok := r.templateManager.GetTemplate("message")
	if !ok {
		return fmt.Errorf("template 'message' not found")
	}
	
	output, renderErr := lipbalm.Render(template, message, r.templateManager.GetStyles(), templateFuncs())
	if renderErr != nil {
		return renderErr
	}
	_, writeErr := fmt.Fprintln(r.Writer, output)
	return writeErr
}
