package editor

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// OpenInEditor opens a temporary file in the user's preferred editor and returns the edited content
func OpenInEditor(initialContent string) (string, error) {
	// Get editor from environment
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		// Default to common editors
		if _, err := exec.LookPath("vim"); err == nil {
			editor = "vim"
		} else if _, err := exec.LookPath("nano"); err == nil {
			editor = "nano"
		} else if _, err := exec.LookPath("ed"); err == nil {
			editor = "ed"
		} else {
			return "", fmt.Errorf("no editor found; please set $EDITOR environment variable")
		}
	}

	// Create temporary file with .txt extension
	tmpfile, err := os.CreateTemp("", "tdh-edit-*.txt")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer func() { _ = os.Remove(tmpfile.Name()) }() // Clean up

	// Write initial content if provided
	if initialContent != "" {
		if _, err := tmpfile.WriteString(initialContent); err != nil {
			_ = tmpfile.Close()
			return "", fmt.Errorf("failed to write initial content: %w", err)
		}
	}
	if err := tmpfile.Close(); err != nil {
		return "", fmt.Errorf("failed to close temporary file: %w", err)
	}

	// Open editor
	cmd := exec.Command(editor, tmpfile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("editor command failed: %w", err)
	}

	// Read the edited content
	content, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		return "", fmt.Errorf("failed to read edited content: %w", err)
	}

	// Process the content
	processed := ProcessTodoText(string(content))
	return processed, nil
}

// ProcessTodoText processes the raw text from the editor according to the rules:
// - Trim whitespace from each line
// - Remove blank lines before the first line
// - Remove blank lines after the last line
// - Preserve blank lines between text lines
func ProcessTodoText(text string) string {
	lines := strings.Split(text, "\n")

	// Trim whitespace from each line
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}

	// Find first non-empty line
	firstNonEmpty := -1
	for i, line := range lines {
		if line != "" {
			firstNonEmpty = i
			break
		}
	}

	// If all lines are empty, return empty string
	if firstNonEmpty == -1 {
		return ""
	}

	// Find last non-empty line
	lastNonEmpty := -1
	for i := len(lines) - 1; i >= 0; i-- {
		if lines[i] != "" {
			lastNonEmpty = i
			break
		}
	}

	// Extract the relevant lines
	relevantLines := lines[firstNonEmpty : lastNonEmpty+1]

	// Join with newlines
	return strings.Join(relevantLines, "\n")
}

// CreateTempFileWithContent creates a temporary file with the given content
// This is useful for testing with editors like ed that need a file
func CreateTempFileWithContent(content string) (string, error) {
	tmpfile, err := os.CreateTemp("", "tdh-test-*.txt")
	if err != nil {
		return "", err
	}

	if _, err := io.WriteString(tmpfile, content); err != nil {
		_ = tmpfile.Close()
		_ = os.Remove(tmpfile.Name())
		return "", err
	}

	if err := tmpfile.Close(); err != nil {
		_ = os.Remove(tmpfile.Name())
		return "", err
	}

	return tmpfile.Name(), nil
}

// GetEditorPath returns the path to the editor, checking EDITOR and VISUAL env vars
func GetEditorPath() (string, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		return "", fmt.Errorf("no editor found; please set $EDITOR environment variable")
	}

	// Resolve the editor path if it's just a name
	if !filepath.IsAbs(editor) {
		if path, err := exec.LookPath(editor); err == nil {
			editor = path
		}
	}

	return editor, nil
}
