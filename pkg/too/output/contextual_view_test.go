package output

import (
	"testing"

	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildContextualView(t *testing.T) {
	// Helper to create a todo
	createTodo := func(uid, text string) *models.Todo {
		return &models.Todo{
			UID:  uid,
			Text: text,
		}
	}

	// Helper to create hierarchical todo
	createHierarchical := func(todo *models.Todo) *models.HierarchicalTodo {
		return &models.HierarchicalTodo{
			Todo: todo,
		}
	}

	tests := []struct {
		name               string
		hierarchy          []*models.HierarchicalTodo
		highlightID        string
		expectedNil        bool
		checkFunc          func(t *testing.T, result *ContextualNode)
	}{
		{
			name: "highlight not found returns nil",
			hierarchy: []*models.HierarchicalTodo{
				createHierarchical(createTodo("1", "Item 1")),
			},
			highlightID: "non-existent",
			expectedNil: true,
		},
		{
			name: "single item highlighted",
			hierarchy: []*models.HierarchicalTodo{
				createHierarchical(createTodo("1", "Item 1")),
			},
			highlightID: "1",
			checkFunc: func(t *testing.T, result *ContextualNode) {
				assert.Equal(t, "1", result.Todo.UID)
				assert.Empty(t, result.SiblingsBefore)
				assert.Empty(t, result.SiblingsAfter)
				assert.False(t, result.ShowEllipsisBefore)
				assert.False(t, result.ShowEllipsisAfter)
			},
		},
		{
			name: "first item in list of 5",
			hierarchy: []*models.HierarchicalTodo{
				createHierarchical(createTodo("1", "Item 1")),
				createHierarchical(createTodo("2", "Item 2")),
				createHierarchical(createTodo("3", "Item 3")),
				createHierarchical(createTodo("4", "Item 4")),
				createHierarchical(createTodo("5", "Item 5")),
			},
			highlightID: "1",
			checkFunc: func(t *testing.T, result *ContextualNode) {
				assert.Equal(t, "1", result.Todo.UID)
				assert.Empty(t, result.SiblingsBefore)
				assert.Len(t, result.SiblingsAfter, 2) // Should show 2 items after
				assert.Equal(t, "2", result.SiblingsAfter[0].UID)
				assert.Equal(t, "3", result.SiblingsAfter[1].UID)
				assert.False(t, result.ShowEllipsisBefore)
				assert.True(t, result.ShowEllipsisAfter) // More items exist after
			},
		},
		{
			name: "middle item in list of 5",
			hierarchy: []*models.HierarchicalTodo{
				createHierarchical(createTodo("1", "Item 1")),
				createHierarchical(createTodo("2", "Item 2")),
				createHierarchical(createTodo("3", "Item 3")),
				createHierarchical(createTodo("4", "Item 4")),
				createHierarchical(createTodo("5", "Item 5")),
			},
			highlightID: "3",
			checkFunc: func(t *testing.T, result *ContextualNode) {
				assert.Equal(t, "3", result.Todo.UID)
				assert.Len(t, result.SiblingsBefore, 2) // Items 1 and 2
				assert.Len(t, result.SiblingsAfter, 2)  // Items 4 and 5
				assert.False(t, result.ShowEllipsisBefore) // No items before 1
				assert.False(t, result.ShowEllipsisAfter)  // No items after 5
			},
		},
		{
			name: "last item in list of 5",
			hierarchy: []*models.HierarchicalTodo{
				createHierarchical(createTodo("1", "Item 1")),
				createHierarchical(createTodo("2", "Item 2")),
				createHierarchical(createTodo("3", "Item 3")),
				createHierarchical(createTodo("4", "Item 4")),
				createHierarchical(createTodo("5", "Item 5")),
			},
			highlightID: "5",
			checkFunc: func(t *testing.T, result *ContextualNode) {
				assert.Equal(t, "5", result.Todo.UID)
				assert.Len(t, result.SiblingsBefore, 2) // Should show 2 items before
				assert.Equal(t, "3", result.SiblingsBefore[0].UID)
				assert.Equal(t, "4", result.SiblingsBefore[1].UID)
				assert.Empty(t, result.SiblingsAfter)
				assert.True(t, result.ShowEllipsisBefore) // More items exist before
				assert.False(t, result.ShowEllipsisAfter)
			},
		},
		{
			name: "nested item highlighted",
			hierarchy: []*models.HierarchicalTodo{
				{
					Todo: createTodo("1", "Parent"),
					Children: []*models.HierarchicalTodo{
						createHierarchical(createTodo("1.1", "Child 1")),
						createHierarchical(createTodo("1.2", "Child 2")),
						createHierarchical(createTodo("1.3", "Child 3")),
					},
				},
			},
			highlightID: "1.2",
			checkFunc: func(t *testing.T, result *ContextualNode) {
				require.NotNil(t, result)
				assert.Equal(t, "1", result.Todo.UID) // Root should be parent
				assert.Empty(t, result.SiblingsBefore)
				assert.Empty(t, result.SiblingsAfter)
				
				// Check child context
				require.Len(t, result.Children, 1)
				child := result.Children[0]
				assert.Equal(t, "1.2", child.Todo.UID)
				assert.Len(t, child.SiblingsBefore, 1)
				assert.Equal(t, "1.1", child.SiblingsBefore[0].UID)
				assert.Len(t, child.SiblingsAfter, 1)
				assert.Equal(t, "1.3", child.SiblingsAfter[0].UID)
			},
		},
		{
			name: "deeply nested item",
			hierarchy: []*models.HierarchicalTodo{
				{
					Todo: createTodo("1", "Root"),
					Children: []*models.HierarchicalTodo{
						{
							Todo: createTodo("1.1", "Level 1"),
							Children: []*models.HierarchicalTodo{
								createHierarchical(createTodo("1.1.1", "Deep 1")),
								createHierarchical(createTodo("1.1.2", "Deep 2")),
								createHierarchical(createTodo("1.1.3", "Deep 3")),
								createHierarchical(createTodo("1.1.4", "Deep 4")),
							},
						},
					},
				},
			},
			highlightID: "1.1.3",
			checkFunc: func(t *testing.T, result *ContextualNode) {
				require.NotNil(t, result)
				assert.Equal(t, "1", result.Todo.UID)
				
				require.Len(t, result.Children, 1)
				level1 := result.Children[0]
				assert.Equal(t, "1.1", level1.Todo.UID)
				
				require.Len(t, level1.Children, 1)
				deepNode := level1.Children[0]
				assert.Equal(t, "1.1.3", deepNode.Todo.UID)
				assert.Len(t, deepNode.SiblingsBefore, 2)
				assert.Len(t, deepNode.SiblingsAfter, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildContextualView(tt.hierarchy, tt.highlightID)
			
			if tt.expectedNil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

func TestFindPathToHighlight(t *testing.T) {
	hierarchy := []*models.HierarchicalTodo{
		{
			Todo: &models.Todo{UID: "1"},
			Children: []*models.HierarchicalTodo{
				{
					Todo: &models.Todo{UID: "1.1"},
					Children: []*models.HierarchicalTodo{
						{Todo: &models.Todo{UID: "1.1.1"}},
					},
				},
			},
		},
		{
			Todo: &models.Todo{UID: "2"},
		},
	}

	tests := []struct {
		name        string
		highlightID string
		expected    []string
	}{
		{
			name:        "root level",
			highlightID: "2",
			expected:    []string{"2"},
		},
		{
			name:        "nested level",
			highlightID: "1.1",
			expected:    []string{"1", "1.1"},
		},
		{
			name:        "deeply nested",
			highlightID: "1.1.1",
			expected:    []string{"1", "1.1", "1.1.1"},
		},
		{
			name:        "not found",
			highlightID: "non-existent",
			expected:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findPathToHighlight(hierarchy, tt.highlightID, []string{})
			assert.Equal(t, tt.expected, result)
		})
	}
}