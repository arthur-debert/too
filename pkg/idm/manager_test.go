package idm

import (
	"fmt"
	"testing"
)

// mockManagedStoreAdapter is a test implementation of the ManagedStoreAdapter.
type mockManagedStoreAdapter struct {
	items    map[string]*mockItem // map[uid]*item
	nextUID  int
	rootUIDs []string
}

type mockItem struct {
	uid      string
	parentID string
	children []string
}

func newMockManagedStoreAdapter() *mockManagedStoreAdapter {
	return &mockManagedStoreAdapter{
		items:    make(map[string]*mockItem),
		nextUID:  1,
		rootUIDs: []string{},
	}
}

func (m *mockManagedStoreAdapter) newUID() string {
	uid := fmt.Sprintf("uid%d", m.nextUID)
	m.nextUID++
	return uid
}

// --- Read Methods ---

func (m *mockManagedStoreAdapter) GetChildren(parentUID string) ([]string, error) {
	if parentUID == "root" {
		return m.rootUIDs, nil
	}
	parent, ok := m.items[parentUID]
	if !ok {
		// Return empty slice for scopes that exist but have no children
		return []string{}, nil
	}
	return parent.children, nil
}

func (m *mockManagedStoreAdapter) GetScopes() ([]string, error) {
	scopes := []string{"root"}
	for uid, item := range m.items {
		if len(item.children) > 0 {
			scopes = append(scopes, uid)
		}
	}
	return scopes, nil
}

func (m *mockManagedStoreAdapter) GetAllUIDs() ([]string, error) {
	uids := make([]string, 0, len(m.items))
	for uid := range m.items {
		uids = append(uids, uid)
	}
	return uids, nil
}

// --- Write Methods ---

func (m *mockManagedStoreAdapter) AddItem(parentUID string) (string, error) {
	uid := m.newUID()
	item := &mockItem{uid: uid, parentID: parentUID, children: []string{}}
	m.items[uid] = item

	if parentUID == "root" {
		m.rootUIDs = append(m.rootUIDs, uid)
	} else {
		parent, ok := m.items[parentUID]
		if !ok {
			return "", fmt.Errorf("parent %s not found", parentUID)
		}
		parent.children = append(parent.children, uid)
	}
	return uid, nil
}

func (m *mockManagedStoreAdapter) RemoveItem(uid string) error {
	// Simplified remove for testing; doesn't handle children correctly.
	delete(m.items, uid)
	return nil
}

func (m *mockManagedStoreAdapter) MoveItem(uid, newParentUID string) error {
	item, ok := m.items[uid]
	if !ok {
		return fmt.Errorf("item %s not found", uid)
	}

	// Remove from old parent
	oldParentID := item.parentID
	if oldParentID == "root" {
		newRoots := []string{}
		for _, rootUID := range m.rootUIDs {
			if rootUID != uid {
				newRoots = append(newRoots, rootUID)
			}
		}
		m.rootUIDs = newRoots
	} else {
		oldParent, ok := m.items[oldParentID]
		if ok {
			newChildren := []string{}
			for _, childUID := range oldParent.children {
				if childUID != uid {
					newChildren = append(newChildren, childUID)
				}
			}
			oldParent.children = newChildren
		}
	}

	// Add to new parent
	item.parentID = newParentUID
	if newParentUID == "root" {
		m.rootUIDs = append(m.rootUIDs, uid)
	} else {
		newParent, ok := m.items[newParentUID]
		if !ok {
			return fmt.Errorf("new parent %s not found", newParentUID)
		}
		newParent.children = append(newParent.children, uid)
	}
	return nil
}

func (m *mockManagedStoreAdapter) SetStatus(uid, status string) error { return nil } // No-op for now
func (m *mockManagedStoreAdapter) SetPinned(uid string, isPinned bool) error { return nil } // No-op for now

// --- Tests ---

func TestManager_Add(t *testing.T) {
	adapter := newMockManagedStoreAdapter()
	manager, err := NewManager(adapter)
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	// Add a root item
	newUID, newHID, err := manager.Add("root")
	if err != nil {
		t.Fatalf("manager.Add() failed: %v", err)
	}

	if newUID != "uid1" {
		t.Errorf("Expected new UID to be 'uid1', got '%s'", newUID)
	}
	if newHID != 1 {
		t.Errorf("Expected new HID to be 1, got %d", newHID)
	}

	// Verify registry state
	resolvedUID, err := manager.Registry().ResolveHID("root", 1)
	if err != nil {
		t.Fatalf("ResolveHID failed: %v", err)
	}
	if resolvedUID != "uid1" {
		t.Errorf("Expected resolved UID to be 'uid1', got '%s'", resolvedUID)
	}

	// Add a child item
	newChildUID, newChildHID, err := manager.Add("uid1")
	if err != nil {
		t.Fatalf("manager.Add() for child failed: %v", err)
	}
	if newChildUID != "uid2" {
		t.Errorf("Expected new child UID to be 'uid2', got '%s'", newChildUID)
	}
	if newChildHID != 1 {
		t.Errorf("Expected new child HID to be 1, got %d", newChildHID)
	}

	// Verify registry state for child
	resolvedChildUID, err := manager.Registry().ResolveHID("uid1", 1)
	if err != nil {
		t.Fatalf("ResolveHID for child failed: %v", err)
	}
	if resolvedChildUID != "uid2" {
		t.Errorf("Expected resolved child UID to be 'uid2', got '%s'", resolvedChildUID)
	}
}

func TestManager_Move(t *testing.T) {
	adapter := newMockManagedStoreAdapter()
	_, _ = adapter.AddItem("root")      // uid1
	_, _ = adapter.AddItem("root")      // uid2
	_, _ = adapter.AddItem("uid1") // uid3

	manager, err := NewManager(adapter)
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	// Move uid3 from uid1 to uid2
	err = manager.Move("uid3", "uid1", "uid2")
	if err != nil {
		t.Fatalf("manager.Move() failed: %v", err)
	}

	// Verify old scope (uid1) is now empty
	_, err = manager.Registry().ResolveHID("uid1", 1)
	if err == nil {
		t.Error("Expected error resolving from old scope, but got nil")
	}

	// Verify new scope (uid2) has the item
	resolvedUID, err := manager.Registry().ResolveHID("uid2", 1)
	if err != nil {
		t.Fatalf("ResolveHID in new scope failed: %v", err)
	}
	if resolvedUID != "uid3" {
		t.Errorf("Expected resolved UID to be 'uid3', got '%s'", resolvedUID)
	}
}
