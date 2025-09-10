package store

import (
	"sort"

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
		if todo.Status == models.StatusPending {
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
