package styles

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Base styles
	Base = lipgloss.NewStyle()

	// Status styles
	StatusDone = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Bold(true)

	StatusPending = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)

	// Symbol styles
	SymbolDone    = StatusDone.SetString("✓")
	SymbolPending = StatusPending.SetString("✕")

	// Position number style
	Position = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Align(lipgloss.Right).
			Width(6)

	// Separator style
	Separator = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			SetString(" | ")

	// Text styles
	TodoText = lipgloss.NewStyle()

	// Hashtag highlighting
	Hashtag = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFF00")).
		Bold(true)

	// Summary styles
	Summary = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#CCCCCC")).
		MarginTop(1)

	// Error styles
	Error = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF0000")).
		Bold(true)

	// Success styles
	Success = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00"))

	// Info styles
	Info = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#0088FF"))
)
