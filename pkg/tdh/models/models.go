package models

import (
	"fmt"
	"strconv"
	"strings"
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
	ParentID string     `json:"parentId"` // UUID of parent item, empty for top-level items
	Position int        `json:"position"` // Sequential position relative to siblings
	Text     string     `json:"text"`
	Status   TodoStatus `json:"status"`
	Modified time.Time  `json:"modified"`
	Items    []*Todo    `json:"items"` // Child todo items
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
		Status:   StatusPending,
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
		Status:   t.Status,
		Modified: t.Modified,
		Items:    make([]*Todo, len(t.Items)),
	}

	// Deep clone child items
	for i, item := range t.Items {
		clone.Items[i] = item.Clone()
	}

	return clone
}

// SetStatus changes the todo's status while maintaining invariants.
// If skipReorder is false (default), it triggers position reset at the appropriate level.
func (t *Todo) SetStatus(status TodoStatus, collection *Collection, skipReorder ...bool) {
	// Handle optional parameter
	skip := false
	if len(skipReorder) > 0 {
		skip = skipReorder[0]
	}

	// Track if status actually changed
	oldStatus := t.Status
	statusChanged := oldStatus != status

	// Update status and timestamp
	t.Status = status
	t.Modified = time.Now()

	// Maintain invariant: done items have position 0
	if status == StatusDone {
		t.Position = 0
	}
	// Note: If changing to pending and position is 0, it will be set by reorder

	// Trigger reorder unless skipped or status unchanged
	if !skip && statusChanged {
		if t.ParentID != "" {
			collection.ResetSiblingPositions(t.ParentID)
		} else {
			collection.ResetRootPositions()
		}
	}
}

// MarkComplete marks the todo as done and maintains invariants.
func (t *Todo) MarkComplete(collection *Collection, skipReorder ...bool) {
	t.SetStatus(StatusDone, collection, skipReorder...)
}

// MarkPending marks the todo as pending and maintains invariants.
func (t *Todo) MarkPending(collection *Collection, skipReorder ...bool) {
	t.SetStatus(StatusPending, collection, skipReorder...)
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

// Reorder sorts todos by their current position and reassigns sequential positions.
func (c *Collection) Reorder() {
	ReorderTodos(c.Todos)
}

// ResetSiblingPositions resets positions for all siblings of the todo with the given parent ID.
// This only affects todos at one level (children of the same parent).
func (c *Collection) ResetSiblingPositions(parentID string) {
	parent := c.FindItemByID(parentID)
	if parent != nil && len(parent.Items) > 0 {
		// Reset positions only for active (pending) items
		ResetActivePositions(parent.Items)
	}
}

// ResetRootPositions resets positions for all root-level todos.
func (c *Collection) ResetRootPositions() {
	if len(c.Todos) > 0 {
		// Reset positions only for active (pending) items
		ResetActivePositions(c.Todos)
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

// FindItemByPositionPath finds a todo item by its dot-notation position path (e.g., "1.2.3")
func (c *Collection) FindItemByPositionPath(path string) (*Todo, error) {
	if path == "" {
		return nil, fmt.Errorf("empty path")
	}

	positions, err := parsePositionPath(path)
	if err != nil {
		return nil, err
	}

	return findItemByPositions(c.Todos, positions)
}

// parsePositionPath converts a dot-notation path like "1.2.3" into a slice of positions
func parsePositionPath(path string) ([]int, error) {
	parts := strings.Split(path, ".")
	positions := make([]int, len(parts))

	for i, part := range parts {
		pos, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return nil, fmt.Errorf("invalid position '%s' in path: %w", part, err)
		}
		if pos < 1 {
			return nil, fmt.Errorf("position must be >= 1, got %d", pos)
		}
		positions[i] = pos
	}

	return positions, nil
}

// findItemByPositions recursively finds an item using a slice of positions
func findItemByPositions(todos []*Todo, positions []int) (*Todo, error) {
	if len(positions) == 0 {
		return nil, fmt.Errorf("no positions provided")
	}

	position := positions[0]

	// Find the todo at the current position
	var found *Todo
	for _, todo := range todos {
		if todo.Position == position {
			found = todo
			break
		}
	}

	if found == nil {
		return nil, fmt.Errorf("no item found at position %d", position)
	}

	// If this is the last position, return the found item
	if len(positions) == 1 {
		return found, nil
	}

	// Otherwise, recursively search in the found item's children
	return findItemByPositions(found.Items, positions[1:])
}
