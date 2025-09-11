package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// TodoStatus represents the status of a todo item
// DEPRECATED: Use workflow statuses instead
type TodoStatus string

const (
	// StatusPending indicates the todo is not yet completed
	StatusPending TodoStatus = "pending"
	// StatusDone indicates the todo has been completed
	StatusDone TodoStatus = "done"
)

// Todo represents a single task in the to-do list.
type Todo struct {
	ID       string            `json:"id"`                 // UUID for stable internal reference
	ParentID string            `json:"parentId"`           // UUID of parent item, empty for top-level items
	Position int               `json:"position,omitempty"` // Position managed by IDM
	Text     string            `json:"text"`
	Statuses map[string]string `json:"statuses,omitempty"` // Multi-dimensional status for workflow features
	Modified time.Time         `json:"modified"`
	Items    []*Todo           `json:"items"` // Child todo items
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
// If parentID is empty, adds to root level. Otherwise, adds as child of the specified parent.
func (c *Collection) CreateTodo(text string, parentID string) (*Todo, error) {
	newTodo := &Todo{
		ID:       uuid.New().String(),
		ParentID: parentID,
		Text:     text,
		Statuses: map[string]string{"completion": string(StatusPending)}, // Set workflow status
		Modified: time.Now(),
		Items:    []*Todo{},
	}

	if parentID == "" {
		// Add to root level
		newTodo.Position = c.findHighestPosition(c.Todos) + 1
		c.Todos = append(c.Todos, newTodo)
	} else {
		// Find parent and add as child
		parent := c.FindItemByID(parentID)
		if parent == nil {
			return nil, fmt.Errorf("parent todo with ID %s not found", parentID)
		}
		newTodo.Position = c.findHighestPosition(parent.Items) + 1
		parent.Items = append(parent.Items, newTodo)
	}

	return newTodo, nil
}

// findHighestPosition finds the highest position in a slice of todos
func (c *Collection) findHighestPosition(todos []*Todo) int {
	var highest = 0
	for _, todo := range todos {
		if todo.Position > highest {
			highest = todo.Position
		}
	}
	return highest
}

// Clone creates a deep copy of the todo.
func (t *Todo) Clone() *Todo {
	clone := &Todo{
		ID:       t.ID,
		ParentID: t.ParentID,
		Position: t.Position,
		Text:     t.Text,
		Statuses: make(map[string]string),
		Modified: t.Modified,
		Items:    make([]*Todo, len(t.Items)),
	}

	// Clone statuses map
	for k, v := range t.Statuses {
		clone.Statuses[k] = v
	}

	// Deep clone child items
	for i, item := range t.Items {
		clone.Items[i] = item.Clone()
	}

	return clone
}

// EnsureStatuses initializes the Statuses map if it's nil and ensures backward compatibility.
func (t *Todo) EnsureStatuses() {
	if t.Statuses == nil {
		t.Statuses = make(map[string]string)
		// Default to pending status
		t.Statuses["completion"] = string(StatusPending)
	}
}

// GetWorkflowStatus gets a status dimension value.
func (t *Todo) GetWorkflowStatus(dimension string) (string, bool) {
	t.EnsureStatuses()
	
	if value, exists := t.Statuses[dimension]; exists {
		return value, true
	}
	
	return "", false
}

// SetModified updates the modified timestamp to the current time.
func (t *Todo) SetModified() {
	t.Modified = time.Now()
}

// GetStatus returns the todo's completion status from the workflow statuses.
func (t *Todo) GetStatus() TodoStatus {
	t.EnsureStatuses()
	if status, exists := t.Statuses["completion"]; exists {
		return TodoStatus(status)
	}
	// Default to pending if no status set
	return StatusPending
}

// IsComplete returns true if the todo is marked as done.
func (t *Todo) IsComplete() bool {
	return t.GetStatus() == StatusDone
}

// IsPending returns true if the todo is marked as pending.
func (t *Todo) IsPending() bool {
	return t.GetStatus() == StatusPending
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

// Reorder resets positions for active (pending) todos, giving them sequential positions starting from 1.
// Done todos are left with position 0.
func (c *Collection) Reorder() {
	ResetActivePositions(&c.Todos)
}

// ResetSiblingPositions resets positions for all siblings of the todo with the given parent ID.
// This only affects todos at one level (children of the same parent).
func (c *Collection) ResetSiblingPositions(parentID string) {
	parent := c.FindItemByID(parentID)
	if parent != nil && len(parent.Items) > 0 {
		// Reset positions only for active (pending) items
		ResetActivePositions(&parent.Items)
	}
}

// ResetRootPositions resets positions for all root-level todos.
func (c *Collection) ResetRootPositions() {
	if len(c.Todos) > 0 {
		// Reset positions only for active (pending) items
		ResetActivePositions(&c.Todos)
	}
}

// MigrateCollection ensures all todos have proper IDs and structure for nested lists
func MigrateCollection(c *Collection) {
	for _, todo := range c.Todos {
		migrateTodo(todo)
	}
}

// migrateTodo recursively migrates a todo and its children
func migrateTodo(t *Todo) {
	// Ensure todo has an ID
	if t.ID == "" {
		t.ID = uuid.New().String()
	}

	// Ensure ParentID is set (empty for top-level)
	// ParentID is already empty string by default for top-level items

	// Ensure Items is initialized
	if t.Items == nil {
		t.Items = []*Todo{}
	}

	// Recursively migrate child items
	for _, child := range t.Items {
		if child.ParentID == "" {
			child.ParentID = t.ID
		}
		migrateTodo(child)
	}
}

// FindItemByID finds a todo item by its ID, searching recursively through the tree
func (c *Collection) FindItemByID(id string) *Todo {
	return findItemByIDInSlice(c.Todos, id)
}

// findItemByIDInSlice recursively searches for a todo by ID in a slice
func findItemByIDInSlice(todos []*Todo, id string) *Todo {
	for _, todo := range todos {
		if todo.ID == id {
			return todo
		}
		// Recursively search in children
		if found := findItemByIDInSlice(todo.Items, id); found != nil {
			return found
		}
	}
	return nil
}

// FindItemByShortID finds a todo item by its short ID, searching recursively.
func (c *Collection) FindItemByShortID(shortID string) (*Todo, error) {
	var found *Todo
	var count int
	c.Walk(func(t *Todo) {
		if strings.HasPrefix(t.ID, shortID) {
			found = t
			count++
		}
	})

	if count == 0 {
		return nil, fmt.Errorf("no todo found with reference '%s'", shortID)
	}
	if count > 1 {
		return nil, fmt.Errorf("multiple todos found with ambiguous reference '%s'", shortID)
	}
	return found, nil
}

// Walk traverses the entire todo tree and calls the given function for each todo.
func (c *Collection) Walk(fn func(*Todo)) {
	for _, todo := range c.Todos {
		walk(todo, fn)
	}
}

func walk(t *Todo, fn func(*Todo)) {
	fn(t)
	for _, child := range t.Items {
		walk(child, fn)
	}
}

// GetShortID returns the first 7 characters of the todo's UUID.
func (t *Todo) GetShortID() string {
	if len(t.ID) >= 7 {
		return t.ID[:7]
	}
	return t.ID
}

// ListActive returns only active (pending) todos from the collection.
// This implements behavioral propagation: when a parent is done, all its
// descendants are hidden regardless of their status.
func (c *Collection) ListActive() []*Todo {
	return filterTodos(c.Todos, func(t *Todo) bool {
		return t.GetStatus() == StatusPending
	}, false) // Don't recurse into done items
}

// ListArchived returns only archived (done) todos from the collection.
// When showing archived items, behavioral propagation stops - we don't
// show the children of done items.
func (c *Collection) ListArchived() []*Todo {
	return filterTodos(c.Todos, func(t *Todo) bool {
		return t.GetStatus() == StatusDone
	}, false) // Don't recurse into done items
}

// ListAll returns all todos from the collection regardless of status.
// This shows the complete tree structure including any inconsistent states
// (e.g., pending children under done parents).
func (c *Collection) ListAll() []*Todo {
	return cloneTodos(c.Todos)
}

// filterTodos recursively filters todos based on a predicate function.
// If recurseIntoDone is false, it stops recursion at done items (behavioral propagation).
func filterTodos(todos []*Todo, predicate func(*Todo) bool, recurseIntoDone bool) []*Todo {
	var filtered []*Todo

	for _, todo := range todos {
		if predicate(todo) {
			// Clone the todo to avoid modifying the original
			filteredTodo := &Todo{
				ID:       todo.ID,
				ParentID: todo.ParentID,
				Position: todo.Position,
				Text:     todo.Text,
				Statuses: make(map[string]string),
				Modified: todo.Modified,
				Items:    []*Todo{},
			}

			// Clone statuses map
			for k, v := range todo.Statuses {
				filteredTodo.Statuses[k] = v
			}

			// If this todo is done and we're not recursing into done items,
			// stop here (behavioral propagation)
			if todo.GetStatus() == StatusDone && !recurseIntoDone {
				filtered = append(filtered, filteredTodo)
			} else {
				// Recursively filter children
				filteredTodo.Items = filterTodos(todo.Items, predicate, recurseIntoDone)
				filtered = append(filtered, filteredTodo)
			}
		}
	}

	return filtered
}

// cloneTodos creates a deep copy of a slice of todos
func cloneTodos(todos []*Todo) []*Todo {
	cloned := make([]*Todo, len(todos))
	for i, todo := range todos {
		cloned[i] = todo.Clone()
	}
	return cloned
}

// AllTodos returns all todos in the collection as a flat slice.
func (c *Collection) AllTodos() []*Todo {
	var allTodos []*Todo
	c.Walk(func(t *Todo) {
		allTodos = append(allTodos, t)
	})
	return allTodos
}

// FindHighestPosition finds the highest position in a slice of todos (public version).
func (c *Collection) FindHighestPosition(todos []*Todo) int {
	return c.findHighestPosition(todos)
}

// RemoveTodo removes a todo by ID from the collection.
func (c *Collection) RemoveTodo(id string) error {
	todo := c.FindItemByID(id)
	if todo == nil {
		return fmt.Errorf("todo with ID %s not found", id)
	}
	
	// Remove from parent's items list
	if todo.ParentID == "" {
		// Remove from root level
		for i, rootTodo := range c.Todos {
			if rootTodo.ID == id {
				c.Todos = append(c.Todos[:i], c.Todos[i+1:]...)
				return nil
			}
		}
	} else {
		// Remove from parent's items
		parent := c.FindItemByID(todo.ParentID)
		if parent != nil {
			for i, child := range parent.Items {
				if child.ID == id {
					parent.Items = append(parent.Items[:i], parent.Items[i+1:]...)
					return nil
				}
			}
		}
	}
	
	return fmt.Errorf("failed to remove todo with ID %s", id)
}