package styles

import (
	"github.com/charmbracelet/lipgloss"
)

// Semantic color palette
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
