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
		assert.Equal(t, expectedPath, result.Text)
	})

	t.Run("shows default path when collection path is empty", func(t *testing.T) {
		// Execute with empty collection path
		opts := datapath.Options{CollectionPath: ""}
		result, err := datapath.Execute(opts)

		// Assert
		testutil.AssertNoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.Text)
		// Should be .todos.json in current dir or home
		assert.True(t,
			filepath.Base(result.Text) == ".todos.json",
			"Expected path to end with .todos.json, got %s", result.Text)
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
		assert.True(t, filepath.IsAbs(result.Text), "Expected absolute path, got %s", result.Text)
	})
}
