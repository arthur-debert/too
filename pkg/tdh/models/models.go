package models

import (
	"time"

	"github.com/google/uuid"
)

// TodoStatus represents the status of a todo item
type TodoStatus string

const (
	// StatusPending indicates the todo is not yet completed
	StatusPending TodoStatus = "pending"
	// StatusDone indicates the todo has been completed
	StatusDone TodoStatus = "done"
)

// Todo represents a single task in the to-do list.
type Todo struct {
	ID       string     `json:"id"`       // UUID for stable internal reference
	Position int        `json:"position"` // Sequential position for user interaction
	Text     string     `json:"text"`
	Status   TodoStatus `json:"status"`
	Modified time.Time  `json:"modified"`
}

// Collection represents a list of todos.
type Collection struct {
	Todos []*Todo `json:"todos"`
}

// NewCollection creates a new collection.
func NewCollection() *Collection {
	return &Collection{
		Todos: []*Todo{},
	}
}

// CreateTodo creates a new todo with the given text and adds it to the collection.
func (c *Collection) CreateTodo(text string) *Todo {
	// Find the highest position
	var highestPosition = 0
	for _, todo := range c.Todos {
		if todo.Position > highestPosition {
			highestPosition = todo.Position
		}
	}

	newTodo := &Todo{
		ID:       uuid.New().String(),
		Position: highestPosition + 1,
		Text:     text,
		Status:   StatusPending,
		Modified: time.Now(),
	}
	c.Todos = append(c.Todos, newTodo)
	return newTodo
}

// Toggle changes the status of a todo between pending and done.
func (t *Todo) Toggle() {
	if t.Status == StatusDone {
		t.Status = StatusPending
	} else {
		t.Status = StatusDone
	}
	t.Modified = time.Now()
}

// Clone creates a deep copy of the todo.
func (t *Todo) Clone() *Todo {
	return &Todo{
		ID:       t.ID,
		Position: t.Position,
		Text:     t.Text,
		Status:   t.Status,
		Modified: t.Modified,
	}
}

// Clone creates a deep copy of the collection.
func (c *Collection) Clone() *Collection {
	clone := &Collection{
		Todos: make([]*Todo, len(c.Todos)),
	}
	for i, todo := range c.Todos {
		clone.Todos[i] = todo.Clone()
	}
	return clone
}
