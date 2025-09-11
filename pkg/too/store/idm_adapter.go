package store

import (
	"fmt"
	"sort"

	"github.com/arthur-debert/too/pkg/idm"
	"github.com/arthur-debert/too/pkg/too/models"
)

// RootScope is a special constant used to identify the root of the todo tree
// in the IDM registry.
const RootScope = "root"

// IDMStoreAdapter bridges the gap between the generic idm.Registry and too's
// specific data storage implementation.
type IDMStoreAdapter struct {
	store      Store
	collection *models.Collection
}

// NewIDMStoreAdapter creates a new adapter. It loads the collection from the
// provided store once and caches it for subsequent calls to avoid repeated I/O.
func NewIDMStoreAdapter(store Store) (*IDMStoreAdapter, error) {
	collection, err := store.Load()
	if err != nil {
		return nil, err
	}
	return &IDMStoreAdapter{
		store:      store,
		collection: collection,
	}, nil
}

// GetChildren implements the idm.StoreAdapter interface. It returns an ordered
// list of UIDs for a given parent UID (scope). It only returns children with
// a "pending" status, as they are the only ones with user-facing HIDs.
func (a *IDMStoreAdapter) GetChildren(parentUID string) ([]string, error) {
	var targetTodos []*models.Todo

	if parentUID == RootScope {
		targetTodos = a.collection.Todos
	} else {
		parent := a.collection.FindItemByID(parentUID)
		if parent == nil {
			// Return an empty slice instead of an error, as a scope may exist
			// but have no children.
			return []string{}, nil
		}
		targetTodos = parent.Items
	}

	// Filter for pending todos, as they are the only ones with HIDs.
	pendingTodos := make([]*models.Todo, 0)
	for _, todo := range targetTodos {
		if todo.GetStatus() == models.StatusPending {
			pendingTodos = append(pendingTodos, todo)
		}
	}

	// Sort by position to ensure a stable order for HIDs.
	sort.Slice(pendingTodos, func(i, j int) bool {
		return pendingTodos[i].Position < pendingTodos[j].Position
	})

	// Extract the UIDs.
	uids := make([]string, len(pendingTodos))
	for i, todo := range pendingTodos {
		uids[i] = todo.ID
	}

	return uids, nil
}

// GetScopes implements the idm.StoreAdapter interface. It returns all possible
// scopes, which includes the RootScope and the UID of every todo that has children.
func (a *IDMStoreAdapter) GetScopes() ([]string, error) {
	scopes := map[string]struct{}{
		RootScope: {},
	}

	a.collection.Walk(func(t *models.Todo) {
		if len(t.Items) > 0 {
			scopes[t.ID] = struct{}{}
		}
	})

	scopeList := make([]string, 0, len(scopes))
	for scope := range scopes {
		scopeList = append(scopeList, scope)
	}

	return scopeList, nil
}

// GetAllUIDs implements the idm.StoreAdapter interface. It returns all UIDs
// in the collection, regardless of status.
func (a *IDMStoreAdapter) GetAllUIDs() ([]string, error) {
	var uids []string
	a.collection.Walk(func(t *models.Todo) {
		uids = append(uids, t.ID)
	})
	return uids, nil
}

// --- ManagedStoreAdapter Write Methods ---

// AddItem implements the idm.ManagedStoreAdapter interface. It creates a new
// todo item with the given parent and returns its new UID.
func (a *IDMStoreAdapter) AddItem(parentUID string) (string, error) {
	var parentID string
	if parentUID != RootScope {
		parentID = parentUID
	}
	
	todo, err := a.collection.CreateTodo("", parentID)
	if err != nil {
		return "", fmt.Errorf("failed to create todo: %w", err)
	}
	
	return todo.ID, nil
}

// RemoveItem implements the idm.ManagedStoreAdapter interface. It permanently
// deletes an item and all its descendants.
func (a *IDMStoreAdapter) RemoveItem(uid string) error {
	todo := a.collection.FindItemByID(uid)
	if todo == nil {
		return fmt.Errorf("todo with UID %s not found", uid)
	}

	// Find the parent and remove from its Items slice
	if todo.ParentID != "" {
		parent := a.collection.FindItemByID(todo.ParentID)
		if parent != nil {
			for i, item := range parent.Items {
				if item.ID == uid {
					parent.Items = append(parent.Items[:i], parent.Items[i+1:]...)
					break
				}
			}
		}
	} else {
		// Remove from root todos
		for i, item := range a.collection.Todos {
			if item.ID == uid {
				a.collection.Todos = append(a.collection.Todos[:i], a.collection.Todos[i+1:]...)
				break
			}
		}
	}

	return nil
}

// MoveItem implements the idm.ManagedStoreAdapter interface. It changes an
// item's parent.
func (a *IDMStoreAdapter) MoveItem(uid, newParentUID string) error {
	todo := a.collection.FindItemByID(uid)
	if todo == nil {
		return fmt.Errorf("todo with UID %s not found", uid)
	}

	// Find the new parent (if not moving to root)
	var newParent *models.Todo
	var newParentID string
	if newParentUID != RootScope {
		newParent = a.collection.FindItemByID(newParentUID)
		if newParent == nil {
			return fmt.Errorf("new parent with UID %s not found", newParentUID)
		}
		newParentID = newParentUID
	}

	// Remove from old location
	if todo.ParentID != "" {
		oldParent := a.collection.FindItemByID(todo.ParentID)
		if oldParent != nil {
			for i, item := range oldParent.Items {
				if item.ID == uid {
					oldParent.Items = append(oldParent.Items[:i], oldParent.Items[i+1:]...)
					break
				}
			}
		}
	} else {
		// Remove from root todos
		for i, item := range a.collection.Todos {
			if item.ID == uid {
				a.collection.Todos = append(a.collection.Todos[:i], a.collection.Todos[i+1:]...)
				break
			}
		}
	}

	// Add to new location
	todo.ParentID = newParentID
	if newParent != nil {
		newParent.Items = append(newParent.Items, todo)
	} else {
		a.collection.Todos = append(a.collection.Todos, todo)
	}

	// Reorder the collection to fix positions
	a.collection.Reorder()

	return nil
}

// SetStatus implements the idm.ManagedStoreAdapter interface. It changes the
// status of an item (e.g., "active", "deleted").
func (a *IDMStoreAdapter) SetStatus(uid, status string) error {
	todo := a.collection.FindItemByID(uid)
	if todo == nil {
		return fmt.Errorf("todo with UID %s not found", uid)
	}
	
	// Map IDM status constants to workflow statuses
	todo.EnsureStatuses()
	switch status {
	case idm.StatusActive:
		todo.Statuses["completion"] = string(models.StatusPending)
	case idm.StatusDeleted:
		todo.Statuses["completion"] = string(models.StatusDone)
	default:
		return fmt.Errorf("unknown status: %s", status)
	}
	
	// Update modified timestamp
	todo.SetModified()
	
	return nil
}

// SetPinned implements the idm.ManagedStoreAdapter interface. It marks an
// item as pinned or not. Note: too doesn't currently support pinned items,
// so this is a no-op for now.
func (a *IDMStoreAdapter) SetPinned(uid string, isPinned bool) error {
	// too doesn't currently support pinned items, so this is a no-op
	return nil
}

// GetParent implements the idm.ManagedStoreAdapter interface. It returns the
// parent UID of the given item.
func (a *IDMStoreAdapter) GetParent(uid string) (string, error) {
	if uid == RootScope {
		return "", nil // Root has no parent
	}
	
	todo := a.collection.FindItemByID(uid)
	if todo == nil {
		return "", fmt.Errorf("todo with UID %s not found", uid)
	}
	
	if todo.ParentID == "" {
		return RootScope, nil
	}
	
	return todo.ParentID, nil
}
