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
		Dimensions: map[string]interface{}{"status": "completed"},
	}
	return n.store.Update(uuid, updates)
}

// ReopenByUUID marks a completed todo as pending by its UUID
func (n *NanoStoreAdapter) ReopenByUUID(uuid string) error {
	updates := nanostore.UpdateRequest{
		Dimensions: map[string]interface{}{"status": "pending"},
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
		Dimensions: map[string]interface{}{},
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
	// Resolve to UUID first
	uuid, err := n.store.ResolveUUID(userFacingID)
	if err != nil {
		return fmt.Errorf("failed to resolve ID '%s': %w", userFacingID, err)
	}
	
	updates := nanostore.UpdateRequest{
		Dimensions: map[string]interface{}{"status": "completed"},
	}
	return n.store.Update(uuid, updates)
}

// Reopen marks a completed todo as pending
func (n *NanoStoreAdapter) Reopen(userFacingID string) error {
	// Resolve to UUID first
	uuid, err := n.store.ResolveUUID(userFacingID)
	if err != nil {
		return fmt.Errorf("failed to resolve ID '%s': %w", userFacingID, err)
	}
	
	updates := nanostore.UpdateRequest{
		Dimensions: map[string]interface{}{"status": "pending"},
	}
	return n.store.Update(uuid, updates)
}

// Update modifies a todo's text
func (n *NanoStoreAdapter) Update(userFacingID string, text string) error {
	// Resolve to UUID first
	uuid, err := n.store.ResolveUUID(userFacingID)
	if err != nil {
		return fmt.Errorf("failed to resolve ID '%s': %w", userFacingID, err)
	}
	
	updates := nanostore.UpdateRequest{
		Title: &text,
	}
	return n.store.Update(uuid, updates)
}

// Move changes a todo's parent
func (n *NanoStoreAdapter) Move(userFacingID string, newParentID *string) error {
	// Resolve the todo's UUID first
	uuid, err := n.store.ResolveUUID(userFacingID)
	if err != nil {
		return fmt.Errorf("failed to resolve ID '%s': %w", userFacingID, err)
	}
	
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
		Dimensions: map[string]interface{}{},
	}
	if newParentUUID != nil {
		updates.Dimensions["parent_uuid"] = *newParentUUID
	} else {
		updates.Dimensions["parent_uuid"] = ""
	}

	return n.store.Update(uuid, updates)
}

// Delete removes a todo and optionally its children
func (n *NanoStoreAdapter) Delete(userFacingID string, cascade bool) error {
	// Resolve to UUID first
	uuid, err := n.store.ResolveUUID(userFacingID)
	if err != nil {
		return fmt.Errorf("failed to resolve ID '%s': %w", userFacingID, err)
	}
	return n.store.Delete(uuid, cascade)
}

// DeleteCompleted removes all completed todos
func (n *NanoStoreAdapter) DeleteCompleted() (int, error) {
	// List all completed todos
	opts := nanostore.ListOptions{
		Filters: map[string]interface{}{"status": "completed"},
	}
	docs, err := n.store.List(opts)
	if err != nil {
		return 0, fmt.Errorf("failed to list completed todos: %w", err)
	}
	
	// Delete them one by one
	count := 0
	for _, doc := range docs {
		if err := n.store.Delete(doc.UUID, true); err != nil {
			return count, fmt.Errorf("failed to delete todo %s: %w", doc.UUID, err)
		}
		count++
	}
	
	return count, nil
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
// This version searches across ALL statuses, not just the default "pending"
func (n *NanoStoreAdapter) ResolvePositionPath(userFacingID string) (string, error) {
	// Try to resolve the ID as-is first (for cases where user provided explicit prefixes)
	uuid, err := n.store.ResolveUUID(userFacingID)
	if err == nil {
		return uuid, nil
	}
	
	// If that fails and the ID doesn't have status prefixes, try different combinations
	if !strings.Contains(userFacingID, "c") && !strings.Contains(userFacingID, "p") {
		// For hierarchical IDs like "1.1", we need to try different status combinations
		// since parents and children might have different statuses
		combinations := n.generateStatusCombinations(userFacingID)
		
		for _, combination := range combinations {
			uuid, err = n.store.ResolveUUID(combination)
			if err == nil {
				return uuid, nil
			}
		}
		
		// If no combinations worked, return a specific error
		return "", fmt.Errorf("could not resolve '%s' with any status combination (tried %d combinations)", userFacingID, len(combinations))
	}
	
	// Return a different error format to distinguish from nanostore errors
	return "", fmt.Errorf("failed to resolve reference '%s': %w", userFacingID, err)
}

// generateStatusCombinations generates different status prefix combinations for a hierarchical ID
func (n *NanoStoreAdapter) generateStatusCombinations(userFacingID string) []string {
	parts := strings.Split(userFacingID, ".")
	numLevels := len(parts)
	
	// For each level, we can have pending (no prefix) or completed ("c" prefix)
	// Generate all 2^numLevels combinations
	var combinations []string
	
	// Use bit manipulation to generate all combinations
	for i := 0; i < (1 << numLevels); i++ {
		var combination []string
		for j := 0; j < numLevels; j++ {
			part := parts[j]
			// If bit j is set, use completed prefix
			if (i >> j) & 1 == 1 {
				part = "c" + part
			}
			combination = append(combination, part)
		}
		combinations = append(combinations, strings.Join(combination, "."))
	}
	
	return combinations
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
		PositionPath: doc.SimpleID,
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