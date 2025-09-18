package models

import (
	"time"

	"github.com/arthur-debert/nanostore/nanostore"
)

// TodoStatus represents the status of a todo item
type TodoStatus string

const (
	// StatusPending indicates the todo is not yet completed
	StatusPending TodoStatus = "pending"
	// StatusDone indicates the todo has been completed
	StatusDone TodoStatus = "done"
)

// TodoDeclarative represents a todo item using nanostore's declarative API
type TodoDeclarative struct {
	nanostore.Document
	
	// Status dimension with enum values and prefix for completed todos
	Status   string `values:"pending,completed" prefix:"completed=c" default:"pending"`
	// Parent relationship for hierarchical todos
	ParentID string `dimension:"parent_uuid,ref"`
	
	// Non-dimension fields - stored as custom data
	Text        string
	Description string // Optional extended description
	Modified    time.Time
}

// Todo represents a todo item with nanostore backing (legacy model for compatibility)
type Todo struct {
	UID          string            `json:"uid"`       // Stable unique identifier
	ParentID     string            `json:"parentId"`  // Parent UID, empty for root items
	Text         string            `json:"text"`      // Todo content
	PositionPath string            `json:"-"`         // User-facing ID like "1", "1.2", "c1"
	Statuses     map[string]string `json:"statuses"`  // Status dimensions
	Modified     time.Time         `json:"modified"`  // Last modification timestamp
}

// GetStatus returns the todo's completion status (declarative model)
func (t *TodoDeclarative) GetStatus() TodoStatus {
	switch t.Status {
	case "completed":
		return StatusDone
	case "pending":
		return StatusPending
	default:
		return StatusPending
	}
}

// Complete marks the todo as completed
func (t *TodoDeclarative) Complete() {
	t.Status = "completed"
	t.Modified = time.Now()
}

// Reopen marks the todo as pending
func (t *TodoDeclarative) Reopen() {
	t.Status = "pending"
	t.Modified = time.Now()
}

// UpdateText updates the todo text
func (t *TodoDeclarative) UpdateText(text string) {
	t.Text = text
	t.Modified = time.Now()
}

// IsCompleted returns true if the todo is completed
func (t *TodoDeclarative) IsCompleted() bool {
	return t.Status == "completed"
}

// ToLegacy converts declarative model to legacy Todo model for backward compatibility
func (t *TodoDeclarative) ToLegacy() *Todo {
	status := "pending"
	if t.Status == "completed" {
		status = "done"
	}
	
	return &Todo{
		UID:          t.UUID,
		ParentID:     t.ParentID,
		Text:         t.Text,
		PositionPath: t.SimpleID,
		Statuses: map[string]string{
			"completion": status,
		},
		Modified: t.Modified,
	}
}

// FromLegacy creates a declarative model from legacy Todo
func FromLegacy(legacy *Todo) *TodoDeclarative {
	status := "pending"
	if legacy.GetStatus() == StatusDone {
		status = "completed"
	}
	
	return &TodoDeclarative{
		Document: nanostore.Document{
			UUID:     legacy.UID,
			SimpleID: legacy.PositionPath,
			Title:    legacy.Text,
		},
		Status:   status,
		ParentID: legacy.ParentID,
		Text:     legacy.Text,
		Modified: legacy.Modified,
	}
}

// GetStatus returns the todo's completion status (legacy model)
func (t *Todo) GetStatus() TodoStatus {
	if t.Statuses == nil {
		return StatusPending
	}
	if status, exists := t.Statuses["completion"]; exists {
		return TodoStatus(status)
	}
	return StatusPending
}

