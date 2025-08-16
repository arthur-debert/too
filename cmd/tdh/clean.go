package main

import (
	"github.com/arthur-debert/tdh/pkg/tdh"
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:     msgCleanUse,
	Short:   msgCleanShort,
	Long:    msgCleanLong,
	GroupID: "misc",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get collection path from flag
		collectionPath, _ := cmd.Flags().GetString("data-path")

		// Call business logic
		result, err := tdh.Clean(tdh.CleanOptions{
			CollectionPath: collectionPath,
		})
		if err != nil {
			return err
		}

		// Render output
		renderer, err := getRenderer()
		if err != nil {
			return err
		}
		return renderer.RenderClean(result)
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
