package styles

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Base styles
	Base = lipgloss.NewStyle()

	// Status styles
	StatusDone = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#2B8A3E", Dark: "#37B24D"}).
			Bold(true)

	StatusPending = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#C92A2A", Dark: "#F03E3E"}).
			Bold(true)

	// Symbol styles
	SymbolDone    = StatusDone.SetString("✓")
	SymbolPending = StatusPending.SetString("✕")

	// Position number style
	Position = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#495057", Dark: "#ADB5BD"}).
			Align(lipgloss.Right).
			Width(6)

	// Separator style
	Separator = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#868E96", Dark: "#6C757D"}).
			SetString(" | ")

	// Text styles
	TodoText = lipgloss.NewStyle()

	// Hashtag highlighting
	Hashtag = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#F59F00", Dark: "#FCC419"}).
		Bold(true)

	// Summary styles
	Summary = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#868E96", Dark: "#CED4DA"}).
		MarginTop(1)

	// Error styles
	Error = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#C92A2A", Dark: "#F03E3E"}).
		Bold(true)

	// Success styles
	Success = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#2B8A3E", Dark: "#37B24D"})

	// Info styles
	Info = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#1971C2", Dark: "#339AF0"})
)
