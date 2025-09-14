package datapath_test

import (
	"path/filepath"
	"testing"

	"github.com/arthur-debert/too/pkg/too/commands/datapath"
	"github.com/arthur-debert/too/pkg/too/testutil"
	"github.com/stretchr/testify/assert"
)

func TestDataPath(t *testing.T) {
	t.Run("shows specified collection path", func(t *testing.T) {
		// Setup
		_, expectedPath := testutil.CreatePopulatedStore(t)

		// Execute
		opts := datapath.Options{CollectionPath: expectedPath}
		result, err := datapath.Execute(opts)

		// Assert
		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedPath, result.Path)
	})

	t.Run("shows default path when collection path is empty", func(t *testing.T) {
		// Execute with empty collection path
		opts := datapath.Options{CollectionPath: ""}
		result, err := datapath.Execute(opts)

		// Assert
		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.Path)
		// Should be .todos.db in current dir or home
		assert.True(t,
			filepath.Base(result.Path) == ".todos.db",
			"Expected path to end with .todos.db, got %s", result.Path)
	})

	t.Run("returns absolute path", func(t *testing.T) {
		// Setup
		_, dbPath := testutil.CreatePopulatedStore(t)

		// Execute
		opts := datapath.Options{CollectionPath: dbPath}
		result, err := datapath.Execute(opts)

		// Assert
		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, filepath.IsAbs(result.Path), "Expected absolute path, got %s", result.Path)
	})
}
