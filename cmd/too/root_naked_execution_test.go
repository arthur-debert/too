package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleNakedExecution(t *testing.T) {
	// Save original os.Args
	originalArgs := os.Args
	defer func() {
		os.Args = originalArgs
	}()

	tests := []struct {
		name     string
		args     []string
		wantArgs []string
	}{
		// List command cases (no arguments)
		{
			name:     "no arguments defaults to list",
			args:     []string{"too"},
			wantArgs: []string{"too", "list"},
		},
		{
			name:     "only flags defaults to list",
			args:     []string{"too", "--all"},
			wantArgs: []string{"too", "list", "--all"},
		},
		{
			name:     "multiple flags defaults to list",
			args:     []string{"too", "--all", "-v"},
			wantArgs: []string{"too", "list", "--all", "-v"},
		},
		{
			name:     "format flag with value defaults to list",
			args:     []string{"too", "--format", "json"},
			wantArgs: []string{"too", "list", "--format", "json"},
		},
		{
			name:     "data-path flag defaults to list",
			args:     []string{"too", "--data-path", "/tmp/test.json"},
			wantArgs: []string{"too", "list", "--data-path", "/tmp/test.json"},
		},

		// Add command cases (with arguments)
		{
			name:     "single argument triggers add",
			args:     []string{"too", "Buy milk"},
			wantArgs: []string{"too", "add", "Buy milk"},
		},
		{
			name:     "multiple arguments trigger add",
			args:     []string{"too", "Buy", "milk", "and", "eggs"},
			wantArgs: []string{"too", "add", "Buy", "milk", "and", "eggs"},
		},
		{
			name:     "arguments with flags trigger add",
			args:     []string{"too", "--to", "1", "Buy", "groceries"},
			wantArgs: []string{"too", "add", "--to", "1", "Buy", "groceries"},
		},
		{
			name:     "mixed flags and arguments trigger add",
			args:     []string{"too", "Fix", "bug", "--urgent"},
			wantArgs: []string{"too", "add", "Fix", "bug", "--urgent"},
		},
		{
			name:     "arguments before flags trigger add",
			args:     []string{"too", "Task", "description", "--format", "json"},
			wantArgs: []string{"too", "add", "Task", "description", "--format", "json"},
		},

		// Special cases (no modification expected, but we still inject)
		{
			name:     "help flag alone",
			args:     []string{"too", "--help"},
			wantArgs: []string{"too", "--help"}, // Should not modify
		},
		{
			name:     "version flag alone",
			args:     []string{"too", "--version"},
			wantArgs: []string{"too", "--version"}, // Should not modify
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set os.Args
			os.Args = tt.args

			// Run handleNakedExecution
			err := handleNakedExecution()
			assert.NoError(t, err)

			// Verify the result
			assert.Equal(t, tt.wantArgs, os.Args)
		})
	}
}

func TestIsUnknownCommandError(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		want    bool
	}{
		{
			name: "cobra unknown command error",
			err:  fmt.Errorf("unknown command \"foo\" for \"too\""),
			want: true,
		},
		{
			name: "alternative unknown command format",
			err:  fmt.Errorf("Error: unknown command \"bar\""),
			want: true,
		},
		{
			name: "other error",
			err:  fmt.Errorf("some other error"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isUnknownCommandError(tt.err)
			assert.Equal(t, tt.want, got, "isUnknownCommandError(%v) = %v, want %v", tt.err, got, tt.want)
		})
	}
}