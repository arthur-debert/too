package too_test

import (
	"path/filepath"
	"testing"

	"github.com/arthur-debert/too/pkg/too"
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper to create test engine
func createTestEngine(t *testing.T) *too.NanoEngine {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.json")
	
	engine, err := too.NewNanoEngine(dbPath)
	require.NoError(t, err)
	
	return engine
}

func TestParentAutoCompleteWhenChildCompleted(t *testing.T) {
	// Setup: Create parent with single child
	engine := createTestEngine(t)
	defer engine.Close()

	// Add parent and child
	parent, err := engine.Add("Parent Task", nil)
	require.NoError(t, err)
	
	parentRef := parent.PositionPath
	child, err := engine.Add("Child Task", &parentRef)
	require.NoError(t, err)

	// Complete the child - parent should auto-complete
	_, err = engine.MutateAttributeByUUID(child.UID, models.AttributeCompletion, string(models.StatusDone))
	require.NoError(t, err)

	// Verify parent is now completed
	updatedParent, err := engine.GetTodoByUID(parent.UID)
	require.NoError(t, err)
	assert.Equal(t, models.StatusDone, updatedParent.GetStatus())
}

func TestParentStaysPendingWhenSomeChildrenPending(t *testing.T) {
	// Setup: Create parent with two children
	engine := createTestEngine(t)
	defer engine.Close()

	// Add parent and two children
	parent, err := engine.Add("Parent Task", nil)
	require.NoError(t, err)
	
	parentRef := parent.PositionPath
	child1, err := engine.Add("Child Task 1", &parentRef)
	require.NoError(t, err)
	
	child2, err := engine.Add("Child Task 2", &parentRef)
	require.NoError(t, err)

	// Complete only one child
	_, err = engine.MutateAttributeByUUID(child1.UID, models.AttributeCompletion, string(models.StatusDone))
	require.NoError(t, err)

	// Verify parent is still pending
	updatedParent, err := engine.GetTodoByUID(parent.UID)
	require.NoError(t, err)
	assert.Equal(t, models.StatusPending, updatedParent.GetStatus())

	// Complete second child - now parent should auto-complete
	_, err = engine.MutateAttributeByUUID(child2.UID, models.AttributeCompletion, string(models.StatusDone))
	require.NoError(t, err)

	// Verify parent is now completed
	updatedParent, err = engine.GetTodoByUID(parent.UID)
	require.NoError(t, err)
	assert.Equal(t, models.StatusDone, updatedParent.GetStatus())
}

func TestChildrenAutoCompleteWhenParentCompleted(t *testing.T) {
	// Setup: Create parent with children
	engine := createTestEngine(t)
	defer engine.Close()

	// Add parent and two children
	parent, err := engine.Add("Parent Task", nil)
	require.NoError(t, err)
	
	parentRef := parent.PositionPath
	child1, err := engine.Add("Child Task 1", &parentRef)
	require.NoError(t, err)
	
	child2, err := engine.Add("Child Task 2", &parentRef)
	require.NoError(t, err)

	// Complete the parent - children should auto-complete
	_, err = engine.MutateAttributeByUUID(parent.UID, models.AttributeCompletion, string(models.StatusDone))
	require.NoError(t, err)

	// Verify children are now completed
	updatedChild1, err := engine.GetTodoByUID(child1.UID)
	require.NoError(t, err)
	assert.Equal(t, models.StatusDone, updatedChild1.GetStatus())

	updatedChild2, err := engine.GetTodoByUID(child2.UID)
	require.NoError(t, err)
	assert.Equal(t, models.StatusDone, updatedChild2.GetStatus())
}

func TestDeepHierarchyAutoUpdate(t *testing.T) {
	// Setup: Create grandparent -> parent -> child hierarchy
	engine := createTestEngine(t)
	defer engine.Close()

	// Add grandparent
	grandparent, err := engine.Add("Grandparent Task", nil)
	require.NoError(t, err)
	
	// Add parent under grandparent
	grandparentRef := grandparent.PositionPath
	parent, err := engine.Add("Parent Task", &grandparentRef)
	require.NoError(t, err)
	
	// Add child under parent
	parentRef := parent.PositionPath
	child, err := engine.Add("Child Task", &parentRef)
	require.NoError(t, err)

	// Complete the child - should bubble up to grandparent
	_, err = engine.MutateAttributeByUUID(child.UID, models.AttributeCompletion, string(models.StatusDone))
	require.NoError(t, err)

	// Verify all levels are completed
	updatedChild, err := engine.GetTodoByUID(child.UID)
	require.NoError(t, err)
	assert.Equal(t, models.StatusDone, updatedChild.GetStatus())

	updatedParent, err := engine.GetTodoByUID(parent.UID)
	require.NoError(t, err)
	assert.Equal(t, models.StatusDone, updatedParent.GetStatus())

	updatedGrandparent, err := engine.GetTodoByUID(grandparent.UID)
	require.NoError(t, err)
	assert.Equal(t, models.StatusDone, updatedGrandparent.GetStatus())
}

func TestParentReopenWhenChildReopened(t *testing.T) {
	// Setup: Create parent with completed child, then reopen child
	engine := createTestEngine(t)
	defer engine.Close()

	// Add parent and child
	parent, err := engine.Add("Parent Task", nil)
	require.NoError(t, err)
	
	parentRef := parent.PositionPath
	child, err := engine.Add("Child Task", &parentRef)
	require.NoError(t, err)

	// Complete child (which auto-completes parent)
	_, err = engine.MutateAttributeByUUID(child.UID, models.AttributeCompletion, string(models.StatusDone))
	require.NoError(t, err)

	// Verify both are completed
	updatedParent, err := engine.GetTodoByUID(parent.UID)
	require.NoError(t, err)
	assert.Equal(t, models.StatusDone, updatedParent.GetStatus())

	// Reopen child - parent should auto-reopen
	_, err = engine.MutateAttributeByUUID(child.UID, models.AttributeCompletion, string(models.StatusPending))
	require.NoError(t, err)

	// Verify parent is now pending
	updatedParent, err = engine.GetTodoByUID(parent.UID)
	require.NoError(t, err)
	assert.Equal(t, models.StatusPending, updatedParent.GetStatus())
}

func TestNestedChildrenAutoComplete(t *testing.T) {
	// Setup: Create parent with nested children
	engine := createTestEngine(t)
	defer engine.Close()

	// Add parent
	parent, err := engine.Add("Parent Task", nil)
	require.NoError(t, err)
	
	// Add child under parent
	parentRef := parent.PositionPath
	child, err := engine.Add("Child Task", &parentRef)
	require.NoError(t, err)
	
	// Add grandchild under child
	childRef := child.PositionPath
	grandchild, err := engine.Add("Grandchild Task", &childRef)
	require.NoError(t, err)

	// Complete parent - should auto-complete all descendants
	_, err = engine.MutateAttributeByUUID(parent.UID, models.AttributeCompletion, string(models.StatusDone))
	require.NoError(t, err)

	// Verify all descendants are completed
	updatedChild, err := engine.GetTodoByUID(child.UID)
	require.NoError(t, err)
	assert.Equal(t, models.StatusDone, updatedChild.GetStatus())

	updatedGrandchild, err := engine.GetTodoByUID(grandchild.UID)
	require.NoError(t, err)
	assert.Equal(t, models.StatusDone, updatedGrandchild.GetStatus())
}