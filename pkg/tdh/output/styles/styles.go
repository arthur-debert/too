package styles

import (
	"github.com/charmbracelet/lipgloss"
)

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
