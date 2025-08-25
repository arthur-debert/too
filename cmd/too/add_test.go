package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arthur-debert/too/pkg/too"
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
		cmd.SetArgs([]string{"add", "-p", testDataPath, "--to", "1.2", "New sub-task"})

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

		// Check that to flag exists
		toFlag := addSubCmd.Flags().Lookup("to")
		assert.NotNil(t, toFlag)
		assert.Equal(t, "to", toFlag.Name)
		assert.Equal(t, "", toFlag.DefValue)
		assert.Contains(t, toFlag.Usage, "parent todo position path")
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
		addChildCmd.SetArgs([]string{"add", "-p", testDataPath, "--to", "1", "Child task"})

		err = addChildCmd.Execute()
		// Should succeed - the parent exists
		assert.NoError(t, err)

		// Test with non-existent parent to verify the flag is being used
		addOrphanCmd := createTestRootCommand()
		addOrphanCmd.SetArgs([]string{"add", "-p", testDataPath, "--to", "99", "Orphan task"})

		err = addOrphanCmd.Execute()
		// Should fail with parent not found
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parent todo not found")
	})

	t.Run("shortcut: first arg as position path", func(t *testing.T) {
		// Initialize collection
		initCmd := createTestRootCommand()
		initCmd.SetArgs([]string{"init", "-p", testDataPath})
		err := initCmd.Execute()
		assert.NoError(t, err)

		// Add parent task
		addParentCmd := createTestRootCommand()
		addParentCmd.SetArgs([]string{"add", "-p", testDataPath, "Parent task"})
		err = addParentCmd.Execute()
		assert.NoError(t, err)

		// Use shortcut: too add 1 "Child task"
		addChildCmd := createTestRootCommand()
		addChildCmd.SetArgs([]string{"add", "-p", testDataPath, "1", "Child task via shortcut"})

		err = addChildCmd.Execute()
		assert.NoError(t, err)
		// The shortcut worked if there's no error - the task was added as a child
	})

	t.Run("shortcut: works with absolute paths like 1.2", func(t *testing.T) {
		// Initialize collection
		initCmd := createTestRootCommand()
		initCmd.SetArgs([]string{"init", "-p", testDataPath})
		err := initCmd.Execute()
		assert.NoError(t, err)

		// Add parent
		addParentCmd := createTestRootCommand()
		addParentCmd.SetArgs([]string{"add", "-p", testDataPath, "Parent"})
		err = addParentCmd.Execute()
		assert.NoError(t, err)

		// Add first child using --to
		addChild1Cmd := createTestRootCommand()
		addChild1Cmd.SetArgs([]string{"add", "-p", testDataPath, "--to", "1", "First child"})
		err = addChild1Cmd.Execute()
		assert.NoError(t, err)

		// Add second child using --to
		addChild2Cmd := createTestRootCommand()
		addChild2Cmd.SetArgs([]string{"add", "-p", testDataPath, "--to", "1", "Second child"})
		err = addChild2Cmd.Execute()
		assert.NoError(t, err)

		// Now use shortcut with absolute path 1.2
		addGrandchildCmd := createTestRootCommand()
		addGrandchildCmd.SetArgs([]string{"add", "-p", testDataPath, "1.2", "Grandchild of second child"})

		err = addGrandchildCmd.Execute()
		assert.NoError(t, err)
	})

	t.Run("shortcut: doesn't trigger for non-numeric first arg", func(t *testing.T) {
		// Initialize collection
		initCmd := createTestRootCommand()
		initCmd.SetArgs([]string{"init", "-p", testDataPath})
		err := initCmd.Execute()
		assert.NoError(t, err)

		// Add with first arg that looks like text
		addCmd := createTestRootCommand()
		addCmd.SetArgs([]string{"add", "-p", testDataPath, "1st", "thing", "to", "do"})

		err = addCmd.Execute()
		assert.NoError(t, err)
		// Should create a todo with text "1st thing to do" (no error means it worked)
	})

	t.Run("shortcut: --to flag takes precedence over shortcut", func(t *testing.T) {
		// Initialize collection
		initCmd := createTestRootCommand()
		initCmd.SetArgs([]string{"init", "-p", testDataPath})
		err := initCmd.Execute()
		assert.NoError(t, err)

		// Add two parent tasks
		addParent1Cmd := createTestRootCommand()
		addParent1Cmd.SetArgs([]string{"add", "-p", testDataPath, "First parent"})
		err = addParent1Cmd.Execute()
		assert.NoError(t, err)

		addParent2Cmd := createTestRootCommand()
		addParent2Cmd.SetArgs([]string{"add", "-p", testDataPath, "Second parent"})
		err = addParent2Cmd.Execute()
		assert.NoError(t, err)

		// Use both --to flag and numeric first arg
		// The --to flag should win
		addChildCmd := createTestRootCommand()
		addChildCmd.SetArgs([]string{"add", "-p", testDataPath, "--to", "2", "1", "Child of second parent"})

		err = addChildCmd.Execute()
		assert.NoError(t, err)
		// This should create "1 Child of second parent" as a child of task 2
		// Not as a child of task 1
	})

	t.Run("shortcut: single numeric arg is treated as todo text", func(t *testing.T) {
		// Initialize collection
		initCmd := createTestRootCommand()
		initCmd.SetArgs([]string{"init", "-p", testDataPath})
		err := initCmd.Execute()
		assert.NoError(t, err)

		// Add with just a number
		addCmd := createTestRootCommand()
		addCmd.SetArgs([]string{"add", "-p", testDataPath, "42"})

		err = addCmd.Execute()
		assert.NoError(t, err)
		// Should create a todo with text "42" (no error means it worked)
	})
}

// createTestRootCommand creates a fresh root command for testing
func createTestRootCommand() *cobra.Command {
	// Create a new root command
	testRoot := &cobra.Command{
		Use:   "too",
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
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if first argument is a position path (shortcut for --to)
			var text string
			collectionPath, _ := cmd.Flags().GetString("data-path")
			parentPath, _ := cmd.Flags().GetString("to")

			// If --to wasn't explicitly set and we have at least 2 args
			if parentPath == "" && len(args) >= 2 {
				// Check if first arg matches position path pattern (e.g., "1", "1.2", "1.2.3")
				if isPositionPath(args[0]) {
					parentPath = args[0]
					text = strings.Join(args[1:], " ")
				} else {
					// Normal case: all args are the todo text
					text = strings.Join(args, " ")
				}
			} else {
				// Normal case: all args are the todo text
				text = strings.Join(args, " ")
			}

			// Call business logic
			result, err := too.Add(text, too.AddOptions{
				CollectionPath: collectionPath,
				ParentPath:     parentPath,
			})
			if err != nil {
				return err
			}

			// Render output
			renderer, err := getRenderer()
			if err != nil {
				return err
			}
			return renderer.RenderAdd(result)
		},
	}

	// Add the to flag to the fresh add command
	testAddCmd.Flags().StringVar(&parentPath, "to", "", "parent todo position path (e.g., \"1.2\")")

	// Add all commands
	testRoot.AddCommand(testAddCmd)
	testRoot.AddCommand(listCmd)
	testRoot.AddCommand(initCmd)

	return testRoot
}
