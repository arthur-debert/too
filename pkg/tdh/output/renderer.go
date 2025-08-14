package output

import (
	"fmt"
	"io"
	"os"

	"github.com/arthur-debert/tdh/pkg/lipbaml"
	"github.com/arthur-debert/tdh/pkg/tdh"
	"github.com/arthur-debert/tdh/pkg/tdh/models"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// LipbamlRenderer is a renderer that uses lipbaml for styled output
type LipbamlRenderer struct {
	writer   io.Writer
	useColor bool
	styles   lipbaml.StyleMap
}

// NewLipbamlRenderer creates a new lipbaml-based renderer
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
	lipbaml.SetDefaultRenderer(lipglossRenderer)

	// Define the style map using colors similar to the template functions
	styles := lipbaml.StyleMap{
		// Basic colors matching template functions
		"green": lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#2B8A3E", Dark: "#37B24D"}),
		"red": lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#C92A2A", Dark: "#F03E3E"}),
		"gray": lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#495057", Dark: "#ADB5BD"}),
		"yellow": lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#F59F00", Dark: "#FCC419"}),
		"cyan": lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#1971C2", Dark: "#339AF0"}),

		// Status-specific styles
		"done": lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#2B8A3E", Dark: "#37B24D"}).
			Bold(true),
		"pending": lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#C92A2A", Dark: "#F03E3E"}).
			Bold(true),

		// Component styles
		"position": lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#495057", Dark: "#ADB5BD"}),
		"success": lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#2B8A3E", Dark: "#37B24D"}),
		"error": lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#C92A2A", Dark: "#F03E3E"}).
			Bold(true),
		"info": lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#1971C2", Dark: "#339AF0"}),
	}

	return &LipbamlRenderer{
		writer:   w,
		useColor: useColor,
		styles:   styles,
	}, nil
}

// RenderAdd renders the add command result using lipbaml
func (r *LipbamlRenderer) RenderAdd(result *tdh.AddResult) error {
	// Template matching add_result.tmpl:
	// {{green "Added todo"}} {{gray (printf "#%d" .Todo.Position)}}: {{.Todo.Text}}
	template := `<success>Added todo</success> <position>#{{.Todo.Position}}</position>: {{.Todo.Text}}`

	output, err := lipbaml.Render(template, result, r.styles)
	if err != nil {
		return fmt.Errorf("failed to render add result: %w", err)
	}

	_, err = fmt.Fprintln(r.writer, output)
	return err
}

// RenderToggle renders the toggle command result using lipbaml
func (r *LipbamlRenderer) RenderToggle(result *tdh.ToggleResult) error {
	// Template matching toggle_result.tmpl:
	// {{cyan "Toggled todo"}} {{gray (printf "#%d" .Todo.Position)}} from {{red .OldStatus}} to {{green .NewStatus}}
	template := `<cyan>Toggled todo</cyan> <gray>#{{.Todo.Position}}</gray> from <red>{{.OldStatus}}</red> to <green>{{.NewStatus}}</green>`

	output, err := lipbaml.Render(template, result, r.styles)
	if err != nil {
		return fmt.Errorf("failed to render toggle result: %w", err)
	}

	_, err = fmt.Fprintln(r.writer, output)
	return err
}

// RenderModify renders the modify command result using lipbaml
func (r *LipbamlRenderer) RenderModify(result *tdh.ModifyResult) error {
	// Template matching modify_result.tmpl
	template := `<info>Modified todo</info> <position>#{{.Todo.Position}}</position>: {{.Todo.Text}}`

	output, err := lipbaml.Render(template, result, r.styles)
	if err != nil {
		return fmt.Errorf("failed to render modify result: %w", err)
	}

	_, err = fmt.Fprintln(r.writer, output)
	return err
}

// RenderInit renders the init command result using lipbaml
func (r *LipbamlRenderer) RenderInit(result *tdh.InitResult) error {
	template := `{{.Message}}`

	output, err := lipbaml.Render(template, result, r.styles)
	if err != nil {
		return fmt.Errorf("failed to render init result: %w", err)
	}

	_, err = fmt.Fprintln(r.writer, output)
	return err
}

// renderTodoItem renders a single todo item (helper function)
func (r *LipbamlRenderer) renderTodoItem(todo *models.Todo) (string, error) {
	// Template matching todo_item.tmpl:
	// {{gray (padPosition .Position)}} | {{if isDone .}}{{green "✓"}}{{else}}{{red "✕"}}{{end}} {{.Text}}
	var template string
	if todo.Status == models.StatusDone {
		template = fmt.Sprintf(`<gray>%6d</gray> | <done>✓</done> {{.Text}}`, todo.Position)
	} else {
		template = fmt.Sprintf(`<gray>%6d</gray> | <pending>✕</pending> {{.Text}}`, todo.Position)
	}

	return lipbaml.Render(template, todo, r.styles)
}

// RenderClean renders the clean command result using lipbaml
func (r *LipbamlRenderer) RenderClean(result *tdh.CleanResult) error {
	if result.RemovedCount == 0 {
		template := `<yellow>No finished todos to clean</yellow>`
		output, err := lipbaml.Render(template, nil, r.styles)
		if err != nil {
			return fmt.Errorf("failed to render clean result: %w", err)
		}
		_, _ = fmt.Fprintln(r.writer, output)
	} else {
		template := `<green>Removed {{.RemovedCount}} finished todo(s)</green>`
		output, err := lipbaml.Render(template, result, r.styles)
		if err != nil {
			return fmt.Errorf("failed to render clean header: %w", err)
		}
		_, _ = fmt.Fprintln(r.writer, output)

		// Show removed todos
		for _, todo := range result.RemovedTodos {
			itemTemplate := fmt.Sprintf(`  - <gray>#%d</gray>: {{.Text}}`, todo.Position)
			output, err := lipbaml.Render(itemTemplate, todo, r.styles)
			if err != nil {
				return fmt.Errorf("failed to render removed todo: %w", err)
			}
			_, _ = fmt.Fprintln(r.writer, output)
		}
	}

	// Show remaining active count
	summaryTemplate := `<cyan>{{.ActiveCount}} active todo(s) remaining</cyan>`
	summary, err := lipbaml.Render(summaryTemplate, result, r.styles)
	if err != nil {
		return fmt.Errorf("failed to render clean summary: %w", err)
	}
	_, err = fmt.Fprintln(r.writer, summary)
	return err
}

// RenderReorder renders the reorder command result using lipbaml
func (r *LipbamlRenderer) RenderReorder(result *tdh.ReorderResult) error {
	if result.ReorderedCount == 0 {
		template := `<yellow>All todos are already in sequential order</yellow>`
		output, err := lipbaml.Render(template, nil, r.styles)
		if err != nil {
			return fmt.Errorf("failed to render reorder result: %w", err)
		}
		_, err = fmt.Fprintln(r.writer, output)
		return err
	}

	// Show reorder count
	template := `<green>Reordered {{.ReorderedCount}} todo(s) to sequential positions</green>`
	output, err := lipbaml.Render(template, result, r.styles)
	if err != nil {
		return fmt.Errorf("failed to render reorder header: %w", err)
	}
	_, _ = fmt.Fprintln(r.writer, output)

	// Show current order if todos are provided
	if len(result.Todos) > 0 {
		_, _ = fmt.Fprintln(r.writer)
		header := `<cyan>Current order:</cyan>`
		output, err := lipbaml.Render(header, nil, r.styles)
		if err != nil {
			return fmt.Errorf("failed to render reorder list header: %w", err)
		}
		_, _ = fmt.Fprintln(r.writer, output)

		for _, todo := range result.Todos {
			output, err := r.renderTodoItem(todo)
			if err != nil {
				return fmt.Errorf("failed to render todo item: %w", err)
			}
			_, _ = fmt.Fprintln(r.writer, output)
		}
	}

	return nil
}

// RenderSearch renders the search command result using lipbaml
func (r *LipbamlRenderer) RenderSearch(result *tdh.SearchResult) error {
	if len(result.MatchedTodos) == 0 {
		template := `<yellow>No todos found matching '{{.Query}}'</yellow>`
		output, err := lipbaml.Render(template, result, r.styles)
		if err != nil {
			return fmt.Errorf("failed to render search result: %w", err)
		}
		_, err = fmt.Fprintln(r.writer, output)
		return err
	}

	// Show match count and query
	headerTemplate := `<cyan>Found {{len .MatchedTodos}} todo(s) matching '{{.Query}}':</cyan>`
	header, err := lipbaml.Render(headerTemplate, result, r.styles)
	if err != nil {
		return fmt.Errorf("failed to render search header: %w", err)
	}
	_, _ = fmt.Fprintln(r.writer, header)

	// Show matched todos
	for _, todo := range result.MatchedTodos {
		output, err := r.renderTodoItem(todo)
		if err != nil {
			return fmt.Errorf("failed to render todo item: %w", err)
		}
		_, _ = fmt.Fprintln(r.writer, output)
	}

	return nil
}

// RenderList renders the list command result using lipbaml
func (r *LipbamlRenderer) RenderList(result *tdh.ListResult) error {
	if len(result.Todos) == 0 {
		template := `<yellow>No todos found</yellow>`
		output, err := lipbaml.Render(template, nil, r.styles)
		if err != nil {
			return fmt.Errorf("failed to render empty list: %w", err)
		}
		_, err = fmt.Fprintln(r.writer, output)
		return err
	}

	// Render each todo item
	for _, todo := range result.Todos {
		output, err := r.renderTodoItem(todo)
		if err != nil {
			return fmt.Errorf("failed to render todo item: %w", err)
		}
		_, _ = fmt.Fprintln(r.writer, output)
	}

	// Render summary
	summaryTemplate := `<cyan>{{.TotalCount}} todo(s), {{.DoneCount}} done</cyan>`
	summary, err := lipbaml.Render(summaryTemplate, result, r.styles)
	if err != nil {
		return fmt.Errorf("failed to render summary: %w", err)
	}
	_, err = fmt.Fprintln(r.writer, summary)
	return err
}

// RenderError renders an error message
func (r *LipbamlRenderer) RenderError(err error) error {
	template := `<error>Error:</error> {{.}}`

	output, renderErr := lipbaml.Render(template, err.Error(), r.styles)
	if renderErr != nil {
		return renderErr
	}

	_, writeErr := fmt.Fprintln(r.writer, output)
	return writeErr
}
