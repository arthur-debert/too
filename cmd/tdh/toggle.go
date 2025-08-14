package main

import (
	"strconv"

	"github.com/arthur-debert/tdh/pkg/tdh"
	"github.com/arthur-debert/tdh/pkg/tdh/output"
	"github.com/spf13/cobra"
)

var toggleCmd = &cobra.Command{
	Use:     "toggle <id>",
	Aliases: []string{"t"},
	Short:   "Toggle the status of a todo",
	Long:    `Toggle the status of a todo between pending and done.`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Parse ID
		id, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}

		// Get collection path from flag
		collectionPath, _ := cmd.Flags().GetString("collection")

		// Call business logic
		result, err := tdh.Toggle(id, tdh.ToggleOptions{
			CollectionPath: collectionPath,
		})
		if err != nil {
			return err
		}

		// Render output
		renderer := output.NewRenderer(nil)
		return renderer.RenderToggle(result)
	},
}

func init() {
	rootCmd.AddCommand(toggleCmd)
}
