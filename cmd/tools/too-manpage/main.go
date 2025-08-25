package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func main() {
	// Create a root command that matches the main too command structure
	rootCmd := &cobra.Command{
		Use:   "too",
		Short: "A simple command-line todo list manager",
		Long: `too is a simple command-line todo list manager that helps you track tasks.
It stores todos in a JSON file and provides commands to add, modify, toggle, and search todos.`,
		Version: "1.0.0", // This should match your actual version
	}

	// Add all the subcommands
	rootCmd.AddCommand(
		&cobra.Command{
			Use:     "init",
			Aliases: []string{"i"},
			Short:   "Initialize a new todo collection",
			Long:    `Initialize a new todo collection in the specified location or the default location (~/.todos.json).`,
		},
		&cobra.Command{
			Use:     "add <text>",
			Aliases: []string{"a"},
			Short:   "Add a new todo",
			Long:    `Add a new todo with the specified text.`,
			Args:    cobra.MinimumNArgs(1),
		},
		&cobra.Command{
			Use:     "modify <id> <text>",
			Aliases: []string{"m"},
			Short:   "Modify the text of an existing todo",
			Long:    `Modify the text of an existing todo by its ID.`,
			Args:    cobra.MinimumNArgs(2),
		},
		&cobra.Command{
			Use:     "toggle <id>",
			Aliases: []string{"t"},
			Short:   "Toggle the status of a todo",
			Long:    `Toggle the status of a todo between pending and done.`,
			Args:    cobra.ExactArgs(1),
		},
		&cobra.Command{
			Use:   "clean",
			Short: "Remove finished todos",
			Long:  `Remove all todos marked as done from the collection.`,
		},
		&cobra.Command{
			Use:     "reorder <id1> <id2>",
			Aliases: []string{"r"},
			Short:   "Swap the position of two todos",
			Long:    `Swap the position of two todos by their IDs.`,
			Args:    cobra.ExactArgs(2),
		},
		&cobra.Command{
			Use:     "search <query>",
			Aliases: []string{"s"},
			Short:   "Search for todos",
			Long:    `Search for todos containing the specified text.`,
			Args:    cobra.MinimumNArgs(1),
		},
		&cobra.Command{
			Use:   "list",
			Short: "List all todos",
			Long:  `List all todos in the collection.`,
		},
		&cobra.Command{
			Use:   "version",
			Short: "Print the version number",
			Long:  `Print the version number of too`,
		},
	)

	// Add search command with its flag
	searchCmd := rootCmd.Commands()[6] // Get the search command
	searchCmd.Flags().BoolP("case-sensitive", "s", false, "Perform case-sensitive search")

	// Add persistent flags
	rootCmd.PersistentFlags().StringP("collection", "c", "", "path to todo collection (default: $HOME/.todos.json)")
	rootCmd.PersistentFlags().CountP("verbose", "v", "Increase verbosity (-v, -vv, -vvv)")

	// Set up man page header
	now := time.Now()
	header := &doc.GenManHeader{
	Title:   "TOO",
		Section: "1",
		Date:    &now,
		Source:  "too",
		Manual:  "User Commands",
	}

	// Generate man page to stdout
	if err := doc.GenMan(rootCmd, header, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating man page: %v\n", err)
		os.Exit(1)
	}
}
