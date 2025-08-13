package models

import (
	"time"
)

// Todo represents a single task in the to-do list.
type Todo struct {
	ID       int64  `json:"id"`
	Text     string `json:"text"`
	Status   string `json:"status"`
	Modified string `json:"modified"`
}

// Collection represents a list of todos.
type Collection struct {
	Todos []*Todo
	Path  string
}

// NewCollection creates a new collection with the given path.
func NewCollection(path string) *Collection {
	return &Collection{
		Path:  path,
		Todos: []*Todo{},
	}
}

// CreateTodo creates a new todo with the given text and adds it to the collection.
func (c *Collection) CreateTodo(text string) *Todo {
	var highestID int64 = 0
	for _, todo := range c.Todos {
		if todo.ID > highestID {
			highestID = todo.ID
		}
	}

	newTodo := &Todo{
		ID:       highestID + 1,
		Text:     text,
		Status:   "pending",
		Modified: time.Now().Local().String(),
	}
	c.Todos = append(c.Todos, newTodo)
	return newTodo
}

// Toggle changes the status of a todo between "pending" and "done".
func (t *Todo) Toggle() {
	if t.Status == "done" {
		t.Status = "pending"
	} else {
		t.Status = "done"
	}
}
