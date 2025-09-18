package store_test

import (
	"path/filepath"
	"testing"

	"github.com/arthur-debert/too/pkg/too/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to create test adapter for hierarchy tests
func createHierarchyTestAdapter(t *testing.T) (*store.NanoStoreAdapter, func()) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.json")
	
	adapter, err := store.NewNanoStoreAdapter(dbPath)
	require.NoError(t, err)
	
	cleanup := func() {
		adapter.Close()
	}
	
	return adapter, cleanup
}

func TestHierarchyQueries(t *testing.T) {
	adapter, cleanup := createHierarchyTestAdapter(t)
	defer cleanup()

	// Create a hierarchy: Root -> Child1, Child2 -> Grandchild
	root, err := adapter.Add("Root Task", nil)
	require.NoError(t, err)
	
	rootRef := root.PositionPath
	child1, err := adapter.Add("Child Task 1", &rootRef)
	require.NoError(t, err)
	
	_, err = adapter.Add("Child Task 2", &rootRef)
	require.NoError(t, err)
	
	child1Ref := child1.PositionPath
	grandchild, err := adapter.Add("Grandchild Task", &child1Ref)
	require.NoError(t, err)

	t.Run("GetChildrenOf returns direct children only", func(t *testing.T) {
		children, err := adapter.GetChildrenOf(root.PositionPath)
		require.NoError(t, err)
		
		assert.Len(t, children, 2)
		
		// Verify we got both children
		childTexts := make(map[string]bool)
		for _, child := range children {
			childTexts[child.Text] = true
		}
		assert.True(t, childTexts["Child Task 1"])
		assert.True(t, childTexts["Child Task 2"])
		
		// Verify grandchild is not included
		for _, child := range children {
			assert.NotEqual(t, "Grandchild Task", child.Text)
		}
	})

	t.Run("GetDescendantsOf returns all descendants", func(t *testing.T) {
		descendants, err := adapter.GetDescendantsOf(root.PositionPath)
		require.NoError(t, err)
		
		assert.Len(t, descendants, 3) // Child1, Child2, Grandchild
		
		// Verify we got all descendants
		descendantTexts := make(map[string]bool)
		for _, desc := range descendants {
			descendantTexts[desc.Text] = true
		}
		assert.True(t, descendantTexts["Child Task 1"])
		assert.True(t, descendantTexts["Child Task 2"])
		assert.True(t, descendantTexts["Grandchild Task"])
	})

	t.Run("GetSiblingsOf returns todos with same parent", func(t *testing.T) {
		siblings, err := adapter.GetSiblingsOf(child1.PositionPath)
		require.NoError(t, err)
		
		assert.Len(t, siblings, 1)
		assert.Equal(t, "Child Task 2", siblings[0].Text)
		
		// Child1 should not be included in its own siblings
		for _, sibling := range siblings {
			assert.NotEqual(t, child1.UID, sibling.UID)
		}
	})

	t.Run("GetSiblingsOf with no siblings returns empty", func(t *testing.T) {
		siblings, err := adapter.GetSiblingsOf(grandchild.PositionPath)
		require.NoError(t, err)
		
		assert.Len(t, siblings, 0)
	})

	t.Run("GetChildrenOf with no children returns empty", func(t *testing.T) {
		children, err := adapter.GetChildrenOf(grandchild.PositionPath)
		require.NoError(t, err)
		
		assert.Len(t, children, 0)
	})

	t.Run("GetChildrenOf with invalid ID returns error", func(t *testing.T) {
		_, err := adapter.GetChildrenOf("nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to resolve parent ID")
	})
}

func TestHierarchyQueriesWithRootSiblings(t *testing.T) {
	adapter, cleanup := createHierarchyTestAdapter(t)
	defer cleanup()

	// Create multiple root-level todos
	root1, err := adapter.Add("Root Task 1", nil)
	require.NoError(t, err)
	
	_, err = adapter.Add("Root Task 2", nil)
	require.NoError(t, err)
	
	_, err = adapter.Add("Root Task 3", nil)
	require.NoError(t, err)

	t.Run("Root todos have siblings", func(t *testing.T) {
		siblings, err := adapter.GetSiblingsOf(root1.PositionPath)
		require.NoError(t, err)
		
		assert.Len(t, siblings, 2)
		
		// Verify we got the other root todos
		siblingTexts := make(map[string]bool)
		for _, sibling := range siblings {
			siblingTexts[sibling.Text] = true
		}
		assert.True(t, siblingTexts["Root Task 2"])
		assert.True(t, siblingTexts["Root Task 3"])
		
		// Root1 should not be included in its own siblings
		for _, sibling := range siblings {
			assert.NotEqual(t, root1.UID, sibling.UID)
		}
	})
}