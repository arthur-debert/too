package internal

// Query defines parameters for finding todos.
// Nil fields are ignored.
type Query struct {
	Status        *string
	TextContains  *string
	CaseSensitive bool
}
