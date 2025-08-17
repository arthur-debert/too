package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddCommandEditorFlag(t *testing.T) {
	t.Run("editor flag validation", func(t *testing.T) {
		// Create a temporary script that acts as an editor
		tmpDir := t.TempDir()
		editorScript := filepath.Join(tmpDir, "test-editor")

		// Write a simple editor script that adds content
		scriptContent := `#!/bin/sh
echo "Test todo from editor" > "$1"
`
		err := os.WriteFile(editorScript, []byte(scriptContent), 0755)
		require.NoError(t, err)

		// Set the editor environment variable
		oldEditor := os.Getenv("EDITOR")
		err = os.Setenv("EDITOR", editorScript)
		require.NoError(t, err)
		defer func() { _ = os.Setenv("EDITOR", oldEditor) }()

		// Test that add command accepts --editor flag
		cmd := rootCmd
		cmd.SetArgs([]string{"add", "--editor"})

		// We can't actually execute the command in tests because it would
		// try to open an editor, but we can verify the flag is registered
		addCmd, _, err := cmd.Find([]string{"add"})
		require.NoError(t, err)

		editorFlag := addCmd.Flag("editor")
		assert.NotNil(t, editorFlag)
		assert.Equal(t, "e", editorFlag.Shorthand)
		assert.Equal(t, "open todo in editor for crafting", editorFlag.Usage)
	})

	t.Run("editor flag with position path", func(t *testing.T) {
		cmd := rootCmd
		addCmd, _, err := cmd.Find([]string{"add"})
		require.NoError(t, err)

		// Verify we can use editor flag with position path
		err = addCmd.ParseFlags([]string{"--editor"})
		assert.NoError(t, err)

		// Verify the flag value
		editorFlag := addCmd.Flag("editor")
		assert.Equal(t, "true", editorFlag.Value.String())
	})
}

func TestEditCommandEditorFlag(t *testing.T) {
	t.Run("editor flag validation", func(t *testing.T) {
		cmd := rootCmd
		editCmd, _, err := cmd.Find([]string{"edit"})
		require.NoError(t, err)

		editorFlag := editCmd.Flag("editor")
		assert.NotNil(t, editorFlag)
		assert.Equal(t, "e", editorFlag.Shorthand)
		assert.Equal(t, "open todo in editor for editing", editorFlag.Usage)
	})
}
