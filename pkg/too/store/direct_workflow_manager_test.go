package store_test

import (
	"sort"
	"testing"

	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/store"
	"github.com/arthur-debert/too/pkg/too/store/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDirectWorkflowManager(t *testing.T) {
	t.Run("Add todo without adapters", func(t *testing.T) {
		// Create memory store
		memStore := internal.NewMemoryStore()
		
		// Create direct workflow manager
		manager, err := store.NewDirectWorkflowManager(memStore, "memory://test")
		require.NoError(t, err)
		
		// Add a todo
		uid, err := manager.Add(store.RootScope, "Test todo")
		require.NoError(t, err)
		assert.NotEmpty(t, uid)
		
		// Verify todo was created
		collection := manager.GetCollection()
		todo := collection.FindItemByID(uid)
		require.NotNil(t, todo)
		assert.Equal(t, "Test todo", todo.Text)
		assert.Equal(t, models.StatusPending, todo.GetStatus())
		
		// Save and verify persistence
		err = manager.Save()
		require.NoError(t, err)
	})
	
	t.Run("Set status without adapters", func(t *testing.T) {
		// Create memory store with existing todo
		memStore := internal.NewMemoryStore()
		collection := models.NewCollection()
		todo, _ := collection.CreateTodo("Test todo", "")
		err := memStore.Save(collection)
		require.NoError(t, err)
		
		// Create direct workflow manager
		manager, err := store.NewDirectWorkflowManager(memStore, "memory://test")
		require.NoError(t, err)
		
		// Set status to done
		err = manager.SetStatus(todo.ID, "completion", "done")
		require.NoError(t, err)
		
		// Verify status was updated
		status, err := manager.GetStatus(todo.ID, "completion")
		require.NoError(t, err)
		assert.Equal(t, "done", status)
		
		// Verify on the todo itself
		updatedTodo := manager.GetCollection().FindItemByID(todo.ID)
		assert.Equal(t, models.StatusDone, updatedTodo.GetStatus())
	})
	
	t.Run("Resolve position path without adapters", func(t *testing.T) {
		// Create memory store with nested todos
		memStore := internal.NewMemoryStore()
		collection := models.NewCollection()
		
		parent, _ := collection.CreateTodo("Parent", "")
		child1 := &models.Todo{
			ID:       "child1",
			ParentID: parent.ID,
			Text:     "Child 1",
			Statuses: map[string]string{"completion": string(models.StatusPending)},
			Items:    []*models.Todo{},
		}
		child2 := &models.Todo{
			ID:       "child2",
			ParentID: parent.ID,
			Text:     "Child 2",
			Statuses: map[string]string{"completion": string(models.StatusPending)},
			Items:    []*models.Todo{},
		}
		parent.Items = []*models.Todo{child1, child2}
		
		err := memStore.Save(collection)
		require.NoError(t, err)
		
		// Create direct workflow manager
		manager, err := store.NewDirectWorkflowManager(memStore, "memory://test")
		require.NoError(t, err)
		
		// Rebuild the parent scope to ensure children are registered
		adapter := &directTestAdapter{collection: manager.GetCollection()}
		err = manager.GetRegistry().RebuildScope(adapter, parent.ID)
		require.NoError(t, err)
		
		// Resolve position paths
		uid, err := manager.ResolvePositionPath(store.RootScope, "1")
		require.NoError(t, err)
		assert.Equal(t, parent.ID, uid)
		
		uid, err = manager.ResolvePositionPath(parent.ID, "2")
		require.NoError(t, err)
		assert.Equal(t, child2.ID, uid)
	})
}

// directTestAdapter is a test adapter for IDM operations
type directTestAdapter struct {
	collection *models.Collection
}

func (a *directTestAdapter) GetChildren(parentUID string) ([]string, error) {
	var todos []*models.Todo
	if parentUID == store.RootScope {
		todos = a.collection.Todos
	} else {
		parent := a.collection.FindItemByID(parentUID)
		if parent == nil {
			return []string{}, nil
		}
		todos = parent.Items
	}

	// Only return active items for HID assignment
	var children []string
	for _, todo := range todos {
		if todo.GetStatus() == models.StatusPending {
			children = append(children, todo.ID)
		}
	}

	// Sort by position
	sort.Slice(children, func(i, j int) bool {
		todoI := a.collection.FindItemByID(children[i])
		todoJ := a.collection.FindItemByID(children[j])
		if todoI == nil || todoJ == nil {
			return false
		}
		// Without Position field, maintain order by array index
		// Find the index of each todo in the parent's Items slice
		parent := a.collection.FindItemByID(todoI.ParentID)
		if parent == nil {
			return false
		}
		indexI, indexJ := -1, -1
		for idx, child := range parent.Items {
			if child.ID == todoI.ID {
				indexI = idx
			}
			if child.ID == todoJ.ID {
				indexJ = idx
			}
		}
		return indexI < indexJ
	})

	return children, nil
}

func (a *directTestAdapter) GetScopes() ([]string, error) {
	var scopes []string
	scopes = append(scopes, store.RootScope)
	a.collection.Walk(func(todo *models.Todo) {
		scopes = append(scopes, todo.ID)
	})
	return scopes, nil
}

func (a *directTestAdapter) GetAllUIDs() ([]string, error) {
	var uids []string
	a.collection.Walk(func(todo *models.Todo) {
		uids = append(uids, todo.ID)
	})
	return uids, nil
}