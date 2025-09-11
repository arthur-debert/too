package models

// ReorderTodos is no longer needed as IDM handles all positioning.
// Kept as empty function for backward compatibility during migration.
func ReorderTodos(todos []*Todo) {
	// No-op: IDM handles positioning
}

// ResetActivePositions is no longer needed as IDM handles all positioning.
// Kept as empty function for backward compatibility during migration.
func ResetActivePositions(todos *[]*Todo) {
	// No-op: IDM handles positioning
}
