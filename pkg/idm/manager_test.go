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
	status   string // For soft delete tests
	isPinned bool   // For pinned items tests
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
	// Handle the special "pinned" scope
	if parentUID == ScopePinned {
		pinnedUIDs := []string{}
		for uid, item := range m.items {
			if item.isPinned {
				pinnedUIDs = append(pinnedUIDs, uid)
			}
		}
		return pinnedUIDs, nil
	}

	var childrenSource []string
	if parentUID == "root" {
		childrenSource = m.rootUIDs
	} else {
		parent, ok := m.items[parentUID]
		if !ok {
			return []string{}, nil
		}
		childrenSource = parent.children
	}

	// For soft-delete, only return active children
	activeChildren := []string{}
	for _, uid := range childrenSource {
		if item, ok := m.items[uid]; ok && item.status == StatusActive {
			activeChildren = append(activeChildren, uid)
		}
	}
	return activeChildren, nil
}

func (m *mockManagedStoreAdapter) GetScopes() ([]string, error) {
	scopes := []string{"root", ScopePinned}
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
	item := &mockItem{uid: uid, parentID: parentUID, children: []string{}, status: StatusActive}
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
	item, ok := m.items[uid]
	if !ok {
		return fmt.Errorf("item %s not found", uid)
	}

	// Remove from parent's children list
	if item.parentID == "root" {
		newRoots := []string{}
		for _, rootUID := range m.rootUIDs {
			if rootUID != uid {
				newRoots = append(newRoots, rootUID)
			}
		}
		m.rootUIDs = newRoots
	} else if parent, ok := m.items[item.parentID]; ok {
		newChildren := []string{}
		for _, childUID := range parent.children {
			if childUID != uid {
				newChildren = append(newChildren, childUID)
			}
		}
		parent.children = newChildren
	}

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

func (m *mockManagedStoreAdapter) SetStatus(uid, status string) error {
	item, ok := m.items[uid]
	if !ok {
		return fmt.Errorf("item %s not found", uid)
	}
	item.status = status
	return nil
}

func (m *mockManagedStoreAdapter) SetPinned(uid string, isPinned bool) error {
	item, ok := m.items[uid]
	if !ok {
		return fmt.Errorf("item %s not found", uid)
	}
	item.isPinned = isPinned
	return nil
}

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
	_, _ = adapter.AddItem("root") // uid1
	_, _ = adapter.AddItem("root") // uid2
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

func TestManager_SoftDelete(t *testing.T) {
	adapter := newMockManagedStoreAdapter()
	_, _ = adapter.AddItem("root") // uid1
	_, _ = adapter.AddItem("root") // uid2

	manager, err := NewManager(adapter)
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	// Pre-check: uid2 should be at HID 2
	resolvedUID, err := manager.Registry().ResolveHID("root", 2)
	if err != nil || resolvedUID != "uid2" {
		t.Fatalf("Pre-check failed, uid2 not at HID 2")
	}

	// Soft delete uid1
	err = manager.SoftDelete("uid1", "root")
	if err != nil {
		t.Fatalf("SoftDelete failed: %v", err)
	}

	// Verify uid1 is gone from the registry scope
	_, err = manager.Registry().ResolveHID("root", 2)
	if err == nil {
		t.Error("Expected error resolving HID 2 after delete, but got nil")
	}

	// Verify uid2 is now at HID 1
	resolvedUID, err = manager.Registry().ResolveHID("root", 1)
	if err != nil || resolvedUID != "uid2" {
		t.Errorf("Expected uid2 to be at HID 1 after delete, but got %s", resolvedUID)
	}
}

func TestManager_Restore(t *testing.T) {
	adapter := newMockManagedStoreAdapter()
	_, _ = adapter.AddItem("root") // uid1
	_ = adapter.items["uid1"]
	adapter.items["uid1"].status = StatusDeleted // Manually set as deleted

	manager, err := NewManager(adapter)
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	// Pre-check: scope should be empty
	_, err = manager.Registry().ResolveHID("root", 1)
	if err == nil {
		t.Fatal("Pre-check failed, scope should be empty")
	}

	// Restore uid1
	err = manager.Restore("uid1", "root")
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	// Verify uid1 is back at HID 1
	resolvedUID, err := manager.Registry().ResolveHID("root", 1)
	if err != nil || resolvedUID != "uid1" {
		t.Errorf("Expected uid1 to be restored at HID 1, but got %s", resolvedUID)
	}
}

func TestManager_Purge(t *testing.T) {
	adapter := newMockManagedStoreAdapter()
	_, _ = adapter.AddItem("root") // uid1
	manager, err := NewManager(adapter)
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	// Purge the item
	err = manager.Purge("uid1")
	if err != nil {
		t.Fatalf("Purge failed: %v", err)
	}

	// Verify item is gone from adapter
	if _, ok := adapter.items["uid1"]; ok {
		t.Error("Item should have been removed from the adapter after purge")
	}
}

func TestManager_Pin(t *testing.T) {
	adapter := newMockManagedStoreAdapter()
	_, _ = adapter.AddItem("root") // uid1
	_, _ = adapter.AddItem("root") // uid2

	manager, err := NewManager(adapter)
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	// Pin uid2
	err = manager.Pin("uid2")
	if err != nil {
		t.Fatalf("Pin failed: %v", err)
	}

	// Verify uid2 is in the pinned scope
	resolvedUID, err := manager.Registry().ResolveHID(ScopePinned, 1)
	if err != nil || resolvedUID != "uid2" {
		t.Errorf("Expected uid2 to be at HID 1 in pinned scope, but got %s", resolvedUID)
	}

	// Verify it's still in its original scope
	resolvedUID, err = manager.Registry().ResolveHID("root", 2)
	if err != nil || resolvedUID != "uid2" {
		t.Errorf("Expected uid2 to still be at HID 2 in root scope, but got %s", resolvedUID)
	}

	// Unpin uid2
	err = manager.Unpin("uid2")
	if err != nil {
		t.Fatalf("Unpin failed: %v", err)
	}

	// Verify pinned scope is now empty
	_, err = manager.Registry().ResolveHID(ScopePinned, 1)
	if err == nil {
		t.Error("Expected pinned scope to be empty after unpinning")
	}
}

// --- Error Cases ---

func TestManager_ErrorCases(t *testing.T) {
	t.Run("Add with non-existent parent", func(t *testing.T) {
		adapter := newMockManagedStoreAdapter()
		manager, err := NewManager(adapter)
		if err != nil {
			t.Fatalf("NewManager() failed: %v", err)
		}

		_, _, err = manager.Add("non-existent")
		if err == nil {
			t.Error("Expected error when adding to non-existent parent")
		}
	})

	t.Run("Move non-existent item", func(t *testing.T) {
		adapter := newMockManagedStoreAdapter()
		manager, err := NewManager(adapter)
		if err != nil {
			t.Fatalf("NewManager() failed: %v", err)
		}

		err = manager.Move("non-existent", "root", "root")
		if err == nil {
			t.Error("Expected error when moving non-existent item")
		}
	})

	t.Run("SetStatus on non-existent item", func(t *testing.T) {
		adapter := newMockManagedStoreAdapter()
		manager, err := NewManager(adapter)
		if err != nil {
			t.Fatalf("NewManager() failed: %v", err)
		}

		err = manager.SoftDelete("non-existent", "root")
		if err == nil {
			t.Error("Expected error when soft-deleting non-existent item")
		}
	})

	t.Run("Pin non-existent item", func(t *testing.T) {
		adapter := newMockManagedStoreAdapter()
		manager, err := NewManager(adapter)
		if err != nil {
			t.Fatalf("NewManager() failed: %v", err)
		}

		err = manager.Pin("non-existent")
		if err == nil {
			t.Error("Expected error when pinning non-existent item")
		}
	})
}