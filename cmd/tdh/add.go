package main

import (
	"strings"

	"github.com/arthur-debert/tdh/pkg/tdh"
	"github.com/arthur-debert/tdh/pkg/tdh/output"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:     "add <text>",
	Aliases: []string{"a", "new", "create"},
	Short:   "Add a new todo",
	Long:    `Add a new todo with the specified text.`,
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Join all arguments as the todo text
		text := strings.Join(args, " ")

		// Get collection path from flag
		collectionPath, _ := cmd.Flags().GetString("data-path")

		// Call business logic
		result, err := tdh.Add(text, tdh.AddOptions{
			CollectionPath: collectionPath,
		})
		if err != nil {
			return err
		}

		// Render output
		renderer := output.NewRenderer(nil)
		return renderer.RenderAdd(result)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
