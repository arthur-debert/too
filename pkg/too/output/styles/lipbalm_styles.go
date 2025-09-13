package styles

import (
	"github.com/arthur-debert/too/pkg/lipbalm"
	"github.com/charmbracelet/lipgloss"
)

// GetLipbalmStyleMap returns the style map for lipbalm template rendering
func GetLipbalmStyleMap() lipbalm.StyleMap {
	// Start with lipbalm's defaults
	styles := lipbalm.DefaultStyles()
	
	// Add/override too-specific styles
	styles["todo-done"] = lipgloss.NewStyle().
		Foreground(SUCCESS_COLOR).
		Bold(true)
	styles["todo-pending"] = lipgloss.NewStyle().
		Foreground(ERROR_COLOR).
		Bold(true)
	styles["position"] = lipgloss.NewStyle().
		Foreground(SUBDUED_TEXT)
	
	return styles
}