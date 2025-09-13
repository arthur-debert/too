package output

import (
	"embed"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/arthur-debert/too/pkg/lipbalm"
	"github.com/arthur-debert/too/pkg/too"
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/output/styles"
)

//go:embed templates/*.tmpl
var engineTemplateFS embed.FS

// Engine wraps lipbalm's RenderEngine with too-specific functionality
type Engine struct {
	lipbalmEngine *lipbalm.RenderEngine
}

// NewEngine creates a new output engine
func NewEngine() (*Engine, error) {
	// Get too's style map
	styleMap := styles.GetLipbalmStyleMap()

	// Create lipbalm engine with templates
	lipbalmEngine, err := lipbalm.WithTemplates(engineTemplateFS, "templates", styleMap)
	if err != nil {
		return nil, fmt.Errorf("failed to create lipbalm engine: %w", err)
	}

	// Configure too-specific callbacks
	lipbalmEngine.Config().Callbacks = lipbalm.RenderCallbacks{
		// Pre-process to handle hierarchical todos
		PreProcess: func(format string, data interface{}) interface{} {
			// Special handling for terminal format to build hierarchy
			if format == "term" {
				switch v := data.(type) {
				case *too.ListResult:
					// For terminal, we want hierarchical display
					return &TodoListWithMessage{
						Todos:      v.Todos,
						TotalCount: v.TotalCount,
						DoneCount:  v.DoneCount,
					}
				case *too.SearchResult:
					return &TodoListWithMessage{
						Message:     fmt.Sprintf("Found %d match%s", len(v.MatchedTodos), pluralize(len(v.MatchedTodos))),
						MessageType: messageTypeForCount(len(v.MatchedTodos)),
						Todos:       v.MatchedTodos,
						TotalCount:  v.TotalCount,
					}
				case *too.ChangeResult:
					highlightID := ""
					if len(v.AffectedTodos) > 0 {
						highlightID = v.AffectedTodos[0].UID
					}
					return &TodoListWithMessage{
						Message:     v.Message,
						MessageType: messageTypeForCommand(v.Command, v.AffectedTodos),
						Todos:       v.AllTodos,
						TotalCount:  v.TotalCount,
						DoneCount:   v.DoneCount,
						HighlightID: highlightID,
					}
				}
			}
			return data
		},

		// Custom field renderers
		CustomFields: map[string]lipbalm.FieldRenderer{
			// Custom markdown rendering for todos
			"__markdown__": func(format, fieldName string, value interface{}) (string, bool) {
				if format != "markdown" {
					return "", false
				}

				switch v := value.(type) {
				case *too.ChangeResult:
					return renderChangeAsMarkdown(v), true
				case *too.ListResult:
					return renderListAsMarkdown(v), true
				case *too.SearchResult:
					return renderSearchAsMarkdown(v), true
				case *too.MessageResult:
					return renderMessageAsMarkdown(v), true
				case *too.ListFormatsResult:
					return renderFormatsAsMarkdown(v), true
				}

				return "", false
			},

			// Custom CSV rendering for hierarchical todos
			"[]*models.IDMTodo": func(format, fieldName string, value interface{}) (string, bool) {
				if format != "csv" {
					return "", false
				}
				
				// TODO: Implement custom CSV rendering for hierarchical todos
				return "", false
			},
		},
	}

	return &Engine{
		lipbalmEngine: lipbalmEngine,
	}, nil
}

// Render renders data in the specified format
func (e *Engine) Render(w io.Writer, format string, data interface{}) error {
	return e.lipbalmEngine.Render(w, format, data)
}

// RenderError renders an error
func (e *Engine) RenderError(w io.Writer, format string, err error) error {
	return e.lipbalmEngine.RenderError(w, format, err)
}

// SetFormat sets the default format
func (e *Engine) SetFormat(format string) error {
	return e.lipbalmEngine.SetFormat(format)
}

// GetFormat returns the current default format
func (e *Engine) GetFormat() string {
	return e.lipbalmEngine.GetFormat()
}

// ListFormats returns available formats
func (e *Engine) ListFormats() []string {
	return e.lipbalmEngine.ListFormats()
}

// Helper functions

func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "es"
}

func messageTypeForCount(count int) string {
	if count == 0 {
		return "warning"
	}
	return "info"
}

func messageTypeForCommand(command string, affected []*models.IDMTodo) string {
	switch command {
	case "edit", "modify":
		return "info"
	case "reopen":
		return "warning"
	case "clean":
		if len(affected) == 0 {
			return "warning"
		}
		return "success"
	default:
		return "success"
	}
}

// Markdown rendering functions

func renderChangeAsMarkdown(result *too.ChangeResult) string {
	var sb strings.Builder

	// Show affected todos summary
	if len(result.AffectedTodos) > 0 {
		verb := result.Command
		if !strings.HasSuffix(verb, "ed") {
			verb = verb + "ed"
		}
		sb.WriteString(fmt.Sprintf("%s %d todo(s)\n\n", strings.Title(verb), len(result.AffectedTodos)))
	}

	// Render all todos
	if len(result.AllTodos) > 0 {
		sb.WriteString(renderTodosAsMarkdown(result.AllTodos))
		sb.WriteString(fmt.Sprintf("\n---\n%d todo(s), %d done\n", result.TotalCount, result.DoneCount))
	}

	return sb.String()
}

func renderListAsMarkdown(result *too.ListResult) string {
	if len(result.Todos) == 0 {
		return "No todos\n"
	}

	var sb strings.Builder
	sb.WriteString(renderTodosAsMarkdown(result.Todos))
	
	if result.TotalCount > 0 {
		sb.WriteString(fmt.Sprintf("\n---\n%d todo(s), %d done\n", result.TotalCount, result.DoneCount))
	}

	return sb.String()
}

func renderSearchAsMarkdown(result *too.SearchResult) string {
	if len(result.MatchedTodos) == 0 {
		return "No matching todos found\n"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d matching todo(s):\n\n", len(result.MatchedTodos)))
	sb.WriteString(renderTodosAsMarkdown(result.MatchedTodos))

	return sb.String()
}

func renderMessageAsMarkdown(result *too.MessageResult) string {
	switch result.Level {
	case "error":
		return fmt.Sprintf("**Error:** %s\n", result.Text)
	case "warning":
		return fmt.Sprintf("**Warning:** %s\n", result.Text)
	case "success":
		return fmt.Sprintf("✓ %s\n", result.Text)
	default:
		return fmt.Sprintf("%s\n", result.Text)
	}
}

func renderFormatsAsMarkdown(result *too.ListFormatsResult) string {
	var sb strings.Builder
	sb.WriteString("Available output formats:\n")
	
	for _, format := range result.Formats {
		sb.WriteString(fmt.Sprintf("- **%s**: %s\n", format.Name, format.Description))
	}

	return sb.String()
}

func renderTodosAsMarkdown(todos []*models.IDMTodo) string {
	// Build hierarchical structure
	hierarchical := BuildHierarchy(todos)
	return renderHierarchicalTodosAsMarkdown(hierarchical, 0)
}

func renderHierarchicalTodosAsMarkdown(todos []*HierarchicalTodo, indent int) string {
	var sb strings.Builder
	indentStr := strings.Repeat("   ", indent)

	for i, todo := range todos {
		checkbox := "[ ]"
		if todo.GetStatus() == models.StatusDone {
			checkbox = "[x]"
		}

		// Format multi-line text properly
		text := formatMultilineMarkdown(todo.Text, indentStr)
		sb.WriteString(fmt.Sprintf("%s%d. %s %s\n", indentStr, i+1, checkbox, text))

		// Render nested todos
		if len(todo.Children) > 0 {
			sb.WriteString(renderHierarchicalTodosAsMarkdown(todo.Children, indent+1))
		}
	}

	return sb.String()
}

func formatMultilineMarkdown(text string, indentStr string) string {
	lines := strings.Split(text, "\n")
	if len(lines) <= 1 {
		return text
	}

	continuationIndent := indentStr + "      "
	var result strings.Builder
	
	for i, line := range lines {
		if i == 0 {
			result.WriteString(line)
		} else {
			result.WriteString("\n" + continuationIndent + line)
		}
	}
	
	return result.String()
}

// GetGlobalEngine returns a singleton engine instance
var globalEngine *Engine

func GetGlobalEngine() (*Engine, error) {
	if globalEngine == nil {
		var err error
		globalEngine, err = NewEngine()
		if err != nil {
			return nil, err
		}
	}
	return globalEngine, nil
}

// Quick render functions for convenience

// QuickRender renders data using the global engine
func QuickRender(format string, data interface{}) error {
	engine, err := GetGlobalEngine()
	if err != nil {
		return err
	}
	return engine.Render(os.Stdout, format, data)
}

// QuickRenderError renders an error using the global engine
func QuickRenderError(format string, err error) error {
	engine, renderErr := GetGlobalEngine()
	if renderErr != nil {
		return renderErr
	}
	return engine.RenderError(os.Stdout, format, err)
}