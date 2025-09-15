package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/arthur-debert/nanostore/nanostore"
	"github.com/arthur-debert/too/pkg/too/models"
)

// NanoStoreAdapter wraps nanostore to provide too-specific functionality
type NanoStoreAdapter struct {
	store nanostore.Store
}

// NewNanoStoreAdapter creates a new adapter instance
func NewNanoStoreAdapter(dbPath string) (*NanoStoreAdapter, error) {
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

	// Create store with custom config for todo management
	config := nanostore.Config{
		Dimensions: []nanostore.DimensionConfig{
			{
				Name:         "status",
				Type:         nanostore.Enumerated,
				Values:       []string{"pending", "completed"},
				Prefixes:     map[string]string{"completed": "c"},
				DefaultValue: "pending",
			},
			{
				Name:     "parent_uuid",
				Type:     nanostore.Hierarchical,
				RefField: "parent_uuid",
			},
		},
	}
	store, err := nanostore.New(dbPath, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create nanostore: %w", err)
	}

	return &NanoStoreAdapter{store: store}, nil
}

// Close releases resources
func (n *NanoStoreAdapter) Close() error {
	return n.store.Close()
}

// CompleteByUUID marks a todo as completed by its UUID
func (n *NanoStoreAdapter) CompleteByUUID(uuid string) error {
	updates := nanostore.UpdateRequest{
		Dimensions: map[string]string{"status": "completed"},
	}
	return n.store.Update(uuid, updates)
}

// ReopenByUUID marks a completed todo as pending by its UUID
func (n *NanoStoreAdapter) ReopenByUUID(uuid string) error {
	updates := nanostore.UpdateRequest{
		Dimensions: map[string]string{"status": "pending"},
	}
	return n.store.Update(uuid, updates)
}

// UpdateByUUID modifies a todo's text by its UUID
func (n *NanoStoreAdapter) UpdateByUUID(uuid string, text string) error {
	updates := nanostore.UpdateRequest{
		Title: &text,
	}
	return n.store.Update(uuid, updates)
}

// MoveByUUID changes a todo's parent by its UUID
func (n *NanoStoreAdapter) MoveByUUID(uuid string, newParentID *string) error {
	// Resolve new parent if provided
	var newParentUUID *string
	if newParentID != nil && *newParentID != "" {
		parentUUID, err := n.store.ResolveUUID(*newParentID)
		if err != nil {
			return fmt.Errorf("failed to resolve new parent ID: %w", err)
		}
		newParentUUID = &parentUUID
	}

	updates := nanostore.UpdateRequest{
		Dimensions: map[string]string{},
	}
	if newParentUUID != nil {
		updates.Dimensions["parent_uuid"] = *newParentUUID
	} else {
		updates.Dimensions["parent_uuid"] = ""
	}
	return n.store.Update(uuid, updates)
}

// Add creates a new todo item
func (n *NanoStoreAdapter) Add(text string, parentID *string) (*models.Todo, error) {
	// If parentID is a user-facing ID, resolve it first
	var parentUUID *string
	if parentID != nil && *parentID != "" {
		uuid, err := n.store.ResolveUUID(*parentID)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve parent ID '%s': %w", *parentID, err)
		}
		parentUUID = &uuid
	}

	// Add the document
	dimensions := make(map[string]interface{})
	if parentUUID != nil {
		dimensions["parent_uuid"] = *parentUUID
	}
	uuid, err := n.store.Add(text, dimensions)
	if err != nil {
		return nil, fmt.Errorf("failed to add todo: %w", err)
	}

	// Get the document to return with its user-facing ID
	doc, err := n.getDocument(uuid)
	if err != nil {
		return nil, err
	}

	return n.documentToTodo(doc), nil
}

// Complete marks a todo as completed
func (n *NanoStoreAdapter) Complete(userFacingID string) error {
	updates := nanostore.UpdateRequest{
		Dimensions: map[string]string{"status": "completed"},
	}
	return n.store.Update(userFacingID, updates)
}

// Reopen marks a completed todo as pending
func (n *NanoStoreAdapter) Reopen(userFacingID string) error {
	updates := nanostore.UpdateRequest{
		Dimensions: map[string]string{"status": "pending"},
	}
	return n.store.Update(userFacingID, updates)
}

// Update modifies a todo's text
func (n *NanoStoreAdapter) Update(userFacingID string, text string) error {
	updates := nanostore.UpdateRequest{
		Title: &text,
	}
	return n.store.Update(userFacingID, updates)
}

// Move changes a todo's parent
func (n *NanoStoreAdapter) Move(userFacingID string, newParentID *string) error {
	// Resolve new parent if provided
	var newParentUUID *string
	if newParentID != nil && *newParentID != "" {
		parentUUID, err := n.store.ResolveUUID(*newParentID)
		if err != nil {
			return fmt.Errorf("failed to resolve new parent ID '%s': %w", *newParentID, err)
		}
		newParentUUID = &parentUUID
	}

	updates := nanostore.UpdateRequest{
		Dimensions: map[string]string{},
	}
	if newParentUUID != nil {
		updates.Dimensions["parent_uuid"] = *newParentUUID
	} else {
		updates.Dimensions["parent_uuid"] = ""
	}

	return n.store.Update(userFacingID, updates)
}

// Delete removes a todo and optionally its children
func (n *NanoStoreAdapter) Delete(userFacingID string, cascade bool) error {
	return n.store.Delete(userFacingID, cascade)
}

// DeleteCompleted removes all completed todos
func (n *NanoStoreAdapter) DeleteCompleted() (int, error) {
	return n.store.DeleteByDimension("status", "completed")
}

// List returns todos based on options
func (n *NanoStoreAdapter) List(showAll bool) ([]*models.Todo, error) {
	opts := nanostore.ListOptions{
		Filters: make(map[string]interface{}),
	}
	if !showAll {
		// Only show pending items
		opts.Filters["status"] = "pending"
	}

	docs, err := n.store.List(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list todos: %w", err)
	}

	todos := make([]*models.Todo, len(docs))
	for i, doc := range docs {
		todos[i] = n.documentToTodo(doc)
	}

	return todos, nil
}

// Search finds todos matching the query
func (n *NanoStoreAdapter) Search(query string, showAll bool) ([]*models.Todo, error) {
	opts := nanostore.ListOptions{
		FilterBySearch: query,
		Filters:        make(map[string]interface{}),
	}
	if !showAll {
		opts.Filters["status"] = "pending"
	}

	docs, err := n.store.List(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to search todos: %w", err)
	}

	todos := make([]*models.Todo, len(docs))
	for i, doc := range docs {
		todos[i] = n.documentToTodo(doc)
	}

	return todos, nil
}

// ResolvePositionPath converts a user-facing ID to UUID
func (n *NanoStoreAdapter) ResolvePositionPath(userFacingID string) (string, error) {
	return n.store.ResolveUUID(userFacingID)
}

// GetByUUID retrieves a todo by its UUID
func (n *NanoStoreAdapter) GetByUUID(uuid string) (*models.Todo, error) {
	doc, err := n.getDocument(uuid)
	if err != nil {
		return nil, err
	}
	return n.documentToTodo(doc), nil
}

// getDocument retrieves a single document by UUID
func (n *NanoStoreAdapter) getDocument(uuid string) (nanostore.Document, error) {
	// List all and find the one with matching UUID
	docs, err := n.store.List(nanostore.ListOptions{})
	if err != nil {
		return nanostore.Document{}, fmt.Errorf("failed to get document: %w", err)
	}

	for _, doc := range docs {
		if doc.UUID == uuid {
			return doc, nil
		}
	}

	return nanostore.Document{}, fmt.Errorf("document not found: %s", uuid)
}

// nanostoreStatusToTodoStatus converts nanostore status to todo status
func (n *NanoStoreAdapter) nanostoreStatusToTodoStatus(status string) string {
	switch status {
	case "completed":
		return string(models.StatusDone)
	case "pending":
		return string(models.StatusPending)
	default:
		return string(models.StatusPending)
	}
}

// getDocumentStatus extracts status from document dimensions
func (n *NanoStoreAdapter) getDocumentStatus(doc nanostore.Document) string {
	if status, ok := doc.Dimensions["status"].(string); ok {
		return status
	}
	return "pending"
}

// documentToTodo converts a nanostore Document to a Todo
func (n *NanoStoreAdapter) documentToTodo(doc nanostore.Document) *models.Todo {
	todo := &models.Todo{
		UID:          doc.UUID,
		Text:         doc.Title,
		PositionPath: doc.UserFacingID,
		ParentID:     "",
		Statuses: map[string]string{
			"completion": n.nanostoreStatusToTodoStatus(n.getDocumentStatus(doc)),
		},
		Modified: doc.UpdatedAt,
	}

	// Set ParentID if has parent
	if parentUUID, ok := doc.Dimensions["parent_uuid"].(string); ok && parentUUID != "" {
		todo.ParentID = parentUUID
	}

	return todo
}