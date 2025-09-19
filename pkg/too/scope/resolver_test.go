package scope

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolverGitDetection(t *testing.T) {
	t.Run("finds git root from subdirectory", func(t *testing.T) {
		tmpDir := t.TempDir()
		gitDir := filepath.Join(tmpDir, ".git")
		err := os.Mkdir(gitDir, 0755)
		require.NoError(t, err)

		subDir := filepath.Join(tmpDir, "src", "pkg")
		err = os.MkdirAll(subDir, 0755)
		require.NoError(t, err)

		resolver := NewResolver(false)
		scope, err := resolver.Resolve(subDir)
		require.NoError(t, err)

		assert.False(t, scope.IsGlobal)
		assert.Equal(t, filepath.Join(tmpDir, ".todos.json"), scope.Path)
		assert.Equal(t, tmpDir, scope.GitRoot)
	})

	t.Run("returns empty when no git repo", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		resolver := NewResolver(false)
		scope, err := resolver.Resolve(tmpDir)
		require.NoError(t, err)

		assert.True(t, scope.IsGlobal)
		assert.Contains(t, scope.Path, "todos.json")
		assert.Empty(t, scope.GitRoot)
	})

	t.Run("handles nested git repositories", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		// Create outer git repo
		outerGit := filepath.Join(tmpDir, ".git")
		err := os.Mkdir(outerGit, 0755)
		require.NoError(t, err)

		// Create inner git repo
		innerDir := filepath.Join(tmpDir, "inner")
		err = os.Mkdir(innerDir, 0755)
		require.NoError(t, err)
		innerGit := filepath.Join(innerDir, ".git")
		err = os.Mkdir(innerGit, 0755)
		require.NoError(t, err)

		resolver := NewResolver(false)
		
		// From inner directory should find inner git
		scope, err := resolver.Resolve(innerDir)
		require.NoError(t, err)
		assert.Equal(t, innerDir, scope.GitRoot)
		assert.Equal(t, filepath.Join(innerDir, ".todos.json"), scope.Path)
	})
}

func TestResolverForceGlobal(t *testing.T) {
	t.Run("force global ignores git repo", func(t *testing.T) {
		tmpDir := t.TempDir()
		gitDir := filepath.Join(tmpDir, ".git")
		err := os.Mkdir(gitDir, 0755)
		require.NoError(t, err)

		resolver := NewResolver(true) // force global
		scope, err := resolver.Resolve(tmpDir)
		require.NoError(t, err)

		assert.True(t, scope.IsGlobal)
		assert.Contains(t, scope.Path, "todos.json")
		assert.Empty(t, scope.GitRoot)
	})
}

func TestResolverXDGPaths(t *testing.T) {
	t.Run("uses XDG_DATA_HOME when set", func(t *testing.T) {
		tmpDir := t.TempDir()
		xdgPath := filepath.Join(tmpDir, "xdg-data")
		
		// Save and restore XDG_DATA_HOME
		oldXDG := os.Getenv("XDG_DATA_HOME")
		if err := os.Setenv("XDG_DATA_HOME", xdgPath); err != nil {
			t.Fatalf("Failed to set XDG_DATA_HOME: %v", err)
		}
		defer func() {
			if err := os.Setenv("XDG_DATA_HOME", oldXDG); err != nil {
				t.Errorf("Failed to restore XDG_DATA_HOME: %v", err)
			}
		}()

		resolver := NewResolver(true)
		scope, err := resolver.Resolve(tmpDir)
		require.NoError(t, err)

		expectedPath := filepath.Join(xdgPath, "too", "todos.json")
		assert.Equal(t, expectedPath, scope.Path)
	})

	t.Run("defaults to ~/.local/share when XDG_DATA_HOME not set", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		// Clear XDG_DATA_HOME
		oldXDG := os.Getenv("XDG_DATA_HOME")
		if err := os.Unsetenv("XDG_DATA_HOME"); err != nil {
			t.Fatalf("Failed to unset XDG_DATA_HOME: %v", err)
		}
		defer func() {
			if err := os.Setenv("XDG_DATA_HOME", oldXDG); err != nil {
				t.Errorf("Failed to restore XDG_DATA_HOME: %v", err)
			}
		}()

		resolver := NewResolver(true)
		scope, err := resolver.Resolve(tmpDir)
		require.NoError(t, err)

		home, _ := os.UserHomeDir()
		expectedPath := filepath.Join(home, ".local", "share", "too", "todos.json")
		assert.Equal(t, expectedPath, scope.Path)
	})
}

func TestFindGitRoot(t *testing.T) {
	t.Run("finds git root from current dir", func(t *testing.T) {
		tmpDir := t.TempDir()
		gitDir := filepath.Join(tmpDir, ".git")
		err := os.Mkdir(gitDir, 0755)
		require.NoError(t, err)

		root, err := findGitRoot(tmpDir)
		require.NoError(t, err)
		assert.Equal(t, tmpDir, root)
	})

	t.Run("finds git root from deeply nested dir", func(t *testing.T) {
		tmpDir := t.TempDir()
		gitDir := filepath.Join(tmpDir, ".git")
		err := os.Mkdir(gitDir, 0755)
		require.NoError(t, err)

		deepDir := filepath.Join(tmpDir, "a", "b", "c", "d", "e")
		err = os.MkdirAll(deepDir, 0755)
		require.NoError(t, err)

		root, err := findGitRoot(deepDir)
		require.NoError(t, err)
		assert.Equal(t, tmpDir, root)
	})

	t.Run("returns empty when no git repo found", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		root, err := findGitRoot(tmpDir)
		require.NoError(t, err)
		assert.Empty(t, root)
	})

	t.Run("handles relative paths", func(t *testing.T) {
		tmpDir := t.TempDir()
		gitDir := filepath.Join(tmpDir, ".git")
		err := os.Mkdir(gitDir, 0755)
		require.NoError(t, err)

		// Change to tmpDir and use relative path
		origWd, _ := os.Getwd()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("Failed to change directory: %v", err)
		}
		defer func() {
			if err := os.Chdir(origWd); err != nil {
				t.Errorf("Failed to restore working directory: %v", err)
			}
		}()

		root, err := findGitRoot(".")
		require.NoError(t, err)
		// Use filepath.EvalSymlinks to handle macOS /var -> /private/var
		expectedRoot, _ := filepath.EvalSymlinks(tmpDir)
		actualRoot, _ := filepath.EvalSymlinks(root)
		assert.Equal(t, expectedRoot, actualRoot)
	})
}

func TestEnsureGitignore(t *testing.T) {
	t.Run("adds to empty gitignore", func(t *testing.T) {
		tmpDir := t.TempDir()
		
		err := EnsureGitignore(tmpDir)
		require.NoError(t, err)

		content, err := os.ReadFile(filepath.Join(tmpDir, ".gitignore"))
		require.NoError(t, err)
		assert.Contains(t, string(content), ".todos.json")
		assert.True(t, endsWithNewline(content))
	})

	t.Run("adds to existing gitignore without newline", func(t *testing.T) {
		tmpDir := t.TempDir()
		gitignorePath := filepath.Join(tmpDir, ".gitignore")
		
		// Create gitignore without trailing newline
		err := os.WriteFile(gitignorePath, []byte("*.log\n*.tmp"), 0644)
		require.NoError(t, err)

		err = EnsureGitignore(tmpDir)
		require.NoError(t, err)

		content, err := os.ReadFile(gitignorePath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "*.log\n*.tmp\n.todos.json\n")
	})

	t.Run("adds to existing gitignore with newline", func(t *testing.T) {
		tmpDir := t.TempDir()
		gitignorePath := filepath.Join(tmpDir, ".gitignore")
		
		// Create gitignore with trailing newline
		err := os.WriteFile(gitignorePath, []byte("*.log\n*.tmp\n"), 0644)
		require.NoError(t, err)

		err = EnsureGitignore(tmpDir)
		require.NoError(t, err)

		content, err := os.ReadFile(gitignorePath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "*.log\n*.tmp\n.todos.json\n")
	})

	t.Run("skips if already present", func(t *testing.T) {
		tmpDir := t.TempDir()
		gitignorePath := filepath.Join(tmpDir, ".gitignore")
		
		// Create gitignore with .todos.json already
		original := []byte("*.log\n.todos.json\n*.tmp\n")
		err := os.WriteFile(gitignorePath, original, 0644)
		require.NoError(t, err)

		err = EnsureGitignore(tmpDir)
		require.NoError(t, err)

		content, err := os.ReadFile(gitignorePath)
		require.NoError(t, err)
		assert.Equal(t, string(original), string(content))
	})

	t.Run("recognizes absolute path pattern", func(t *testing.T) {
		tmpDir := t.TempDir()
		gitignorePath := filepath.Join(tmpDir, ".gitignore")
		
		// Create gitignore with /.todos.json pattern
		original := []byte("*.log\n/.todos.json\n*.tmp\n")
		err := os.WriteFile(gitignorePath, original, 0644)
		require.NoError(t, err)

		err = EnsureGitignore(tmpDir)
		require.NoError(t, err)

		content, err := os.ReadFile(gitignorePath)
		require.NoError(t, err)
		assert.Equal(t, string(original), string(content))
	})
}