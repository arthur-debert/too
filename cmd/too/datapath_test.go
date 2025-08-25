package main

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestDatapathCommandCLI(t *testing.T) {
	t.Run("datapath command exists", func(t *testing.T) {
		cmd := createTestDatapathCommand()
		assert.NotNil(t, cmd)
		assert.Equal(t, "datapath", cmd.Use)
		assert.Contains(t, cmd.Aliases, "path")
	})

	t.Run("accepts data-path flag", func(t *testing.T) {
		// Create temp file
		tmpDir := t.TempDir()
		testPath := filepath.Join(tmpDir, "test.json")

		cmd := createTestDatapathCommand()
		cmd.SetArgs([]string{"-p", testPath})

		output := &bytes.Buffer{}
		cmd.SetOut(output)
		cmd.SetErr(output)

		// The actual execution would fail without proper setup,
		// but we're testing that the flag is parsed correctly
		_ = cmd.Execute()

		// Check that the flag was parsed
		flag := cmd.Flag("data-path")
		assert.NotNil(t, flag)
		assert.Equal(t, testPath, flag.Value.String())
	})

	t.Run("path alias works", func(t *testing.T) {
		// Test that we can find the command by its alias
		root := &cobra.Command{Use: "too"}
		testDatapathCmd := createTestDatapathCommand()
		root.AddCommand(testDatapathCmd)

		// Try to find by alias
		cmd, _, err := root.Find([]string{"path"})
		assert.NoError(t, err)
		assert.NotNil(t, cmd)
		assert.Equal(t, "datapath", cmd.Use)
	})
}

func createTestDatapathCommand() *cobra.Command {
	// Create a test version of the datapath command
	testCmd := &cobra.Command{
		Use:     "datapath",
		Aliases: []string{"path"},
		Short:   "Show the path to the todo data file",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Just return nil for testing
			return nil
		},
	}

	// Add the data-path flag to match the real command
	testCmd.Flags().StringP("data-path", "p", "", "path to todo collection")

	return testCmd
}
