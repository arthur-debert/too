package models

import "fmt"

// MigrateToIDM converts a hierarchical Collection to a flat IDMCollection.
// This migration flattens the tree structure and prepares the data for pure IDM management.
func MigrateToIDM(collection *Collection) *IDMCollection {
	idmCollection := NewIDMCollection()
	
	// Recursively flatten all todos from the hierarchical structure
	for _, todo := range collection.Todos {
		flattenTodoToIDM(todo, idmCollection)
	}
	
	return idmCollection
}

// flattenTodoToIDM recursively converts a hierarchical Todo to flat IDMTodo items.
func flattenTodoToIDM(todo *Todo, collection *IDMCollection) {
	// Convert the current todo to IDMTodo
	idmTodo := &IDMTodo{
		UID:      todo.ID,           // Keep the same UID
		ParentID: todo.ParentID,     // Preserve parent relationship
		Text:     todo.Text,
		Statuses: make(map[string]string),
		Modified: todo.Modified,
	}
	
	// Copy statuses
	if todo.Statuses != nil {
		for k, v := range todo.Statuses {
			idmTodo.Statuses[k] = v
		}
	} else {
		// Ensure backward compatibility for old todos without statuses
		idmTodo.Statuses["completion"] = string(StatusPending)
	}
	
	// Add to flat collection
	collection.AddItem(idmTodo)
	
	// Recursively process children
	for _, child := range todo.Items {
		flattenTodoToIDM(child, collection)
	}
}

// MigrateFromIDM converts a flat IDMCollection back to a hierarchical Collection.
// This is used for backward compatibility or when hierarchical structure is needed.
func MigrateFromIDM(idmCollection *IDMCollection) *Collection {
	collection := NewCollection()
	
	// Create a map for fast lookup
	todoMap := make(map[string]*Todo)
	
	// First pass: create all Todo objects
	for _, idmTodo := range idmCollection.Items {
		todo := &Todo{
			ID:       idmTodo.UID,
			ParentID: idmTodo.ParentID,
			Text:     idmTodo.Text,
			Statuses: make(map[string]string),
			Modified: idmTodo.Modified,
			Items:    []*Todo{}, // Initialize empty children
		}
		
		// Copy statuses
		for k, v := range idmTodo.Statuses {
			todo.Statuses[k] = v
		}
		
		todoMap[todo.ID] = todo
	}
	
	// Second pass: build hierarchy
	for _, todo := range todoMap {
		if todo.ParentID == "" {
			// Root level todo
			collection.Todos = append(collection.Todos, todo)
		} else {
			// Child todo - add to parent's Items
			if parent, exists := todoMap[todo.ParentID]; exists {
				parent.Items = append(parent.Items, todo)
			}
		}
	}
	
	return collection
}

// ValidateIDMCollection checks that an IDMCollection is valid for IDM usage.
// Returns any validation errors found.
func ValidateIDMCollection(collection *IDMCollection) []error {
	var errors []error
	uidMap := make(map[string]bool)
	
	for _, item := range collection.Items {
		// Check for duplicate UIDs
		if uidMap[item.UID] {
			errors = append(errors, fmt.Errorf("duplicate UID found: %s", item.UID))
		}
		uidMap[item.UID] = true
		
		// Check UID is not empty
		if item.UID == "" {
			errors = append(errors, fmt.Errorf("empty UID found in item with text: %s", item.Text))
		}
		
		// Check ParentID references valid item (if not empty)
		if item.ParentID != "" && !uidMap[item.ParentID] {
			// We need to defer this check since items might not be processed in order
			// For now, we'll check this in a second pass
		}
	}
	
	// Second pass: check parent references
	for _, item := range collection.Items {
		if item.ParentID != "" && !uidMap[item.ParentID] {
			errors = append(errors, fmt.Errorf("item %s references non-existent parent %s", item.UID, item.ParentID))
		}
	}
	
	return errors
}

// DetectCircularReferences checks for circular parent-child relationships in an IDMCollection.
func DetectCircularReferences(collection *IDMCollection) []error {
	var errors []error
	
	for _, item := range collection.Items {
		if hasCircularReference(item.UID, collection, make(map[string]bool)) {
			errors = append(errors, fmt.Errorf("circular reference detected involving item %s", item.UID))
		}
	}
	
	return errors
}

// hasCircularReference recursively checks if an item has a circular parent reference.
func hasCircularReference(startUID string, collection *IDMCollection, visited map[string]bool) bool {
	if visited[startUID] {
		return true // Found a cycle
	}
	
	item := collection.FindByUID(startUID)
	if item == nil || item.ParentID == "" {
		return false // No parent or item not found
	}
	
	visited[startUID] = true
	result := hasCircularReference(item.ParentID, collection, visited)
	delete(visited, startUID) // Clean up for other paths
	
	return result
}