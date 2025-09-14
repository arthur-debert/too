package output

import (
	"embed"
	"fmt"
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

	// Create template manager with domain-specific functions
	tm := lipbalm.NewTemplateManager(styleMap, templateFuncs())
	if err := tm.AddTemplatesFromEmbed(engineTemplateFS, "templates"); err != nil {
		return nil, fmt.Errorf("failed to load templates: %w", err)
	}

	// Create lipbalm engine with custom config
	lipbalmEngine := lipbalm.New(&lipbalm.Config{
		AutoDetectTerminal: true,
		Styles:             styleMap,
		TemplateManager:    tm,
	})

	// Configure too-specific callbacks
	lipbalmEngine.Config().Callbacks = lipbalm.RenderCallbacks{
		// Pre-process to handle hierarchical todos
		PreProcess: func(format string, data interface{}) interface{} {
			// Special handling for terminal format to build hierarchy
			if format == "term" {
				switch v := data.(type) {
				case *too.ChangeResult:
					highlightID := ""
					if len(v.AffectedTodos) > 0 {
						highlightID = v.AffectedTodos[0].UID
					}
					return &TodoListWithMessage{
						Message:     v.Message,
						MessageType: v.MessageType(),
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
				case *too.ListFormatsResult:
					return renderFormatsAsMarkdown(v), true
				}

				return "", false
			},

			// Custom CSV rendering for hierarchical todos
			"[]*models.Todo": func(format, fieldName string, value interface{}) (string, bool) {
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

// GetLipbalmEngine returns the underlying lipbalm engine for direct use
func (e *Engine) GetLipbalmEngine() *lipbalm.RenderEngine {
	return e.lipbalmEngine
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


func renderFormatsAsMarkdown(result *too.ListFormatsResult) string {
	var sb strings.Builder
	sb.WriteString("Available output formats:\n")
	
	for _, format := range result.Formats {
		sb.WriteString(fmt.Sprintf("- **%s**: %s\n", format.Name, format.Description))
	}

	return sb.String()
}

func renderTodosAsMarkdown(todos []*models.Todo) string {
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