package output

import (
	"github.com/arthur-debert/too/pkg/too/commands/formats"
	"github.com/arthur-debert/too/pkg/too/formatter"
)

func init() {
	// Set the function to retrieve formatter info to avoid import cycles
	formats.GetFormatterInfoFunc = func() []*formatter.Info {
		engine, err := GetGlobalEngine()
		if err != nil {
			return []*formatter.Info{}
		}
		
		// Convert format names to Info structs
		var infos []*formatter.Info
		for _, format := range engine.ListFormats() {
			info := &formatter.Info{
				Name:        format,
				Description: getFormatDescription(format),
			}
			infos = append(infos, info)
		}
		
		return infos
	}
}

func getFormatDescription(format string) string {
	switch format {
	case "json":
		return "JSON output for programmatic consumption"
	case "yaml":
		return "YAML output for programmatic consumption"
	case "csv":
		return "CSV output for spreadsheet applications"
	case "markdown":
		return "Markdown output for documentation and notes"
	case "term":
		return "Rich terminal output with colors and formatting (default)"
	case "plain":
		return "Plain text output without formatting"
	default:
		return format + " output format"
	}
}