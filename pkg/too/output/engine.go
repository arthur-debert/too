package output

import (
	"embed"
	"fmt"
	"text/template"

	"github.com/arthur-debert/too/pkg/lipbalm"
	"github.com/arthur-debert/too/pkg/too"
	"github.com/arthur-debert/too/pkg/too/commands/formats"
	"github.com/arthur-debert/too/pkg/too/models"
)

//go:embed templates/*.tmpl
var engineTemplateFS embed.FS

// templateFuncs returns too-specific template functions
func templateFuncs() template.FuncMap {
	// Start with lipbalm's default functions
	funcs := lipbalm.DefaultTemplateFuncs()
	
	// Add too-specific functions
	funcs["isDone"] = func(todo interface{}) bool {
		switch t := todo.(type) {
		case *models.Todo:
			return t.GetStatus() == models.StatusDone
		case *models.HierarchicalTodo:
			return t.Todo.GetStatus() == models.StatusDone
		default:
			return false
		}
	}
	funcs["getSymbol"] = GetStatusSymbol
	funcs["buildHierarchy"] = models.BuildHierarchy
	
	return funcs
}

// Engine wraps lipbalm's RenderEngine with too-specific functionality
type Engine struct {
	lipbalmEngine *lipbalm.RenderEngine
}

// NewEngine creates a new output engine
func NewEngine() (*Engine, error) {
	// Get too's style map
	styleMap := GetLipbalmStyleMap()

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
				case *formats.Result:
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

