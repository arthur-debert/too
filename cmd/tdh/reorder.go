package main

import (
	"github.com/arthur-debert/tdh/pkg/tdh"
	"github.com/arthur-debert/tdh/pkg/tdh/output"
	"github.com/spf13/cobra"
)

var reorderCmd = &cobra.Command{
	Use:     msgReorderUse,
	Aliases: aliasesReorder,
	Short:   msgReorderShort,
	Long:    msgReorderLong,
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
