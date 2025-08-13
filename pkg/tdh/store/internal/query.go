package internal

// Query defines parameters for finding todos.
// All fields are optional - nil pointer fields are ignored during filtering.
// Multiple criteria are combined with AND logic (all must match).
//
// This struct is designed for easy extensibility - new query parameters
// can be added as needed without breaking existing code.
type Query struct {
	// Status filters todos by their status (e.g., "done", "pending").
	// Nil means no status filtering.
	Status *string

	// TextContains filters todos whose text contains this substring.
	// Nil means no text filtering.
	TextContains *string

	// CaseSensitive controls whether text search is case-sensitive.
	// Defaults to false (case-insensitive search).
	// Note: This enhancement was added beyond the original design to improve usability.
	CaseSensitive bool
}
