package main

import (
	"strconv"

	"github.com/arthur-debert/tdh/pkg/tdh"
	"github.com/arthur-debert/tdh/pkg/tdh/output"
	"github.com/spf13/cobra"
)

var reorderCmd = &cobra.Command{
	Use:     "reorder <id1> <id2>",
	Aliases: []string{"r"},
	Short:   "Swap the position of two todos",
	Long:    `Swap the position of two todos by their IDs.`,
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Parse IDs
		idA, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}
		idB, err := strconv.Atoi(args[1])
		if err != nil {
			return err
		}

		// Get collection path from flag
		collectionPath, _ := cmd.Flags().GetString("data-path")

		// Call business logic
		result, err := tdh.Reorder(idA, idB, tdh.ReorderOptions{
			CollectionPath: collectionPath,
		})
		if err != nil {
			return err
		}

		// Render output
		renderer := output.NewRenderer(nil)
		return renderer.RenderReorder(result)
	},
}

func init() {
	rootCmd.AddCommand(reorderCmd)
}
