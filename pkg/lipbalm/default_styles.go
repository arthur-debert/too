package lipbalm

import (
	"github.com/charmbracelet/lipgloss"
)

// Common colors for CLI applications
var (
	ColorSuccess = lipgloss.Color("#22c55e") // Green
	ColorError   = lipgloss.Color("#ef4444") // Red
	ColorWarning = lipgloss.Color("#f59e0b") // Amber
	ColorInfo    = lipgloss.Color("#3b82f6") // Blue
	ColorMuted   = lipgloss.Color("#64748b") // Slate
	ColorAccent  = lipgloss.Color("#8b5cf6") // Purple
)

// DefaultStyles returns a set of common styles for CLI applications
func DefaultStyles() StyleMap {
	return StyleMap{
		// Semantic message styles
		"success": lipgloss.NewStyle().
			Foreground(ColorSuccess),
		"error": lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true),
		"warning": lipgloss.NewStyle().
			Foreground(ColorWarning),
		"info": lipgloss.NewStyle().
			Foreground(ColorInfo),

		// Common UI element styles
		"muted": lipgloss.NewStyle().
			Foreground(ColorMuted).
			Faint(true),
		"highlighted": lipgloss.NewStyle().
			Bold(true),
		"subdued": lipgloss.NewStyle().
			Foreground(ColorMuted),
		"accent": lipgloss.NewStyle().
			Foreground(ColorAccent),
		"count": lipgloss.NewStyle().
			Foreground(ColorInfo),
		"label": lipgloss.NewStyle().
			Foreground(ColorMuted),
		"value": lipgloss.NewStyle(),
		
		// Layout helpers
		"indent-1": lipgloss.NewStyle().PaddingLeft(2),
		"indent-2": lipgloss.NewStyle().PaddingLeft(4),
		"indent-3": lipgloss.NewStyle().PaddingLeft(6),
	}
}

// MergeStyles combines multiple style maps, with later maps overriding earlier ones
func MergeStyles(maps ...StyleMap) StyleMap {
	result := make(StyleMap)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}