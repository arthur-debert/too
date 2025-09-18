package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/arthur-debert/nanostore/nanostore"
	"github.com/arthur-debert/too/pkg/too/models"
)

// DeclarativeAdapter wraps nanostore's declarative API for todo management
type DeclarativeAdapter struct {
	store *nanostore.TypedStore[models.TodoDeclarative]
}

// NewDeclarativeAdapter creates a new declarative adapter instance
func NewDeclarativeAdapter(dbPath string) (*DeclarativeAdapter, error) {
	// Expand ~ to home directory
	if strings.HasPrefix(dbPath, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		dbPath = filepath.Join(home, dbPath[2:])
	}

	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Create typed store - nanostore will automatically configure dimensions from struct tags
	store, err := nanostore.NewFromType[models.TodoDeclarative](dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create declarative store: %w", err)
	}

	return &DeclarativeAdapter{store: store}, nil
}

// Close releases resources
func (d *DeclarativeAdapter) Close() error {
	return d.store.Close()
}

// Add creates a new todo item using the declarative API
func (d *DeclarativeAdapter) Add(text string, parentID *string) (*models.Todo, error) {
	// Create the declarative todo
	todo := &models.TodoDeclarative{
		Text:     text,
		Modified: time.Now(),
	}
	
	// Set parent if provided
	if parentID != nil && *parentID != "" {
		// Try to get parent to validate it exists
		_, err := d.store.Get(*parentID)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve parent ID '%s': %w", *parentID, err)
		}
		todo.ParentID = *parentID
	}

	// Create using declarative API
	id, err := d.store.Create(text, todo)
	if err != nil {
		return nil, fmt.Errorf("failed to add todo: %w", err)
	}

	// Get the created todo and convert to legacy format
	created, err := d.store.Get(id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created todo: %w", err)
	}

	return created.ToLegacy(), nil
}

// Complete marks a todo as completed
func (d *DeclarativeAdapter) Complete(userFacingID string) error {
	todo, err := d.store.Get(userFacingID)
	if err != nil {
		return err
	}
	todo.Complete()
	return d.store.Update(userFacingID, todo)
}

// CompleteByUUID marks a todo as completed by its UUID
func (d *DeclarativeAdapter) CompleteByUUID(uuid string) error {
	// For now, UUID and SimpleID are the same in this API
	return d.Complete(uuid)
}

// Reopen marks a completed todo as pending
func (d *DeclarativeAdapter) Reopen(userFacingID string) error {
	todo, err := d.store.Get(userFacingID)
	if err != nil {
		return err
	}
	todo.Reopen()
	return d.store.Update(userFacingID, todo)
}

// ReopenByUUID marks a completed todo as pending by its UUID
func (d *DeclarativeAdapter) ReopenByUUID(uuid string) error {
	// For now, UUID and SimpleID are the same in this API
	return d.Reopen(uuid)
}

// Update modifies a todo's text
func (d *DeclarativeAdapter) Update(userFacingID string, text string) error {
	todo, err := d.store.Get(userFacingID)
	if err != nil {
		return err
	}
	todo.UpdateText(text)
	return d.store.Update(userFacingID, todo)
}

// UpdateByUUID modifies a todo's text by its UUID
func (d *DeclarativeAdapter) UpdateByUUID(uuid string, text string) error {
	// For now, UUID and SimpleID are the same in this API
	return d.Update(uuid, text)
}

// Move changes a todo's parent
func (d *DeclarativeAdapter) Move(userFacingID string, newParentID *string) error {
	// Validate new parent exists if provided
	if newParentID != nil && *newParentID != "" {
		_, err := d.store.Get(*newParentID)
		if err != nil {
			return fmt.Errorf("failed to resolve new parent ID '%s': %w", *newParentID, err)
		}
	}

	todo, err := d.store.Get(userFacingID)
	if err != nil {
		return err
	}
	
	if newParentID != nil && *newParentID != "" {
		todo.ParentID = *newParentID
	} else {
		todo.ParentID = ""
	}
	todo.Modified = time.Now()
	
	return d.store.Update(userFacingID, todo)
}

// MoveByUUID changes a todo's parent by its UUID
func (d *DeclarativeAdapter) MoveByUUID(uuid string, newParentID *string) error {
	// For now, UUID and SimpleID are the same in this API
	return d.Move(uuid, newParentID)
}

// Delete removes a todo and optionally its children
func (d *DeclarativeAdapter) Delete(userFacingID string, cascade bool) error {
	return d.store.Delete(userFacingID, cascade)
}

// DeleteCompleted removes all completed todos
func (d *DeclarativeAdapter) DeleteCompleted() (int, error) {
	// Get all completed todos first
	completed, err := d.store.Query().Status("completed").Find()
	if err != nil {
		return 0, err
	}
	
	// Delete each one
	count := 0
	for _, todo := range completed {
		if err := d.store.Delete(todo.SimpleID, false); err == nil {
			count++
		}
	}
	
	return count, nil
}

// List returns todos based on options
func (d *DeclarativeAdapter) List(showAll bool) ([]*models.Todo, error) {
	var todos []models.TodoDeclarative
	var err error
	
	if showAll {
		todos, err = d.store.Query().Find()
	} else {
		todos, err = d.store.Query().Status("pending").Find()
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to list todos: %w", err)
	}

	// Convert to legacy format
	result := make([]*models.Todo, len(todos))
	for i, todo := range todos {
		result[i] = todo.ToLegacy()
	}

	return result, nil
}

// Search finds todos matching the query
func (d *DeclarativeAdapter) Search(query string, showAll bool) ([]*models.Todo, error) {
	var todos []models.TodoDeclarative
	var err error
	
	if showAll {
		todos, err = d.store.Query().Search(query).Find()
	} else {
		todos, err = d.store.Query().Status("pending").Search(query).Find()
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to search todos: %w", err)
	}

	// Convert to legacy format
	result := make([]*models.Todo, len(todos))
	for i, todo := range todos {
		result[i] = todo.ToLegacy()
	}

	return result, nil
}

// ResolvePositionPath converts a user-facing ID to UUID
func (d *DeclarativeAdapter) ResolvePositionPath(userFacingID string) (string, error) {
	todo, err := d.store.Get(userFacingID)
	if err != nil {
		return "", fmt.Errorf("failed to resolve reference '%s': %w", userFacingID, err)
	}
	return todo.UUID, nil
}

// GetByUUID retrieves a todo by its UUID
func (d *DeclarativeAdapter) GetByUUID(uuid string) (*models.Todo, error) {
	todo, err := d.store.Get(uuid)
	if err != nil {
		return nil, err
	}
	return todo.ToLegacy(), nil
}