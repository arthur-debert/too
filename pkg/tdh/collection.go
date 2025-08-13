package tdh

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"time"
)

type Collection struct {
	Todos []*Todo
	Path  string
}

// NewCollection creates a new collection with the given path
func NewCollection(path string) *Collection {
	return &Collection{
		Path:  path,
		Todos: []*Todo{},
	}
}

// CreateStoreFileIfNeeded creates the store file if it doesn't exist
func CreateStoreFileIfNeeded(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		w, err := os.Create(path)
		if err != nil {
			return false, err
		}
		defer func() { _ = w.Close() }()
		_, err = w.WriteString("[]")
		return true, err
	}

	if err != nil {
		return false, err
	}

	// File exists
	return false, nil
}

func (c *Collection) RemoveAtIndex(item int) {
	s := *c
	s.Todos = append(s.Todos[:item], s.Todos[item+1:]...)
	*c = s
}

// Load loads todos from the collection's path
func (c *Collection) Load() error {
	file, err := os.OpenFile(c.Path, os.O_RDONLY, 0600)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, start with empty collection
			c.Todos = []*Todo{}
			return nil
		}
		return err
	}

	defer func() { _ = file.Close() }()

	err = json.NewDecoder(file).Decode(&c.Todos)
	if err != nil {
		return err
	}

	// Ensure non-nil slice
	if c.Todos == nil {
		c.Todos = []*Todo{}
	}
	return nil
}

// Save writes todos to the collection's path
func (c *Collection) Save() error {
	file, err := os.OpenFile(c.Path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}

	defer func() { _ = file.Close() }()

	data, err := json.MarshalIndent(&c.Todos, "", "  ")
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	return err
}

func (c *Collection) ListPendingTodos() {
	for i := len(c.Todos) - 1; i >= 0; i-- {
		if c.Todos[i].Status != "pending" {
			c.RemoveAtIndex(i)
		}
	}
}

func (c *Collection) ListDoneTodos() {
	for i := len(c.Todos) - 1; i >= 0; i-- {
		if c.Todos[i].Status != "done" {
			c.RemoveAtIndex(i)
		}
	}
}

// CreateTodo creates a new todo with the given text
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

// Find finds a todo by ID
func (c *Collection) Find(id int) (foundedTodo *Todo, err error) {
	id64 := int64(id)
	founded := false
	for _, todo := range c.Todos {
		if id64 == todo.ID {
			foundedTodo = todo
			founded = true
		}
	}
	if !founded {
		err = errors.New("The todo with the id " + strconv.FormatInt(id64, 10) + " was not found.")
	}
	return
}

// RemoveFinishedTodos removes all done todos and returns count of active todos
func (c *Collection) RemoveFinishedTodos() int {
	var activeTodos []*Todo
	for _, todo := range c.Todos {
		if todo.Status != "done" {
			activeTodos = append(activeTodos, todo)
		}
	}
	c.Todos = activeTodos
	return len(activeTodos)
}

// Swap swaps the position of two todos by their IDs
func (c *Collection) Swap(idA, idB int) error {
	var positionA, positionB = -1, -1
	idA64 := int64(idA)
	idB64 := int64(idB)

	for i, todo := range c.Todos {
		if todo.ID == idA64 {
			positionA = i
		}
		if todo.ID == idB64 {
			positionB = i
		}
	}

	if positionA == -1 || positionB == -1 {
		return errors.New("one or both todos not found")
	}

	// Swap the todos
	c.Todos[positionA], c.Todos[positionB] = c.Todos[positionB], c.Todos[positionA]
	// Swap the IDs
	c.Todos[positionA].ID, c.Todos[positionB].ID = c.Todos[positionB].ID, c.Todos[positionA].ID
	return nil
}
