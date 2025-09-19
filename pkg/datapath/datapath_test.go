package datapath_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/arthur-debert/too/pkg/too/commands/datapath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveCollectionPath(t *testing.T) {
	t.Run("empty path uses scope-based resolution", func(t *testing.T) {
		// Save current env
		oldEnv := os.Getenv("TODO_DB_PATH")
		if err := os.Unsetenv("TODO_DB_PATH"); err != nil {
			t.Fatalf("Failed to unset TODO_DB_PATH: %v", err)
		}
		defer func() {
			if err := os.Setenv("TODO_DB_PATH", oldEnv); err != nil {
				t.Errorf("Failed to restore TODO_DB_PATH: %v", err)
			}
		}()
		
		path := datapath.ResolveCollectionPath("")
		// Should return a path (either global or project)
		assert.NotEmpty(t, path)
	})

	t.Run("explicit path stays as-is", func(t *testing.T) {
		path := datapath.ResolveCollectionPath("my-todos.json")
		assert.Equal(t, "my-todos.json", path)
	})

	t.Run("absolute path stays absolute", func(t *testing.T) {
		absPath := "/tmp/todos.json"
		path := datapath.ResolveCollectionPath(absPath)
		assert.Equal(t, absPath, path)
	})

	t.Run("tilde expansion for home directory", func(t *testing.T) {
		home, err := os.UserHomeDir()
		require.NoError(t, err)

		path := datapath.ResolveCollectionPath("~/my-todos.json")
		expected := filepath.Join(home, "my-todos.json")
		assert.Equal(t, expected, path)
	})
	
	t.Run("respects TODO_DB_PATH env var", func(t *testing.T) {
		// Save current env
		oldEnv := os.Getenv("TODO_DB_PATH")
		testPath := "/custom/path/todos.json"
		if err := os.Setenv("TODO_DB_PATH", testPath); err != nil {
			t.Fatalf("Failed to set TODO_DB_PATH: %v", err)
		}
		defer func() {
			if err := os.Setenv("TODO_DB_PATH", oldEnv); err != nil {
				t.Errorf("Failed to restore TODO_DB_PATH: %v", err)
			}
		}()
		
		path := datapath.ResolveCollectionPath("")
		assert.Equal(t, testPath, path)
	})
}

func TestResolveScopedPath(t *testing.T) {
	t.Run("project scope in git repo", func(t *testing.T) {
		tmpDir := t.TempDir()
		gitDir := filepath.Join(tmpDir, ".git")
		err := os.Mkdir(gitDir, 0755)
		require.NoError(t, err)

		// Change to git repo
		origWd, _ := os.Getwd()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("Failed to change directory: %v", err)
		}
		defer func() {
			if err := os.Chdir(origWd); err != nil {
				t.Errorf("Failed to restore working directory: %v", err)
			}
		}()

		path, isGlobal := datapath.ResolveScopedPath(false)
		assert.False(t, isGlobal)
		// Use EvalSymlinks to handle macOS /var -> /private/var
		expectedPath, _ := filepath.EvalSymlinks(filepath.Join(tmpDir, ".todos.json"))
		actualPath, _ := filepath.EvalSymlinks(path)
		assert.Equal(t, expectedPath, actualPath)
	})

	t.Run("global scope when forced", func(t *testing.T) {
		tmpDir := t.TempDir()
		gitDir := filepath.Join(tmpDir, ".git")
		err := os.Mkdir(gitDir, 0755)
		require.NoError(t, err)

		// Change to git repo
		origWd, _ := os.Getwd()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("Failed to change directory: %v", err)
		}
		defer func() {
			if err := os.Chdir(origWd); err != nil {
				t.Errorf("Failed to restore working directory: %v", err)
			}
		}()

		path, isGlobal := datapath.ResolveScopedPath(true)
		assert.True(t, isGlobal)
		assert.Contains(t, path, "todos.json")
		assert.NotEqual(t, filepath.Join(tmpDir, ".todos.json"), path)
	})

	t.Run("global scope when not in git repo", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Change to non-git directory
		origWd, _ := os.Getwd()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("Failed to change directory: %v", err)
		}
		defer func() {
			if err := os.Chdir(origWd); err != nil {
				t.Errorf("Failed to restore working directory: %v", err)
			}
		}()

		path, isGlobal := datapath.ResolveScopedPath(false)
		assert.True(t, isGlobal)
		assert.Contains(t, path, "todos.json")
	})
}

func TestEnsureProjectGitignore(t *testing.T) {
	t.Run("adds todos to gitignore", func(t *testing.T) {
		tmpDir := t.TempDir()
		gitDir := filepath.Join(tmpDir, ".git")
		err := os.Mkdir(gitDir, 0755)
		require.NoError(t, err)

		// Change to git repo
		origWd, _ := os.Getwd()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("Failed to change directory: %v", err)
		}
		defer func() {
			if err := os.Chdir(origWd); err != nil {
				t.Errorf("Failed to restore working directory: %v", err)
			}
		}()

		err = datapath.EnsureProjectGitignore()
		require.NoError(t, err)

		content, err := os.ReadFile(filepath.Join(tmpDir, ".gitignore"))
		require.NoError(t, err)
		assert.Contains(t, string(content), ".todos.json")
	})

	t.Run("no error when not in git repo", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Change to non-git directory
		origWd, _ := os.Getwd()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("Failed to change directory: %v", err)
		}
		defer func() {
			if err := os.Chdir(origWd); err != nil {
				t.Errorf("Failed to restore working directory: %v", err)
			}
		}()

		err := datapath.EnsureProjectGitignore()
		// Should not error even when not in git repo
		assert.NoError(t, err)
	})
}