package search_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/arthur-debert/too/pkg/too/commands/search"
	"github.com/arthur-debert/too/pkg/too/models"
	"github.com/arthur-debert/too/pkg/too/testutil"
	"github.com/stretchr/testify/assert"
)

func TestSearchCommand(t *testing.T) {
	t.Run("finds todos with case-insensitive search by default", func(t *testing.T) {
		// Create store with various todos using testutil
		store := testutil.CreatePopulatedStore(t,
			"Buy milk from store",
			"Call Mike about project",
			"MILK the cows",
			"Send email to team",
			"Review milestone report",
		)

		// Execute search for "milk"
		opts := search.Options{
			CollectionPath: store.Path(),
			CaseSensitive:  false,
		}
		result, err := search.Execute("milk", opts)

		// Verify results
		testutil.AssertNoError(t, err)
		assert.Equal(t, "milk", result.Query)
		assert.Equal(t, 2, len(result.MatchedTodos))
		assert.Equal(t, 5, result.TotalCount)

		// Verify matched todos using testutil
		var matchedTexts []string
		for _, todo := range result.MatchedTodos {
			matchedTexts = append(matchedTexts, todo.Text)
		}
		assert.Contains(t, matchedTexts, "Buy milk from store")
		assert.Contains(t, matchedTexts, "MILK the cows")
	})

	t.Run("performs case-sensitive search when enabled", func(t *testing.T) {
		// Create store with case variations
		store := testutil.CreatePopulatedStore(t,
			"TODO: implement feature",
			"todo: fix bug",
			"ToDo: review PR",
			"Need to do something",
		)

		// Execute case-sensitive search
		opts := search.Options{
			CollectionPath: store.Path(),
			CaseSensitive:  true,
		}
		result, err := search.Execute("todo", opts)

		// Verify only exact case match
		testutil.AssertNoError(t, err)
		assert.Equal(t, 1, len(result.MatchedTodos))
		assert.Equal(t, "todo: fix bug", result.MatchedTodos[0].Text)
		assert.Equal(t, 4, result.TotalCount)
	})

	t.Run("searches across both pending and done todos", func(t *testing.T) {
		// Create store with mixed status todos using testutil
		store := testutil.CreateStoreWithSpecs(t, []testutil.TodoSpec{
			{Text: "Pending: buy coffee", Status: models.StatusPending},
			{Text: "Done: bought coffee maker", Status: models.StatusDone},
			{Text: "Pending: meeting at coffee shop", Status: models.StatusPending},
			{Text: "Done: cleaned coffee pot", Status: models.StatusDone},
			{Text: "Pending: order tea", Status: models.StatusPending},
		})

		// Search for "coffee"
		opts := search.Options{
			CollectionPath: store.Path(),
			CaseSensitive:  false,
		}
		result, err := search.Execute("coffee", opts)

		// Verify results include both statuses
		testutil.AssertNoError(t, err)
		assert.Equal(t, 4, len(result.MatchedTodos))
		assert.Equal(t, 5, result.TotalCount)

		// Count by status
		pendingCount := 0
		doneCount := 0
		for _, todo := range result.MatchedTodos {
			switch todo.GetStatus() {
			case models.StatusPending:
				pendingCount++
			case models.StatusDone:
				doneCount++
			}
		}
		assert.Equal(t, 2, pendingCount)
		assert.Equal(t, 2, doneCount)
	})

	t.Run("handles partial word matches", func(t *testing.T) {
		// Create store with todos
		store := testutil.CreatePopulatedStore(t,
			"Implement authentication",
			"Add author field",
			"Authorize user access",
			"Create bibliography",
		)

		// Search for "auth"
		opts := search.Options{
			CollectionPath: store.Path(),
			CaseSensitive:  false,
		}
		result, err := search.Execute("auth", opts)

		// Verify partial matches
		testutil.AssertNoError(t, err)
		assert.Equal(t, 3, len(result.MatchedTodos))

		// Verify specific matches
		var matchedTexts []string
		for _, todo := range result.MatchedTodos {
			matchedTexts = append(matchedTexts, todo.Text)
		}
		assert.Contains(t, matchedTexts, "Implement authentication")
		assert.Contains(t, matchedTexts, "Add author field")
		assert.Contains(t, matchedTexts, "Authorize user access")
		assert.NotContains(t, matchedTexts, "Create bibliography")
	})

	t.Run("returns empty results for no matches", func(t *testing.T) {
		// Create store with todos
		store := testutil.CreatePopulatedStore(t,
			"First todo",
			"Second todo",
			"Third todo",
		)

		// Search for non-existent text
		opts := search.Options{
			CollectionPath: store.Path(),
			CaseSensitive:  false,
		}
		result, err := search.Execute("xyz", opts)

		// Verify empty results
		testutil.AssertNoError(t, err)
		assert.Equal(t, "xyz", result.Query)
		assert.Empty(t, result.MatchedTodos)
		assert.Equal(t, 3, result.TotalCount) // Total count still shows all todos
	})

	t.Run("returns error for empty query", func(t *testing.T) {
		// Create store
		store := testutil.CreatePopulatedStore(t, "Some todo")

		// Try empty search
		opts := search.Options{
			CollectionPath: store.Path(),
			CaseSensitive:  false,
		}
		result, err := search.Execute("", opts)

		// Verify error
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "search query cannot be empty")
	})

	t.Run("handles special characters in search", func(t *testing.T) {
		// Create store with special characters
		store := testutil.CreatePopulatedStore(t,
			"Task with (parentheses)",
			"Email: user@example.com",
			"Price: $99.99",
			"C++ programming task",
			"Normal task without special chars",
		)

		// Test various special character searches
		testCases := []struct {
			query         string
			expectedCount int
		}{
			{"(parentheses)", 1},
			{"@example", 1},
			{"$99", 1},
			{"C++", 1},
		}

		opts := search.Options{
			CollectionPath: store.Path(),
			CaseSensitive:  false,
		}

		for _, tc := range testCases {
			result, err := search.Execute(tc.query, opts)
			testutil.AssertNoError(t, err)
			assert.Equal(t, tc.expectedCount, len(result.MatchedTodos),
				"Query '%s' should match %d todos", tc.query, tc.expectedCount)
		}
	})

	t.Run("returns error when store operation fails", func(t *testing.T) {
		// Create a read-only directory to force a store error
		dir := t.TempDir()
		dbPath := filepath.Join(dir, "todos.json")

		// Write invalid JSON to cause a Find error
		err := os.WriteFile(dbPath, []byte("invalid json"), 0644)
		assert.NoError(t, err)

		opts := search.Options{
			CollectionPath: dbPath,
			CaseSensitive:  false,
		}
		result, err := search.Execute("test", opts)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
