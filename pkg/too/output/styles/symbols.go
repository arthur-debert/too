package styles

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