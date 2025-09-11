package models

import (
	"time"

	"github.com/google/uuid"
)

// IDMTodo represents a todo item in the pure IDM data model.
// This eliminates the hierarchical Items field in favor of IDM-managed parent-child relationships.
// The IDM Registry maintains all positional and hierarchical information separately.
type IDMTodo struct {
	UID      string            `json:"uid"`                // Stable unique identifier (same as old ID)
	ParentID string            `json:"parentId,omitempty"` // Parent UID, empty for root items
	Text     string            `json:"text"`               // Todo content
	Statuses map[string]string `json:"statuses,omitempty"` // Multi-dimensional workflow statuses
	Modified time.Time         `json:"modified"`           // Last modification timestamp
	// NO Items field - hierarchy is managed by IDM Registry scopes
}

// IDMCollection represents a flat collection of IDM todos.
// Parent-child relationships are managed by the IDM Registry, not embedded in the data structure.
type IDMCollection struct {
	Items []*IDMTodo `json:"items"` // Flat list of all todos
}

// NewIDMTodo creates a new IDM todo with the given text and parent ID.
func NewIDMTodo(text string, parentID string) *IDMTodo {
	return &IDMTodo{
		UID:      uuid.New().String(),
		ParentID: parentID,
		Text:     text,
		Statuses: map[string]string{"completion": string(StatusPending)},
		Modified: time.Now(),
	}
}

// NewIDMCollection creates a new empty IDM collection.
func NewIDMCollection() *IDMCollection {
	return &IDMCollection{
		Items: []*IDMTodo{},
	}
}

// FindByUID finds a todo by its UID in the flat collection.
func (c *IDMCollection) FindByUID(uid string) *IDMTodo {
	for _, item := range c.Items {
		if item.UID == uid {
			return item
		}
	}
	return nil
}

// AddItem adds a new todo to the collection.
func (c *IDMCollection) AddItem(item *IDMTodo) {
	c.Items = append(c.Items, item)
}

// RemoveItem removes a todo by UID from the collection.
func (c *IDMCollection) RemoveItem(uid string) bool {
	for i, item := range c.Items {
		if item.UID == uid {
			c.Items = append(c.Items[:i], c.Items[i+1:]...)
			return true
		}
	}
	return false
}

// GetChildren returns all direct children of the given parent UID.
// This method operates on the flat structure to support IDM operations.
func (c *IDMCollection) GetChildren(parentUID string) []*IDMTodo {
	var children []*IDMTodo
	for _, item := range c.Items {
		if item.ParentID == parentUID {
			children = append(children, item)
		}
	}
	return children
}

// GetDescendants returns all descendants (children, grandchildren, etc.) of the given parent UID.
func (c *IDMCollection) GetDescendants(parentUID string) []*IDMTodo {
	var descendants []*IDMTodo
	
	// Get direct children
	children := c.GetChildren(parentUID)
	for _, child := range children {
		descendants = append(descendants, child)
		// Recursively get grandchildren
		descendants = append(descendants, c.GetDescendants(child.UID)...)
	}
	
	return descendants
}

// Clone creates a deep copy of the IDM todo.
func (t *IDMTodo) Clone() *IDMTodo {
	clone := &IDMTodo{
		UID:      t.UID,
		ParentID: t.ParentID,
		Text:     t.Text,
		Statuses: make(map[string]string),
		Modified: t.Modified,
	}
	
	// Clone statuses map
	for k, v := range t.Statuses {
		clone.Statuses[k] = v
	}
	
	return clone
}

// Clone creates a deep copy of the IDM collection.
func (c *IDMCollection) Clone() *IDMCollection {
	clone := &IDMCollection{
		Items: make([]*IDMTodo, len(c.Items)),
	}
	for i, item := range c.Items {
		clone.Items[i] = item.Clone()
	}
	return clone
}

// EnsureStatuses initializes the Statuses map if it's nil.
func (t *IDMTodo) EnsureStatuses() {
	if t.Statuses == nil {
		t.Statuses = make(map[string]string)
		// Default to pending status
		t.Statuses["completion"] = string(StatusPending)
	}
}

// GetWorkflowStatus gets a status dimension value.
func (t *IDMTodo) GetWorkflowStatus(dimension string) (string, bool) {
	t.EnsureStatuses()
	
	if value, exists := t.Statuses[dimension]; exists {
		return value, true
	}
	
	return "", false
}

// SetModified updates the modified timestamp to the current time.
func (t *IDMTodo) SetModified() {
	t.Modified = time.Now()
}

// GetStatus returns the todo's completion status from the workflow statuses.
func (t *IDMTodo) GetStatus() TodoStatus {
	t.EnsureStatuses()
	if status, exists := t.Statuses["completion"]; exists {
		return TodoStatus(status)
	}
	// Default to pending if no status set
	return StatusPending
}

// IsComplete returns true if the todo is marked as done.
func (t *IDMTodo) IsComplete() bool {
	return t.GetStatus() == StatusDone
}

// IsPending returns true if the todo is marked as pending.
func (t *IDMTodo) IsPending() bool {
	return t.GetStatus() == StatusPending
}

// GetShortID returns the first 7 characters of the todo's UID.
func (t *IDMTodo) GetShortID() string {
	if len(t.UID) >= 7 {
		return t.UID[:7]
	}
	return t.UID
}

// AllItems returns all todos in the collection as a slice (already flat).
func (c *IDMCollection) AllItems() []*IDMTodo {
	return c.Items
}

// Count returns the total number of items in the collection.
func (c *IDMCollection) Count() int {
	return len(c.Items)
}