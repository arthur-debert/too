package idm

import (
	"fmt"
	"testing"
)

// mockStoreAdapter is a test implementation of the StoreAdapter interface.
// It allows us to simulate a data store for testing the Registry.
type mockStoreAdapter struct {
	data map[string][]string // map[parentUID][]childUIDs
}

// newMockStoreAdapter creates a new mock adapter.
func newMockStoreAdapter() *mockStoreAdapter {
	return &mockStoreAdapter{
		data: make(map[string][]string),
	}
}

// GetChildren implements the StoreAdapter interface for the mock.
func (m *mockStoreAdapter) GetChildren(parentUID string) ([]string, error) {
	children, ok := m.data[parentUID]
	if !ok {
		return nil, fmt.Errorf("parent UID '%s' not found in mock store", parentUID)
	}
	return children, nil
}

// GetScopes implements the StoreAdapter interface for the mock.
func (m *mockStoreAdapter) GetScopes() ([]string, error) {
	scopes := make([]string, 0, len(m.data))
	for scope := range m.data {
		scopes = append(scopes, scope)
	}
	return scopes, nil
}

// --- Test Cases ---

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	if r.scopes == nil {
		t.Fatal("NewRegistry() failed to initialize the scopes map")
	}
}

func TestRebuildScope(t *testing.T) {
	adapter := newMockStoreAdapter()
	adapter.data["root"] = []string{"uid1", "uid2", "uid3"}

	r := NewRegistry()
	err := r.RebuildScope(adapter, "root")

	if err != nil {
		t.Fatalf("RebuildScope() returned an unexpected error: %v", err)
	}

	if len(r.scopes["root"]) != 3 {
		t.Errorf("Expected scope 'root' to have 3 UIDs, but got %d", len(r.scopes["root"]))
	}

	if r.scopes["root"][0] != "uid1" {
		t.Errorf("Expected first UID to be 'uid1', but got '%s'", r.scopes["root"][0])
	}
}

func TestAdd(t *testing.T) {
	r := NewRegistry()
	r.scopes["active"] = []string{"uidA"}

	newHID := r.Add("active", "uidB")
	if newHID != 2 {
		t.Errorf("Expected new HID to be 2, but got %d", newHID)
	}

	if len(r.scopes["active"]) != 2 || r.scopes["active"][1] != "uidB" {
		t.Error("Add() failed to append the new UID correctly")
	}

	newHID = r.Add("new_scope", "uidC")
	if newHID != 1 {
		t.Errorf("Expected new HID for a new scope to be 1, but got %d", newHID)
	}
	if len(r.scopes["new_scope"]) != 1 || r.scopes["new_scope"][0] != "uidC" {
		t.Error("Add() failed to create a new scope correctly")
	}
}

func TestRemove(t *testing.T) {
	r := NewRegistry()
	r.scopes["active"] = []string{"uidA", "uidB", "uidC"}

	r.Remove("active", "uidB")

	if len(r.scopes["active"]) != 2 {
		t.Fatalf("Expected scope 'active' to have 2 UIDs after removal, but got %d", len(r.scopes["active"]))
	}

	if r.scopes["active"][0] != "uidA" || r.scopes["active"][1] != "uidC" {
		t.Errorf("Remove() resulted in incorrect order: got %v", r.scopes["active"])
	}

	// Test removing a non-existent UID
	r.Remove("active", "uidZ")
	if len(r.scopes["active"]) != 2 {
		t.Error("Removing a non-existent UID should not change the scope")
	}

	// Test removing from a non-existent scope
	r.Remove("inactive", "uidA") // Should not panic
}

func TestResolveHID(t *testing.T) {
	r := NewRegistry()
	r.scopes["root"] = []string{"uid1", "uid2", "uid3"}

	// Test valid HID
	uid, err := r.ResolveHID("root", 2)
	if err != nil {
		t.Fatalf("ResolveHID() returned an unexpected error for a valid HID: %v", err)
	}
	if uid != "uid2" {
		t.Errorf("Expected to resolve HID 2 to 'uid2', but got '%s'", uid)
	}

	// Test invalid HID (out of bounds)
	_, err = r.ResolveHID("root", 4)
	if err == nil {
		t.Error("Expected an error for an out-of-bounds HID, but got nil")
	}

	// Test invalid HID (zero)
	_, err = r.ResolveHID("root", 0)
	if err == nil {
		t.Error("Expected an error for HID 0, but got nil")
	}

	// Test scope not found
	_, err = r.ResolveHID("nonexistent", 1)
	if err == nil {
		t.Error("Expected an error for a non-existent scope, but got nil")
	}
}

func TestGetUIDs(t *testing.T) {
	r := NewRegistry()
	r.scopes["active"] = []string{"uidA", "uidB"}
	r.scopes["pinned"] = []string{"uidC"}
	r.scopes["archived"] = []string{"uidD", "uidE"}

	// Test single scope
	uids := r.GetUIDs("active")
	if len(uids) != 2 || uids[0] != "uidA" || uids[1] != "uidB" {
		t.Errorf("GetUIDs() with single scope failed. Got: %v", uids)
	}

	// Test multiple scopes
	uids = r.GetUIDs("pinned", "active")
	expected := []string{"uidC", "uidA", "uidB"}
	if len(uids) != 3 {
		t.Fatalf("GetUIDs() with multiple scopes returned wrong number of UIDs. Got: %d", len(uids))
	}
	for i, uid := range expected {
		if uids[i] != uid {
			t.Errorf("GetUIDs() with multiple scopes failed at index %d. Expected '%s', got '%s'", i, uid, uids[i])
		}
	}

	// Test with non-existent scope
	uids = r.GetUIDs("active", "nonexistent")
	if len(uids) != 2 {
		t.Errorf("GetUIDs() should ignore non-existent scopes. Expected 2 UIDs, got %d", len(uids))
	}

	// Test empty call
	uids = r.GetUIDs()
	if len(uids) != 0 {
		t.Errorf("GetUIDs() with no arguments should return an empty slice. Got: %d", len(uids))
	}
}
