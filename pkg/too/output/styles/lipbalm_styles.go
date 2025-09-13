package styles

import (
	"github.com/arthur-debert/too/pkg/lipbalm"
	"github.com/charmbracelet/lipgloss"
)

// GetLipbalmStyleMap returns the style map for lipbalm template rendering
func GetLipbalmStyleMap() lipbalm.StyleMap {
	return lipbalm.StyleMap{
		// Status and result styles
		"success": lipgloss.NewStyle().
			Foreground(SUCCESS_COLOR),
		"error": lipgloss.NewStyle().
			Foreground(ERROR_COLOR).
			Bold(true),
		"warning": lipgloss.NewStyle().
			Foreground(WARNING_COLOR),
		"info": lipgloss.NewStyle().
			Foreground(INFO_COLOR),

		// Todo state styles
		"todo-done": lipgloss.NewStyle().
			Foreground(SUCCESS_COLOR).
			Bold(true),
		"todo-pending": lipgloss.NewStyle().
			Foreground(ERROR_COLOR).
			Bold(true),

		// UI element styles
		"position": lipgloss.NewStyle().
			Foreground(SUBDUED_TEXT),
		"muted": lipgloss.NewStyle().
			Foreground(VERY_FAINT_TEXT).
			Faint(true),
		"highlighted-todo": lipgloss.NewStyle().
			Bold(true),
		"subdued": lipgloss.NewStyle().
			Foreground(SUBDUED_TEXT),
		"accent": lipgloss.NewStyle().
			Foreground(ACCENT_COLOR),
		"count": lipgloss.NewStyle().
			Foreground(INFO_COLOR),
		"label": lipgloss.NewStyle().
			Foreground(SUBDUED_TEXT),
		"value": lipgloss.NewStyle().
			Foreground(PRIMARY_TEXT),
	}
}