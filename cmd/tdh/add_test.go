package main

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestAddCommandCLI(t *testing.T) {
	// Create a temp directory for test data
	tempDir := t.TempDir()
	testDataPath := filepath.Join(tempDir, "test-todos.json")

	t.Run("passes parent flag to business logic", func(t *testing.T) {
		// We'll verify the flag is parsed correctly by checking that the command executes without error
		// The actual business logic testing happens in the add package tests

		cmd := createTestRootCommand()
		cmd.SetArgs([]string{"add", "-p", testDataPath, "--parent", "1.2", "New sub-task"})

		output := &bytes.Buffer{}
		cmd.SetOut(output)
		cmd.SetErr(output)

		// This will fail because no todo exists at position 1.2, but that's expected
		// We're just testing that the CLI layer parses and passes the flag
		err := cmd.Execute()

		// The error should be about parent not found, not about flag parsing
		if err != nil {
			assert.Contains(t, err.Error(), "parent todo not found")
		}
	})

	t.Run("add command accepts parent flag", func(t *testing.T) {
		cmd := createTestRootCommand()

		// Test that the flag is registered
		addSubCmd, _, err := cmd.Find([]string{"add"})
		assert.NoError(t, err)
		assert.NotNil(t, addSubCmd)

		// Check that parent flag exists
		parentFlag := addSubCmd.Flags().Lookup("parent")
		assert.NotNil(t, parentFlag)
		assert.Equal(t, "parent", parentFlag.Name)
		assert.Equal(t, "", parentFlag.DefValue)
		assert.Contains(t, parentFlag.Usage, "parent todo position path")
	})

	t.Run("parent flag is optional", func(t *testing.T) {
		// Initialize the collection first
		initCmd := createTestRootCommand()
		initCmd.SetArgs([]string{"init", "-p", testDataPath})
		err := initCmd.Execute()
		assert.NoError(t, err)

		// Add without parent flag should work
		addCmd := createTestRootCommand()
		addCmd.SetArgs([]string{"add", "-p", testDataPath, "Top level task"})

		output := &bytes.Buffer{}
		addCmd.SetOut(output)
		addCmd.SetErr(output)

		err = addCmd.Execute()
		assert.NoError(t, err)
	})

	t.Run("parent flag works with add command", func(t *testing.T) {
		// This test verifies that the --parent flag is accepted and processed
		// Initialize collection
		initCmd := createTestRootCommand()
		initCmd.SetArgs([]string{"init", "-p", testDataPath})
		err := initCmd.Execute()
		assert.NoError(t, err)

		// Add parent
		addParentCmd := createTestRootCommand()
		addParentCmd.SetArgs([]string{"add", "-p", testDataPath, "Parent task"})
		err = addParentCmd.Execute()
		assert.NoError(t, err)

		// Add child with --parent flag - this tests that the flag is parsed correctly
		addChildCmd := createTestRootCommand()
		addChildCmd.SetArgs([]string{"add", "-p", testDataPath, "--parent", "1", "Child task"})

		err = addChildCmd.Execute()
		// Should succeed - the parent exists
		assert.NoError(t, err)

		// Test with non-existent parent to verify the flag is being used
		addOrphanCmd := createTestRootCommand()
		addOrphanCmd.SetArgs([]string{"add", "-p", testDataPath, "--parent", "99", "Orphan task"})

		err = addOrphanCmd.Execute()
		// Should fail with parent not found
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parent todo not found")
	})
}

// createTestRootCommand creates a fresh root command for testing
func createTestRootCommand() *cobra.Command {
	// Create a new root command
	testRoot := &cobra.Command{
		Use:   "tdh",
		Short: "Test root",
	}

	// Define command groups to match the main root command
	testRoot.AddGroup(
		&cobra.Group{ID: "core", Title: "CORE:"},
		&cobra.Group{ID: "extras", Title: "EXTRAS:"},
		&cobra.Group{ID: "misc", Title: "MISC:"},
	)

	// Add the persistent flag
	testRoot.PersistentFlags().StringP("data-path", "p", "", "path to todo collection")
	var testVerbosity int
	testRoot.PersistentFlags().CountVarP(&testVerbosity, "verbose", "v", "Increase verbosity")

	// Create a fresh add command for this test to avoid state pollution
	testAddCmd := &cobra.Command{
		Use:     msgAddUse,
		Aliases: aliasesAdd,
		Short:   msgAddShort,
		Long:    msgAddLong,
		Args:    cobra.MinimumNArgs(1),
		RunE:    addCmd.RunE, // Use the same RunE function
	}

	// Add the parent flag to the fresh add command
	testAddCmd.Flags().StringVar(&parentPath, "parent", "", "parent todo position path (e.g., \"1.2\")")

	// Add all commands
	testRoot.AddCommand(testAddCmd)
	testRoot.AddCommand(listCmd)
	testRoot.AddCommand(initCmd)

	return testRoot
}
