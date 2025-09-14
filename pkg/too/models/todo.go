package models

import "time"

// TodoStatus represents the status of a todo item
type TodoStatus string

const (
	// StatusPending indicates the todo is not yet completed
	StatusPending TodoStatus = "pending"
	// StatusDone indicates the todo has been completed
	StatusDone TodoStatus = "done"
)

// Todo represents a todo item with nanostore backing
type Todo struct {
	UID          string            `json:"uid"`       // Stable unique identifier
	ParentID     string            `json:"parentId"`  // Parent UID, empty for root items
	Text         string            `json:"text"`      // Todo content
	PositionPath string            `json:"-"`         // User-facing ID like "1", "1.2", "c1"
	Statuses     map[string]string `json:"statuses"`  // Status dimensions
	Modified     time.Time         `json:"modified"`  // Last modification timestamp
}

// GetStatus returns the todo's completion status
func (t *Todo) GetStatus() TodoStatus {
	if t.Statuses == nil {
		return StatusPending
	}
	if status, exists := t.Statuses["completion"]; exists {
		return TodoStatus(status)
	}
	return StatusPending
}

