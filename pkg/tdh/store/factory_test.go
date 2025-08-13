package store_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/arthur-debert/tdh/pkg/tdh/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T) {
	// Save original values
	originalWd, _ := os.Getwd()
	originalEnv := os.Getenv("TODO_DB_PATH")

	// Restore after test
	t.Cleanup(func() {
		_ = os.Chdir(originalWd)
		if originalEnv != "" {
			_ = os.Setenv("TODO_DB_PATH", originalEnv)
		} else {
			_ = os.Unsetenv("TODO_DB_PATH")
		}
		store.ResetCache()
	})
}

func TestPathResolution(t *testing.T) {
	t.Run("should find .todos in current directory", func(t *testing.T) {
		setupTest(t)
		// Create temp directory with .todos file
		dir, err := os.MkdirTemp("", "tdh-path-test")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(dir) }()

		todosPath := filepath.Join(dir, ".todos")
		err = os.WriteFile(todosPath, []byte("[]"), 0600)
		require.NoError(t, err)

		// Change to test directory
		err = os.Chdir(dir)
		require.NoError(t, err)

		// Test
		s := store.NewStore("")
		// Resolve symlinks for comparison (macOS temp dirs may have symlinks)
		expectedPath, _ := filepath.EvalSymlinks(todosPath)
		actualPath, _ := filepath.EvalSymlinks(s.Path())
		assert.Equal(t, expectedPath, actualPath)
	})

	t.Run("should find .todos in parent directory", func(t *testing.T) {
		setupTest(t)
		// Create parent directory with .todos file
		parentDir, err := os.MkdirTemp("", "tdh-parent-test")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(parentDir) }()

		todosPath := filepath.Join(parentDir, ".todos")
		err = os.WriteFile(todosPath, []byte("[]"), 0600)
		require.NoError(t, err)

		// Create subdirectory
		subDir := filepath.Join(parentDir, "subdir")
		err = os.Mkdir(subDir, 0755)
		require.NoError(t, err)

		// Change to subdirectory
		err = os.Chdir(subDir)
		require.NoError(t, err)

		// Test
		s := store.NewStore("")
		// Resolve symlinks for comparison (macOS temp dirs may have symlinks)
		expectedPath, _ := filepath.EvalSymlinks(todosPath)
		actualPath, _ := filepath.EvalSymlinks(s.Path())
		assert.Equal(t, expectedPath, actualPath)
	})

	t.Run("should use TODO_DB_PATH environment variable", func(t *testing.T) {
		setupTest(t)
		// Set environment variable
		envPath := "/custom/path/todos.json"
		err := os.Setenv("TODO_DB_PATH", envPath)
		require.NoError(t, err)

		// Change to directory without .todos
		dir, err := os.MkdirTemp("", "tdh-env-test")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(dir) }()

		err = os.Chdir(dir)
		require.NoError(t, err)

		// Test
		s := store.NewStore("")
		assert.Equal(t, envPath, s.Path())
	})

	t.Run("should fall back to home directory", func(t *testing.T) {
		setupTest(t)
		// Clear environment variable
		err := os.Unsetenv("TODO_DB_PATH")
		require.NoError(t, err)

		// Change to directory without .todos
		dir, err := os.MkdirTemp("", "tdh-home-test")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(dir) }()

		err = os.Chdir(dir)
		require.NoError(t, err)

		// Test
		s := store.NewStore("")
		home, _ := os.UserHomeDir()
		expectedPath := filepath.Join(home, ".todos.json")
		assert.Equal(t, expectedPath, s.Path())
	})

	t.Run("should use provided path if not empty", func(t *testing.T) {
		setupTest(t)
		customPath := "/custom/todos.json"
		s := store.NewStore(customPath)
		assert.Equal(t, customPath, s.Path())
	})
}
