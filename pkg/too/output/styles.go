package output

import (
	"github.com/arthur-debert/too/pkg/lipbalm"
	"github.com/charmbracelet/lipgloss"
)

// Color palette
var (
	// Text colors
	PRIMARY_TEXT = lipgloss.AdaptiveColor{
		Light: "#000000",
		Dark:  "#FFFFFF",
	}

	SUBDUED_TEXT = lipgloss.AdaptiveColor{
		Light: "#495057",
		Dark:  "#ADB5BD",
	}

	MUTED_TEXT = lipgloss.AdaptiveColor{
		Light: "#868E96",
		Dark:  "#6C757D",
	}

	FAINT_TEXT = lipgloss.AdaptiveColor{
		Light: "#868E96",
		Dark:  "#CED4DA",
	}

	VERY_FAINT_TEXT = lipgloss.AdaptiveColor{
		Light: "#D6DADD", // ~20% darker than white (more faint)
		Dark:  "#3A3D42", // ~20% lighter than black (more faint)
	}

	// Status colors
	SUCCESS_COLOR = lipgloss.AdaptiveColor{
		Light: "#2B8A3E",
		Dark:  "#37B24D",
	}

	ERROR_COLOR = lipgloss.AdaptiveColor{
		Light: "#C92A2A",
		Dark:  "#F03E3E",
	}

	WARNING_COLOR = lipgloss.AdaptiveColor{
		Light: "#F59F00",
		Dark:  "#FCC419",
	}

	INFO_COLOR = lipgloss.AdaptiveColor{
		Light: "#1971C2",
		Dark:  "#339AF0",
	}

	// Accent colors
	ACCENT_COLOR = lipgloss.AdaptiveColor{
		Light: "#F59F00",
		Dark:  "#FCC419",
	}
)

// StatusSymbols maps status names to their Unicode symbols
var StatusSymbols = map[string]string{
	"pending": "○", // Hollow circle for pending
	"done":    "●", // Filled circle for done
	"mixed":   "◐", // Half-filled circle for mixed states
	"deleted": "⊘", // Circle with slash for deleted
}

// GetStatusSymbol returns the symbol for a given status, with a fallback
func GetStatusSymbol(status string) string {
	if symbol, ok := StatusSymbols[status]; ok {
		return symbol
	}
	return "○" // Default to pending symbol
}

// Lipgloss styles
var (
	// Base styles
	Base = lipgloss.NewStyle()

	// Status styles
	StatusDone = lipgloss.NewStyle().
			Foreground(SUCCESS_COLOR).
			Bold(true)

	StatusPending = lipgloss.NewStyle().
			Foreground(ERROR_COLOR).
			Bold(true)

	// Symbol styles
	SymbolDone    = StatusDone.SetString("✓")
	SymbolPending = StatusPending.SetString("✕")

	// Position number style
	Position = lipgloss.NewStyle().
			Foreground(SUBDUED_TEXT).
			Align(lipgloss.Right).
			Width(6)

	// Separator style
	Separator = lipgloss.NewStyle().
			Foreground(MUTED_TEXT).
			SetString(" | ")

	// Text styles
	TodoText = lipgloss.NewStyle()

	// Hashtag highlighting
	Hashtag = lipgloss.NewStyle().
		Foreground(ACCENT_COLOR).
		Bold(true)

	// Summary styles
	Summary = lipgloss.NewStyle().
		Foreground(FAINT_TEXT).
		MarginTop(1)

	// Error styles
	Error = lipgloss.NewStyle().
		Foreground(ERROR_COLOR).
		Bold(true)

	// Success styles
	Success = lipgloss.NewStyle().
		Foreground(SUCCESS_COLOR)

	// Info styles
	Info = lipgloss.NewStyle().
		Foreground(INFO_COLOR)
)

// GetLipbalmStyleMap returns the style map for lipbalm template rendering
func GetLipbalmStyleMap() lipbalm.StyleMap {
	// Start with lipbalm's defaults
	styles := lipbalm.DefaultStyles()
	
	// Semantic styles for todos
	// Active (pending) todos - prominent display
	styles["active-todo"] = lipgloss.NewStyle().
		Foreground(PRIMARY_TEXT)
	
	// Completed todos - subdued display
	styles["completed-todo"] = lipgloss.NewStyle().
		Foreground(MUTED_TEXT)
	
	// Highlighted todo (for change feedback)
	styles["highlighted-todo"] = lipgloss.NewStyle().
		Foreground(PRIMARY_TEXT)
	
	// Override the default muted style to use our MUTED_TEXT
	styles["muted"] = lipgloss.NewStyle().
		Foreground(MUTED_TEXT)
	
	// Subdued style for summary text
	styles["subdued"] = lipgloss.NewStyle().
		Foreground(SUBDUED_TEXT)
	
	// Message styles - use same color as highlighted items
	styles["success"] = lipgloss.NewStyle().
		Foreground(PRIMARY_TEXT)
	styles["info"] = lipgloss.NewStyle().
		Foreground(PRIMARY_TEXT)
	styles["warning"] = lipgloss.NewStyle().
		Foreground(PRIMARY_TEXT)
	styles["error"] = lipgloss.NewStyle().
		Foreground(PRIMARY_TEXT)
	
	return styles
}