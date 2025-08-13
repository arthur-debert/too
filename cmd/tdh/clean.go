package main

import (
	"github.com/arthur-debert/tdh/pkg/tdh"
	"github.com/arthur-debert/tdh/pkg/tdh/display"
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove finished todos",
	Long:  `Remove all todos marked as done from the collection.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get collection path from flag
		collectionPath, _ := cmd.Flags().GetString("collection")

		// Call business logic
		result, err := tdh.Clean(tdh.CleanOptions{
			CollectionPath: collectionPath,
		})
		if err != nil {
			return err
		}

		// Render output
		renderer := display.NewRenderer(nil)
		return renderer.RenderClean(result)
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
