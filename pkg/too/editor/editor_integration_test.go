package editor_test

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/arthur-debert/too/pkg/too/editor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenInEditor_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a mock editor script that we can control
	tmpDir := t.TempDir()
	mockEditor := filepath.Join(tmpDir, "mock-editor")

	var scriptContent string
	if runtime.GOOS == "windows" {
		mockEditor += ".bat"
		scriptContent = `@echo off
echo Test content from mock editor > %1
echo Second line >> %1
`
	} else {
		scriptContent = `#!/bin/sh
cat > "$1" << 'EOF'
Test content from mock editor
Second line
EOF
`
	}

	err := os.WriteFile(mockEditor, []byte(scriptContent), 0755)
	require.NoError(t, err)

	// Save and restore EDITOR env var
	oldEditor := os.Getenv("EDITOR")
	err = os.Setenv("EDITOR", mockEditor)
	require.NoError(t, err)
	defer func() { _ = os.Setenv("EDITOR", oldEditor) }()

	t.Run("creates new content", func(t *testing.T) {
		result, err := editor.OpenInEditor("")
		require.NoError(t, err)
		assert.Equal(t, "Test content from mock editor\nSecond line", result)
	})

	t.Run("edits existing content", func(t *testing.T) {
		result, err := editor.OpenInEditor("Initial content")
		require.NoError(t, err)
		// Our mock editor ignores initial content and always writes the same thing
		assert.Equal(t, "Test content from mock editor\nSecond line", result)
	})
}

func TestOpenInEditor_Errors(t *testing.T) {
	t.Run("no editor available", func(t *testing.T) {
		// Clear all editor env vars
		oldEditor := os.Getenv("EDITOR")
		oldVisual := os.Getenv("VISUAL")
		err := os.Unsetenv("EDITOR")
		require.NoError(t, err)
		err = os.Unsetenv("VISUAL")
		require.NoError(t, err)
		defer func() {
			_ = os.Setenv("EDITOR", oldEditor)
			_ = os.Setenv("VISUAL", oldVisual)
		}()

		// Set PATH to empty to ensure no default editors are found
		oldPath := os.Getenv("PATH")
		err = os.Setenv("PATH", "")
		require.NoError(t, err)
		defer func() { _ = os.Setenv("PATH", oldPath) }()

		_, err = editor.OpenInEditor("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no editor found")
	})

	t.Run("editor command fails", func(t *testing.T) {
		// Set editor to a command that will fail
		oldEditor := os.Getenv("EDITOR")
		err := os.Setenv("EDITOR", "false") // 'false' command always exits with error
		require.NoError(t, err)
		defer func() { _ = os.Setenv("EDITOR", oldEditor) }()

		_, err = editor.OpenInEditor("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "editor command failed")
	})
}

func TestGetEditorPath(t *testing.T) {
	t.Run("from EDITOR env var", func(t *testing.T) {
		oldEditor := os.Getenv("EDITOR")
		err := os.Setenv("EDITOR", "vim")
		require.NoError(t, err)
		defer func() { _ = os.Setenv("EDITOR", oldEditor) }()

		path, err := editor.GetEditorPath()
		require.NoError(t, err)
		assert.Contains(t, path, "vim")
	})

	t.Run("from VISUAL env var", func(t *testing.T) {
		oldEditor := os.Getenv("EDITOR")
		oldVisual := os.Getenv("VISUAL")
		err := os.Unsetenv("EDITOR")
		require.NoError(t, err)
		err = os.Setenv("VISUAL", "nano")
		require.NoError(t, err)
		defer func() {
			_ = os.Setenv("EDITOR", oldEditor)
			_ = os.Setenv("VISUAL", oldVisual)
		}()

		path, err := editor.GetEditorPath()
		require.NoError(t, err)
		assert.Contains(t, path, "nano")
	})

	t.Run("no editor set", func(t *testing.T) {
		oldEditor := os.Getenv("EDITOR")
		oldVisual := os.Getenv("VISUAL")
		_ = os.Unsetenv("EDITOR")
		_ = os.Unsetenv("VISUAL")
		defer func() {
			_ = os.Setenv("EDITOR", oldEditor)
			_ = os.Setenv("VISUAL", oldVisual)
		}()

		_, err := editor.GetEditorPath()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no editor found")
	})
}

func TestCreateTempFileWithContent(t *testing.T) {
	t.Run("creates file with content", func(t *testing.T) {
		content := "Test content\nLine 2"
		filename, err := editor.CreateTempFileWithContent(content)
		require.NoError(t, err)
		defer func() { _ = os.Remove(filename) }()

		// Verify file exists and has correct content
		readContent, err := os.ReadFile(filename)
		require.NoError(t, err)
		assert.Equal(t, content, string(readContent))
	})

	t.Run("creates empty file", func(t *testing.T) {
		filename, err := editor.CreateTempFileWithContent("")
		require.NoError(t, err)
		defer func() { _ = os.Remove(filename) }()

		// Verify file exists and is empty
		stat, err := os.Stat(filename)
		require.NoError(t, err)
		assert.Equal(t, int64(0), stat.Size())
	})
}

// Example of how to use the editor package in tests
func ExampleOpenInEditor() {
	// Create a mock editor for testing
	mockEditor := "/tmp/mock-editor.sh"
	script := `#!/bin/sh
echo "Example todo" > "$1"
`
	_ = os.WriteFile(mockEditor, []byte(script), 0755)
	defer func() { _ = os.Remove(mockEditor) }()

	// Set the mock editor
	_ = os.Setenv("EDITOR", mockEditor)

	// Open editor
	content, err := editor.OpenInEditor("")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Content: %s\n", content)
	// Output: Content: Example todo
}
