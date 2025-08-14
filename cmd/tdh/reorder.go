package main

import (
	"github.com/arthur-debert/tdh/pkg/tdh"
	"github.com/arthur-debert/tdh/pkg/tdh/output"
	"github.com/spf13/cobra"
)

var reorderCmd = &cobra.Command{
	Use:     "reorder",
	Aliases: []string{"r"},
	Short:   "Reorder todos by sorting and reassigning sequential positions (alias: r)",
	Long:    `Reorder todos by sorting them by their current position and reassigning sequential positions starting from 1.`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get collection path from flag
		collectionPath, _ := cmd.Flags().GetString("data-path")

		// Call business logic
		result, err := tdh.Reorder(tdh.ReorderOptions{
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
